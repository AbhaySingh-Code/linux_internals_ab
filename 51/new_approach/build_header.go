package main

import (
	"encoding/binary"
	"fmt"
)

func buildSYNHeader(srcPort, dstPort uint16, seq uint32) []byte {
	h := make([]byte, 20)
	binary.BigEndian.PutUint16(h[0:2], srcPort)
	binary.BigEndian.PutUint16(h[2:4], dstPort)
	binary.BigEndian.PutUint32(h[4:8], seq)
	h[12] = 0x50                                // offset
	h[13] = 0x02                                // syn
	binary.BigEndian.PutUint16(h[14:16], 64240) // window size
	return h
}

func main() {
	h := buildSYNHeader(54321, 80, 11051999)
	fmt.Printf("% x\n", h)
}
