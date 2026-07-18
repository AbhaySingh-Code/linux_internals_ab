package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"syscall"
)

func main() {
	// 1. Configure networking parameters
	srcIPStr := "192.168.1.151" // Change to your machine's actual local interface IP
	dstIPStr := "192.168.1.1"
	srcPort := uint16(54321)
	dstPort := uint16(80)
	customSeq := uint32(11051999)

	srcIP := net.ParseIP(srcIPStr).To4()
	dstIP := net.ParseIP(dstIPStr).To4()

	// 2. Open a raw native TCP Socket via Linux System Calls (No libpcap required)
	// AF_INET = IPv4, SOCK_RAW = Raw Access, IPPROTO_TCP = Handle TCP segments
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_TCP)
	if err != nil {
		log.Fatalf("Failed to open raw socket (Run as root/sudo!): %v", err)
	}
	defer syscall.Close(fd)

	// Set a read timeout so the loop doesn't hang indefinitely if the packet drops
	tv := syscall.Timeval{Sec: 3, Usec: 0}
	syscall.SetsockoptTimeval(fd, syscall.SOL_SOCKET, syscall.SO_RCVTIMEO, &tv)

	// 3. Manually craft the TCP Header binary payload slice
	// A standard TCP header without options is exactly 20 bytes long
	tcpHeader := make([]byte, 20)
	binary.BigEndian.PutUint16(tcpHeader[0:2], srcPort)   // Source Port
	binary.BigEndian.PutUint16(tcpHeader[2:4], dstPort)   // Destination Port
	binary.BigEndian.PutUint32(tcpHeader[4:8], customSeq) // Sequence Number
	binary.BigEndian.PutUint32(tcpHeader[8:12], 0)        // Acknowledgment Number (0 for SYN)
	tcpHeader[12] = 0x50                                  // Data Offset: 5 words (5 * 4 = 20 bytes), Reserved flags = 0
	tcpHeader[13] = 0x02                                  // Flags: SYN bit explicitly set to 1 (00000010)
	binary.BigEndian.PutUint16(tcpHeader[14:16], 64240)   // Window Size
	// tcpHeader[16:18] is the Checksum space (initialized to 0 before calculating)
	binary.BigEndian.PutUint16(tcpHeader[18:20], 0) // Urgent Pointer

	// 4. Calculate the TCP Checksum using the structural Pseudo Header math
	checksum := calcTCPChecksum(tcpHeader, srcIP, dstIP)
	binary.BigEndian.PutUint16(tcpHeader[16:18], checksum)

	// Change syscall.SockaddrIn to syscall.SockaddrInet4
	sockAddr := &syscall.SockaddrInet4{
		Port: int(dstPort),
		Addr: [4]byte{dstIP[0], dstIP[1], dstIP[2], dstIP[3]},
	}

	// 6. Transmit the handcrafted TCP packet into the network card
	err = syscall.Sendto(fd, tcpHeader, 0, sockAddr)
	if err != nil {
		log.Fatalf("Failed to transmit manual header packet: %v", err)
	}
	fmt.Printf("[SYN Sent] Outbound Raw TCP packet fired to %s:%d with Seq: %d\n", dstIPStr, dstPort, customSeq)

	// 7. Read Loop: Catch incoming network raw byte responses
	// Raw sockets read at Layer 3, meaning the buffer will contain the IP header followed by the TCP header
	recvBuf := make([]byte, 4096)
	fmt.Println("Listening for raw packet responses directly on the socket file descriptor...")

	for {
		n, _, err := syscall.Recvfrom(fd, recvBuf, 0)
		if err != nil {
			log.Fatalf("Timeout or connection read error: %v", err)
		}

		// The IPv4 header tells us how big it is via its lower 4 bits of the first byte (IHL)
		ipHeaderLen := int(recvBuf[0]&0x0F) * 4
		if n < ipHeaderLen+20 {
			continue // Packet too short to be a valid TCP segment response
		}

		// Extract the TCP Segment boundary out of the full IP packet data
		tcpSegment := recvBuf[ipHeaderLen:n]
		respSrcPort := binary.BigEndian.Uint16(tcpSegment[0:2])
		respDstPort := binary.BigEndian.Uint16(tcpSegment[2:4])

		// Ensure this packet belongs to our active data channel connection
		if respSrcPort == dstPort && respDstPort == srcPort {
			respSeq := binary.BigEndian.Uint32(tcpSegment[4:8])
			respAck := binary.BigEndian.Uint32(tcpSegment[8:12])
			flags := tcpSegment[13]

			// Read flag bit states via bitwise masks
			isSYN := (flags & 0x02) != 0
			isACK := (flags & 0x10) != 0
			isRST := (flags & 0x04) != 0

			fmt.Println("\n--- [MANUALLY INTERCEPTED RAW TCP HEADER] ---")
			fmt.Printf("Source Port:      %d\n", respSrcPort)
			fmt.Printf("Destination Port: %d\n", respDstPort)
			fmt.Printf("Sequence Number:  %d\n", respSeq)
			fmt.Printf("Acknowledgment:   %d (Expected: %d)\n", respAck, customSeq+1)
			fmt.Printf("Raw Flag Byte:    0x%02x (SYN=%t, ACK=%t, RST=%t)\n", flags, isSYN, isACK, isRST)
			fmt.Println("----------------------------------------------")
			break
		}
	}
}

// Checksum calculation requires pairing the TCP segment payload with an L3 Pseudo-Header block
func calcTCPChecksum(tcpHeader []byte, srcIP, dstIP []byte) uint16 {
	pseudoHeader := make([]byte, 12)
	copy(pseudoHeader[0:4], srcIP)
	copy(pseudoHeader[4:8], dstIP)
	pseudoHeader[8] = 0                   // Reserved zero space
	pseudoHeader[9] = syscall.IPPROTO_TCP // Protocol Type
	binary.BigEndian.PutUint16(pseudoHeader[10:12], uint16(len(tcpHeader)))

	// Merge pseudo header structure bytes with target raw header bytes to compute full data block sum
	buffer := append(pseudoHeader, tcpHeader...)

	var sum uint32
	for i := 0; i < len(buffer)-1; i += 2 {
		sum += uint32(binary.BigEndian.Uint16(buffer[i : i+2]))
	}
	if len(buffer)%2 != 0 {
		sum += uint32(buffer[len(buffer)-1]) << 8
	}

	for sum>>16 != 0 {
		sum = (sum & 0xFFFF) + (sum >> 16)
	}

	return uint16(^sum)
}
