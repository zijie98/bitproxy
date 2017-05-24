package main

import (
	"crypto/rc4"
	"crypto/sha1"
	"fmt"
	"golang.org/x/crypto/pbkdf2"
)

func main() {
	pass := pbkdf2.Key([]byte("password"), []byte("HELLO"), 4096, 32, sha1.New)
	fmt.Printf("%x\n", pass)

	key := []byte("key")
	src := []byte("hello")
	des := make([]byte, len(src))

	fmt.Println(src)

	c, _ := rc4.NewCipher(key)
	c.XORKeyStream(des, src)
	fmt.Println(des)

	c, _ = rc4.NewCipher(key)
	c.XORKeyStream(src, des)
	fmt.Println(src)
}
