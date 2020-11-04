package croninit

import (
	"flag"
	"os"

	// "fmt"
	"path/filepath"

	"github.com/google/logger"
	"github.com/qiusnay/gocron/model"
)

var verbose = flag.Bool("verbose", false, "print info level logs to stdout")

const (
	CronNormal      int64 = 10000 // 正常
	CronRunning     int64 = 10006 //执行中
	CronSucess      int64 = 10001 // 成功
	CronError       int64 = 10002 // 失败
	CronTimeOut     int64 = 10003 // 超时
	CronForceKill   int64 = 10004 // 超时强杀
	CronHealthCheck int64 = 10005 // 健康检查
)

var BASEPATH string

func Init(logPath string) {
	lf, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		logger.Fatalf("Failed to open log file: %v", err)
	}

	BASEPATH, _ = filepath.Abs(filepath.Dir("../"))

	// defer lf.Close()
	// defer logger.Init("Logger", *verbose, true, lf).Close()
	logger.Init("Logger", *verbose, true, lf)

	//defer db.Close()
	model.Dbinit()

	model.Redis.InitPool()
}
