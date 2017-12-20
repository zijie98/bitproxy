/**
tcp udp proxy
**/
package proxy

import (
	"net"
	"time"

	"github.com/molisoft/bitproxy/services"
	"github.com/molisoft/bitproxy/utils"
)

const READ_TIMEOUT = 60

type StreamProxy struct {
	localPort     uint
	localNet      NetProtocol
	remoteHost    string
	remotePort    uint
	rate          uint
	enableTraffic bool
	enableBlack   bool

	ln   net.Listener
	log  *utils.Logger
	done bool
}

func (this *StreamProxy) Start() (err error) {
	this.ln, err = net.Listen(string(this.localNet), utils.JoinHostPort("", this.localPort))
	if err != nil {
		this.log.Info("Can't Listen port: ", this.localPort, " ", err)
		return err
	}
	this.log.Info("Listen port", this.localPort)
	for !this.done {
		conn, err := this.ln.Accept()
		if err != nil {
			// this.log.Info("Can't Accept: ", err)
			time.Sleep(100 * time.Millisecond)
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
	if !this.enableBlack {
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
	this.done = true
	if this.ln == nil {
		return nil
	}
	return this.ln.Close()
}

func (this *StreamProxy) LocalPort() uint {
	return this.localPort
}

func (this *StreamProxy) Traffic() (uint64, error) {
	return services.GetTraffic(this.localPort)
}

func (this *StreamProxy) handle(localConn net.Conn) {
	var done = make(chan bool, 2)

	defer func() {
		err := localConn.Close()
		if err != nil {
			this.log.Info("local_conn close error ", err.Error())
		}
	}()

	remoteConn, err := net.Dial(
		string(this.localNet),
		utils.JoinHostPort(this.remoteHost, this.remotePort),
	)
	if err != nil {
		this.log.Info("Dial remote host error: ", err)
		return
	}
	defer func() {
		err := remoteConn.Close()
		if err != nil {
			this.log.Info("Remote close error ", err.Error())
		}
	}()

	// 读取到数据后通知、相当于超时机制
	var readNotify = make(utils.ReadNotify, 5)
	go func() {
		for !this.done {
			select {
			case <-time.After(READ_TIMEOUT * time.Second):
				this.log.Info("Timeout")
				done <- true
				return
			case <-readNotify:
				break
			}
		}
	}()
	readAfterFunc := func(n int64, e error) {
		readNotify <- n
	}

	var copyData = func(dsc net.Conn, src net.Conn, limit *utils.Limit) {
		_, err := utils.Copy(dsc, src, limit, nil, nil, readAfterFunc, nil, this.trafficStats, 600*time.Second)
		if err != nil {
			done <- true
		}
	}
	go copyData(remoteConn, localConn, nil)
	go copyData(localConn, remoteConn, this.Limit())
	<-done
}

// 流量统计
func (this *StreamProxy) trafficStats(n int64, e error) {
	if this.enableTraffic == false {
		return
	}
	if n > 0 {
		services.AddTrafficStats(this.localPort, n)
	}
}

// 流量限制
func (this *StreamProxy) Limit() *utils.Limit {
	return &utils.Limit{
		Rate: this.rate,
	}
}

func NewStreamProxy(localNet NetProtocol, localPort uint, remoteHost string, remotePort uint, rate uint, enableTraffic bool) Proxyer {
	return &StreamProxy{
		localNet:      localNet,
		localPort:     localPort,
		remoteHost:    remoteHost,
		remotePort:    remotePort,
		enableTraffic: enableTraffic,
		rate:          rate,
		log:           utils.NewLogger("StreamProxy"),
	}
}
