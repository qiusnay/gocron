package rpc

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/qiusnay/gocron/model"
	"github.com/qiusnay/gocron/service/rpc/etcd"
	gocron "github.com/qiusnay/gocron/service/rpc/protofile"
	cron "github.com/robfig/cron/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/resolver"
)

var (
	EtcdAddr = flag.String("EtcdAddr", "127.0.0.1:2379", "register etcd address")
)

// RPC调用执行任务
type CronClient struct{}

var Parser = cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)

func (c *CronClient) Run(jobModel model.FlCron, taskId int64) (result model.TaskResult, err string) {
	flag.Parse()
	r := etcd.NewResolver(*EtcdAddr)
	resolver.Register(r)
	// https://github.com/grpc/grpc/blob/master/doc/naming.md
	// The gRPC client library will use the specified scheme to pick the right resolver plugin and pass it the fully qualified name string.
	conn, rpcerr := grpc.Dial(r.Scheme()+"://author/"+*ServiceName, grpc.WithBalancerName("round_robin"), grpc.WithInsecure())
	if rpcerr != nil {
		panic(err)
	}

	jobModel.Taskid = taskId
	s, _ := Parser.Parse(jobModel.Rule)
	expreTime := s.Next(time.Now()).Unix() - time.Now().Unix() - 5 // 5秒预留空间

	resultChan := make(chan model.TaskResult)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(expreTime)*time.Second)
		defer cancel()
		client := gocron.NewTaskClient(conn)
		resp, err := client.Run(ctx, &gocron.TaskRequest{
			Command:   jobModel.Cmd,
			Timeout:   expreTime,
			Jobid:     jobModel.Jobid,
			Taskid:    jobModel.Taskid,
			Querytype: jobModel.Querytype,
		})
		if err == nil {
			fmt.Printf("Reply is %+v\n", resp)
			resultChan <- model.TaskResult{Result: resp.Output, Err: resp.Err, Host: resp.Host, Status: resp.Status, Endtime: resp.Endtime}
		} else {
			fmt.Printf("call server error:%s\n", err)
			resultChan <- model.TaskResult{Result: "", Err: err.Error(), Host: "", Status: 10001, Endtime: time.Now().Format("2006-01-02 15:04:05")}
		}

	}()
	var aggregationErr string = ""
	rpcResult := <-resultChan
	if rpcResult.Err != "" {
		aggregationErr = rpcResult.Err
	}
	return rpcResult, aggregationErr
}
