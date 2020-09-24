package model

import (
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

type FlCron struct {
	Id         int   `gorm:"primary_key;AUTO_INCREMENT" json:"id"`
	Taskid     int   `gorm:"type:int(50);not null;index:IX_taskid" json:"taskid"`
	JobName   string  `gorm:"type:varchar(550);not null" json:"job_name"`
	Params     string `gorm:"type:varchar(200);not null" json:"params"`
	Rule       string `gorm:"type:varchar(100);" json:"rule"`
	Runyear    string `gorm:"type:int(2);" json:"runyear"`
	Type       string `gorm:"type:varchar(50);not null" json:"type"`
	Name       string `gorm:"type:varchar(500);" json:"name"`
	State       string `gorm:"type:int(50);not null" json:"state"`
	Author      string `gorm:"type:varchar(50);" json:"author"`
	AdminMobile string `gorm:"type:varchar(500);" json:"admin_mobile"`
	AdminEmail  string `gorm:"type:varchar(500);" json:"admin_email"`
	Remark      string `gorm:"type:varchar(50);" json:"remark"`
}
func (FlCron) TableName() string {
	return "tb_cron_schedule"
}