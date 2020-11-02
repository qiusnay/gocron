#### gocron 任务调度系统
cron任务调度系统分为dispacher(作业分发者),executor(作业执行者),monitor(作业监控者)
 * 任务分发者 主要负责对作业的rule规则解析出下一次执行时间.并将当前作业加入任务调度器中定时执行
 * 任务执行者 负责任务执行的主体.
 * 任务监控者 对当前的每一个作业做健康检查.并对新增的作业动态添加进任务调度器中执行.

cron系统底层通过rpcx微服务调度框架以及etcd服务注册与发现.由多个dispacher生成cron任务.通过client发送至etcd服务中心,etcd接到服务后,按轮训的方式分配服务提供者来执行.服务提供者执行完后,返回给client端执行结果,写入DB.

### gocron 系统的优点
* 支持多台机器的client端任务分发,支持多台机器的服务提供者.
* 通过etcd服务注册中心来达到客服端与服务端解耦的问题.
* 通过redis分布式锁,实现同一个task只会被创建一次
* 支持指定机器,指定机器组执行作业.
* 支持执行日志的后台查看.
* 支持邮件,短信等告警业务
* 超时间机制的处理 : 放弃并告警,强制终止,同步运行
* 支持平滑重启,发布迭代不影响具体业务

#### todo
 * monitor 功能实现
 * k8s 引入平滑重启 , 文件描述符继承平滑重启
 * 短信功能实现
 * 超时机制的处理

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

go mod 机器变更后需要修改GOPROXY
sudo go env -w GOPROXY=https://goproxy.cn,https://mirrors.aliyun.com/goproxy/,direct

#### 环境要求
>  MySQL, MAC

#### 程序使用的组件
* 服务发现与注册 [Etcd](https://github.com/etcd-io/etcd)
* 定时任务调度 [Cron](https://github.com/jakecoffman/cron)
* GORM [gorm](https://github.com/go-gorm/gorm)
* RPCX框架 [gRPC](https://github.com/smallnest/rpcx)

