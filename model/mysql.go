package model

import (
	"fmt"
	"time"
	"github.com/qiusnay/gocron/comm"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var db *gorm.DB

type mysql struct {
}

func (mydb *mysql) DBinit() {
	dbConf := comm.GetConfig("database", "")
	db, err := gorm.Open("mysql", fmt.Sprintf("%s:%s@(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", dbConf["username"], dbConf["password"], dbConf["host"], dbConf["port"], dbConf["database"]))
	if err != nil {
		panic("连接数据库失败")
	}
	defer db.Close()
}

func (mydb *mysql) Migrate() {
	db.AutoMigrate(&Cron{}) //存在就自动适配表，也就说原先没字段的就增加字段
}



func DB() *gorm.DB {
	return db.New()
}

type Cron struct {
	Id         uint   `gorm:"primary_key;AUTO_INCREMENT" json:"id"`
	Taskid     uint   `gorm:"type:int(50);not null;index:IX_taskid" json:"taskid"`
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
func (Cron) TableName() string {
	return "tb_cron_schedule"
}


type User struct {
    Birthday     time.Time
    Age          int
    Name         string  `gorm:"size:255"`       // string默认长度为255, 使用这种tag重设。
    Num          int     `gorm:"AUTO_INCREMENT"` // 自增

    CreditCard        CreditCard      // One-To-One (拥有一个 - CreditCard表的UserID作外键)
    Emails            []Email         // One-To-Many (拥有多个 - Email表的UserID作外键)

    BillingAddress    Address         // One-To-One (属于 - 本表的BillingAddressID作外键)
    BillingAddressID  int

    ShippingAddress   Address         // One-To-One (属于 - 本表的ShippingAddressID作外键)
    ShippingAddressID int

    IgnoreMe          int `gorm:"-"`   // 忽略这个字段
    Languages         []Language `gorm:"many2many:user_languages;"` // Many-To-Many , 'user_languages'是连接表
}

type Email struct {
    ID      int
    UserID  int     `gorm:"index"` // 外键 (属于), tag `index`是为该列创建索引
    Email   string  `gorm:"type:varchar(100);unique_index"` // `type`设置sql类型, `unique_index` 为该列设置唯一索引
    Subscribed bool
}

type Address struct {
    ID       int
    Address1 string         `gorm:"not null;unique"` // 设置字段为非空并唯一
    Address2 string         `gorm:"type:varchar(100);unique"`
    Post     string `gorm:"not null"`
}

type Language struct {
    ID   int
    Name string `gorm:"index:idx_name_code"` // 创建索引并命名，如果找到其他相同名称的索引则创建组合索引
    Code string `gorm:"index:idx_name_code"` // `unique_index` also works
}

type CreditCard struct {
    UserID  uint
    Number  string
}
