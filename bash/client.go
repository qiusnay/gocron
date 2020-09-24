package main

import(
	"flag"
	"os"
	"github.com/qiusnay/gocron/service"
	"github.com/google/logger"
	"github.com/qiusnay/gocron/model"
)

var verbose = flag.Bool("verbose", false, "print info level logs to stdout")

const logPath = "/Users/qiusnay/project/golang/gocron/log/gocron.log"


func main() {
	lf, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
	  logger.Fatalf("Failed to open log file: %v", err)
	}
	defer lf.Close()
	defer logger.Init("Logger", *verbose, true, lf).Close()


	db, err := model.Dbinit()
	if err != nil {
		logger.Error("err open databases", err)
		return
	}
	defer db.Close()


	// 初始化定时任务
	var serviceCron = service.FlCron{}
	serviceCron.Initialize()
}