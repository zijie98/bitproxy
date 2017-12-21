package services

import (
	"fmt"

	"github.com/garyburd/redigo/redis"
)

type Stats struct {
	Port    uint
	Traffic int64
}

var stats = false // 是否启动统计
var trafficStats = make(chan *Stats, 512)
var deleteTrafficStats = make(chan *Stats, 128)

func AddTrafficStats(port uint, traffic int64) {
	if !stats {
		return
	}
	trafficStats <- &Stats{Port: port, Traffic: traffic}
}

func DeleteTrafficStats(port uint) {
	if !stats {
		return
	}
	deleteTrafficStats <- &Stats{Port: port}
}

func GetTraffic(port uint) (int64, error) {
	return getTraffic(port)
}

// 开始统计
func StartStats() {
	stats = true
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
	//go func() {
	//	for {
	//		select {
	//		case <-time.After(10 * time.Second):
	//			persistent()
	//		}
	//	}
	//}()
}

// 持久化
func persistent() {
	conn, err := getRedisConn()
	if err != nil {
		return
	}
	conn.Send("SAVE")
}

func getTraffic(port uint) (int64, error) {
	conn, err := getRedisConn()
	if err != nil {
		return 0, err
	}
	key := statsKey(port)
	reply, err := conn.Do("GET", key)
	if err != nil {
		return 0, err
	}
	return redis.Int64(reply, err)
}

func statsToRedis(s *Stats) error {
	conn, err := getRedisConn()
	if err != nil {
		return err
	}
	key := statsKey(s.Port)
	_, err = conn.Do("INCRBY", key, s.Traffic)
	if err != nil {
		return err
	}
	return nil
}

func deleteFromTheRedis(s *Stats) error {
	conn, err := getRedisConn()
	if err != nil {
		return err
	}
	key := statsKey(s.Port)
	_, err = conn.Do("DEL", key)
	if err != nil {
		return err
	}
	return nil
}

func statsKey(port uint) string {
	return fmt.Sprintf("TRAFFIC.STATS.%d", port)
}
