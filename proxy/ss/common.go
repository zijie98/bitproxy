package ss

import "time"

const (
	SALT = "HELLO WORLD, HELLO MOLI"

	CryptXor     = "xor"
	CryptSalsa20 = "salsa20"

	sockbuf = 4194304

	readTimeout = time.Duration(500) * time.Second
)

const (
	socks5Version = 5
	socks5Connect = 1

	typeIpv4   = 1
	typeIpv6   = 4
	typeDomain = 3
)
