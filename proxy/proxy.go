package proxy

type NetProtocol string

// 连接类型
const (
	TCP_PROTOCOL NetProtocol = "tcp"
	UDP_PROTOCOL NetProtocol = "udp"
	KCP_PROTOCOL NetProtocol = "kcp"
)

//	代理程序的接口
//
type Proxyer interface {
	Start() error
	Stop() error
	LocalPort() uint
	Traffic() (uint64, error)
}

//	句柄接口
//
type ProxyHandler interface {
	Port() uint
	Start() error
	Stop() error
	GetConfig() interface{}
	GetTraffic() (uint64, error)
}
