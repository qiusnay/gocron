package model

import (
	"github.com/qiusnay/gocron/config"
	"gorm.io/gorm"
	"gorm.io/driver/mysql"
	_ "github.com/go-sql-driver/mysql"
)

type Mysql struct {

}

func init() {
	config.LoadConfig() // 读取DB配置
	db, err := gorm.Open("mysql", "user:password@/dbname?charset=utf8&parseTime=True&loc=Local")
  	defer db.Close()_ "github.com/go-sql-driver/mysql"
}