package main

import(
	"github.com/qiusnay/gocron/service/rpcx"
	"github.com/qiusnay/gocron/init"
)
func main() {
	croninit.Init();
	
	// 初始化定时任务
	rpcx.Start()
}
