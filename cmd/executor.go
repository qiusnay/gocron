package main

import (
	"time"

	croninit "github.com/qiusnay/gocron/init"
	"github.com/qiusnay/gocron/service/rpc"
)

var logPath = "../log/executor." + time.Now().Format("2006-01-02") + ".log"

func main() {
	croninit.Init(logPath)
	var Executor = rpc.MyServer{}
	Executor.Start()
}
