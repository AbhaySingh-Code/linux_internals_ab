package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"golang.org/x/sys/unix"
)

func main() {

	if len(os.Args) < 3 {
		fmt.Printf("[Usage] %s <pid> <memory>\n", os.Args[0])
		os.Exit(1)
	}

	//Parse pid and memory address
	pid, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatalf("[-] failed to parse pid error : %v", err)
		os.Exit(1)
	}

	addr64, err := strconv.ParseUint(os.Args[2], 0, 64)
	if err != nil {
		log.Fatalf("[-] Failed to convert address string. Error : %v", err)
		os.Exit(1)
	}

	fmt.Printf("[+] Target PID is : %d and Target Address is : 0x%X\n", pid, addr64)

	newPayload := []byte("GOPHER!!\x00")

	localIov := []unix.Iovec{
		{
			Base: &newPayload[0],
			Len:  uint64(len(newPayload)),
		},
	}

	remoteIov := []unix.RemoteIovec{
		{
			Base: uintptr(addr64),
			Len:  len(newPayload),
		},
	}

	fmt.Println("[+] Executing process_vm_write .... \n")

	bytesWritten, err := unix.ProcessVMWritev(pid, localIov, remoteIov, 0)
	if err != nil {
		log.Fatalf("[-] ProcessVMWritev failed: %v\n", err)
	}

	fmt.Printf("[+] Successfully overwrote %d bytes in the target process\n", bytesWritten)
}
