package libs

import (
	"time"

	"fmt"
	"github.com/garyburd/redigo/redis"
	"rkproxy/utils"
)

type Stats struct {
	Port    uint
	Traffic int
}

var trafficStats = make(chan *Stats, 512)
var RedisPool *redis.Pool

func InitRedis(host string, port uint) {
	RedisPool = &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial:        func() (redis.Conn, error) { return redis.Dial("tcp", utils.JoinHostPort(host, port)) },
	}
	StartStats()
}

func AddTrafficStats(port uint, traffic int) {
	trafficStats <- &Stats{port, traffic}
}

func DeleteTrafficStats(port uint) error {
	key := statsKey(port)
	conn := RedisPool.Get()
	return conn.Send("DEL", key)
}

func StartStats() {
	go func() {
		select {
		case s := <-trafficStats:
			statsToRedis(s)
		}
	}()
}

func statsToRedis(s *Stats) error {
	conn := RedisPool.Get()
	key := statsKey(s.Port)
	return conn.Send("INCRBY", key, s.Traffic)
}

func statsKey(port uint) string {
	return fmt.Sprintf("TRAFFIC.STATS.%d", port)
}
