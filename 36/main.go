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

func main(){

	//Check if args are passed correctly
	if len(os.Args) < 3{
		log.Fatalf("Usage: %s <pid> <memory>\nExample: %s 1234 0x441243124", os.Args[0], os.Args[0])
	}

	//Parse command line arguments
	pid, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatalf("Invalid PID: %v", err)
	}

	//Parse target address
	addrStr := strings.TrimPrefix(os.Args[2], "0x")
	addr, err := strconv.ParseUint(addrStr, 16, 64)
	if err != nil {
		log.Fatalf("Invalid memory address : %v", err)
	}

	//Critical: Lock this goroutine to its current operating system thread
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	fmt.Printf("[+] Attaching to process with PID: %d ......\n", pid)

	//1. SYSCALL: PtraceAttach
	// This sends a SIGSTOP to the target process
	err = syscall.PtraceAttach(pid)
	if err != nil {
		log.Fatalf("PtraceAttach failed: %v (Are you running as root)", err)
	}

	//We must wait for the tracee to actually change status to stop
	var status syscall.WaitStatus
	wpid, err := syscall.Wait4(pid, &status, 0, nil)
	if err != nil {
		log.Fatalf("Wait4 failed: %v", err)
	}
	fmt.Printf("[+] Target process %d successfully stopped and intercepted\n", wpid)

	//2. PtracePeekData reads a single word from the tracee's virtual memory address
	outBuffer := make([]byte, 16)
	n, err := syscall.PtracePeekData(pid, uintptr(addr), outBuffer)
	if err != nil {
		fmt.Printf("[-] PtracePeekData failed at address 0x%x: %v\n", addr,err)
	} else {
		fmt.Printf("[+] Read %d bytes from 0x%x\n", n, addr)
		fmt.Printf("[+] Raw memory hex data: %x\n", outBuffer)
		fmt.Printf("[+] Raw memory string data (ASCII Representation): %q\n", string(outBuffer))
	}

	//3. Read register values with syscall.PtraceRegs
	var regs syscall.PtraceRegs
	
	err = syscall.PtraceGetRegs(pid, &regs)
	if err != nil {
		log.Fatalf("failed to get regs : %v", err)
	}
	
	fmt.Printf("\n ----- [CPU Register Inspection] -----\n")
	// RIP points to the memory address of the next CPU instruction to execute
	fmt.Printf("[+] RIP (Instruction pointer) : 0x%x\n", regs.Rip)

	// RSP points to the top of the current stack frame
	fmt.Printf("[+] RSP (Stack Pointer) : 0x%x\n", regs.Rsp)

	//RAX often holds syscall numbers or function return values
	fmt.Printf("[+] RAX (Accumulater Register) : 0x%x\n", regs.Rax)
	fmt.Printf("--------------------------\n")

	//Clean and detach to return control fully back to the system
	err = syscall.PtraceDetach(pid)
	if err != nil {
		fmt.Printf("[+] Failed to detach ptrace : %v", err)
	}
	fmt.Printf("[+] Successfully detached ptrace from pid : %d\n", pid)
}