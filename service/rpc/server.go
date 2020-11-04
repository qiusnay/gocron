package rpc

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/logger"
	croninit "github.com/qiusnay/gocron/init"
	"github.com/qiusnay/gocron/model"
	"github.com/qiusnay/gocron/service/rpc/etcd"
	gocron "github.com/qiusnay/gocron/service/rpc/protofile"
	"github.com/qiusnay/gocron/utils"
	"google.golang.org/grpc"
)

var (
	addr        = flag.String("addr", utils.GetLocalIP()+":8973", "server address")
	etcdAddr    = flag.String("etcdAddr", "127.0.0.1:2379", "etcd address")
	ServiceName = flag.String("ServiceName", "task", "service name")
)

type CronResponse struct {
	Host   string
	Code   int64
	Result string
	Err    error
}

type MyServer struct{}

func (s *MyServer) Start() {
	flag.Parse()
	//定义rpc服务端
	lis, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("failed to listen: %s", err)
	}
	defer lis.Close()

	GrpcServer := grpc.NewServer()
	defer GrpcServer.GracefulStop()
	//注册当前服务到grpc中
	gocron.RegisterTaskServer(GrpcServer, &MyServer{})
	//注册当前服务到etcd中
	go etcd.Register(*EtcdAddr, *ServiceName, *addr, 5)

	//进程终止信号，注销etcd上的服务
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGQUIT)
	go func() {
		s := <-ch
		etcd.UnRegister(*ServiceName, *addr)

		if i, ok := s.(syscall.Signal); ok {
			os.Exit(int(i))
		} else {
			os.Exit(0)
		}
	}()

	//拉起rpc服务
	if err := GrpcServer.Serve(lis); err != nil {
		fmt.Printf("failed to serve: %s", err)
	}
}

//服务入口
func (s *MyServer) Run(ctx context.Context, req *gocron.TaskRequest) (*gocron.TaskResponse, error) {
	// logger.Info(fmt.Sprintf("taskid : %s, 接收通道 : %+v", req.Taskid, ctx))
	queryResult := CronResponse{}
	queryCmd := AssembleCmd(req)
	//执行前更新状态
	s.BeforeExecJob(req.Taskid)
	switch req.Querytype {
	case "wget":
	case "curl":
		rpccurl := RpcServiceCurl{}
		queryResult = rpccurl.ExecCurl(ctx, queryCmd)
		break
	default:
		rpcshell := RpcServiceShell{}
		queryResult = rpcshell.ExecShell(ctx, queryCmd, req.Taskid)
	}
	queryResult.Host = utils.GetLocalIP()
	//更新DB执行日志
	s.AfterExecJob(queryResult, req)
	//写文件日志
	logger.Info(fmt.Sprintf("execute cmd end: [cmd: %s err: %s, status : %d]", queryCmd, queryResult.Err, queryResult.Code))
	return &gocron.TaskResponse{Err: queryResult.Err.Error(), Output: queryResult.Result, Status: queryResult.Code, Host: queryResult.Host}, nil
}

//开始执行任务
func (s *MyServer) BeforeExecJob(TaskId int64) {
	var TaskResult = model.TaskResult{}
	TaskResult.Status = croninit.CronRunning
	TaskResult.Err = "作业执行中"
	new(model.VLog).UpdateTaskLog(TaskId, TaskResult)
}

//执行完后更新日志
func (s *MyServer) AfterExecJob(queryResult CronResponse, req *gocron.TaskRequest) {
	var TaskResult = model.TaskResult{}
	TaskResult.Result = queryResult.Result
	TaskResult.Host = queryResult.Host
	TaskResult.Status = queryResult.Code
	TaskResult.Endtime = time.Now().Format("2006-01-02 15:04:05")
	logger.Info(fmt.Sprintf("AfterExecJob : %v, taskid : %s", queryResult.Err, req.Taskid))
	if queryResult.Err.Error() != "" {
		TaskResult.Err = queryResult.Err.Error()
	} else {
		TaskResult.Err = "执行成功"
	}
	_, err := new(model.VLog).UpdateTaskLog(req.Taskid, TaskResult)
	if err != nil {
		logger.Error("任务结束#更新任务日志失败-", err)
	}
	jobModel := model.VCron{}
	JobInfo, _ := jobModel.GetJobInfo(req.GetJobid())

	// 发送邮件
	go SendNotification(JobInfo[0], TaskResult)
}

// 发送任务结果通知
func SendNotification(jobModel model.VCron, taskResult model.TaskResult) {
	if taskResult.Err == "succss" {
		return // 执行失败才发送通知
	}
	//发送邮件
	// notify.SendCronAlarmMail(taskResult, jobModel)
}

//命令组装
func AssembleCmd(cron *gocron.TaskRequest) string {
	return cron.Command + " > " + GetLogFile(cron.Jobid, cron.Taskid)
}

//获取日志文件
func GetLogFile(Jobid int64, Taskid int64) string {
	//设置日志目录
	LogDir := croninit.BASEPATH + "/log/cronlog/" + time.Now().Format("2006-01-02")
	if !utils.IsDir(LogDir) {
		// mkdir($LogDir, 0777, true);
		os.MkdirAll(LogDir, os.ModePerm)
	}
	return LogDir + "/cron-task-" + utils.Int64toString(Jobid) + "-" + utils.Int64toString(Taskid) + "-log.log"
}
