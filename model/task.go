package model

import (
	"time"
	"encoding/json"
)


type TaskResult struct {
	Result     string
	Err        error
	Host     string
	Status   int
	Endtime   string
}

const (
	Disabled int = 0 // 禁用
	Failure  int = 10001 // 失败
	Enabled  int = 1 // 启用
	Running  int = 10000 // 运行中
	Finish   int = 10002 // 完成
	Cancel   int = 3 // 取消
)

// 获取所有激活任务
func (task *FlCron) GetAllJobList() ([]FlCron, error) {
	list := make([]FlCron, 0)
	dberr := DB.Where("state = ?", "1").Find(&list).Error
	return list, dberr
}

// 获取依赖任务列表
func (task *FlCron) GetDependencyTaskList(ids string) ([]Task, error) {
	list := make([]Task, 0)
	if ids == "" {
		return list, nil
	}
	idList := strings.Split(ids, ",")
	taskIds := make([]interface{}, len(idList))
	for i, v := range idList {
		taskIds[i] = v
	}
	fields := "t.*"
	err := Db.Alias("t").
		Where("t.level = ?", TaskLevelChild).
		In("t.id", taskIds).
		Cols(fields).
		Find(&list)

	if err != nil {
		return list, err
	}

	return task.setHostsForTasks(list)
}

// 创建任务日志
func (task *FlLog) CreateTaskLog(taskModel FlCron) (int64, error) {
	jobdata, _ := json.Marshal(taskModel)
	taskLogModel := FlLog{
		Jobid : taskModel.Jobid,
		JobName : taskModel.JobName,
		Cmd : taskModel.Cmd,
		Runat : taskModel.Runat,
		Jobdata : string(jobdata),
		Createtime : time.Now(),
		Starttime : time.Now(),
		Status : Running,
	}
	DB.Create(&taskLogModel)
	return int64(taskLogModel.Id), nil
}

// 更新任务日志
func (task *FlLog) UpdateTaskLog(taskLogId int64, taskResult TaskResult) (int64, error) {
	taskLogModel := new(FlLog)
	var status int
	if taskResult.Err != nil {
		status = Failure
	} else {
		status = Finish
	}
    ts,_ := time.ParseInLocation("2006-01-02 15:04:05",taskResult.Endtime, time.Local)
	upResult := DB.Model(&taskLogModel).Where("id = ?", taskLogId).Updates(map[string]interface{}{
		"status": status, 
		"endtime": ts, //taskResult.Endtime, 
		"runat" : taskResult.Host, 
		"updatetime": time.Now(),
	})
	return upResult.RowsAffected , upResult.Error
}