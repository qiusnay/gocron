package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/logger"
	croninit "github.com/qiusnay/gocron/init"
	"github.com/qiusnay/gocron/service/cron"
)

var logPath = "../log/dispacher." + time.Now().Format("2006-01-02") + ".log"

func main() {
	croninit.Init(logPath)
	// 初始化定时任务
	var serviceCron = cron.VCron{}
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
