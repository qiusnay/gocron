#### gocron 任务调度系统
gocron 任务调度系统基于coffman的cron库搭建.DB框架采用gorm,底层的任务调度通过rpcx框架,client与server端的通信实现.

#### todo
 * (完成)http命令支持
 * (完成)当新添加了一个作业后.一直在运行的进程不会退出.而且不会自动载入新的作业运行  A : 通过页面添加作业时,直接通过addFunc操作写入当前的作业定时器里.同时写入DB
 * 多机器,多个group功能支持,采用无中心化设计,多分发者多消费者,任何一台机器宕机,不影响整体系统的运行
 *    (己完成)1.多个任务分发者 - 采用redis锁实现.因为不同机器之间的routine无法通信.只能用分布式锁来实现.
 *    (己完成)2.多个任务消费者 - 采用rpcx服务注册发现机制(etcd)实现
 * 如何做到发布可以随意启动的问题,架构设计脱钩方案
 * 支持rabbitmq设计
 * 支持任务间耦合设计
 * go mod 机器变更后需要修改GOPROXY
 * sudo go env -w GOPROXY=https://goproxy.cn,https://mirrors.aliyun.com/goproxy/,direct

#### 运行步骤

* cron作业分发者
`go run dispacher.go`
* cron作业执行者
`go run executor.go`
* cron作业监控者
`go run monitor.go`

#### 配置: /conf/conf.ini
[database]
  * username = root
  * password = root
  * host     = localhost
  * port     = 3306
  * database = cron
  * charset  = utf8
[redis]
  * host     = 192.168.100.60
  * port     = 6381
  * max_idle = 10
  * max_active = 60

[machine]
  * phpgroup     = 192.168.2.74|192.168.2.75|192.168.75.115|192.168.75.213
  * pcrongroup   = 192.168.3.206|192.168.3.198
  * cron74       = 192.168.100.74
  * cron75       = 192.168.100.75

[alarm_mail_list]
  * cron_mail = admin@cron.com test@qq.com test2@qq.com
  * from_mail = alarm@cron.it
  * cron_url = http://cron.admin.it

#### etcd docker 安装
  * rm -rf /tmp/etcd-data.tmp
  * mkdir -p /tmp/etcd-data.tmp
  * docker rmi gcr.io/etcd-development/etcd:v3.4.13 
  * docker run \
  -p 2379:2379 \
  -p 2380:2380 \
  --mount type=bind,source=/tmp/etcd-data.tmp,destination=/etcd-data \
  --name etcd-gcr-v3.4.13 \
  gcr.io/etcd-development/etcd:v3.4.13 \
  /usr/local/bin/etcd \
  --name s1 \
  --data-dir /etcd-data \
  --listen-client-urls http://0.0.0.0:2379 \
  --advertise-client-urls http://0.0.0.0:2379 \
  --listen-peer-urls http://0.0.0.0:2380 \
  --initial-advertise-peer-urls http://0.0.0.0:2380 \
  --initial-cluster s1=http://0.0.0.0:2380 \
  --initial-cluster-token tkn \
  --initial-cluster-state new \
  --log-level info \
  --logger zap \
  --log-outputs stderr

  * docker exec etcd-gcr-v3.4.13 /bin/sh -c "/usr/local/bin/etcd --version"
  * docker exec etcd-gcr-v3.4.13 /bin/sh -c "/usr/local/bin/etcdctl version"
  * docker exec etcd-gcr-v3.4.13 /bin/sh -c "/usr/local/bin/etcdctl endpoint health"
  * docker exec etcd-gcr-v3.4.13 /bin/sh -c "/usr/local/bin/etcdctl put foo bar"
  * docker exec etcd-gcr-v3.4.13 /bin/sh -c "/usr/local/bin/etcdctl get foo"


#### 环境要求
>  MySQL, MAC

#### 程序使用的组件
* 服务发现与注册 [Etcd](https://github.com/etcd-io/etcd)
* 定时任务调度 [Cron](https://github.com/jakecoffman/cron)
* GORM [gorm](https://github.com/go-gorm/gorm)
* RPCX框架 [gRPC](https://github.com/smallnest/rpcx)

