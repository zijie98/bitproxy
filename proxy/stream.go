/**
tcp udp proxy
**/
package proxy

import (
	"net"
	"time"

	"rkproxy/proxy/ss"
	"rkproxy/services"
	"rkproxy/utils"
)

const READ_TIMEOUT = 60

type StreamProxy struct {
	local_port     uint
	local_net      ss.NetProtocol
	remote_host    string
	remote_port    uint
	rate           uint
	enable_traffic bool
	enable_black   bool

	ln  net.Listener
	log *utils.Logger
}

func (this *StreamProxy) Start() (err error) {
	this.ln, err = net.Listen(string(this.local_net), utils.JoinHostPort("", this.local_port))
	if err != nil {
		this.log.Info("Can't Listen port: ", this.local_port, " ", err)
		return err
	}
	this.log.Info("Listen port", this.local_port)
	for {
		conn, err := this.ln.Accept()
		if err != nil {
			this.log.Info("Can't Accept: ", err)
			continue
		}
		this.log.Info("Client ip", conn.RemoteAddr().String())
		if this.isBlack(conn.RemoteAddr()) {
			conn.Close()
			continue
		}
		go this.handle(conn)
	}
	return nil
}

func (this *StreamProxy) isBlack(addr net.Addr) bool {
	if !this.enable_black {
		return false
	}
	ip, _, _ := net.SplitHostPort(addr.String())
	if services.Wall.IsBlack(ip) {
		this.log.Info("Ip was black ", ip)
		return true
	}
	services.Filter <- services.RequestAt{
		Ip: ip,
		At: time.Now(),
	}
	return false
}

func (this *StreamProxy) Stop() error {
	if this.ln == nil {
		return nil
	}
	return this.ln.Close()
}

func (this *StreamProxy) LocalPort() uint {
	return this.local_port
}

func (this *StreamProxy) Traffic() (uint64, error) {
	return services.GetTraffic(this.local_port)
}

func (this *StreamProxy) handle(local_conn net.Conn) {
	var done = make(chan bool, 2)

	defer func() {
		err := local_conn.Close()
		if err != nil {
			this.log.Info("local_conn close error ", err.Error())
		}
	}()

	remote_conn, err := net.Dial(
		string(this.local_net),
		utils.JoinHostPort(this.remote_host, this.remote_port),
	)
	if err != nil {
		this.log.Info("Dial remote host error: ", err)
		return
	}
	defer func() {
		err := remote_conn.Close()
		if err != nil {
			this.log.Info("Remote close error ", err.Error())
		}
	}()

	// 读取到数据后通知、相当于超时机制
	var read_notify = make(utils.ReadNotify, 5)
	go func() {
		for {
			select {
			case <-time.After(READ_TIMEOUT * time.Second):
				this.log.Info("Timeout")
				done <- true
				return
			case <-read_notify:
				break
			}
		}
	}()
	readAfterFunc := func(n int64, e error) {
		read_notify <- n
	}

	var copy_data = func(dsc net.Conn, src net.Conn, limit *utils.Limit) {
		_, err := utils.Copy(dsc, src, limit, nil, nil, readAfterFunc, nil, this.trafficStats, nil)
		if err != nil {
			done <- true
		}
	}
	go copy_data(remote_conn, local_conn, nil)
	go copy_data(local_conn, remote_conn, this.Limit())
	<-done
}

// 流量统计
func (this *StreamProxy) trafficStats(n int64, e error) {
	if this.enable_traffic == false {
		return
	}
	services.AddTrafficStats(this.local_port, n)
}

// 流量限制
func (this *StreamProxy) Limit() *utils.Limit {
	return &utils.Limit{
		Rate: this.rate,
	}
}

func NewStreamProxy(local_net ss.NetProtocol, local_port uint, remote_host string, remote_port uint, rate uint) *StreamProxy {
	return &StreamProxy{
		local_net:   local_net,
		local_port:  local_port,
		remote_host: remote_host,
		remote_port: remote_port,
		rate:        rate,
		log:         utils.NewLogger("StreamProxy"),
	}
}
