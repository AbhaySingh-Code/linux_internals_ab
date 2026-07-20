package main

import (
	"fmt"
	"log"
	"net"
	"time"
)

func main() {

	//Server: listen for udp packets
	go func() {
		conn, err := net.ListenPacket("udp", ":5555")
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()

		buf := make([]byte, 2000)
		for {
			n, addr, err := conn.ReadFrom(buf)
			if err != nil {
				log.Println("server read: ", err)
				return
			}
			fmt.Printf("server got %d bytes from %s: %q\n", n, addr, buf[:n])

			conn.WriteTo(buf[:n], addr)
		}
	}()

	time.Sleep(50 * time.Millisecond)

	//Client send a packet and wait for reply
	conn, err := net.Dial("udp", "127.0.0.1:5555")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	conn.Write([]byte("hello tunnel"))

	buf := make([]byte, 2000)
	n, err := conn.Read(buf)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Client got reply: %q\n", buf[:n])
}
