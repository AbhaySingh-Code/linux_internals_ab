package main

import (
	"log"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func main() {
	ipLayer := &layers.IPv4{
		SrcIP:    net.ParseIP("192.168.1.151"),
		DstIP:    net.ParseIP("192.168.1.1"),
		Protocol: layers.IPProtocolICMPv4,
		Version:  4,
		TTL:      64,
	}

	tcpLayer := &layers.TCP{
		SrcPort: layers.TCPPort(12345),
		DstPort: layers.TCPPort(80),
		Seq:     11051999,
		SYN:     true,
		Window:  64240,
	}

	if err := tcpLayer.SetNetworkLayerForChecksum(ipLayer); err != nil {
		log.Fatalf("Failed to set checksum layer: %v", err)
	}

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		ComputeChecksums: true,
		FixLengths:       true,
	}

	err := gopacket.SerializeLayers(buf, opts, tcpLayer)
	if err != nil {
		log.Fatalf("Failed to serialize: %v", err)
	}

	// Send the raw bytes via a raw socket connection
	conn, err := net.Dial("ip4:tcp", "192.168.1.1")
	if err != nil {
		log.Fatalf("Failed to open raw socket: %v", err)
	}
	defer conn.Close()

	_, err = conn.Write(buf.Bytes())
	if err != nil {
		log.Fatalf("Failed to send packet: %v", err)
	}

	log.Println("Custom TCP SYN Packet sent successfully!")
}
