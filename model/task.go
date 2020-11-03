package model

import (
	"encoding/json"
	"time"
)

type TaskResult struct {
	Result  string
	Err     string
	Host    string
	Status  int64
	Endtime string
}

const (
	Disabled int = 0     // 禁用
	Failure  int = 10001 // 失败
	Enabled  int = 1     // 启用
	Running  int = 10000 // 运行中
	Finish   int = 10002 // 完成
	Cancel   int = 3     // 取消

	EmailNotify int = 1
	ForceKill   int = 2
	HealthCheck int = 3
	BothRun     int = 4
)

// 获取所有激活任务
func (task *FlCron) GetAllJobList() ([]FlCron, error) {
	list := make([]FlCron, 0)
	dberr := DB.Where("state = ?", "1").Find(&list).Error
	return list, dberr
}

func (task *FlCron) GetJobInfo(Jobid int64) ([]FlCron, error) {
	jobinfo := make([]FlCron, 0)
	dberr := DB.Where("jobid = ?", Jobid).First(&jobinfo).Error
	return jobinfo, dberr
}

// 创建任务日志
func (task *FlLog) CreateTaskLog(taskModel FlCron) (int64, error) {
	jobdata, _ := json.Marshal(taskModel)
	ts, _ := time.ParseInLocation("2006-01-02 15:04:05", "2006-01-02 15:04:05", time.Local)
	taskLogModel := FlLog{
		Jobid:      taskModel.Jobid,
		JobName:    taskModel.JobName,
		Cmd:        taskModel.Cmd,
		Runat:      taskModel.Runat,
		Jobdata:    string(jobdata),
		Createtime: time.Now(),
		Updatetime: time.Now(),
		Endtime:    ts,
		Starttime:  time.Now(),
		Status:     Running,
	}
	DB.Create(&taskLogModel)
	return int64(taskLogModel.Id), nil
}

// 更新任务日志
func (task *FlLog) UpdateTaskLog(taskLogId int64, taskResult TaskResult) (int64, error) {
	taskLogModel := new(FlLog)
	ts, _ := time.ParseInLocation("2006-01-02 15:04:05", taskResult.Endtime, time.Local)
	upResult := DB.Model(&taskLogModel).Where("id = ?", taskLogId).Updates(map[string]interface{}{
		"status":     taskResult.Status,
		"error":      taskResult.Err,
		"endtime":    ts, //taskResult.Endtime,
		"runat":      taskResult.Host,
		"updatetime": time.Now(),
	})
	return upResult.RowsAffected, upResult.Error
}
