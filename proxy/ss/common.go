package ss

import (
	"net"
	"time"

	"github.com/molisoft/bitproxy/utils"
)

// ss common
const (
	SOCKS5_VERSION = 5

	TYPE_IPV4   = 1
	TYPE_IPV6   = 4
	TYPE_DOMAIN = 3
)

// ss server
const (
	SERVER_RAW_TYPE       = 0
	SERVER_RAW_DOMAIN_LEN = 1

	IPV4_INDEX   = 1
	IPV6_INDEX   = 1
	DOMAIN_INDEX = 2

	SERVER_LEN_TYPE_IPV4        = 1 + net.IPv4len + 2
	SERVER_LEN_TYPE_IPV6        = 1 + net.IPv6len + 2
	SERVER_LEN_TYPE_DOMAIN_BASE = 1 + 1 + 2
)

// ss client
const (
	CLIENT_LEN_TYPE_IPV4        = 3 + 1 + net.IPv4len + 2
	CLIENT_LEN_TYPE_IPV6        = 3 + 1 + net.IPv6len + 2
	CLIENT_LEN_TYPE_DOMAIN_BASE = 3 + 1 + 1 + 2
	CLIENT_RAW_TYPE             = 3
	CLIENT_RAW_ADDR             = 4
)

type NetProtocol string

// 连接类型
const (
	TCP_PROTOCOL NetProtocol = "tcp"
	UDP_PROTOCOL NetProtocol = "udp"
	KCP_PROTOCOL NetProtocol = "kcp"
)

var (
	IvPool utils.BytePool // 建立IV使用的内存池
	RWPool utils.BytePool // 建立读写流内存池
)

func init() {
	IvPool.Init(0, 1024)
	RWPool.Init(10*time.Minute, 1024)
}
