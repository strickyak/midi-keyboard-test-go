// +build main

/*
$ go run quick-input-test.go
(0x0,0x0)
(0x0,0x0) 3 144 60 74
(0x0,0x0) 3 128 60 64
(0x0,0x0) 3 144 62 84
(0x0,0x0) 3 128 62 64
(0x0,0x0) 3 144 64 98
(0x0,0x0) 3 128 64 64
^Csignal: interrupt
*/
package main

import (
	"os"
)

func main() {
	r, err := os.Open("/dev/midi1")
	println(err)

	bb := make([]byte, 3)
	for {
		n, err := r.Read(bb)
		println(err, n, bb[0], bb[1], bb[2])
	}
}
