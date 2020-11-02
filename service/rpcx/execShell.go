package rpcx

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"syscall"

	"github.com/google/logger"
)

type RpcServiceShell struct {
	Result string
	Err    error
}

// 执行shell命令，可设置执行超时时间
func (c *RpcServiceShell) ExecShell(ctx context.Context, command string, taskid string) (string, error) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error(err)
		}
	}()
	// logger.Info(fmt.Sprintf("开始执行 exec pre %s, 当前时间 %s", taskid, time.Now().Format("2006-01-02 15:04:05")))
	cmd := exec.Command("/bin/bash", "-c", command)
	// logger.Info(fmt.Sprintf("开始执行 exec tail %s, 当前时间 %s", taskid, time.Now().Format("2006-01-02 15:04:05")))
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	resultChan := make(chan CronService)
	go func() {
		// logger.Info(fmt.Sprintf("go rountine %s, 当前时间 %s", taskid, time.Now().Format("2006-01-02 15:04:05")))
		Result, err := cmd.CombinedOutput()
		resultChan <- CronService{string(Result), err}
	}()
	select {
	case <-ctx.Done():
		logger.Info(fmt.Sprintf("执行超时 %s, 被客户端强制终止", taskid))
		if cmd.Process.Pid > 0 {
			syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		}
		return "", errors.New("timeout killed")
	case result := <-resultChan:
		logger.Info(fmt.Sprintf("通道正常返回 %s", taskid))
		return result.Result, result.Err
	}
}
