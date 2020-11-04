package main

import (
	"time"

	croninit "github.com/qiusnay/gocron/init"
	"github.com/qiusnay/gocron/service/cron"
)

var logPath = "../log/dispacher." + time.Now().Format("2006-01-02") + ".log"

func main() {
	croninit.Init(logPath)
	var Dispacher = cron.VCron{}
	Dispacher.Start()
}
