package etcd

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"go.etcd.io/etcd/clientv3"
)

// Register register service with name as prefix to etcd, multi etcd addr should use ; to split
//etcdAddr etcd地址
//name 服务名称
//addr 服务的地址
//ttl 服务注册在etcd键的过期时间
func Register(etcdAddr, name string, addr string, ttl int64) error {
	var err error
	//创建etcd的连接
	if cli == nil {
		cli, err = clientv3.New(clientv3.Config{
			Endpoints:   strings.Split(etcdAddr, ";"),
			DialTimeout: 15 * time.Second,
		})
		if err != nil {
			fmt.Printf("connect to etcd err:%s", err)
			return err
		}
	}
	//时间定时器返回channel
	ticker := time.NewTicker(time.Second * time.Duration(ttl))

	//定时上传服务到etcd，并设置键的过期时间
	//时间定时器的时间与过期时间相同（防止无服务挂了，etcd上数据还在）
	go func() {
		for {
			getResp, err := cli.Get(context.Background(), "/"+schema+"/"+name+"/"+addr)
			//fmt.Printf("getResp:%+v\n",getResp)
			if err != nil {
				log.Println(err)
				fmt.Printf("Register:%s", err)
			} else if getResp.Count == 0 {
				err = withAlive(name, addr, ttl)
				if err != nil {
					log.Println(err)
					fmt.Printf("keep alive:%s", err)
				}
			} else {
				//fmt.Printf("getResp:%+v, do nothing\n",getResp)
			}

			<-ticker.C
		}
	}()

	return nil
}

func withAlive(name string, addr string, ttl int64) error {
	leaseResp, err := cli.Grant(context.Background(), ttl)
	if err != nil {
		return err
	}

	//fmt.Printf("key:%v\n", "/"+schema+"/"+name+"/"+addr)
	_, err = cli.Put(context.Background(), "/"+schema+"/"+name+"/"+addr, addr, clientv3.WithLease(leaseResp.ID))
	if err != nil {
		fmt.Printf("put etcd error:%s", err)
		return err
	}

	_, err = cli.KeepAlive(context.Background(), leaseResp.ID)
	if err != nil {
		fmt.Printf("keep alive error:%s", err)
		return err
	}
	return nil
}

// UnRegister remove service from etcd
func UnRegister(name string, addr string) {
	if cli != nil {
		cli.Delete(context.Background(), "/"+schema+"/"+name+"/"+addr)
	}
}
