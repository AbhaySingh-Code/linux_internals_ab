package main 

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"syscall"
)

const (
	direntBufSize = 4096
	fileBufSize = 8192
)

func main() {
	fmt.Printf("%-8s %-50s\n", "PID", "COMMAND LINE")
	fmt.Println(strings.Repeat("-", 60))

	//1. Open the /proc directory using the openat/open system call -> Returns a raw file descriptor (int)
	procFd, err := syscall.Open("/proc", syscall.O_RDONLY|syscall.O_DIRECTORY, 0)
	if err != nil {
		fmt.Printf("Failed to open /proc via raw API: %v\n", err)
	}
	defer syscall.Close(procFd)

	// Buffer to store the raw linux_dirent64 structures
	buf := make([]byte, direntBufSize)

	for {
		//2. Invoke the getdents64 systemcall via Go's syscall wrapper 
		n, err := syscall.ReadDirent(procFd, buf)
		if err != nil {
			fmt.Printf("Error reading directory entries: %v\n", err)
			break
		}
		if n <= 0 {
			break // End of directory stream
		}

		// Properly pass an empty target slice to hold parsed names
		var names []string

		// Parse the raw byte buffer into individual directory entry names
		_, _, names = syscall.ParseDirent(buf[:n], -1, names)

		for _, name := range names {
			//If its not a numeric directory, its not a process
			pid, err := strconv.Atoi(name)
			if err != nil {
				continue
			}

			// 3. Read commandline using low-level open/read system calls
			cmdline := getCmdLineRaw(pid)
			if cmdline == "" {
				continue
			}
			fmt.Printf("%-8d %-50s\n", pid, cmdline)

			//4. Show raw maps output for our own process
			if pid == syscall.Getpid() {
				fmt.Println("\n[!] Memory maps via Raw Linux API:")
				displayMapsRaw(pid)
				fmt.Println(strings.Repeat("-", 60))
			}
		}
	}
}

func getCmdLineRaw(pid int) string {
	path := fmt.Sprintf("/proc/%d/cmdline", pid)
	//sys_open
	fd, err := syscall.Open(path, syscall.O_RDONLY, 0)
	if err != nil {
		return ""
	}
	defer syscall.Close(fd)

	buf := make([]byte, fileBufSize)
	//sys_read
	n, err := syscall.Read(fd, buf)
	if err != nil || n <= 0 {
		return ""
	}

	//clean up null bytes into spaces
	actualData := buf[:n]
	for i , b := range actualData {
		if b == 0 {
			actualData[i] = ' '
		}
	}
	return strings.TrimSpace(string(actualData))
}

// displayMapsRaw opens and read /proc/[pid]/maps using raw system call
func displayMapsRaw(pid int){
	path := fmt.Sprintf("/proc/%d/maps", pid)

	fd, err := syscall.Open(path, syscall.O_RDONLY, 0)
	if err != nil {
		fmt.Printf(" Error opening maps: %v\n", err)
		return 
	}
	defer syscall.Close(fd)

	buf := make([]byte, fileBufSize)
	n, err := syscall.Read(fd, buf)
	if err != nil || n <= 0 {
		return 
	}

	//split lines manually from raw byte buffer
	lines := bytes.Split(buf[:n], []byte("\n"))
	for i := 0; i < 5 && i < len(lines); i++ {
		if len (lines[i]) > 0 {
			fmt.Printf(" %s\n", string(lines[i]))
		}
	}
}