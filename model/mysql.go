package model

import (
	"github.com/qiusnay/gocron/config"
	"github.com/jinzhu/gorm"
    _ "github.com/jinzhu/gorm/dialects/mysql"
)

type Mysql struct {
}

func init() {
	config.LoadConfig() // 读取DB配置

	db, err := gorm.Open("mysql", "root:root@tcp(192.168.100.60:3306)/51fanli_cron?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		panic("连接数据库失败")
	}
	defer db.Close()
}
