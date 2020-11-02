package rpcx

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/google/logger"
	"github.com/qiusnay/gocron/model"
	"github.com/qiusnay/gocron/utils"
	"github.com/smallnest/rpcx/server"

	croninit "github.com/qiusnay/gocron/init"
	metrics "github.com/rcrowley/go-metrics"
	"github.com/smallnest/rpcx/serverplugin"
)

var (
	addr     = flag.String("addr", utils.GetLocalIP()+":8973", "server address")
	etcdAddr = flag.String("etcdAddr", "127.0.0.1:2379", "etcd address")
	basePath = flag.String("base", "com/example/rpcx", "prefix path")
)

type CronService struct {
	Result string
	Err    error
}

func Start() {
	flag.Parse()
	s := server.NewServer()

	// addRegistryPlugin(s)

	s.RegisterName("CronService", new(CronService), "")
	go func() {
		err := s.Serve("tcp", *addr)
		if err != nil {
			panic(err)
		}
	}()
}

func addRegistryPlugin(s *server.Server) {
	r := &serverplugin.EtcdV3RegisterPlugin{
		ServiceAddress: "tcp@" + *addr,
		EtcdServers:    []string{*etcdAddr},
		BasePath:       *basePath,
		Metrics:        metrics.NewRegistry(),
		UpdateInterval: time.Minute,
	}
	err := r.Start()
	if err != nil {
		panic(err)
	}
	s.Plugins.Add(r)
}

func (c *CronService) Run(ctx context.Context, req *model.FlCron, res *model.TaskResult) error {
	logger.Info(fmt.Sprintf("接收通道 : %+v", ctx))
	logger.Info(fmt.Sprintf("接收到请求 [taskid: %d cmd: %s err: %s, result : %s, host: %s]", req.Jobid, req.Cmd))
	var out string
	var err error
	queryCmd := AssembleCmd(req)
	switch req.Querytype {
	case "wget":
	case "curl":
		rpccurl := RpcServiceCurl{}
		out, err = rpccurl.ExecCurl(ctx, queryCmd)
		break
	default:
		rpcshell := RpcServiceShell{}
		out, err = rpcshell.ExecShell(ctx, queryCmd, req.Taskid)
	}
	res.Result = out
	res.Host = utils.GetLocalIP()
	res.Endtime = time.Now().Format("2006-01-02 15:04:05")
	if err != nil {
		res.Err = err
		res.Status = 10002
	} else {
		res.Err = nil
		res.Status = 10003
	}
	logger.Info(fmt.Sprintf("execute cmd end: [id: %d cmd: %s err: %s, endtime : %s, host: %s]", req.Id, queryCmd, err, res.Endtime, res.Host))

	return nil
}

func AssembleCmd(cron *model.FlCron) string {
	LogFile := GetLogFile(strconv.Itoa(cron.Jobid), cron.Taskid)
	// if utils.IsFile(LogFile) {
	// 	s, err := os.Stat(LogFile)
	// 	s.Chmod(0664)
	// }
	return cron.Cmd + " > " + LogFile
}

func GetLogFile(Jobid string, Taskid string) string {
	//设置日志目录
	LogDir := croninit.BASEPATH + "/log/cronlog/" + time.Now().Format("2006-01-02")
	if !utils.IsDir(LogDir) {
		// mkdir($LogDir, 0777, true);
		os.MkdirAll(LogDir, os.ModePerm)
	}
	return LogDir + "/cron-task-" + Jobid + "-" + Taskid + "-log.log"
}
