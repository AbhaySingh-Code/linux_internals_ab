package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FindKeyboardDevice scans /proc/bus/input/devices for a keyboard event path
func FindKeyboardDevice() (string, error) {
	file, err := os.Open("/proc/bus/input/devices")
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var currentHandlers string
	var isKeyboard bool

	for scanner.Scan() {
		line := scanner.Text()

		// An empty line signals the end of a device block
		if line == "" {
			if isKeyboard && currentHandlers != "" {
				// Parse event number from handlers (e.g., "sysrq kbd event3")
				fields := strings.Fields(currentHandlers)
				for _, field := range fields {
					if strings.HasPrefix(field, "event") {
						return filepath.Join("/dev/input", field), nil
					}
				}
			}
			// Reset block tracking
			isKeyboard = false
			currentHandlers = ""
			continue
		}

		// Look for device name or EV bitmask indicating a keyboard
		if strings.HasPrefix(line, "N: Name=") {
			name := strings.ToLower(line)
			if strings.Contains(name, "keyboard") || strings.Contains(name, "kbd") {
				isKeyboard = true
			}
		}

		// Read the handlers line
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
	fmt.Printf("Found Keyboard Device Path: %s\n", path)
}
