package cron

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/logger"
	croninit "github.com/qiusnay/gocron/init"
	"github.com/qiusnay/gocron/model"
	"github.com/qiusnay/gocron/service/rpc"
	"github.com/qiusnay/gocron/utils"
	cron "github.com/robfig/cron/v3"
)

var mu sync.Mutex

type VCron struct {
	TaskRunStatus sync.Map
}

var Parser = cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)

type Handler interface {
	Run(Job model.VCron, Taskid int64) (model.TaskResult, error)
}

// 定时任务调度管理器
var Mycron *cron.Cron

// 初始化任务, 从数据库取出所有任务, 添加到定时任务并运行
func (c VCron) Initialize() {
	Mycron = cron.New(c.CronWithNoSeconds())
	Mycron.Start()
	logger.Info("开始初始化定时任务")
	JobList, _ := new(model.VCron).GetAllJobList()
	for _, Job := range JobList {
		//获取当前机器配置列表
		MachineList := utils.GetConfig("machine", Job.Runat)
		//判断当前作业是否可以在当前机器运行
		if !utils.InArray(utils.GetLocalIP(), strings.Split(MachineList[Job.Runat], "|")) {
			continue
		}
		c.AddJobToSchedule(Job)
	}
	logger.Infof("定时任务初始化完成")
}

// 添加任务
func (c VCron) AddJobToSchedule(Job model.VCron) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf(utils.PanicTrace(e))
		}
	}()
	logger.Info(fmt.Sprintf("jobid %d has run at machine %s", Job.Jobid, utils.GetLocalIP()))
	_, err = Mycron.AddFunc(Job.Rule, c.createJob(Job))
	if err != nil {
		logger.Error("添加作业到调度器失败:", err)
	}
	return
}

/**
 * 创建任务
 **/
func (c VCron) createJob(Job model.VCron) cron.FuncJob {
	TaskFunc := func() {
		//1.获取锁,这里是为了判断同一个作业在不同机器之间竞争时,
		// jobid + nexttime 防止多次被dispacher,这里要考虑一下死锁问题
		JobDupKey := c.GetJobDuplicateRedisKey(Job)
		lock, _ := model.Redis.Int("setnx", JobDupKey, 1)
		if lock != 1 {
			logger.Error(fmt.Sprintf("获取lock失败,跳过本机任务分发:jobid : %s, key : %s", Job.Jobid, JobDupKey))
			return
		}
		//创建TASK
		TaskId := c.BeforeExecJob(Job)
		if TaskId <= 0 {
			return
		}
		//设置30天过期
		model.Redis.Int("expire", JobDupKey, 30*86400)

		//2.上一个任务是否还在运行?这里判断是为了判断同一个job的不同task之间的执行超时问题
		PreTaskid, _ := c.TaskRunStatus.Load("current_task_id")
		var NeedCtxTimeOut bool = false
		if PreTaskid != nil && Job.Overflow != model.BothRun { //如果上一个作业还没有执行完,同时该作业并不是同时运行
			TaskTimeoutStatus, err := c.TaskTimeOut(Job, PreTaskid)
			if TaskTimeoutStatus != croninit.CronForceKill {
				QueryResult := model.TaskResult{"", err, utils.GetLocalIP(), croninit.CronTimeOut, time.Now().Format("2006-01-02 15:04:05")}
				c.AfterExecJob(Job, QueryResult, TaskId) //结束当前Task
				return
			}
			NeedCtxTimeOut = true //标识当前作业需要向服务端传递一个带超时的context
		}
		c.TaskRunStatus.Store("current_task_id", TaskId)
		defer c.TaskRunStatus.Delete("current_task_id")

		logger.Info(fmt.Sprintf("开始执行任务 - %s - 命令-%s", Job.JobName, Job.Cmd))
		TaskResult := c.ExecJob(Job, TaskId, NeedCtxTimeOut)
		logger.Info(fmt.Sprintf("任务完成,命令 - %s - 执行结果- %s - 结束时间 - %s", Job.JobName, Job.Cmd, TaskResult.Result, TaskResult.Endtime))
		//释放锁
		model.Redis.Int("del", JobDupKey)
	}
	return TaskFunc
}

// 任务前置操作
func (c VCron) BeforeExecJob(jobModel model.VCron) (taskLogId int64) {
	taskLogModel := new(model.VLog)
	TaskId, err := taskLogModel.CreateTaskLog(jobModel)
	if err != nil {
		logger.Error("任务开始执行#写入任务日志失败-", err)
		return
	}
	return TaskId
}

// 执行具体任务
func (c VCron) ExecJob(Job model.VCron, TaskId int64, NeedCtxTimeOut bool) model.TaskResult {
	s, _ := Parser.Parse(Job.Rule)
	ExpireTime := s.Next(time.Now()).Unix() - time.Now().Unix()
	return new(rpc.CronClient).Run(Job, TaskId, NeedCtxTimeOut, ExpireTime)
}

// 任务执行后置操作
func (c VCron) AfterExecJob(Job model.VCron, TaskResult model.TaskResult, Taskid int64) {
	_, err := new(model.VLog).UpdateTaskLog(Taskid, TaskResult)
	if err != nil {
		logger.Error("任务结束#更新任务日志失败-", err)
	}
}

//设置cron最小粒度为分钟
func (c VCron) CronWithNoSeconds() cron.Option {
	return cron.WithParser(Parser)
}

func (c VCron) GetJobDuplicateRedisKey(Job model.VCron) string {
	StrNextQueryTime := utils.Int64toString(c.GetNextQueryTime(Job.Rule))
	return "cronlock_" + utils.Int64toString(Job.Jobid) + "_" + StrNextQueryTime
}

//根据rule规则获取下一次执行时间
func (c VCron) GetNextQueryTime(Rule string) int64 {
	s, _ := Parser.Parse(Rule)
	return s.Next(time.Now()).Unix()
}

//超时处理
func (c VCron) TaskTimeOut(Job model.VCron, PreTaskid interface{}) (s int64, e string) {
	TaskResult := model.TaskResult{Result: "", Err: ""}
	switch Job.Overflow {
	case model.EmailNotify:
		TaskResult.Err = "当前作业执行超时,若经常出现,请适当调整执行周期"
		go SendNotification(Job, TaskResult)
		return croninit.CronTimeOut, TaskResult.Err
	case model.ForceKill:
		TaskResult.Err = "当前作业执行超时,系统己终止上一个任务的运行,请知晓."
		go SendNotification(Job, TaskResult)
		return croninit.CronForceKill, TaskResult.Err
	case model.HealthCheck:
		return croninit.CronHealthCheck, ""
	}
	return croninit.CronNormal, ""
}

// 发送任务结果通知
func SendNotification(Job model.VCron, TaskResult model.TaskResult) {
	if TaskResult.Err == "succss" {
		return // 执行失败才发送通知
	}
	//发送邮件
	// notify.SendCronAlarmMail(TaskResult, Job)
}
