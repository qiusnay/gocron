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
