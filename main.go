package main

import (
	"flag"
	"log"
	"fmt"
	"os"
)

var (
	h bool
	s bool
	p string
	m bool
	b string
	x bool
	c string
	k bool
	f bool
)

func init() {
	flag.BoolVar(&s, "s", false, "启动Yearning")
	flag.BoolVar(&m, "m", false, "数据初始化(第一次安装时执行)")
	flag.StringVar(&p, "p", "8000", "Yearning端口")
	flag.BoolVar(&h, "help", false, "帮助")
	flag.BoolVar(&f, "f", false, "初始化Admin用户密码")
	flag.BoolVar(&x, "x", false, "表结构修复")
	flag.StringVar(&b, "b", "127.0.0.1", "钉钉/邮件推送时显示的平台地址")
	flag.StringVar(&c, "c", "conf.toml", "配置文件路径")
	flag.BoolVar(&k, "k", false, "用户权限变更为权限组(2.1.7以下升级至2.1.7及以上使用)")
	flag.Usage = usage
	log.SetPrefix("Yearning_error: ")
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Llongfile)
}

func usage() {
	_, err := fmt.Fprintf(os.Stderr, `version: Yearning/2.3.0 author: HenryYee
Usage: Yearning [-m migrate] [-p port] [-s start] [-b web-bind] [-h help] [-c config file]

Options:
 -s  启动Yearning
 -m  数据初始化(第一次安装时执行)
 -p  端口
 -b  钉钉/邮件推送时显示的平台地址
 -x  表结构修复,升级时可以操作。如出现错误可直接忽略。
 -h  帮助
 -c  配置文件路径
 -k  用户权限变更为权限组(2.1.7以下升级至2.1.7及以上使用)
 -f  初始化Admin用户密码
`)
	if err != nil {
		panic(err.Error())
	}
}

func main() {
	flag.Parse()
	if(h) {
		flag.Usage()
	}
	if(m) {
		service.Migrate()
	}

}
