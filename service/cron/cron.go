package cron

import (
	"fmt"
	"strings"
	"strconv"
	"github.com/jakecoffman/cron"
	"github.com/google/logger"
	"github.com/qiusnay/gocron/model"
	"github.com/ouqiang/goutil"
	"github.com/qiusnay/gocron/service/rpcx"
	"github.com/qiusnay/gocron/init"
	"github.com/qiusnay/gocron/utils"
)

const (
	Disabled int = 0 // 禁用
	Failure  int = 10001 // 失败
	Enabled  int = 1 // 启用
	Running  int = 10000 // 运行中
	Finish   int = 10002 // 完成
	Cancel   int = 3 // 取消
)

//定义一个空结构体
type FlCron struct{}

type Handler interface {
	Run(taskModel model.FlCron, taskUniqueId int64) (croninit.TaskResult, error)
}

/****************************************/
// 定时任务调度管理器
var mycron *cron.Cron

// 初始化任务, 从数据库取出所有任务, 添加到定时任务并运行
func (fl FlCron) Initialize() {
	mycron = cron.New()
	mycron.Start()

	logger.Info("开始初始化定时任务")
	taskModel := new(model.FlCron)
	taskList, err := taskModel.GetAllJobList()
	
	if err != nil {
		logger.Error("定时任务初始化,获取任务列表错误: %s", err)
	}
	for _, item := range taskList {
		logger.Infof(fmt.Sprintf("Initialize : %+v", item))
		//获取当前机器配置列表
		machineList := utils.GetConfig("machine", item.Runat)
		//获取当前机器ID
		localIP := utils.GetLocalIP()
		//判断当前作业是否可以在当前机器运行
		if !utils.InArray(localIP, strings.Split(machineList[item.Runat], "|")) {
			continue
		}
		fl.Add(item)
	}
	logger.Infof("定时任务初始化完成")
}

// 添加任务
func (fl FlCron) Add(taskModel model.FlCron) {
	taskModel.Rule = "1 " + taskModel.Rule

	localIP := utils.GetLocalIP()
	logger.Info(fmt.Sprintf("jobid %d has run at machine %s", taskModel.Jobid, localIP))
	
	taskFunc := createJob(taskModel)

	err := goutil.PanicToError(func() {
		mycron.AddFunc(taskModel.Rule, taskFunc, taskModel.JobName)
	})
	if err != nil {
		logger.Error("添加任务到调度器失败#", err)
	}
}

/**
 * 创建任务
 **/
func createJob(taskModel model.FlCron) cron.FuncJob {
	handler := new(rpcx.RPCHandler)
	taskFunc := func() {
		taskLogId := beforeExecJob(taskModel)
		if taskLogId <= 0 {
			return
		}

		//获取锁
		lock, _ := model.Redis.Int("setnx", "cronlock_" + strconv.Itoa(taskModel.Jobid), 1)
		if lock != 1 {
			logger.Error(fmt.Sprintf("获取redis lock 失败 %d, 跳过本机任务分发", lock))
			return
		}
		logger.Error(fmt.Sprintf("获取redis lock 成功 %d", lock))

		logger.Info(fmt.Sprintf("开始执行任务 - %s - 命令-%s", taskModel.JobName, taskModel.Cmd))
		taskResult := execJob(handler, taskModel, taskLogId)

		logger.Info(fmt.Sprintf("任务完成 - %s - 命令- %s - 执行结果- %s - 执行机器 - %s", taskModel.JobName, taskModel.Cmd, taskResult.Result, taskResult.Host))
		afterExecJob(taskModel, taskResult, taskLogId)

		//释放锁
		model.Redis.Int("del", "cronlock_" + strconv.Itoa(taskModel.Jobid))

		logger.Error("释放 redis lock 成功")
	}
	return taskFunc
}

// 执行具体任务
func execJob(handler Handler, taskModel model.FlCron, taskUniqueId int64) croninit.TaskResult {
	ret, err := handler.Run(taskModel, taskUniqueId)
	if err == nil {
		return croninit.TaskResult{Result: ret.Result, Err: ret.Err, Host : ret.Host, Status : ret.Status, Endtime : ret.Endtime}
	}
	return croninit.TaskResult{Result: ret.Result, Err: ret.Err, Host : ret.Host, Status : ret.Status, Endtime : ret.Endtime}
}

// 任务前置操作
func beforeExecJob(taskModel model.FlCron) (taskLogId int64) {
	taskLogId, err := createTaskLog(taskModel)
	if err != nil {
		logger.Error("任务开始执行#写入任务日志失败-", err)
		return
	}
	logger.Info("任务命令-", taskModel.Cmd)

	return taskLogId
}

// 任务执行后置操作
func afterExecJob(taskModel model.FlCron, taskResult croninit.TaskResult, taskLogId int64) {
	_, err := updateTaskLog(taskLogId, taskResult)
	if err != nil {
		logger.Error("任务结束#更新任务日志失败-", err)
	}

	// 发送邮件
	// go SendNotification(taskModel, taskResult)
}