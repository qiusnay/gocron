module github.com/qiusnay/gocron

go 1.15

require (
	github.com/coreos/etcd v3.3.13+incompatible
	github.com/go-gomail/gomail v0.0.0-20160411212932-81ebce5c23df
	github.com/golang/protobuf v1.4.2
	github.com/gomodule/redigo v1.8.2
	github.com/google/logger v1.1.0
	github.com/jakecoffman/cron v0.0.0-20190106200828-7e2009c226a5
	github.com/jinzhu/gorm v1.9.16
	github.com/ouqiang/gocron v1.5.3
	github.com/ouqiang/goutil v1.2.9
	github.com/rcrowley/go-metrics v0.0.0-20200313005456-10cdbea86bc0
	github.com/robfig/cron v1.2.0
	github.com/robfig/cron/v3 v3.0.1
	github.com/sirupsen/logrus v1.4.2
	github.com/smallnest/rpcx v0.0.0-20200924044220-f2cdd4dea15a
	go.etcd.io/etcd v3.3.13+incompatible
	google.golang.org/grpc v1.30.0
	google.golang.org/protobuf v1.23.0
	gorm.io/gorm v1.20.5
)

replace google.golang.org/grpc => google.golang.org/grpc v1.29.0
