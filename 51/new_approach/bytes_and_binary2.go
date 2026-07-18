package main

import (
	"encoding/binary"
	"fmt"
)

func main() {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint16(buf[0:2], 54321)
	binary.BigEndian.PutUint16(buf[2:4], 80)
	fmt.Printf("% x\n", buf)
	fmt.Println(binary.BigEndian.Uint16(buf[0:2]))
	fmt.Println(binary.BigEndian.Uint16(buf[2:4]))
}
