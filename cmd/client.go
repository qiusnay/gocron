package main

import(
	"github.com/qiusnay/gocron/service/cron"
	// "github.com/google/logger"
	"github.com/qiusnay/gocron/init"
)

func main() {
	croninit.Init();

	// 初始化定时任务
	var serviceCron = cron.FlCron{}
	serviceCron.Initialize()
}