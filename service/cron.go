package service

import (
	"sync"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"github.com/jakecoffman/cron"
	"github.com/google/logger"
	"github.com/qiusnay/gocron/model"
	"github.com/ouqiang/goutil"
)

const (
	Disabled int8 = 0 // 禁用
	Failure  int8 = 0 // 失败
	Enabled  int8 = 1 // 启用
	Running  int8 = 1 // 运行中
	Finish   int8 = 2 // 完成
	Cancel   int8 = 3 // 取消
)

//定义一个空结构体
type FlCron struct{}

// 任务计数
type TaskCount struct {
	wg   sync.WaitGroup
	exit chan struct{}
}
// 并发队列
type ConcurrencyQueue struct {
	queue chan struct{}
}

type Handler interface {
	Run(taskModel model.FlCron, taskUniqueId int64) (string, error)
}

type TaskResult struct {
	Result     string
	Err        error
}

/****************************************/

func (tc *TaskCount) Add() {tc.wg.Add(1)}
func (tc *TaskCount) Done() {tc.wg.Done()}
func (tc *TaskCount) Exit() {
	tc.wg.Done()
	<-tc.exit
}
func (tc *TaskCount) Wait() {
	tc.Add()
	tc.wg.Wait()
	close(tc.exit)
}

func (cq *ConcurrencyQueue) Add() {
	cq.queue <- struct{}{}
}
func (cq *ConcurrencyQueue) Done() {
	<-cq.queue
}

var (
	// 定时任务调度管理器
	mycron *cron.Cron

	// 任务计数-正在运行的任务
	taskCount TaskCount

	// 并发队列, 限制同时运行的任务数量
	concurrencyQueue ConcurrencyQueue
)


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
		fl.Add(item)
	}
	logger.Infof("定时任务初始化完成")

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	for {
		s := <-c
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
func (fl FlCron) Add(taskModel model.FlCron) {
	taskModel.Rule = "* " + taskModel.Rule
	
	taskFunc := createJob(taskModel)

	cronName := strconv.Itoa(taskModel.Id)
	err := goutil.PanicToError(func() {
		mycron.AddFunc(taskModel.Rule, taskFunc, cronName)
	})
	if err != nil {
		logger.Error("添加任务到调度器失败#", err)
	}
}

/**
 * 创建任务
 **/
func createJob(taskModel model.FlCron) cron.FuncJob {
	handler := createHandler(taskModel)
	if handler == nil {
		return nil
	}
	taskFunc := func() {
		// taskCount.Add()
		// defer taskCount.Done()

		// taskLogId := beforeExecJob(taskModel)
		// if taskLogId <= 0 {
		// 	return
		// }
		var  taskLogId int64 = 1;
		// concurrencyQueue.Add()
		// defer concurrencyQueue.Done()

		logger.Info(fmt.Sprintf("开始执行任务#%s#命令-%s", taskModel.JobName, taskModel.Params))
		taskResult := execJob(handler, taskModel, taskLogId)
		logger.Info(fmt.Sprintf("任务完成#%s#命令-%s#执行结果-%s", taskModel.JobName, taskModel.Params, taskResult.Result))
		//afterExecJob(taskModel, taskResult, taskLogId)
	}
	return taskFunc
}

// 执行具体任务
func execJob(handler Handler, taskModel model.FlCron, taskUniqueId int64) TaskResult {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("panic#service/task.go:execJob#", err)
		}
	}()
	var output string
	var err error
	output, err = handler.Run(taskModel, taskUniqueId)
	if err == nil {
		return TaskResult{Result: output, Err: err}
	}
	return TaskResult{Result: output, Err: err}
}

// 任务前置操作
// func beforeExecJob(taskModel model.FlCron) (taskLogId int64) {
// 	taskLogId, err := createTaskLog(taskModel, Running)
// 	if err != nil {
// 		logger.Error("任务开始执行#写入任务日志失败-", err)
// 		return
// 	}
// 	logger.Info("任务命令-%s", taskModel.Params)

// 	return taskLogId
// }

// 任务执行后置操作
// func afterExecJob(taskModel model.FlCron, taskResult TaskResult, taskLogId int64) {
// 	_, err := updateTaskLog(taskLogId, taskResult)
// 	if err != nil {
// 		logger.Error("任务结束#更新任务日志失败-", err)
// 	}

// 	// 发送邮件
// 	// go SendNotification(taskModel, taskResult)
// }

func createHandler(taskModel model.FlCron) Handler {
	var handler Handler = nil
	switch taskModel.Type {
		case "wget":
		case "curl":
			handler = new(HTTPHandler)
			break
		default:
			handler = new(RPCHandler)
	}
	return handler
}




