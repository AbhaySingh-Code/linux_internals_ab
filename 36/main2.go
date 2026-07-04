package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"syscall"
)

// findSyscallGadget scans /proc/pid/maps for an executable region
// and searches it for the 2-byte "syscall" opcode (0F 05).
func findSyscallGadget(pid int) (uintptr, error) {
	f, err := os.Open(fmt.Sprintf("/proc/%d/maps", pid))
	if err != nil {
		return 0, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		// Only look at executable regions
		fields := strings.Fields(line)
		if len(fields) < 2 || !strings.Contains(fields[1], "x") {
			continue
		}
		addrRange := strings.Split(fields[0], "-")
		start, _ := strconv.ParseUint(addrRange[0], 16, 64)
		end, _ := strconv.ParseUint(addrRange[1], 16, 64)

		size := end - start
		if size > 0x200000 { // cap scan size for sanity
			size = 0x200000
		}
		buf := make([]byte, size)
		n, err := syscall.PtracePeekData(pid, uintptr(start), buf)
		if err != nil || n == 0 {
			continue
		}
		for i := 0; i < len(buf)-1; i++ {
			if buf[i] == 0x0f && buf[i+1] == 0x05 {
				return uintptr(start) + uintptr(i), nil
			}
		}
	}
	return 0, fmt.Errorf("no syscall gadget found")
}

func doSyscallInjection(pid int, sysno, arg1, arg2, arg3 uint64) (uint64, error) {
	var origRegs syscall.PtraceRegs
	if err := syscall.PtraceGetRegs(pid, &origRegs); err != nil {
		return 0, fmt.Errorf("GetRegs failed: %v", err)
	}

	gadget, err := findSyscallGadget(pid)
	if err != nil {
		return 0, err
	}
	fmt.Printf("[+] Found syscall gadget at 0x%x\n", gadget)

	// Clone regs, redirect RIP, set syscall args
	newRegs := origRegs
	newRegs.Rip = uint64(gadget)
	newRegs.Rax = sysno
	newRegs.Rdi = arg1
	newRegs.Rsi = arg2
	newRegs.Rdx = arg3

	if err := syscall.PtraceSetRegs(pid, &newRegs); err != nil {
		return 0, fmt.Errorf("SetRegs failed: %v", err)
	}

	// Single step exactly one instruction: the syscall itself
	if err := syscall.PtraceSingleStep(pid); err != nil {
		return 0, fmt.Errorf("SingleStep failed: %v", err)
	}
	var ws syscall.WaitStatus
	if _, err := syscall.Wait4(pid, &ws, 0, nil); err != nil {
		return 0, fmt.Errorf("Wait4 failed: %v", err)
	}

	// Read back result (return value lands in RAX)
	var afterRegs syscall.PtraceRegs
	if err := syscall.PtraceGetRegs(pid, &afterRegs); err != nil {
		return 0, fmt.Errorf("GetRegs (post) failed: %v", err)
	}
	result := afterRegs.Rax

	// CRITICAL: restore the original register state exactly
	if err := syscall.PtraceSetRegs(pid, &origRegs); err != nil {
		return 0, fmt.Errorf("restore SetRegs failed: %v", err)
	}

	return result, nil
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <pid>", os.Args[0])
	}
	pid, _ := strconv.Atoi(os.Args[1])

	if err := syscall.PtraceAttach(pid); err != nil {
		log.Fatalf("Attach failed: %v", err)
	}
	var ws syscall.WaitStatus
	syscall.Wait4(pid, &ws, 0, nil)

	// Example: syscall 39 = getpid (no args) — safe, side-effect-free test
	ret, err := doSyscallInjection(pid, 39, 0, 0, 0)
	if err != nil {
		fmt.Printf("[-] Injection failed: %v\n", err)
	} else {
		fmt.Printf("[+] Syscall returned: %d\n", ret)
	}

	syscall.PtraceDetach(pid)
	fmt.Println("[+] Detached, target should resume normally")
}
