package rpcx

import (
	"context"
	"flag"
	// "fmt"
	"github.com/smallnest/rpcx/protocol"
	"github.com/smallnest/rpcx/client"
	"github.com/qiusnay/gocron/model"
	"github.com/qiusnay/gocron/init"
	// "github.com/qiusnay/gocron/service/cron"
	"github.com/google/logger"
)

// RPC调用执行任务
type RPCHandler struct{}

func (h *RPCHandler) Run(taskModel model.FlCron, taskUniqueId int64) (result croninit.TaskResult, err error) {
	flag.Parse()

	d := client.NewPeer2PeerDiscovery("tcp@"+*addr, "")
	opt := client.DefaultOption
	opt.SerializeType = protocol.JSON

	xclient := client.NewXClient("RpcService", client.Failtry, client.RandomSelect, d, opt)
	defer xclient.Close()

	resultChan := make(chan croninit.TaskResult)
	
	go func() {
		reply := &croninit.TaskResult{}
		xclient.Call(context.Background(), "Run", taskModel, reply)
		logger.Error("任务开始执行#写入任务日志失败-", reply.Result)
		resultChan <- croninit.TaskResult{Result: reply.Result, Err: reply.Err, Host : reply.Host, Status : reply.Status, Endtime : reply.Endtime}
	}()
	var aggregationErr error = nil
	rpcResult := <-resultChan
	if rpcResult.Err != nil {
		aggregationErr = rpcResult.Err
	}
	return rpcResult, aggregationErr
}