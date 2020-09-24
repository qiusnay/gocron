package model
import (
	"github.com/google/logger"
	"fmt"
	"time"
	"github.com/qiusnay/gocron/utils"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

const (
	dbPingInterval = 90 * time.Second
	dbMaxLiftTime  = 2 * time.Hour
)

var DB *gorm.DB

func Dbinit()(*gorm.DB, error) {
	dbConf := utils.GetConfig("database", "")
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
	return db, err
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



// 获取所有激活任务
func (task *FlCron) GetAllJobList() ([]FlCron, error) {
	list := make([]FlCron, 0)
	dberr := DB.Where("state = ?", "1").Find(&list).Error
	return list, dberr
}