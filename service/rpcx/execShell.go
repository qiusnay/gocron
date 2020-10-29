package rpcx

import (
	"context"
	"errors"
	"os/exec"
	"syscall"
	// "github.com/qiusnay/gocron/model"
	// "github.com/qiusnay/gocron/utils"
	// "github.com/qiusnay/gocron/init"
)

type RpcServiceShell struct {
	Result string
	Err    error
}

// 执行shell命令，可设置执行超时时间
func (c *RpcServiceShell) ExecShell(ctx context.Context, command string) (string, error) {
	cmd := exec.Command("/bin/bash", "-c", command)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	resultChan := make(chan CronService)
	go func() {
		Result, err := cmd.Output()
		resultChan <- CronService{string(Result), err}
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
