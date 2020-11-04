package rpc

import (
	"context"
	"flag"
	"time"

	"github.com/qiusnay/gocron/model"
	"github.com/qiusnay/gocron/service/rpc/etcd"
	gocron "github.com/qiusnay/gocron/service/rpc/protofile"
	"google.golang.org/grpc"
	"google.golang.org/grpc/resolver"
)

var (
	EtcdAddr = flag.String("EtcdAddr", "127.0.0.1:2379", "register etcd address")
)

// RPC调用执行任务
type CronClient struct{}

func (c *CronClient) Run(Job model.VCron, TaskId int64, ExpireTime int64) model.TaskResult {
	flag.Parse()
	r := etcd.NewResolver(*EtcdAddr)
	resolver.Register(r)
	// https://github.com/grpc/grpc/blob/master/doc/naming.md
	// The gRPC client library will use the specified scheme to pick the right resolver plugin and pass it the fully qualified name string.
	conn, rpcerr := grpc.Dial(r.Scheme()+"://author/"+*ServiceName, grpc.WithBalancerName("round_robin"), grpc.WithInsecure())
	if rpcerr != nil {
		panic(rpcerr)
	}
	Job.Taskid = TaskId
	resultChan := make(chan model.TaskResult)
	go func() {
		ctx := context.Background() //正常作业
		if Job.Overflow == model.ForceKill {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(context.Background(), time.Duration(ExpireTime)*time.Second)
			defer cancel()
		}
		// logger.Info(fmt.Sprintf("client taskid : %s, 发送ctx : %+v", Job.Taskid, ctx))
		client := gocron.NewTaskClient(conn)
		resp, err := client.Run(ctx, &gocron.TaskRequest{ //发送请求
			Command:   Job.Cmd,
			Timeout:   ExpireTime,
			Jobid:     Job.Jobid,
			Taskid:    Job.Taskid,
			Querytype: Job.Querytype,
		})
		// logger.Info(fmt.Sprintf("grpc 返回结果 : %+v, 错误信息 : %+v", resp, err))
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
	rpcResult := <-resultChan
	return rpcResult
}
