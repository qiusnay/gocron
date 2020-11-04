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
	EmailNotify int = 1
	ForceKill   int = 2
	HealthCheck int = 3
	BothRun     int = 4
)

// 获取所有激活任务
func (task *VCron) GetAllJobList() ([]VCron, error) {
	list := make([]VCron, 0)
	dberr := DB.Where("state = ?", "1").Find(&list).Error
	return list, dberr
}

func (task *VCron) GetJobInfo(Jobid int64) ([]VCron, error) {
	jobinfo := make([]VCron, 0)
	dberr := DB.Where("jobid = ?", Jobid).First(&jobinfo).Error
	return jobinfo, dberr
}

// 创建任务日志
func (task *VLog) CreateTaskLog(Job VCron) (int64, error) {
	jobdata, _ := json.Marshal(Job)
	TaskLog := VLog{
		Jobid:      Job.Jobid,
		JobName:    Job.JobName,
		Cmd:        Job.Cmd,
		Runat:      Job.Runat,
		Jobdata:    string(jobdata),
		Createtime: time.Now(),
		Updatetime: time.Now(),
		Starttime:  time.Now(),
		Status:     Job.TaskStatus,
	}
	DB.Create(&TaskLog)
	return int64(TaskLog.Id), nil
}

// 更新任务日志
func (task *VLog) UpdateTaskLog(taskLogId int64, taskResult TaskResult) (int64, error) {
	taskLogModel := new(VLog)
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
