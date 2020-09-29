package rpcx

import (
	"context"
	"flag"
	"fmt"
	"github.com/smallnest/rpcx/protocol"
	"github.com/smallnest/rpcx/client"
	"github.com/qiusnay/gocron/model"
	// "github.com/qiusnay/gocron/service/cron"
	// "github.com/google/logger"
)

// RPC调用执行任务
type RPCHandler struct{}

type TaskResult struct {
	Result     string
	Err        error
}


func (h *RPCHandler) Run(taskModel model.FlCron, taskUniqueId int64) (result string, err error) {
	flag.Parse()

	d := client.NewPeer2PeerDiscovery("tcp@"+*addr, "")
	opt := client.DefaultOption
	opt.SerializeType = protocol.JSON

	xclient := client.NewXClient("RpcService", client.Failtry, client.RandomSelect, d, opt)
	defer xclient.Close()

	resultChan := make(chan TaskResult)
	
	go func() {
		reply := &Reply{Output:"", Err : nil}
		err := xclient.Call(context.Background(), "Run", taskModel, reply)
		errorMessage := ""
		if err != nil {
			errorMessage = fmt.Sprintf("failed to call: %v", err)
		}
		outputMessage := fmt.Sprintf("主机: [%s-%s:%d]-%s-%s", "localhost", "qiusnay", 8088, errorMessage, reply.Output)
		
		resultChan <- TaskResult{Err: err, Result: outputMessage}
	}()
	var aggregationErr error = nil
	aggregationResult := ""
	taskResult := <-resultChan
	aggregationResult += taskResult.Result
	if taskResult.Err != nil {
		aggregationErr = taskResult.Err
	}
	return aggregationResult, aggregationErr
}