package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strconv"

	unix "golang.org/x/sys/unix"
)

func main() {

	if len(os.Args) < 3 {
		fmt.Printf("Usage: %s <target_pid> <memory_address>\n", os.Args[0])
		fmt.Printf("Example: %s 1234 0x7asfsdf114\n", os.Args[0])
		os.Exit(1)
	}

	//Parse pid and memory address
	pid, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatalf("[-] Invalid pid : %v", pid)
	}

	addr64, err := strconv.ParseUint(os.Args[2], 0, 64)
	if err != nil {
		log.Fatalf("[-] Invalid Address Error : %v", err)
	}

	fmt.Printf("[+] Target PID: %d\n", pid)
	fmt.Printf("[+] Target Memory Address : 0x%X\n", addr64)

	//2. Allocate a local buffer to hold the incoming data
	bufferSize := 64
	localBuffer := make([]byte, bufferSize)

	localIov := []unix.Iovec{
		{
			Base: &localBuffer[0],
			Len:  uint64(bufferSize),
		},
	}

	remoteIov := []unix.RemoteIovec{
		{
			Base: uintptr(addr64),
			Len:  bufferSize,
		},
	}

	fmt.Println("[+] Executing process_vm_readv.....")

	bytesRead, err := unix.ProcessVMReadv(pid, localIov, remoteIov, 0)
	if err != nil {
		log.Fatalf("[-] ProcessVmReadv failed: %v\n Note: You may need root/CAP_SYS_PTRACE privileges.\n", err)
		os.Exit(1)
	}

	fmt.Printf("[+] Successfully read %d bytes\n", bytesRead)
	fmt.Printf("[+]---- [Hex Dump]-----\n")
	fmt.Printf(hex.Dump(localBuffer[:bytesRead]))

	fmt.Println("\n------ [ String Representation] ------ \n")
	fmt.Printf("%s\n", string(localBuffer[:bytesRead]))

}
