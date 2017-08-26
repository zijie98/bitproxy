package proxy

import (
	"bufio"
	"fmt"
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

func (c *FtpReturn) IsOpening() bool {
	if c.Code == OPENING {
		return true
	}
	return false
}

type FtpCommand struct {
	Command string
	Params  string
}

const (
	download = iota
	upload
)

type FtpCommandFlag struct {
	uploadOrDownload int
}

func (c *FtpCommandFlag) lastIsUpload() bool {
	if c.uploadOrDownload == upload {
		return true
	}
	return false
}
func (c *FtpCommandFlag) lastIsDownload() bool {
	if c.uploadOrDownload == download {
		return true
	}
	return false
}
func (c *FtpCommandFlag) Mark(command *FtpCommand) {
	if command.isUpload() {
		c.uploadOrDownload = upload
	} else if command.isDownload() {
		c.uploadOrDownload = download
	}
}

func (c *FtpCommand) isUpload() bool {
	if c.Command == C_STOR {
		return true
	}
	return false
}

func (c *FtpCommand) isDownload() bool {
	if c.Command == C_RETR {
		return true
	}
	return false
}

func (c *FtpCommand) isQuit() bool {
	if c.Command == C_QUIT {
		return true
	}
	return false
}

const (
	PASV    = 227
	EPSV    = 229
	QUIT    = 221
	OPENING = 150

	C_STOR = "STOR"
	C_RETR = "RETR"
	C_QUIT = "QUIT"
)

var errQuit = errors.New("Quit")
var errNotMatch = errors.New("Not match to the returned")
var errCommand = errors.New("Error command")

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
	var command_mark FtpCommandFlag

	finish := func() {
		done <- true
	}
	notify := func(n int64, err error) {
		read_notify <- n
	}

	go func() {
		for {
			select {
			case <-time.After(60 * 2 * time.Second):
				finish()
				return
			case <-read_notify:
			}
		}
	}()

	go func() {
		var pasv_resp *FtpReturn
		var pasv_server_conn net.Conn
		var pasv_client_conn net.Conn
		remote := bufio.NewReader(remote_conn)
		defer func() {
			if pasv_server_conn != nil {
				pasv_server_conn.Close()
			}
			if pasv_client_conn != nil {
				pasv_client_conn.Close()
			}
		}()
		for {
			// remote -> self -> local
			str, err := remote.ReadBytes('\n')
			if err != nil {
				break
			}
			read_notify <- 0
			resp, str, err := this.parseResult(str)
			if err != nil && err != errNotMatch {
				this.log.Info("Parse return error: ", err)
				if err == errQuit {
					finish()
					return
				}
				continue
			}
			if resp != nil {
				if resp.IsPassiveMode() {
					pasv_resp = resp
					go func() {
						pasv_client_conn, err = this.listenPasvConnFromClient(resp)
						if err != nil {
							this.log.Info("Lisent client conn error :", err)
							return
						}
						pasv_server_conn, err = this.dialToServer(resp)
						if err != nil {
							this.log.Info("Dial to pasv ftp server error: ", err)
							return
						}
					}()
				}
				if pasv_resp != nil && resp.IsOpening() {
					if command_mark.lastIsDownload() {
						go func() {
							_, e := utils.CopyWithAfter(pasv_client_conn, pasv_server_conn, notify, notify)
							if e != nil {
								pasv_server_conn.Close()
								pasv_server_conn = nil
								pasv_client_conn.Close()
								pasv_client_conn = nil
							}
						}()
					} else {
						go func() {
							_, e := utils.CopyWithAfter(pasv_server_conn, pasv_client_conn, notify, notify)
							if e != nil {
								pasv_server_conn.Close()
								pasv_server_conn = nil
								pasv_client_conn.Close()
								pasv_client_conn = nil
							}
						}()
					}
				}
			}
			local_conn.Write(str)
		}
	}()

	go func() {
		local := bufio.NewReader(local_conn)
		for {
			// local -> self -> remote
			str, err := local.ReadBytes('\n')
			if err != nil {
				break
			}
			command, err := this.parseCommand(str)
			if err != nil {
				this.log.Info("Parse command error ", err)
				continue
			}
			command_mark.Mark(command)
			remote_conn.Write(str)
		}
	}()
	//go copy_data(remote_conn, local_conn)

	<-done
	this.log.Info("exit .. ", local_conn.RemoteAddr().String())
}

func (this *FtpProxy) parseCommand(b []byte) (command *FtpCommand, err error) {
	c := strings.Split(string(b), " ")
	command = &FtpCommand{}
	if len(c) == 0 {
		return nil, errCommand
	}
	if len(c) >= 1 {
		command.Command = c[0] //strings.TrimSpace(c[0])
	}
	if len(c) >= 2 {
		command.Params = c[1]
	}
	return command, nil
}

func (this *FtpProxy) parseResult(b []byte) (resp *FtpReturn, src []byte, err error) {
	src = b
	resp, err = this.parseFtpRetrun(string(src))
	if err != nil {
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

func (this *FtpProxy) listenPasvConnFromClient(resp *FtpReturn) (client_conn net.Conn, err error) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", resp.Port))
	if err != nil {
		this.log.Info("Pasv listen error ", err)
		return
	}
	client_conn, err = ln.Accept()
	if err != nil {
		this.log.Info("Accept client error : ", err)
		//if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
	}
	return
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
