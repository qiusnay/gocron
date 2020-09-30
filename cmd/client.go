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



/**
待办事项 :
 1.当新添加了一个作业后.一直在运行的进程不会退出.而且不会自动载入新的作业运行
 2.如何做到发布可以随意启动的问题,架构设计脱钩方案
 3.多机器,多个group功能支持
 4.http命令支持
 5.支持rabbitmq设计
 6.支持任务间耦合设计
 7.无中心化设计
 8.考虑是需要继续使用redis作锁设计
*/