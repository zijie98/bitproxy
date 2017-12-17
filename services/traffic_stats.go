package services

import (
	"fmt"
	"strconv"
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

func GetTraffic(port uint) (uint64, error) {
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

func getTraffic(port uint) (uint64, error) {
	conn, err := getRedisConn()
	if err != nil {
		return 0, err
	}
	key := statsKey(port)
	reply, err := conn.Do("GET", key)
	if err != nil {
		return 0, err
	}
	return strconv.ParseUint(reply.(string), 10, 64)
}

func statsToRedis(s *Stats) error {
	conn, err := getRedisConn()
	if err != nil {
		return err
	}
	key := statsKey(s.Port)
	return conn.Send("INCRBY", key, s.Traffic)
}

func deleteFromTheRedis(s *Stats) error {
	conn, err := getRedisConn()
	if err != nil {
		return err
	}
	key := statsKey(s.Port)
	return conn.Send("DEL", key)
}

func statsKey(port uint) string {
	return fmt.Sprintf("TRAFFIC.STATS.%d", port)
}
