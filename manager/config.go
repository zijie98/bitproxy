/*
	各种配置
*/

package manager

// FTP代理

// HTTP代理
//type HttpProxyConfig struct {
//
//}

// HTTP 反向代理配置
type HttpReproxyConfig struct {
	LocalPort  int    `json:"local_port"`
	RemoteHost string `json:"remote_host"`
	RemotePort int    `json:"remote_port"`
}

// UDP 代理配置
//type UdpProxyConfig struct {
//	TcpProxyConfig
//}

// TCP 代理配置
type TcpProxyConfig struct {
	LocalPort  int    `json:"local_port"`
	RemoteHost string `json:"remote_host"`
	RemotePort int    `json:"remote_port"`
}

// Shadowsocks 客户端配置
type SsClientConfig struct {
	LocalPort  int    `json:"local_port"`
	RemoteHost string `json:"remote_host"`
	RemotePort int    `json:"remote_port"`
	Password   string `json:"password"`
	Crypt      string `json:"crypt"`
}

// Shadowsocks 服务器端配置
type SsServerConfig struct {
	Crypt     string `json:"crypt"`
	Password  string `json:"password"`
	LocalPort int    `json:"server_port"`
}
