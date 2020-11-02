package cron

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/google/logger"
	"github.com/qiusnay/gocron/model"
	"github.com/qiusnay/gocron/service/rpcx"
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
		// //获取当前机器配置列表
		// machineList := utils.GetConfig("machine", item.Runat)
		// //获取当前机器ID
		// localIP := utils.GetLocalIP()
		// //判断当前作业是否可以在当前机器运行
		// if !utils.InArray(localIP, strings.Split(machineList[item.Runat], "|")) {
		// 	continue
		// }
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
		//1.获取锁,这里是为了判断同一个作业在不同机器之间竞争时,防止多次被dispacher
		// lock, _ := model.Redis.Int("setnx", "cronlock_"+strconv.Itoa(jobModel.Jobid), 1)
		// if lock != 1 {
		// 	logger.Error(fmt.Sprintf("获取redis lock 失败 %d, 跳过本机任务分发", lock))
		// 	return
		// }
		// logger.Error(fmt.Sprintf("获取redis lock 成功 %d", lock))
		//创建TASK
		taskId := c.BeforeExecJob(jobModel)
		if taskId <= 0 {
			return
		}
		//2.上一个任务是否还在运行?这里判断是为了判断同一个job的不同task之间的执行超时问题
		// preTaskid, _ := c.TaskRunStatus.Load("current_task_id")
		// if preTaskid != nil { //如果上一个作业还没有执行完
		// 	return
		// }

		// c.TaskRunStatus.Store("current_task_id", taskId)
		// defer c.TaskRunStatus.Delete("current_task_id")

		logger.Info(fmt.Sprintf("开始执行任务 - %s - 命令-%s", jobModel.JobName, jobModel.Cmd))
		taskResult := execJob(jobModel, taskId)
		logger.Info(fmt.Sprintf("任务完成 - %s - 命令- %s - 执行结果- %s - 执行机器 - %s - 结束时间 - %s", jobModel.JobName, jobModel.Cmd, taskResult.Result, taskResult.Host, taskResult.Endtime))
		afterExecJob(jobModel, taskResult, taskId)
		//释放锁
		model.Redis.Int("del", "cronlock_"+strconv.Itoa(jobModel.Jobid))
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
func execJob(jobModel model.FlCron, taskId int64) model.TaskResult {
	ret, err := new(rpcx.CronClient).Run(jobModel, taskId)
	if err == nil {
		return model.TaskResult{Result: ret.Result, Err: ret.Err, Host: ret.Host, Status: ret.Status, Endtime: ret.Endtime}
	}
	return model.TaskResult{Result: ret.Result, Err: ret.Err, Host: ret.Host, Status: ret.Status, Endtime: ret.Endtime}
}

// 任务执行后置操作
func afterExecJob(jobModel model.FlCron, taskResult model.TaskResult, taskLogId int64) {
	taskLogModel := new(model.FlLog)
	_, err := taskLogModel.UpdateTaskLog(taskLogId, taskResult)
	if err != nil {
		logger.Error("任务结束#更新任务日志失败-", err)
	}
	// 发送邮件
	go SendNotification(jobModel, taskResult)
}

//设置cron最小粒度为分钟
func (fl FlCron) CronWithNoSeconds() cron.Option {
	return cron.WithParser(rpcx.Parser)
}

//超时处理
// func (fl FlCron) TaskTimeOut(jobModel model.FlCron, preTaskid interface{}, ch chan rpcx.RequestCtx) int {
// 	taskResult := model.TaskResult{Result: "", Err: nil}

// 	switch jobModel.Overflow {
// 	case model.EmailNotify:
// 		// taskResult.Err = "当前作业执行超时,若经常出现,请适当调整执行周期"
// 		go SendNotification(jobModel, taskResult)
// 		return -1
// 	case model.ForceKill:
// 		var cl = len(ch)
// 		logger.Info(fmt.Sprintf("通道长度 : %d", cl))
// 		if cl > 0 {
// 			req := <-ch
// 			logger.Info(fmt.Sprintf("通道内容 : %+v", req))
// 			req.Cancel()
// 		}
// 		return 1
// 	}
// 	// case model.HealthCheck:
// 	return 0
// }

// 发送任务结果通知
func SendNotification(jobModel model.FlCron, taskResult model.TaskResult) {
	if taskResult.Err == nil {
		return // 执行失败才发送通知
	}
	//发送邮件
	// notify.SendCronAlarmMail(taskResult, jobModel)
}
