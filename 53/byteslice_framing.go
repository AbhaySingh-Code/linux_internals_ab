package main

import "fmt"

func main() {

	buf := make([]byte, 2000)

	fakePacket := []byte{0x45, 0x00, 0x00, 0x1c, 0xAB, 0xCD} //6 bytes
	n := copy(buf, fakePacket)

	fmt.Println("len(buf):", len(buf))
	fmt.Println("n:", n)

	broken := buf
	fmt.Println("broken lenght sent downstream: ", len(broken))

	correct := buf[:n]
	fmt.Println("Correct lenght sent downstream:", len(correct))

	fmt.Printf("correct payload: % x\n", correct)
}
