package main

import(
	"github.com/qiusnay/gocron/service/rpcx"
	"github.com/qiusnay/gocron/init"
)
func main() {
	//1.检查DB是否已经初始化
	

	//2.cron项目初始化
	croninit.Init();
	
	//3.启动cron服务
	rpcx.Start()
}
