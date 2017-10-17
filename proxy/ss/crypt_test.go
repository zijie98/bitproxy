package ss

import (
	"bytes"
	"crypto/rand"
	"io"
	"testing"
)

func TestNotCrypt(t *testing.T) {
	testCrypt(t, CryptXor)
}

func TestRC4MD5Crypt(t *testing.T) {
	testCrypt(t, CryptRc4md5)
}

func TestChaCha20Crypt(t *testing.T) {
	testCrypt(t, CryptChacha20)
}

func testCrypt(t *testing.T, method string) {
	cipher := CipherMethod[method]
	password := "hello"
	buff := []byte("hello")
	buff2 := make([]byte, len(buff))

	iv := make([]byte, cipher.ivLen)
	io.ReadFull(rand.Reader, iv)

	key := EvpBytesToKey(password, cipher.keyLen)

	stream, err := cipher.newStream(key, iv, Encrypt)
	if err != nil {
		t.Fatal(err)
	}

	if method == CryptXor {
		stream.XORKeyStream(buff2, buff)
		stream.XORKeyStream(buff, buff2)
	} else {
		stream.XORKeyStream(buff2, buff)
		stream, err = cipher.newStream(key, iv, Decrypt)
		stream.XORKeyStream(buff, buff2)
	}

	if bytes.Compare(buff, []byte("hello")) != 0 {
		t.Fatalf("%s加密有问题呢", method)
	}
}
