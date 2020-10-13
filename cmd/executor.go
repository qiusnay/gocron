package main

import(
	"github.com/qiusnay/gocron/service/rpcx"
	"github.com/qiusnay/gocron/init"
)
func main() {
	//1.cron项目初始化
	const logPath = "../log/gocronserver.log"

	croninit.Init(logPath);
	
	//3.启动cron服务
	rpcx.Start()
}
