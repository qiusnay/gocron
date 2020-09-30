package main

import(
	"github.com/qiusnay/gocron/service/cron"
	// "github.com/google/logger"
	"github.com/qiusnay/gocron/init"
)

func main() {
	const logPath = "/Users/qiusnay/project/golang/gocron/log/gocron.log"

	croninit.Init(logPath);
	// 初始化定时任务
	var serviceCron = cron.FlCron{}
	serviceCron.Initialize()
}