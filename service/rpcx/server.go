package rpcx

import (
	"context"
	"fmt"
	"os"
	"errors"
	"os/exec"
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

type RpcService struct{
	Result  string
	Err  error
}


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
	out, err := c.ExecShell(ctx, req.Cmd)
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

// 执行shell命令，可设置执行超时时间
func (c *RpcService) ExecShell(ctx context.Context, command string) (string, error) {
	cmd := exec.Command("/bin/bash", "-c", command)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	resultChan := make(chan RpcService)
	go func() {
		Result, err := cmd.Output()
		resultChan <- RpcService{string(Result), err}
	}()
	select {
	case <-ctx.Done():
		if cmd.Process.Pid > 0 {
			syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		}
		return "", errors.New("timeout killed")
	case result := <-resultChan:
		return result.Result, result.Err
	}
}





