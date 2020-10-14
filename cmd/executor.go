package main

import(
	"os"
	"os/signal"
	"syscall"
	"github.com/qiusnay/gocron/service/rpcx"
	"github.com/qiusnay/gocron/init"
	"github.com/google/logger"
)
func main() {
	//1.cron项目初始化
	const logPath = "../log/gocronserver.log"

	croninit.Init(logPath);
	
	//3.启动cron服务
	rpcx.Start()

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
