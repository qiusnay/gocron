package rpcx

import (
	"context"
	"flag"
	"strconv"

	// "fmt"

	"github.com/qiusnay/gocron/model"
	"github.com/smallnest/rpcx/client"
	// "github.com/qiusnay/gocron/init"
	// "github.com/qiusnay/gocron/service/cron"
	// "github.com/google/logger"
)

// RPC调用执行任务
type CronClient struct{}

func (c *CronClient) Run(taskModel model.FlCron, taskId int64) (result model.TaskResult, err error) {
	flag.Parse()
	taskModel.Taskid = strconv.FormatInt(taskId, 10)
	// d := client.NewPeer2PeerDiscovery("tcp@"+*addr, "")
	d := client.NewEtcdV3Discovery(*basePath, "CronService", []string{*etcdAddr}, nil)
	// xclient := client.NewXClient("RpcService", client.Failtry, client.RandomSelect, d, opt)
	xclient := client.NewXClient("CronService", client.Failover, client.RoundRobin, d, client.DefaultOption)
	defer xclient.Close()

	resultChan := make(chan model.TaskResult)

	go func() {
		reply := &model.TaskResult{}
		xclient.Call(context.Background(), "Run", taskModel, reply)

		resultChan <- model.TaskResult{Result: reply.Result, Err: reply.Err, Host: reply.Host, Status: reply.Status, Endtime: reply.Endtime}
	}()
	var aggregationErr error = nil
	rpcResult := <-resultChan
	if rpcResult.Err != nil {
		aggregationErr = rpcResult.Err
	}
	return rpcResult, aggregationErr
}
