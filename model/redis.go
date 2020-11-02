package model

import (
	"strconv"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/google/logger"
	"github.com/qiusnay/gocron/utils"
)

type RedisDB struct {
	Pool *redis.Pool
}

var Redis *RedisDB

func init() {
	Redis = &RedisDB{}
}

func (db *RedisDB) InitPool() {
	redisConf := utils.GetConfig("redis_local", "")
	host := redisConf["host"]
	logger.Info("收到终端断开信号, 忽略, %v", host)
	port := redisConf["port"]
	maxIdle, _ := strconv.Atoi(redisConf["max_idle"])
	maxActive, _ := strconv.Atoi(redisConf["max_active"])

	db.Pool = &redis.Pool{
		MaxIdle:     maxIdle,
		MaxActive:   maxActive,
		IdleTimeout: 240 * time.Second,
		Wait:        true,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", host+":"+port, redis.DialPassword(""))
			if err != nil {
				return nil, err
			}
			return c, nil
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}

	// 检查redis是否连接有误
	if _, err := db.Pool.Dial(); err != nil {
		panic(err)
	}
}

func (db *RedisDB) Do(command string, args ...interface{}) (interface{}, error) {
	conn := db.Pool.Get()
	defer conn.Close()
	return conn.Do(command, args...)
}

func (db *RedisDB) String(command string, args ...interface{}) (string, error) {
	return redis.String(db.Do(command, args...))
}

func (db *RedisDB) Bool(command string, args ...interface{}) (bool, error) {
	return redis.Bool(db.Do(command, args...))
}

func (db *RedisDB) Strings(command string, args ...interface{}) ([]string, error) {
	return redis.Strings(db.Do(command, args...))
}

func (db *RedisDB) Int(command string, args ...interface{}) (int, error) {
	return redis.Int(db.Do(command, args...))
}

func (db *RedisDB) Ints(command string, args ...interface{}) ([]int, error) {
	return redis.Ints(db.Do(command, args...))
}

func (db *RedisDB) StringMap(command string, args ...interface{}) (map[string]string, error) {
	return redis.StringMap(db.Do(command, args...))
}
