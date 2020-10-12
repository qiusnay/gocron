package rpcx

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"flag"
	"syscall"
	"time"
	"github.com/smallnest/rpcx/server"
	"github.com/google/logger"
	"github.com/qiusnay/gocron/model"
	"github.com/qiusnay/gocron/utils"
	"github.com/qiusnay/gocron/init"
)


var (
	addr = flag.String("addr", "localhost:8972", "server address")
)

func Start() {
	flag.Parse()
	s := server.NewServer()
	s.RegisterName("RpcService", new(RpcService), "")
	logger.Info("server listen on %s", addr)
	go func() {
		err := s.Serve("tcp", *addr)
		if err != nil {
			panic(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	for {
		s := <-c
		logger.Info("收到信号 -- ", s)
		switch s {
		case syscall.SIGHUP:
			logger.Info("收到终端断开信号, 忽略")
		case syscall.SIGINT, syscall.SIGTERM:
			logger.Info("应用准备退出")
			return
		}
	}

}

func (c *RpcService) Run(ctx context.Context, req *model.FlCron, res *croninit.TaskResult) error {
	var out string
	var err error
	switch req.Querytype {
		case "wget":
		case "curl":
			out, err := c.ExecCurl(ctx, req.Cmd)
			break
		default:
			out, err := c.ExecShell(ctx, req.Cmd)
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


