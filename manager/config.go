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
	LocalPort  uint   `json:"local_port"`
	RemoteHost string `json:"remote_host"`
	RemotePort uint   `json:"remote_port"`
}

// TCP/UDP 代理配置
type StreamProxyConfig struct {
	LocalPort  uint   `json:"local_port"`
	LocalNet   string `json:"local_net"`
	RemoteHost string `json:"remote_host"`
	RemotePort uint   `json:"remote_port"`
}

// Shadowsocks 客户端配置
type SsClientConfig struct {
	LocalNet   string `json:"local_net"`
	LocalPort  uint   `json:"local_port"`
	ServerHost string `json:"server_host"`
	ServerPort uint   `json:"server_port"`
	ChannelNet string `json:"channel_net"`
	Password   string `json:"password"`
	Crypt      string `json:"crypt"`
}

// Shadowsocks 服务器端配置
type SsServerConfig struct {
	ChannelNet string `json:"channel_net"`
	Crypt      string `json:"crypt"`
	Password   string `json:"password"`
	LocalPort  uint   `json:"server_port"`
}
