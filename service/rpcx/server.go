package rpcx

import (
	"context"
	"fmt"
	"flag"
	"time"
	"github.com/smallnest/rpcx/server"
	"github.com/google/logger"
	"github.com/qiusnay/gocron/model"
	"github.com/qiusnay/gocron/utils"
	// "github.com/qiusnay/gocron/init"
	"github.com/smallnest/rpcx/serverplugin"
	metrics "github.com/rcrowley/go-metrics"
)

var (
	addr     = flag.String("addr", utils.GetLocalIP() + ":8973", "server address")
	etcdAddr = flag.String("etcdAddr", "10.200.105.49:2379", "etcd address")
	basePath = flag.String("base", "com/example/rpcx", "prefix path")
)

type RpcService struct{
	Result  string
	Err  error
}

func Start() {
	flag.Parse()
	s := server.NewServer()

	addRegistryPlugin(s)

	s.RegisterName("RpcService", new(RpcService), "")
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

func (c *RpcService) Run(ctx context.Context, req *model.FlCron, res *model.TaskResult) error {
	var out string
	var err error
	switch req.Querytype {
		case "wget":
		case "curl":
			rpccurl := RpcServiceCurl{}
			out, err = rpccurl.ExecCurl(ctx, req.Cmd)
			break
		default:
			rpcshell := RpcServiceShell{}
			out, err = rpcshell.ExecShell(ctx, req.Cmd)
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
	logger.Info(fmt.Sprintf("execute cmd end: [id: %d cmd: %s err: %s, result : %s, host: %s]", req.Id, req.Cmd, err, out, res.Host))

	return nil
}


