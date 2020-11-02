package rpcx

import (
	"context"
	"flag"
	"strconv"
	"time"

	// "fmt"

	"github.com/google/logger"
	"github.com/qiusnay/gocron/model"
	cron "github.com/robfig/cron/v3"
	"github.com/smallnest/rpcx/client"
	// "github.com/qiusnay/gocron/init"
	// "github.com/qiusnay/gocron/service/cron"
	// "github.com/google/logger"
)

// RPC调用执行任务
type CronClient struct{}

var Parser = cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)

func (c *CronClient) Run(jobModel model.FlCron, taskId int64) (result model.TaskResult, err error) {
	flag.Parse()
	jobModel.Taskid = strconv.FormatInt(taskId, 10)
	d := client.NewPeer2PeerDiscovery("tcp@"+*addr, "")
	// d := client.NewEtcdV3Discovery(*basePath, "CronService", []string{*etcdAddr}, nil)

	s, _ := Parser.Parse(jobModel.Rule)
	expreTime := s.Next(time.Now()).Unix() - time.Now().Unix() - 5 // 5秒预留空间
	logger.Error("添加任务到调度器失败#", time.Duration(expreTime)*time.Second)

	option := client.DefaultOption
	// 设置了响应服务端的超时时间为5秒
	option.ConnectTimeout = time.Duration(expreTime) * time.Second
	xclient := client.NewXClient("CronService", client.Failfast, client.RandomSelect, d, option)
	// xclient := client.NewXClient("CronService", client.Failover, client.RoundRobin, d, client.DefaultOption)
	defer xclient.Close()

	resultChan := make(chan model.TaskResult)
	go func() {
		reply := &model.TaskResult{}

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(expreTime)*time.Second)
		defer cancel()

		xclient.Call(ctx, "Run", jobModel, reply)
		resultChan <- model.TaskResult{Result: reply.Result, Err: reply.Err, Host: reply.Host, Status: reply.Status, Endtime: reply.Endtime}
	}()
	var aggregationErr error = nil
	rpcResult := <-resultChan
	if rpcResult.Err != nil {
		aggregationErr = rpcResult.Err
	}
	return rpcResult, aggregationErr
}
