#### gocron 任务调度系统
gocron 任务调度系统基于coffman的cron库搭建.DB框架采用gorm,底层的任务调度通过rpcx框架,client与server端的通信实现.

#### todo
 * (完成)当新添加了一个作业后.一直在运行的进程不会退出.而且不会自动载入新的作业运行  A : 通过页面添加作业时,直接通过addFunc操作写入当前的作业定时器里.同时写入DB
 * 多机器,多个group功能支持,采用无中心化设计
 *     如果要不借助redis的话,如何保证在有多个dispacher的时候.同一个作业不被多次解析并调用执行呢(可以用slice分片来实现,但是这样不够灵活)
 *     go mod 机器变更后需要修改GOPROXY
 *     sudo go env -w GOPROXY=https://goproxy.cn,https://mirrors.aliyun.com/goproxy/,direct
 * 
 * 如何做到发布可以随意启动的问题,架构设计脱钩方案
 * (完成)http命令支持
 * 支持rabbitmq设计
 * 支持任务间耦合设计
 * (完成)考虑是需要继续使用redis作锁设计 - A: 需要使用redis锁

#### 运行步骤

* 客户端启动
`go run client.go`
* 服务端启动
`go run server.go`

#### 配置: /conf/conf.ini
[database]
  * username = root
  * password = root
  * host     = localhost
  * port     = 3306
  * database = cron
  * charset  = utf8


#### 环境要求
>  MySQL, MAC

#### 程序使用的组件
* Web框架 [Macaron](http://go-macaron.com/)
* 定时任务调度 [Cron](https://github.com/jakecoffman/cron)
* GORM [gorm](https://github.com/go-gorm/gorm)
* RPCX框架 [gRPC](https://github.com/smallnest/rpcx)

