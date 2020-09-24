package service

// import (
// 	"fmt"
// 	"time"
// 	"github.com/qiusnay/gocron/model"
// )

// // 创建任务日志
// func createTaskLog(taskModel model.FlCron, status Status) (int64, error) {
// 	taskLogModel := new(model.TaskLog)
// 	taskLogModel.TaskId = taskModel.Id
// 	taskLogModel.Name = taskModel.JobName
// 	taskLogModel.Spec = taskModel.Rule
// 	taskLogModel.Protocol = taskModel.Type
// 	taskLogModel.Command = taskModel.Params
// 	if taskModel.Protocol == models.TaskRPC {
// 		aggregationHost := ""
// 		for _, host := range taskModel.Hosts {
// 			aggregationHost += fmt.Sprintf("%s - %s<br>", host.Alias, host.Name)
// 		}
// 		taskLogModel.Hostname = aggregationHost
// 	}
// 	taskLogModel.StartTime = time.Now()
// 	taskLogModel.Status = status
// 	insertId, err := taskLogModel.Create()

// 	return insertId, err
// }

// // 更新任务日志
// func updateTaskLog(taskLogId int64, taskResult TaskResult) (int64, error) {
// 	taskLogModel := new(model.TaskLog)
// 	var status Status
// 	result := taskResult.Result
// 	if taskResult.Err != nil {
// 		status = Failure
// 	} else {
// 		status = Finish
// 	}
// 	return taskLogModel.Update(taskLogId, models.CommonMap{
// 		"status":      status,
// 		"result":      result,
// 	})

// }