package libs

import (
	"time"

	"fmt"
	"github.com/garyburd/redigo/redis"
	"rkproxy/utils"
	"strconv"
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

func GetTraffic(port uint) (uint64, error) {
	return getTraffic(port)
}

// 开始统计
func StartStats() {

	// 累加流量
	go func() {
		for {
			select {
			case s := <-trafficStats:
				statsToRedis(s)
			}
		}
	}()

	// 清空流量
	go func() {
		for {
			select {
			case s := <-deleteTrafficStats:
				deleteFromTheRedis(s)
			}
		}
	}()

	// 定时SAVE持久化redis
	go func() {
		for {
			select {
			case <-time.After(10 * time.Second):
				persistent()
			}
		}
	}()
}

// 持久化
func persistent() {
	conn := RedisPool.Get()
	conn.Send("SAVE")
}

func getTraffic(port uint) (uint64, error) {
	conn := RedisPool.Get()
	key := statsKey(port)
	reply, err := conn.Do("GET", key)
	if err != nil {
		return 0, err
	}
	return strconv.ParseUint(reply.(string), 10, 64)
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
