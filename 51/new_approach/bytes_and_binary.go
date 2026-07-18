package main

import (
	"encoding/binary"
	"fmt"
)

//serializing(encoding) numbers into a raw byte format and then deserializing (decoding) them back.
//54321 -> 0xD431 (in hexadecimal 54321 equals this value)
//80 -> 0x0050 (in hexadecimal 80 equals 0x0050)

func main() {
	buf := make([]byte, 4)                      //byte slice named buf with a length of 4 bytes [00, 00, 00, 00]
	binary.BigEndian.PutUint16(buf[0:2], 54321) //16 bit unsigned int takes 2 bytes of memory, buf[0:2]. This slices the first two slots of buf
	binary.BigEndian.PutUint16(buf[2:4], 80)
	fmt.Printf("% x\n", buf)
	fmt.Println(binary.BigEndian.Uint16(buf[0:2])) //read it back
	fmt.Println(binary.BigEndian.Uint16(buf[2:4]))
}
