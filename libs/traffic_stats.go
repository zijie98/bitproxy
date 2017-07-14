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
var deleteTrafficStats = make(chan *Stats, 128)
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
	trafficStats <- &Stats{Port: port, Traffic: traffic}
}

func DeleteTrafficStats(port uint) {
	deleteTrafficStats <- &Stats{Port: port}
}

func StartStats() {
	go func() {
		select {
		case s := <-trafficStats:
			statsToRedis(s)
		}
	}()

	go func() {
		select {
		case s := <-deleteTrafficStats:
			deleteFromTheRedis(s)
		}
	}()
}

func statsToRedis(s *Stats) error {
	conn := RedisPool.Get()
	key := statsKey(s.Port)
	return conn.Send("INCRBY", key, s.Traffic)
}

func deleteFromTheRedis(s *Stats) error {
	key := statsKey(s.Port)
	conn := RedisPool.Get()
	return conn.Send("DEL", key)
}

func statsKey(port uint) string {
	return fmt.Sprintf("TRAFFIC.STATS.%d", port)
}
