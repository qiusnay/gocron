package server

import (
	"context"
	"fmt"
	"os"
	"errors"
	"os/exec"
	"os/signal"
	"flag"
	"syscall"
	"github.com/smallnest/rpcx/server"
	"github.com/google/logger"
	"github.com/qiusnay/gocron/model"
)

type RpcService struct{}

type Reply struct {
	Output string
	Err    error
}

var (
	addr = flag.String("addr", "localhost:8972", "server address")
)

func Initialize() {
	flag.Parse()
	s := server.NewServer()
	s.RegisterName("RpcService", new(RpcService), "")
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

func (c *RpcService) Run(ctx context.Context, req *model.FlCron, res *Reply) error {
	out, err := c.ExecShell(ctx, req.Params)
	res.Output = out
	if err != nil {
		res.Err = err
	} else {
		res.Err = nil
	}
	logger.Info(fmt.Sprintf("execute cmd end: [id: %d cmd: %s err: %s, result : %s]", req.Id, req.Params, err, out))

	return nil
}

// 执行shell命令，可设置执行超时时间
func (c *RpcService) ExecShell(ctx context.Context, command string) (string, error) {
	cmd := exec.Command("/bin/bash", "-c", command)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	resultChan := make(chan Reply)
	go func() {
		output, err := cmd.Output()
		resultChan <- Reply{string(output), err}
	}()
	select {
	case <-ctx.Done():
		if cmd.Process.Pid > 0 {
			syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		}
		return "", errors.New("timeout killed")
	case result := <-resultChan:
		return result.Output, result.Err
	}
}





