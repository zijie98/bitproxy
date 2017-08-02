package services

import (
	"errors"
	"time"

	"github.com/garyburd/redigo/redis"

	"rkproxy/utils"
)

var RedisPool *redis.Pool

func InitRedis(host string, port uint) {
	RedisPool = &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial:        func() (redis.Conn, error) { return redis.Dial("tcp", utils.JoinHostPort(host, port)) },
	}
}

func getRedisConn() (redis.Conn, error) {
	if RedisPool == nil {
		return nil, errors.New("Redis not init")
	}
	return RedisPool.Get(), nil
}
