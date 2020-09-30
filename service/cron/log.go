package cron

import (
	// "fmt"
	"time"
	"encoding/json"
	"github.com/qiusnay/gocron/model"
	"github.com/qiusnay/gocron/init"
	// "github.com/google/logger"
)

// 创建任务日志
func createTaskLog(taskModel model.FlCron) (int64, error) {
	jobdata, _ := json.Marshal(taskModel)
	taskLogModel := model.FlLog{
		Jobid : taskModel.Jobid,
		JobName : taskModel.JobName,
		Cmd : taskModel.Cmd,
		Runat : taskModel.Runat,
		Jobdata : string(jobdata),
		Createtime : time.Now(),
		Starttime : time.Now(),
		Status : Running,
	}
	model.DB.Create(&taskLogModel)
	return int64(taskLogModel.Id), nil
}

// 更新任务日志
func updateTaskLog(taskLogId int64, taskResult croninit.TaskResult) (int64, error) {
	taskLogModel := new(model.FlLog)
	var status int
	if taskResult.Err != nil {
		status = Failure
	} else {
		status = Finish
	}
    ts,_ := time.ParseInLocation("2006-01-02 15:04:05",taskResult.Endtime, time.Local)
	upResult := model.DB.Model(&taskLogModel).Where("id = ?", taskLogId).Updates(map[string]interface{}{
		"status": status, 
		"endtime": ts, //taskResult.Endtime, 
		"runat" : taskResult.Host, 
		"updatetime": time.Now(),
	})
	return upResult.RowsAffected , upResult.Error
}