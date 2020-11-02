package model

import (
	"fmt"
	"time"

	"github.com/google/logger"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/qiusnay/gocron/utils"
)

const (
	dbPingInterval = 90 * time.Second
	dbMaxLiftTime  = 2 * time.Hour
)

var DB *gorm.DB

func Dbinit() (*gorm.DB, error) {
	dbConf := utils.GetConfig("database_local", "")
	db, err := gorm.Open("mysql", fmt.Sprintf("%s:%s@(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", dbConf["username"], dbConf["password"], dbConf["host"], dbConf["port"], dbConf["database"]))
	if err != nil {
		logger.Infof("database connect erro : %s", err)
		return db, err
		//panic("连接数据库失败")
	}
	DB = db

	// db.LogMode(true)

	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	db.DB().SetMaxIdleConns(10)

	// SetMaxOpenConns sets the maximum number of open connections to the database.
	db.DB().SetMaxOpenConns(50)

	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	db.DB().SetConnMaxLifetime(time.Hour)

	go keepDbAlived(db)
	go Automigrate()

	return db, err
}

func Automigrate() {
	if !DB.HasTable("tb_cron_schedule") {
		DB.Set("gorm:table_options", "ENGINE=InnoDB  DEFAULT CHARSET=utf8 AUTO_INCREMENT=1 comment 'CRON作业表'").CreateTable(&FlCron{})
		DB.Set("gorm:table_options", "ENGINE=InnoDB  DEFAULT CHARSET=utf8 AUTO_INCREMENT=1 comment 'CRON用户表'").CreateTable(&FlUser{})
		DB.Set("gorm:table_options", "ENGINE=InnoDB  DEFAULT CHARSET=utf8 AUTO_INCREMENT=1 comment 'CRON任务调度日志表'").CreateTable(&FlLog{})
		DB.Set("gorm:table_options", "ENGINE=InnoDB  DEFAULT CHARSET=utf8 AUTO_INCREMENT=1 comment 'cron用户修改日志表'").CreateTable(&Fluserlog{})
	} else {
		// fmt.Println("检查更新.......")
		DB.AutoMigrate(&FlCron{})
		DB.AutoMigrate(&FlUser{})
		DB.AutoMigrate(&FlLog{})
		DB.AutoMigrate(&Fluserlog{})
		// fmt.Println("数据已更新!")
	}
}

func keepDbAlived(db *gorm.DB) {
	t := time.Tick(dbPingInterval)
	var err error
	for {
		<-t
		err = db.DB().Ping()
		if err != nil {
			logger.Infof("database ping: %s", err)
		} else {
			logger.Infof("database ping sucess")
		}
	}
}
