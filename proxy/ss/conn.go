package ss

import (
	"crypto/cipher"
	"crypto/rand"
	"io"
	"net"
)

type CryptConn struct {
	net.Conn
	cipher *cipherInfo
	key    []byte
	iv     []byte
	dec    cipher.Stream
	enc    cipher.Stream
}

// crypt_name 加密类型（"chacha20"等）
func NewCryptConn(conn net.Conn, password, cryptName string) (cryptConn *CryptConn, err error) {
	cipher := CipherMethod[cryptName]
	key := EvpBytesToKey(password, cipher.keyLen)

	// 从内存池申请ivLen大小的内存
	iv := IvPool.Get(cipher.ivLen)

	// 生成iv
	_, err = io.ReadFull(rand.Reader, iv)
	if err != nil {
		return
	}

	cipherInfo := &cipherInfo{
		keyLen:    cipher.keyLen,
		ivLen:     cipher.ivLen,
		newStream: cipher.newStream,
	}
	cryptConn = &CryptConn{
		Conn:   conn,
		cipher: cipherInfo,
		key:    key,
		iv:     iv,
	}
	return cryptConn, err
}

// 加密并且写入 conn
func (this *CryptConn) Write(p []byte) (n int, err error) {
	sendIv := false
	if this.enc == nil {
		this.enc, err = this.cipher.newStream(this.key, this.iv, Encrypt)
		sendIv = true
		if err != nil {
			return 0, err
		}
	}
	if sendIv {
		this.Conn.Write(this.iv)
	}
	this.enc.XORKeyStream(p, p)
	return this.Conn.Write(p)
}

// 读取conn并且解密
func (this *CryptConn) Read(p []byte) (n int, err error) {
	if this.dec == nil {
		if _, err = io.ReadFull(this.Conn, this.iv); err != nil {
			return
		}
		this.dec, err = this.cipher.newStream(this.key, this.iv, Decrypt)
		if err != nil {
			return
		}
	}
	n, err = this.Conn.Read(p)
	if n > 0 {
		this.dec.XORKeyStream(p[0:n], p[0:n])
	}
	return
}

func (this *CryptConn) Close() error {
	IvPool.Put(this.iv)
	return this.Conn.Close()
}
