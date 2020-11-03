package cron

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/logger"
	"github.com/qiusnay/gocron/model"
	"github.com/qiusnay/gocron/service/rpc"
	"github.com/qiusnay/gocron/utils"
	cron "github.com/robfig/cron/v3"
)

var mu sync.Mutex

//定义一个空结构体
type FlCron struct {
	TaskRunStatus sync.Map
}

type Handler interface {
	Run(jobModel model.FlCron, taskUniqueId int64) (model.TaskResult, error)
}

/****************************************/
// 定时任务调度管理器
var mycron *cron.Cron

// 初始化任务, 从数据库取出所有任务, 添加到定时任务并运行
func (c FlCron) Initialize() {
	mycron = cron.New(c.CronWithNoSeconds())
	mycron.Start()

	logger.Info("开始初始化定时任务")
	jobModel := new(model.FlCron)
	taskList, err := jobModel.GetAllJobList()
	if err != nil {
		logger.Error("定时任务初始化,获取任务列表错误: %s", err)
	}
	for _, item := range taskList {
		//获取当前机器配置列表
		machineList := utils.GetConfig("machine", item.Runat)
		//获取当前机器ID
		localIP := utils.GetLocalIP()
		//判断当前作业是否可以在当前机器运行
		if !utils.InArray(localIP, strings.Split(machineList[item.Runat], "|")) {
			continue
		}
		c.Add(item)
	}
	logger.Infof("定时任务初始化完成")
}

// 添加任务
func (c FlCron) Add(jobModel model.FlCron) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf(utils.PanicTrace(e))
		}
	}()
	logger.Info(fmt.Sprintf("jobid %d has run at machine %s", jobModel.Jobid, utils.GetLocalIP()))
	_, err = mycron.AddFunc(jobModel.Rule, c.createJob(jobModel))
	if err != nil {
		logger.Error("添加任务到调度器失败#", err)
	}
	return
}

/**
 * 创建任务
 **/
func (c FlCron) createJob(jobModel model.FlCron) cron.FuncJob {
	taskFunc := func() {
		//1.获取锁,这里是为了判断同一个作业在不同机器之间竞争时,
		// jobid + nexttime 防止多次被dispacher,这里要考虑一下死锁问题
		JobDupKey := c.getJobDuplicateRedisKey(jobModel)
		lock, _ := model.Redis.Int("setnx", JobDupKey, 1)
		if lock != 1 {
			logger.Error(fmt.Sprintf("获取redis lock 失败 %d, 跳过本机任务分发", "cronlock_"+strconv.FormatInt(jobModel.Jobid, 10)))
			return
		}
		//创建TASK
		taskId := c.BeforeExecJob(jobModel)
		if taskId <= 0 {
			return
		}
		//设置30天过期
		model.Redis.Int("expire", JobDupKey, 30*86400)

		//2.上一个任务是否还在运行?这里判断是为了判断同一个job的不同task之间的执行超时问题
		preTaskid, _ := c.TaskRunStatus.Load("current_task_id")
		logger.Info(fmt.Sprintf("获取上一次执行的taskid - %v", preTaskid))
		var contextWithDeadline int = 0
		if preTaskid != nil && jobModel.Overflow != model.BothRun { //如果上一个作业还没有执行完,同时该作业并不是同时运行
			taskTimeoutStatus, err := c.TaskTimeOut(jobModel, preTaskid)
			if taskTimeoutStatus != "forcekill" {
				taskResult := model.TaskResult{"", err, utils.GetLocalIP(), rpc.CronTimeOut, time.Now().Format("2006-01-02 15:04:05")}
				c.afterExecJob(jobModel, taskResult, taskId) //结束当前Task
				return
			}
			contextWithDeadline = 1 //标识当前作业需要向服务端传递一个带超时的context
		}
		c.TaskRunStatus.Store("current_task_id", taskId)
		defer c.TaskRunStatus.Delete("current_task_id")

		logger.Info(fmt.Sprintf("开始执行任务 - %s - 命令-%s", jobModel.JobName, jobModel.Cmd))
		taskResult, err := execJob(jobModel, taskId, contextWithDeadline)

		logger.Info(fmt.Sprintf("执行结果 %+v, 错误信息 : %+v", taskResult, err))

		logger.Info(fmt.Sprintf("任务完成 - %s - 命令- %s - 执行结果- %s - 执行机器 - %s - 结束时间 - %s", jobModel.JobName, jobModel.Cmd, taskResult.Result, taskResult.Host, taskResult.Endtime))
		//释放锁
		model.Redis.Int("del", JobDupKey)
	}
	return taskFunc
}

// 任务前置操作
func (fl FlCron) BeforeExecJob(jobModel model.FlCron) (taskLogId int64) {
	taskLogModel := new(model.FlLog)
	taskId, err := taskLogModel.CreateTaskLog(jobModel)
	if err != nil {
		logger.Error("任务开始执行#写入任务日志失败-", err)
		return
	}
	return taskId
}

// 执行具体任务
func execJob(jobModel model.FlCron, taskId int64, contextWithDeadline int) (task model.TaskResult, err string) {
	return new(rpc.CronClient).Run(jobModel, taskId, contextWithDeadline)
}

// 任务执行后置操作
func (fl FlCron) afterExecJob(jobModel model.FlCron, taskResult model.TaskResult, taskLogId int64) {
	taskLogModel := new(model.FlLog)
	_, err := taskLogModel.UpdateTaskLog(taskLogId, taskResult)
	if err != nil {
		logger.Error("任务结束#更新任务日志失败-", err)
	}
}

//设置cron最小粒度为分钟
func (fl FlCron) CronWithNoSeconds() cron.Option {
	return cron.WithParser(rpc.Parser)
}

func (fl FlCron) getJobDuplicateRedisKey(jobModel model.FlCron) string {
	s, _ := rpc.Parser.Parse(jobModel.Rule)
	expreTime := strconv.FormatInt(s.Next(time.Now()).Unix(), 10)
	return "cronlock_" + strconv.FormatInt(jobModel.Jobid, 10) + "_" + expreTime
}

//超时处理
func (fl FlCron) TaskTimeOut(jobModel model.FlCron, preTaskid interface{}) (s string, e string) {
	taskResult := model.TaskResult{Result: "", Err: ""}
	switch jobModel.Overflow {
	case model.EmailNotify:
		taskResult.Err = "当前作业执行超时,若经常出现,请适当调整执行周期"
		go SendNotification(jobModel, taskResult)
		return "", taskResult.Err
	case model.ForceKill:
		taskResult.Err = "当前作业执行超时,系统己终止上一个任务的运行,请知晓."
		go SendNotification(jobModel, taskResult)
		return "forcekill", taskResult.Err
	case model.HealthCheck:
		return "healthcheck", ""
	}
	return "", ""
}

// 发送任务结果通知
func SendNotification(jobModel model.FlCron, taskResult model.TaskResult) {
	if taskResult.Err == "succss" {
		return // 执行失败才发送通知
	}
	//发送邮件
	// notify.SendCronAlarmMail(taskResult, jobModel)
}
