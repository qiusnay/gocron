package rpc

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/google/logger"
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

func (c *CronClient) Run(jobModel model.FlCron, taskId int64, contextWithDeadline int) (result model.TaskResult, err string) {
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

	resultChan := make(chan model.TaskResult)
	go func() {
		var expreTime int64
		var cancel context.CancelFunc
		ctx := context.Background() //正常作业
		if contextWithDeadline == 1 {
			//如果当前Task需要在周期内强制结束服务端正在运行的作业
			s, _ := Parser.Parse(jobModel.Rule)
			expreTime := s.Next(time.Now()).Unix() - time.Now().Unix() - 5 // 5秒预留空间
			ctx, cancel = context.WithTimeout(context.Background(), time.Duration(expreTime)*time.Second)
			defer cancel()
		}
		client := gocron.NewTaskClient(conn)
		resp, err := client.Run(ctx, &gocron.TaskRequest{ //发送请求
			Command:   jobModel.Cmd,
			Timeout:   expreTime,
			Jobid:     jobModel.Jobid,
			Taskid:    jobModel.Taskid,
			Querytype: jobModel.Querytype,
		})
		logger.Info(fmt.Sprintf("grpc 返回结果 : %+v, 错误信息 : %+v", resp, err))
		//返回结构初始化
		GrpcResult := model.TaskResult{Result: "", Err: "", Host: "", Status: 0, Endtime: ""}
		if err != nil {
			GrpcResult.Err = err.Error()
		} else {
			GrpcResult.Result = resp.Output
			GrpcResult.Err = resp.Err
			GrpcResult.Host = resp.Host
			GrpcResult.Status = resp.Status
			GrpcResult.Endtime = resp.Endtime
		}
		resultChan <- GrpcResult //写入通道
	}()
	var aggregationErr string = ""
	rpcResult := <-resultChan
	if rpcResult.Err != "" {
		aggregationErr = rpcResult.Err
	}
	return rpcResult, aggregationErr
}
