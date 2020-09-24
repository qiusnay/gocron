#### gocron 任务调度系统
gocron 任务高度系统基于coffman的cron库搭建.DB框架采用gorm,底层的任务调度通过rpcx框架,client与server端的通信实现.

#### 运行步骤

* 客户端启动
`go run client.go`
* 服务端启动
`go run server.go`

#### 配置: /conf/conf.ini

[database]
username = root
password = root
host     = localhost
port     = 3306
database = cron
charset  = utf8


#### 环境要求
>  MySQL, MAC

## 程序使用的组件
* Web框架 [Macaron](http://go-macaron.com/)
* 定时任务调度 [Cron](https://github.com/jakecoffman/cron)
* GORM [gorm](https://github.com/go-gorm/gorm)
* RPCX框架 [gRPC](https://github.com/smallnest/rpcx)

