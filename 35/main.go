package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"syscall"
	"strings"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatalf("Usage: %s <pid> <memory_address_hex>\nExample: %s 123 0x401000", os.Args[0], os.Args[0])
	}

	//Parse command line arguments
	pid, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatalf("Invalid PID: %v", err)
	}

	//Parse target address (e.g., 0x7ffacga4123412)
	addrStr := strings.TrimPrefix(os.Args[2], "0x")
	addr , err := strconv.ParseUint(addrStr, 16, 64)
	if err != nil {
		log.Fatalf("Invalid hex memory address: %v", err)
	}

	//Critical: Lock this goroutine to its current operating system thread 
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	fmt.Printf("[+] Attaching to process %d .... \n", pid)

	//1. Syscall: PTRACE_ATTACH
	// This sends a SIGSTOP to the target process.
	err = syscall.PtraceAttach(pid)
	if err != nil {
		log.Fatalf("PtraceAttach failed: %v (Are you running as root)", err)
	}

	// We must wait for the tracee to actually change status to stop
	var status syscall.WaitStatus
	wpid, err := syscall.Wait4(pid, &status, 0, nil)
	if err != nil {
		log.Fatalf("Wait4 failed: %v", err)
	}
	fmt.Printf("[+] Target process %d successfully stopped and intercepted\n", wpid)

	//2. Syscall: PTRACE_PEEKDATA
	// Allocates a 8-byte buffer to hold a single 64-bit word read from the target address space
	outBuffer := make([]byte, 8)

	// PtracePeekData reads a single word from the tracee's virtual memory address space
	n, err := syscall.PtracePeekData(pid, uintptr(addr), outBuffer)
	if err != nil {
		fmt.Printf("[-] PtracePeekData failed at address 0x%x: %v\n",addr, err)
	} else {
		fmt.Printf("[+] Read %d bytes from 0x%x\n", n, addr)
		fmt.Printf("[+] Raw memory Hex Data: %x\n", outBuffer)
		fmt.Printf("[+] Raw memory string data (ASCII representation): %q\n", string(outBuffer))
	}

	//3. Syscall. PTRACE_CONT
	// Detach or continue
	// fmt.Println("[+] Resuming target process execution....")
	// err = syscall.PtraceCont(pid, 0)
	// if err != nil {
	// 	log.Fatalf("PtraCont failed: %v", err)
	// }

	//Clean detach to return control fully back to the system
	err = syscall.PtraceDetach(pid)
	if err != nil {
		log.Fatalf("PtraceDetach failed: %v", err)
	}
	fmt.Println("[+] Successfully detached. Process is clean.")
}