package main

import(
	"os"
	"os/signal"
	"syscall"
	"github.com/qiusnay/gocron/service/cron"
	"github.com/google/logger"
	"github.com/qiusnay/gocron/init"
)

func main() {
	const logPath = "../log/gocron.log"

	croninit.Init(logPath);
	// 初始化定时任务
	var serviceCron = cron.FlCron{}
	serviceCron.Initialize()
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	for {
		s := <-c
		logger.Info("收到信号 -- ", s)
		switch s {
		case syscall.SIGHUP:
			logger.Info("收到终端断开信号, 忽略")
		case syscall.SIGINT, syscall.SIGTERM:
			logger.Info("应用准备退出")
			return
		}
	}
}