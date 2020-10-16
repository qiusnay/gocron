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
	// "github.com/qiusnay/gocron/init"
	"github.com/qiusnay/gocron/utils"
)

//定义一个空结构体
type FlCron struct{}

type Handler interface {
	Run(taskModel model.FlCron, taskUniqueId int64) (model.TaskResult, error)
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
	taskModel.Rule = "* " + taskModel.Rule

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
	taskFunc := func() {
		//获取锁
		lock, _ := model.Redis.Int("setnx", "cronlock_" + strconv.Itoa(taskModel.Jobid), 1)
		if lock != 1 {
			logger.Error(fmt.Sprintf("获取redis lock 失败 %d, 跳过本机任务分发", lock))
			return
		}
		logger.Error(fmt.Sprintf("获取redis lock 成功 %d", lock))
		//创建TASK
		taskLogId := beforeExecJob(taskModel)
		if taskLogId <= 0 {
			return
		}

		logger.Info(fmt.Sprintf("开始执行任务 - %s - 命令-%s", taskModel.JobName, taskModel.Cmd))
		taskResult := execJob(taskModel, taskLogId)

		logger.Info(fmt.Sprintf("任务完成 - %s - 命令- %s - 执行结果- %s - 执行机器 - %s", taskModel.JobName, taskModel.Cmd, taskResult.Result, taskResult.Host))
		afterExecJob(taskModel, taskResult, taskLogId)

		//释放锁
		model.Redis.Int("del", "cronlock_" + strconv.Itoa(taskModel.Jobid))

		logger.Error("释放 redis lock 成功")
	}
	return taskFunc
}

// 执行具体任务
func execJob(taskModel model.FlCron, taskUniqueId int64) model.TaskResult {
	ret, err := new(rpcx.RPCHandler).Run(taskModel, taskUniqueId)
	if err == nil {
		return model.TaskResult{Result: ret.Result, Err: ret.Err, Host : ret.Host, Status : ret.Status, Endtime : ret.Endtime}
	}
	return model.TaskResult{Result: ret.Result, Err: ret.Err, Host : ret.Host, Status : ret.Status, Endtime : ret.Endtime}
}

// 任务前置操作
func beforeExecJob(taskModel model.FlCron) (taskLogId int64) {
	taskLogModel := new(model.FlLog)
	taskLogId, err := taskLogModel.CreateTaskLog(taskModel)
	if err != nil {
		logger.Error("任务开始执行#写入任务日志失败-", err)
		return
	}
	logger.Info("任务命令-", taskModel.Cmd)

	return taskLogId
}

// 任务执行后置操作
func afterExecJob(taskModel model.FlCron, taskResult model.TaskResult, taskLogId int64) {
	taskLogModel := new(model.FlLog)
	_, err := taskLogModel.UpdateTaskLog(taskLogId, taskResult)
	if err != nil {
		logger.Error("任务结束#更新任务日志失败-", err)
	}

	// 发送邮件
	go SendNotification(taskModel, taskResult)

	// 执行依赖任务
	go execDependencyTask(taskModel, taskResult)
}

// 执行依赖任务, 多个任务并发执行
func execDependencyTask(taskModel models.Task, taskResult TaskResult) {
	// 父任务才能执行子任务
	if taskModel.Level != models.TaskLevelParent {
		return
	}

	// 是否存在子任务
	dependencyTaskId := strings.TrimSpace(taskModel.DependencyTaskId)
	if dependencyTaskId == "" {
		return
	}

	// 父子任务关系为强依赖, 父任务执行失败, 不执行依赖任务
	if taskModel.DependencyStatus == models.TaskDependencyStatusStrong && taskResult.Err != nil {
		logger.Infof("父子任务为强依赖关系, 父任务执行失败, 不运行依赖任务#主任务ID-%d", taskModel.Id)
		return
	}

	// 获取子任务
	model := new(models.Task)
	tasks, err := model.GetDependencyTaskList(dependencyTaskId)
	if err != nil {
		logger.Errorf("获取依赖任务失败#主任务ID-%d#%s", taskModel.Id, err.Error())
		return
	}
	if len(tasks) == 0 {
		logger.Errorf("依赖任务列表为空#主任务ID-%d", taskModel.Id)
	}
	for _, task := range tasks {
		task.Spec = fmt.Sprintf("依赖任务(主任务ID-%d)", taskModel.Id)
		ServiceTask.Run(task)
	}
}

// 发送任务结果通知
func SendNotification(taskModel models.Task, taskResult TaskResult) {
	var statusName string

	if taskModel.NotifyStatus == 3 {
		// 关键字匹配通知
		if !strings.Contains(taskResult.Result, taskModel.NotifyKeyword) {
			return
		}
	}
	if taskModel.NotifyStatus == 1 && taskResult.Err == nil {
		// 执行失败才发送通知
		return
	}
	if taskModel.NotifyType != 3 && taskModel.NotifyReceiverId == "" {
		return
	}
	if taskResult.Err != nil {
		statusName = "失败"
	} else {
		statusName = "成功"
	}
	// 发送通知
	msg := notify.Message{
		"task_type":        taskModel.NotifyType,
		"task_receiver_id": taskModel.NotifyReceiverId,
		"name":             taskModel.Name,
		"output":           taskResult.Result,
		"status":           statusName,
		"task_id":          taskModel.Id,
	}
	notify.Push(msg)
}