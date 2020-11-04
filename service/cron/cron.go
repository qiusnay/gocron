package cron

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/google/logger"
	croninit "github.com/qiusnay/gocron/init"
	"github.com/qiusnay/gocron/model"
	"github.com/qiusnay/gocron/service/rpc"
	"github.com/qiusnay/gocron/utils"
	cron "github.com/robfig/cron/v3"
)

type VCron struct{}

var Parser = cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)

type Handler interface {
	Run(Job model.VCron, Taskid int64) (model.TaskResult, error)
}

// 定时任务调度管理器
var Mycron *cron.Cron

// 初始化任务, 从数据库取出所有任务, 添加到定时任务并运行
func (c VCron) Start() {
	Mycron = cron.New(c.CronWithNoSeconds())
	Mycron.Start()
	logger.Info("开始初始化定时任务")
	JobList, _ := new(model.VCron).GetAllJobList()
	for _, Job := range JobList {
		//获取当前机器配置列表
		MachineList := utils.GetConfig("machine", Job.Runat)
		//判断当前作业是否可以在当前机器运行
		if !utils.InArray(utils.GetLocalIP(), strings.Split(MachineList[Job.Runat], "|")) {
			logger.Info("不在当前机器执行,跳过本机")
			continue
		}
		c.AddJobToSchedule(Job)
	}
	logger.Infof("定时任务初始化完成")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	for {
		s := <-sig
		logger.Info("收到信号 -- ", s)
		switch s {
		case syscall.SIGHUP:
			logger.Info("收到终端断开信号, 忽略")
		case syscall.SIGINT, syscall.SIGTERM:
			logger.Info("应用准备退出")
			return
		}
	}
}

// 添加任务
func (c VCron) AddJobToSchedule(Job model.VCron) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf(utils.PanicTrace(e))
		}
	}()
	_, err = Mycron.AddFunc(Job.Rule, c.CreateJob(Job))
	if err != nil {
		logger.Error("添加作业到调度器失败:", err)
	}
	return
}

//创建任务
func (c VCron) CreateJob(Job model.VCron) cron.FuncJob {
	TaskFunc := func() {
		//防止被多次dispacher,考虑一下死锁问题
		if !c.GetDispacherRedisLock(Job) {
			return
		}
		TaskId := c.BeforeExecJob(Job) //创建TASK
		if TaskId <= 0 {
			return
		}
		//上一个Task是否完成了?
		if !c.IsLastTaskRunning(Job) {
			TimeOutType, _ := c.DoTaskTimeOut(Job, TaskId)
			if TimeOutType != croninit.CronForceKill {
				return
			}
		}
		logger.Info(fmt.Sprintf("开始执行任务 - %s - 命令-%s", Job.JobName, Job.Cmd))
		TaskResult := c.ExecJob(Job, TaskId)
		logger.Info(fmt.Sprintf("任务完成,命令 - %s - 执行结果- %s - 结束时间 - %s", Job.JobName, Job.Cmd, TaskResult.Result, TaskResult.Endtime))
		c.ReleaseLock(Job.Jobid)
	}
	return TaskFunc
}

// 任务前置操作
func (c VCron) BeforeExecJob(Job model.VCron) (taskLogId int64) {
	Job.TaskStatus = croninit.CronNormal
	TaskId, err := new(model.VLog).CreateTaskLog(Job)
	if err != nil {
		logger.Error("任务开始执行#写入任务日志失败-", err)
		return
	}
	return TaskId
}

// 执行具体任务
func (c VCron) ExecJob(Job model.VCron, TaskId int64) model.TaskResult {
	s, _ := Parser.Parse(Job.Rule)
	ExpireTime := s.Next(time.Now()).Unix() - time.Now().Unix()
	return new(rpc.CronClient).Run(Job, TaskId, ExpireTime)
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

//当前job是否已经有作业在运行
func (c VCron) IsLastTaskRunning(Job model.VCron) bool {
	RunningKey := "cron_task_running_" + utils.Int64toString(Job.Jobid)
	lock, _ := model.Redis.Int("setnx", RunningKey, croninit.CronRunning)
	if lock != 1 {
		return false
	}
	//设置30天过期
	model.Redis.Int("expire", RunningKey, 30*86400)
	return true
}

//释放锁
func (c VCron) ReleaseLock(Jobid int64) {
	model.Redis.Int("del", "cron_task_running_"+utils.Int64toString(Jobid))
}

//防并发锁
func (c VCron) GetDispacherRedisLock(Job model.VCron) bool {
	JobDupKey := "cronlock_" + utils.Int64toString(Job.Jobid) + "_" + utils.Int64toString(c.GetNextQueryTime(Job.Rule))
	lock, _ := model.Redis.Int("setnx", JobDupKey, 1)
	if lock != 1 {
		logger.Error(fmt.Sprintf("获取lock失败,跳过本机任务分发:jobid : %s, key : %s", Job.Jobid, JobDupKey))
		return false
	}
	//设置30天过期
	model.Redis.Int("expire", JobDupKey, 30*86400)
	return true
}

//根据rule规则获取下一次执行时间
func (c VCron) GetNextQueryTime(Rule string) int64 {
	s, _ := Parser.Parse(Rule)
	return s.Next(time.Now()).Unix()
}

//超时处理
func (c VCron) DoTaskTimeOut(Job model.VCron, TaskId int64) (s int64, e string) {
	TaskResult := model.TaskResult{"", "", utils.GetLocalIP(), croninit.CronTimeOut, time.Now().Format("2006-01-02 15:04:05")}
	var TaskStatus int64
	switch Job.Overflow {
	case model.EmailNotify:
		TaskResult.Err = "当前作业执行超时,若经常出现,请适当调整执行周期"
		TaskStatus = croninit.CronTimeOut
		c.AfterExecJob(Job, TaskResult, TaskId) //结束当前Task
		break
	case model.ForceKill:
		TaskResult.Err = "当前作业执行超时,系统己终止上一个任务的运行,请知晓."
		TaskStatus = croninit.CronForceKill
		break
	case model.HealthCheck:
		TaskResult.Err = "健康检查结果正常."
		TaskStatus = croninit.CronHealthCheck
		c.AfterExecJob(Job, TaskResult, TaskId) //结束当前Task
		break
	}
	go SendNotification(Job, TaskResult)
	return TaskStatus, TaskResult.Err
}

// 发送任务结果通知
func SendNotification(Job model.VCron, TaskResult model.TaskResult) {
	//发送邮件
	// notify.SendCronAlarmMail(TaskResult, Job)
}
