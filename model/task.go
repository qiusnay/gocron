package model

// 获取所有激活任务
func (task *FlCron) GetAllJobList() ([]FlCron, error) {
	list := make([]FlCron, 0)
	dberr := DB.Where("state = ?", "1").Find(&list).Error
	return list, dberr
}