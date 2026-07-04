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

func findSyscallGadget(pid int) (uintptr, error) {
	f, err := os.Open(fmt.Sprintf("/proc/%d/maps", pid))
	if err != nil {
		return 0, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 || !strings.Contains(fields[1], "x") { // only executable regions
			continue
		}
		addrRange := strings.Split(fields[0], "-")
		start, _ := strconv.ParseUint(addrRange[0], 16, 64)
		end, _ := strconv.ParseUint(addrRange[1], 16, 64)

		size := end - start
		if size > 0x200000 {
			size = 0x200000 // don't scan huge regions, cap it
		}
		buf := make([]byte, size)
		n, err := syscall.PtracePeekData(pid, uintptr(start), buf)
		if err != nil || n == 0 {
			continue
		}
		for i := 0; i < len(buf)-1; i++ {
			if buf[i] == 0x0f && buf[i+1] == 0x05 { // found "syscall" opcode
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
	fmt.Printf("[+] Using syscall gadget at 0x%x\n", gadget)

	newRegs := origRegs
	newRegs.Rip = uint64(gadget)
	newRegs.Rax = sysno
	newRegs.Rdi = arg1
	newRegs.Rsi = arg2
	newRegs.Rdx = arg3

	if err := syscall.PtraceSetRegs(pid, &newRegs); err != nil {
		return 0, fmt.Errorf("SetRegs failed: %v", err)
	}

	if err := syscall.PtraceSingleStep(pid); err != nil {
		return 0, fmt.Errorf("SingleStep failed: %v", err)
	}
	var ws syscall.WaitStatus
	if _, err := syscall.Wait4(pid, &ws, 0, nil); err != nil {
		return 0, fmt.Errorf("Wait4 failed: %v", err)
	}

	var afterRegs syscall.PtraceRegs
	if err := syscall.PtraceGetRegs(pid, &afterRegs); err != nil {
		return 0, fmt.Errorf("GetRegs (post) failed: %v", err)
	}
	result := afterRegs.Rax

	if err := syscall.PtraceSetRegs(pid, &origRegs); err != nil {
		return 0, fmt.Errorf("restore SetRegs failed: %v", err)
	}

	return result, nil
}

func main() {
	if len(os.Args) < 3 {
		log.Fatalf("Usage: %s <pid> <hex_address_of_string>", os.Args[0])
	}
	pid, _ := strconv.Atoi(os.Args[1])
	addrStr := strings.TrimPrefix(os.Args[2], "0x")
	addr, err := strconv.ParseUint(addrStr, 16, 64)
	if err != nil {
		log.Fatalf("Invalid address: %v", err)
	}

	if err := syscall.PtraceAttach(pid); err != nil {
		log.Fatalf("Attach failed: %v", err)
	}
	var ws syscall.WaitStatus
	syscall.Wait4(pid, &ws, 0, nil)
	fmt.Printf("[+] Attached and stopped pid %d\n", pid)

	const SYS_WRITE = 1
	const STDOUT = 1
	const STRLEN = 15

	fmt.Println("[+] Triggering write() syscall inside the TARGET process...")
	n, err := doSyscallInjection(pid, SYS_WRITE, STDOUT, addr, STRLEN)
	if err != nil {
		fmt.Printf("[-] Injection failed: %v\n", err)
	} else {
		fmt.Printf("[+] write() returned: %d (bytes written)\n", n)
		fmt.Println("[+] Check the TARGET's terminal — it should have printed the string just now, triggered by us, not by its own code.")
	}

	if err := syscall.PtraceDetach(pid); err != nil {
		fmt.Printf("[-] Detach failed: %v\n", err)
	} else {
		fmt.Println("[+] Detached, target resumes its own loop as if nothing happened")
	}
}
