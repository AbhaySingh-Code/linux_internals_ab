package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	EV_KEY = 0x01
)

type inputEvent struct {
	Time  [16]byte
	Type  uint16
	Code  uint16
	Value int32
}

// KeyCodeToChar maps raw kernel key codes to human-readable strings.
func KeyCodeToChar(code uint16) string {
	keyMap := map[uint16]string{
		1:  "[ESC]",
		2:  "1",
		3:  "2",
		4:  "3",
		5:  "4",
		6:  "5",
		7:  "6",
		8:  "7",
		9:  "8",
		10: "9",
		11: "0",
		12: "-",
		13: "=",
		14: "[BACKSPACE]",
		15: "[TAB]",
		16: "q",
		17: "w",
		18: "e",
		19: "r",
		20: "t",
		21: "y",
		22: "u",
		23: "i",
		24: "o",
		25: "p",
		26: "[",
		27: "]",
		28: "[ENTER]",
		29: "[LCTRL]",
		30: "a",
		31: "s",
		32: "d",
		33: "f",
		34: "g",
		35: "h",
		36: "j",
		37: "k",
		38: "l",
		39: ";",
		40: "'",
		41: "`",
		42: "[LSHIFT]",
		43: "\\",
		44: "z",
		45: "x",
		46: "c",
		47: "v",
		48: "b",
		49: "n",
		50: "m",
		51: ",",
		52: ".",
		53: "/",
		54: "[RSHIFT]",
		56: "[LALT]",
		57: "[SPACE]",
		58: "[CAPSLOCK]",
	}

	if val, exists := keyMap[code]; exists {
		return val
	}
	return fmt.Sprintf("UNKNOWN(%d)", code)
}

func FindKeyboardDevice() (string, error) {

	file, err := os.Open("/proc/bus/input/devices")
	if err != nil {
		log.Fatalf("Failed to open file : %v\n", err)
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var isKeyboard bool
	var currentHandlers string

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			if isKeyboard && currentHandlers != "" {
				fields := strings.Fields(currentHandlers)
				for _, field := range fields {
					if strings.HasPrefix(field, "event") {
						return filepath.Join("/dev/input", field), nil
					}
				}
			}
			isKeyboard = false
			currentHandlers = ""
			continue
		}

		if strings.HasPrefix(line, "N: Name=") {
			name := strings.ToLower(line)
			if strings.Contains(name, "keyboard") || strings.Contains(name, "kbd") {
				isKeyboard = true
			}
		}

		if strings.HasPrefix(line, "H: Handlers=") {
			currentHandlers = line[12:]
		}
	}

	return "", fmt.Errorf("no keyboard device found")
}

func main() {
	path, err := FindKeyboardDevice()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Found keyboard device path: %s\n", path)

	file2, err := os.Open(path)
	if err != nil {
		fmt.Printf("Error opening device: %v\n", err)
		return
	}

	fmt.Printf("Listening for input on %s...\n", path)

	var ev inputEvent
	buffer := make([]byte, 24)

	for {
		_, err := file2.Read(buffer)
		if err != nil {
			fmt.Printf("Error reading event : %v\n", err)
			break
		}

		err = binary.Read(bytes.NewReader(buffer), binary.NativeEndian, &ev)
		if err != nil {
			fmt.Printf("Error parsing binary event: %v\n", err)
			continue
		}

		if ev.Type == EV_KEY {
			char := KeyCodeToChar(ev.Code)
			switch ev.Value {
			case 1:
				fmt.Printf("Key pressed: %s (code %d)\n", char, ev.Code)
			case 0:
				fmt.Printf("Key released: %s (code %d)\n", char, ev.Code)
			case 2:
				continue
			}
		}
	}
}
