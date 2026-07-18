package main

import (
	"fmt"
)

func main() {
	var flags byte = 0x02 //SYN packet
	isSYN := flags&0x02 != 0
	isACK := flags&0x10 != 0
	fmt.Println("isSYN:", isSYN, "isAck:", isACK)

	flags = 0x02 | 0x10
	fmt.Printf("%08b\n", flags)
}
