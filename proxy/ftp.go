package proxy

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/kataras/go-errors"
	"rkproxy/utils"
)

type FtpProxy struct {
	local_port  uint
	remote_host string
	remote_port uint

	log *utils.Logger
	ln  net.Listener
}

type FtpReturn struct {
	Ip   string
	Port uint
	Code int
}

func (r *FtpReturn) IsPassiveMode() bool {
	switch r.Code {
	case PASV, EPSV:
		return true
	}
	return false
}

const (
	PASV = 227
	EPSV = 229
	QUIT = 221
)

var errQuit = errors.New("Quit")
var errNotMatch = errors.New("Not match to the returned")

func (this *FtpProxy) handle(local_conn net.Conn) {
	defer func() {
		err := local_conn.Close()
		if err != nil {
			this.log.Info("local_conn close error ", err.Error())
		}
	}()

	remote_conn, err := net.DialTimeout("tcp", utils.JoinHostPort(this.remote_host, this.remote_port), 30*time.Second)
	if err != nil {
		this.log.Info("Dial to remote host error: ", err)
		return
	}
	defer func() {
		err := remote_conn.Close()
		if err != nil {
			this.log.Info("remote_conn close error ", err.Error())
		}
	}()

	var read_notify = make(utils.ReadNotify)
	var done = make(chan bool, 5)
	//var rwMu sync.WaitGroup

	copy_data := func(dsc io.Writer, src io.Reader) {
		_, err := utils.CopyWithNone(dsc, src)
		if err != nil {
			done <- true
		}
	}

	go func() {
		for {
			select {
			case <-time.After(60 * 2 * time.Second):
				done <- true
				return
			case <-read_notify:
			}
		}
	}()

	// remote -> self -> local
	go func() {
		remote := bufio.NewReader(remote_conn)
		for {
			str, err := remote.ReadBytes('\n')
			if err != nil {
				if len(str) == 0 {
					break
				}
			}
			resp, str, err := this.parseResult(str, read_notify)
			//fmt.Println("- ", string(str))
			if err != nil && err != errNotMatch {
				this.log.Info("Parse return error: ", err)
				if err == errQuit {
					done <- true
					return
				}
				continue
			}
			if resp != nil && resp.IsPassiveMode() {
				go this.pasvHandle(resp, read_notify)
				time.Sleep(1 * time.Second)
			}
			local_conn.Write(str)
		}
	}()

	// local -> self -> remote
	go copy_data(remote_conn, local_conn)

	<-done
	this.log.Info("exit .. ", local_conn.RemoteAddr().String())
}

func (this *FtpProxy) parseResult(b []byte, read_notify utils.ReadNotify) (resp *FtpReturn, src []byte, err error) {
	src = b
	resp, err = this.parseFtpRetrun(string(src))
	if err != nil {
		//this.log.Info("Get code error: ", err)
		return nil, nil, err
	}
	switch resp.Code {
	case PASV, EPSV:
		if err != nil {
			this.log.Info("Get pasv error: ", err)
			return nil, nil, err
		}
		if resp.Code == EPSV {
			src = []byte(fmt.Sprintf("229 Entering Extended Passive Mode (|||%d|).\r\n", resp.Port))
		} else if resp.Code == PASV {
			ip, _ := utils.PublicIp()
			ip = strings.Replace(ip, ".", ",", -1)
			port := fmt.Sprintf("%d,%d", resp.Port>>8, resp.Port&0xFF)
			src = []byte(fmt.Sprintf("227 Entering Passive Mode (%s,%s).\r\n", ip, port))
		}
	case QUIT:
		return nil, nil, errQuit
	}
	return
}

func (this *FtpProxy) dialToServer(pasv *FtpReturn) (conn net.Conn, err error) {
	server_address := utils.JoinHostPort(pasv.Ip, pasv.Port)
	conn, err = net.Dial("tcp", server_address)
	return
}

func (this *FtpProxy) pasvHandle(resp *FtpReturn, notify utils.ReadNotify) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", resp.Port))
	if err != nil {
		this.log.Info("Pasv listen error ", err)
		return
	}
	client_conn, err := ln.Accept()
	if err != nil {
		this.log.Info("Accept client error : ", err)
		if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
			time.Sleep(100 * time.Millisecond)
			this.log.Info("Accept client again! ", nerr)
		} else {
			return
		}
	}
	defer func() {
		client_conn.Close()
	}()

	pasv_conn, err := this.dialToServer(resp)
	if err != nil {
		this.log.Info("Dial to pasv ftp server error: ", err)
		return
	}
	defer func() {
		pasv_conn.Close()
	}()

	callback := func() {
		notify <- 0
	}
	utils.CopyWithBefore(client_conn, pasv_conn, nil, callback)
}

func (this *FtpProxy) parseFtpRetrun(str string) (ret *FtpReturn, err error) {
	r := regexp.MustCompile(`\d+`)
	matches := r.FindAllString(str, -1)
	if matches == nil {
		return nil, errNotMatch
	}
	matched_count := len(matches)

	uintBuff := make([]int, matched_count)
	for i, e := range matches {
		uintBuff[i], err = strconv.Atoi(e)
		if err != nil {
			return nil, err
		}
	}
	ret = new(FtpReturn)
	if matched_count >= 1 {
		// code
		ret.Code = int(uintBuff[0])
	}
	if matched_count >= 2 {
		if matched_count == 2 {
			// port
			ret.Port = uint(uintBuff[1])
			return
		}
		if matched_count >= 7 {
			// host
			ip := net.IPv4(byte(uintBuff[1]), byte(uintBuff[2]), byte(uintBuff[3]), byte(uintBuff[4]))
			ret.Ip = ip.String()
			// port
			ret.Port = uint(uintBuff[5]*256 + uintBuff[6])
		}
	}
	return
}

func (this *FtpProxy) Start() (err error) {
	this.log.Info("Listen port", this.local_port)

	this.ln, err = net.Listen("tcp", utils.JoinHostPort("", this.local_port))
	if err != nil {
		this.log.Info("Can't Listen port: ", this.local_port, " ", err)
		return err
	}
	for {
		conn, err := this.ln.Accept()
		if err != nil {
			this.log.Info("Can't Accept: ", this.local_port, " ", err)
		}
		go this.handle(conn)
	}
	return nil
}

func (tihs *FtpProxy) Stop() error {
	return nil
}

func (this *FtpProxy) LocalPort() uint {
	return this.local_port
}

func (this *FtpProxy) Traffic() (uint64, error) {
	return 0, nil
}

func NewFtpProxy(local_port uint, remote_host string, remote_port uint) *FtpProxy {
	return &FtpProxy{
		local_port:  local_port,
		remote_host: remote_host,
		remote_port: remote_port,
		log:         utils.NewLogger("FtpProxy"),
	}
}
