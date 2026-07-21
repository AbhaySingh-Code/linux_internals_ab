package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
)

const (
	EV_KEY     = 0x01
	DevicePath = "/dev/input/event0"
)

type InputEvent struct {
	Time  [16]byte
	Type  uint16
	Code  uint16
	Value int32
}

func main() {
	file, err := os.Open(DevicePath)
	if err != nil {
		fmt.Printf("Error opening device: %v\n", err)
		return
	}

	fmt.Printf("Listening for input on %s...\n", DevicePath)

	var ev InputEvent
	buffer := make([]byte, 24)

	for {
		_, err := file.Read(buffer)
		if err != nil {
			fmt.Printf("Error reading event: %v\n", err)
			break
		}

		err = binary.Read(bytes.NewReader(buffer), binary.NativeEndian, &ev)
		if err != nil {
			fmt.Printf("Error parsing binary event: %v\n", err)
			continue
		}

		if ev.Type == EV_KEY {
			switch ev.Value {
			case 1:
				fmt.Printf("Key pressed: code %d\n", ev.Code)
			case 0:
				fmt.Printf("key Released: code %d\n", ev.Code)
			case 2:
				// key auto repeated while held down
				continue
			}
		}
	}
}
