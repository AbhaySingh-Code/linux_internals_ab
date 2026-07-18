package main

import (
	"encoding/binary"
	"syscall"
	"testing"
)

func TestChecksum(t *testing.T) {
	srcIP := []byte{192, 168, 1, 151}
	dstIP := []byte{192, 168, 1, 1}

	header := make([]byte, 20)
	binary.BigEndian.PutUint16(header[0:2], 54321)    //src port
	binary.BigEndian.PutUint16(header[2:4], 80)       //dst port
	binary.BigEndian.PutUint32(header[4:8], 11051999) //sequence number
	binary.BigEndian.PutUint32(header[8:12], 0)       //ack number
	header[12] = 0x50                                 // data offset
	header[13] = 0x02                                 // SYN flag
	binary.BigEndian.PutUint16(header[14:16], 64240)  // window size

	got := calcTCPChecksum(header, srcIP, dstIP)

	if got == 0 {
		t.Errorf("checksum should not be zero, got %d", got)
	}
}

// Copied from test2.go

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
