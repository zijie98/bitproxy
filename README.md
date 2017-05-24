### RKPROXY

rkproxy是一个 Shadowsocks/TCP/HTTPReproxy 代理软件


### 特性

#### shadowsocks

- 基于UDP实现的数据传输，加快网络环境较差时传输速度
- 只需一个程序，通过配置即可实现客户端-服务端
- 目前支持NOT、salsa20等加密传输（NOT最快）

#### TCP

- TCP数据代理

#### HTTP反向代理

- 简单易用的HTTP反向代理

### 使用教程

程序目录下新建config.json文件
```json
{
    "stream": [
        {
            "local_port": 8080,
            "local_net": "tcp", // tcp or udp
            "remote_host": "baidu.com",
            "remote_port": 80
        }
    ],
    "ss-client":[
        {
            "local_net": "tcp", // tcp or udp
            "local_port": 8081,
            "server_host": "127.0.0.1",
            "server_port": 8082,
            "password": "123",
            "channel_net": "kcp", // kcp udp tcp 客户端与服务器端之间的传输协议
            "crypt": "not"   // not salsa20 chacha20 rc4md5 
        }
    ],
    "ss-server": [
        {
            "channel_net": "kcp", // kcp udp tcp 同上 
            "server_port": 8082,
            "password": "123",
            "crypt": "not"  // 同上
        }
    ]
}
```

### 待续

- 暂时还未完善的测试
- 加入FTP代理
- shadowsocks支持原生的TCP/UDP协议
- 流量统计
- 宽带控制
- API功能
- IP屏蔽