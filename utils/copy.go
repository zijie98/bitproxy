package utils

import (
	"io"
	//"sync"
	"net"
	"time"
)

//var bufpool *sync.Pool
var bufpool BytePool

func init() {
	//bufpool = &sync.Pool{}
	//bufpool.New = func() interface{} {
	//	return make([]byte, 1024 * 1)  // 1kb
	//}
	bufpool.Init(0, 1024)
}

func Copy(dst io.Writer, src io.Reader) (written int64, err error) {
	//fmt.Println("--------", bufpool.entries())
	if wt, ok := src.(io.WriterTo); ok {
		return wt.WriteTo(dst)
	}
	if rt, ok := dst.(io.ReaderFrom); ok {
		return rt.ReadFrom(src)
	}

	//buf := bufpool.Get().([]byte)
	buf := bufpool.Get(1024)
	defer bufpool.Put(buf)
	for {
		src.(net.Conn).SetReadDeadline(time.Now().Add(time.Duration(500) * time.Second))
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
	}
	return written, err
}
