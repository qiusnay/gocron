package croninit

import(
	"flag"
	"os"
	"github.com/google/logger"
	"github.com/qiusnay/gocron/model"
)

var verbose = flag.Bool("verbose", false, "print info level logs to stdout")

const logPath = "/Users/qiusnay/project/golang/gocron/log/gocronserver.log"

func Init() {
	lf, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
	  logger.Fatalf("Failed to open log file: %v", err)
	}
	defer lf.Close()
	defer logger.Init("Logger", *verbose, true, lf).Close()

	model.Dbinit()
	//defer db.Close()
}
