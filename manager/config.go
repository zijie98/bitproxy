package manager

type Proxyer interface {
	Start() error
	Stop() error
	LocalPort() int
}

// FTP代理

// HTTP代理
//type HttpProxyConfig struct {
//
//}

// HTTP 反向代理配置
type (
	HttpReproxyConfig struct {
		LocalPort  int    `json:"local_port"`
		RemoteHost string `json:"remote_host"`
		RemotePort int    `json:"remote_port"`
	}
	HttpReproxyHandle struct {
		Config *HttpReproxyConfig
		Proxy  Proxyer
	}
)

// UDP 代理配置
//type UdpProxyConfig struct {
//	TcpProxyConfig
//}

// TCP 代理配置
type (
	TcpProxyConfig struct {
		LocalPort  int    `json:"local_port"`
		RemoteHost string `json:"remote_host"`
		RemotePort int    `json:"remote_port"`
	}
	TcpProxyHandle struct {
		Config *TcpProxyConfig `json:"tcp"`
		Proxy  Proxyer
	}
)

// Shadowsocks 客户端配置
type (
	SsClientConfig struct {
		LocalPort  int    `json:"local_port"`
		RemoteHost string `json:"remote_host"`
		RemotePort int    `json:"remote_port"`
		Password   string `json:"password"`
		Crypt      string `json:"crypt"`
	}
	SsClientHandle struct {
		Config *SsClientConfig
		Proxy  Proxyer
	}
)

// Shadowsocks 服务器端配置
type (
	SsServerConfig struct {
		Crypt     string `json:"crypt"`
		Password  string `json:"password"`
		LocalPort int    `json:"port"`
	}

	SsServerHandle struct {
		Config *SsServerConfig
		Proxy  Proxyer
	}
)
