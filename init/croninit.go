package croninit

import(
	"flag"
	"os"
	// "fmt"
	"github.com/google/logger"
	"github.com/qiusnay/gocron/model"
)

var verbose = flag.Bool("verbose", false, "print info level logs to stdout")

type TaskResult struct {
	Result     string
	Err        error
	Host     string
	Status   int
	Endtime   string
}

type RpcService struct{
	Result  string
	Err  error
}

func Init(logPath string) {
	lf, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
	  logger.Fatalf("Failed to open log file: %v", err)
	}

	// defer lf.Close()
	// defer logger.Init("Logger", *verbose, true, lf).Close()
	logger.Init("Logger", *verbose, true, lf)

	//defer db.Close()
	model.Dbinit()
}
