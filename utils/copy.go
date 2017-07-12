package utils

import (
	"io"
	//"sync"
	"fmt"
	"net"
	"time"
)

var CopyPool BytePool

const CopyPoolSize = 4096 // 4KB
const maxNBuf = 1024

func init() {
	CopyPool.Init(1*time.Hour, maxNBuf)
}

func Copy(dst io.Writer, src io.Reader, limit *Limiter) (written int64, err error) {
	buf := CopyPool.Get(CopyPoolSize)
	defer CopyPool.Put(buf)

	for {
		src.(net.Conn).SetReadDeadline(time.Now().Add(5 * time.Second))
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw > 0 {
				written += int64(nw)
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er == io.EOF {
			break
		}
		if er != nil {
			err = er
			break
		}
		if limit != nil {
			limit.Sleep()
		}
	}
	return written, err
}

// 1s / (L/B) = sleep
type Limiter struct {
	Rate uint // KB/每秒
}

func (l *Limiter) Sleep() {
	s := 1000 / (l.Rate / (CopyPoolSize / 1024))
	t := time.Duration(s) * time.Microsecond
	fmt.Printf("sleep ... %d ms", t)
	time.Sleep(t)
}
