package rpc

import (
	"context"
	"errors"
	"os/exec"
	"syscall"

	"github.com/google/logger"
)

type RpcServiceShell struct {
	Result string
	Err    error
}

// 执行shell命令，可设置执行超时时间
func (c *RpcServiceShell) ExecShell(ctx context.Context, command string, taskid int64) CronResponse {
	defer func() {
		if err := recover(); err != nil {
			logger.Error(err)
		}
	}()
	cmd := exec.Command("/bin/bash", "-c", command)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	resultChan := make(chan CronResponse)
	go func() {
		err, _ := cmd.CombinedOutput()
		resultChan <- CronResponse{"", CronSucess, "", errors.New(string(err))} // 正常结束
	}()
	select {
	case <-ctx.Done():
		if cmd.Process.Pid > 0 {
			syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		}
		return CronResponse{"", CronTimeOut, "", errors.New("timeout killed,kill the process")}
	case result := <-resultChan:
		return result
	}
}
