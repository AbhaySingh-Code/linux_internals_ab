package main

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/songgao/water"
)

func main() {
	//Create the TUN INterface
	config := water.Config{
		DeviceType: water.TUN,
	}
	config.Name = "tun0" // optional, kernel picks a name if omitted

	iface, err := water.New(config)
	if err != nil {
		log.Fatalf("Unabled to create TUN Interface:", err)
	}
	fmt.Println("Created interface: ", iface.Name())

	runCmd("ip", "addr", "add", "10.0.0.1/24", "dev", iface.Name())
	runCmd("ip", "link", "set", "dev", iface.Name(), "up")

	//Read packets in a loop
	packet := make([]byte, 2000)
	for {
		n, err := iface.Read(packet)
		if err != nil {
			log.Fatalf("read error: %v", err)
		}
		fmt.Printf("Received %d bytes: % x\n", n, packet[:n])
	}
}

func runCmd(cmd string, args ...string) {
	c := exec.Command(cmd, args...)
	if out, err := c.CombinedOutput(); err != nil {
		log.Fatalf("cmd %s %v failed: %v\n%s", cmd, args, err, out)
	}
}
