package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"syscall"
)

func main() {

	if len(os.Args) < 3 {
		fmt.Printf("Usage: %s <pid> <memory_address>", os.Args[0])
		os.Exit(1)
	}

	// Parse pid
	pid, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatalf("Error parsing pid: %v", err)
	}

	//Parse memory address
	addrStr := strings.TrimPrefix(os.Args[2], "0x")
	addr, err := strconv.ParseUint(addrStr, 16, 64)
	if err != nil {
		log.Fatalf("Failed to get memory adddress error : %v", err)
	}

	fmt.Printf("PID: %d, Memory_Address: 0x%X", pid, addr)

	//Critical lock this goroutine to its current operating system thread
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	fmt.Printf("\n[+] Attaching to process with pid : %d ....... \n", pid)

	//1. Attach ptrace
	err = syscall.PtraceAttach(pid)
	if err != nil {
		log.Fatalf("Failed to attach ptrace: %v", err)
	}

	//We must wait for the process to actually chagne status to stop
	var status syscall.WaitStatus
	wpid, err := syscall.Wait4(pid, &status, 0, nil)
	if err != nil {
		log.Fatalf("Wait failed: %v", err)
	}
	fmt.Printf("[+] Target process successfully stopped. WPID : %d", wpid)

	//PTracePeekData
	outBuffer := make([]byte, 16)
	n, err := syscall.PtracePeekData(pid, uintptr(addr), outBuffer)
	if err != nil {
		fmt.Printf("[-] PtracePeekData failed: %v", err)
	} else {
		fmt.Printf("\n[+] Read %d bytes from 0x%X\n", n, addr)
		fmt.Printf("[+] Raw memory string in ASCII: %q\n", string(outBuffer))
	}

	//PtracePokeData
	newMessage := "INSERTEDPOKEDDATA"
	newData := []byte(newMessage)
	fmt.Printf("[+] Preparing to overwrite memory at 0x%X with %q\n", addr, newData)
	n, err = syscall.PtracePokeData(pid, uintptr(addr), newData)
	if err != nil {
		log.Fatalf("Failed to poke data: %v", err)
	}

	//Peeking data again
	buffer2 := make([]byte, 16)
	n, err = syscall.PtracePeekData(pid, uintptr(addr), buffer2)
	if err != nil {
		log.Fatalf("Failed to peek data : %v", err)
	} else {
		fmt.Printf("[+] New data at address is %q\n", string(buffer2))
	}

	//Clean and close
	err = syscall.PtraceDetach(pid)
	if err != nil {
		log.Fatalf("ptace detach failed: %v", err)
	}
	fmt.Printf("Ptrace detach successeded")
}
