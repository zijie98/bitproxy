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

程序目录下新建config.json文件, 参考`config.json.example`

## 待续

#### 功能
- 暂时还未完善的测试
- 加入FTP代理
- shadowsocks支持原生的TCP/UDP协议
- 流量统计
- 宽带控制
- API功能
- IP屏蔽

#### API

- 获取启动的代理
- 启动/停止代理
- 获取/设置屏蔽IP