package model

import (
	"time"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

//作业表
type FlCron struct {
	Id         uint   `gorm:"primary_key;AUTO_INCREMENT" json:"id"`
	Jobid     int   `gorm:"type:int(50);comment:'作业ID';not null;index:IX_jobid" json:"jobid"`
	JobName   string  `gorm:"type:varchar(550);comment:'作业名字';not null" json:"job_name"`
	Cmd     string `gorm:"type:varchar(200);comment:'执行命令';not null" json:"cmd"`
	Rule       string `gorm:"type:varchar(100);comment:'cron规则';" json:"rule"`
	Runyear    string `gorm:"type:int(2);comment:'执行年份 : 1每年,2今年';" json:"runyear"`
	Querytype       string `gorm:"type:varchar(50);comment:'命令类型:wget,curl,phpget,other';not null" json:"querytype"`
	State       string `gorm:"type:int(50);comment:'作业状态';not null" json:"state"`
	Author      string `gorm:"type:varchar(50);comment:'作业属主';" json:"author"`
	AdminMobile string `gorm:"type:varchar(10);comment:'属主电话';" json:"admin_mobile"`
	AdminEmail  string `gorm:"type:varchar(100);comment:'属主邮箱';" json:"admin_email"`
	Remark      string `gorm:"type:varchar(500);comment:'备注';" json:"remark"`
	Createtime  time.Time `gorm:"type:varchar(50);comment:'创建时间';" json:"create_time"`
	Runat       string `gorm:"type:varchar(50);comment:'执行机器'" json:"runat"`
	Overflow  string `gorm:"type:int(2);comment:'超时机制 1:放弃当前任务,继续执行上次(默认),2:强制终止上次,并启动新的任务,3:健康检查,不告警'" json:"overflow"`
}
func (FlCron) TableName() string {
	return "tb_cron_schedule"
}

//用户表
type FlUser struct {
	Id         uint   `gorm:"primary_key;AUTO_INCREMENT" json:"id"`
	Userid     int   `gorm:"type:int(50);comment:'用户ID';not null;index:IX_user_id" json:"user_id"`
	Username   string  `gorm:"type:varchar(550);comment:'用户名字';not null" json:"user_name"`
	Email     string `gorm:"type:varchar(200);comment:'邮箱';not null" json:"email"`
	Password       string `gorm:"type:varchar(100);comment:'密码';" json:"password"`
	Isactive       string `gorm:"type:varchar(50);comment:'是否激活 0:未激活,1:激活';not null" json:"is_active"`
	Createtime  time.Time `gorm:"type:varchar(50);comment:'创建时间';" json:"create_time"`
	LastLogin  time.Time `gorm:"type:varchar(50);comment:'上次登录时间';" json:"last_login"`
}
func (FlUser) TableName() string {
	return "tb_cron_user"
}

//CRON日志表
type FlLog struct {
	Id         uint   `gorm:"primary_key;AUTO_INCREMENT" json:"id"`
	Jobid     int   `gorm:"type:int(50);comment:'作业ID';not null;index:IX_jobid" json:"jobid"`
	JobName   string  `gorm:"type:varchar(550);comment:'作业名字';not null" json:"job_name"`
	Status    int   `gorm:"type:int(50);comment:'任务状态:10000:等待执行,10001:执行中,10002:执行成功,10006:超时锁定,其他:出错';not null;index:IX_taskid" json:"status"`
	Starttime  time.Time `gorm:"type:varchar(50);comment:'开始时间';" json:"starttime"`
	Endtime    time.Time `gorm:"type:varchar(50);comment:'结束时间';not null" json:"endtime"`
	Cmd     string `gorm:"type:varchar(200);comment:'执行命令';not null" json:"cmd"`
	Runat       string `gorm:"type:varchar(50);comment:'执行机器'" json:"runat"`
	Jobdata       string `gorm:"type:varchar(1000);comment:'Job data'" json:"jobdata"`
	Createtime  time.Time `gorm:"type:varchar(50);comment:'创建时间';" json:"createtime"`
	Updatetime  time.Time `gorm:"type:varchar(50);comment:'修改时间';" json:"updatetime"`
}
func (FlLog) TableName() string {
	return "tb_cron_log"
}

//cron用户日志表
type Fluserlog struct {
	Id         uint   `gorm:"primary_key;AUTO_INCREMENT" json:"id"`
	Jobid     int   `gorm:"type:int(50);comment:'作业ID';not null;index:IX_jobid" json:"jobid"`
	JobName   string  `gorm:"type:varchar(550);comment:'作业名字';not null" json:"job_name"`
	Userid     int   `gorm:"type:int(50);comment:'用户ID';not null;index:IX_user_id" json:"user_id"`
	Modifycontent  string `gorm:"type:varchar(1000);comment:'Job data'" json:"modify_content"`
	Mtype      int   `gorm:"type:int(50);comment:'修改类型(creat,modify)';not null;" json:"mtype"`
	Createtime  time.Time `gorm:"type:varchar(50);comment:'创建时间';" json:"createtime"`
}
func (Fluserlog) TableName() string {
	return "tb_cron_user_log"
}