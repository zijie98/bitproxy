package main

import (
	"fmt"
	//"unsafe"
)

type AA struct {
	aa map[int]int
}

func main() {
	a := AA{}
	a.aa = make(map[int]int)
	a.aa[1] = 1
	a.aa[2] = 2
	fmt.Println(a.aa)
}
