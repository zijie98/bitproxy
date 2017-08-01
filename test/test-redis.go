package main

import (
	"fmt"
	"time"

	"github.com/garyburd/redigo/redis"
	"rkproxy/services"
	"rkproxy/utils"
)

func main() {
	services.RedisPool = &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial:        func() (redis.Conn, error) { return redis.Dial("tcp", utils.JoinHostPort("127.0.0.1", 6379)) },
	}
	conn := services.RedisPool.Get()
	key := "test-hash"
	conn.Send("HSET", key, "123.123.123.123", 1)
	conn.Send("HSET", key, "321.123.123.123", 2)

	keys, err := conn.Do("HKEYS", key)
	if err != nil {
		fmt.Println("--- HKEYS error ", err)
		return
	}
	for _, ip := range keys.([]interface{}) {
		fmt.Println(fmt.Sprintf("%s", ip))
	}
}
