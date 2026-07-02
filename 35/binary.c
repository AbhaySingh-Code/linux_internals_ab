#include <stdio.h>
#include <unistd.h>

char secret_message[] = "PEEKABOO";

int main() {
    //Print our own process ID (PID)
    printf("[+] Target Process Running: PID %d\n", getpid());

    //Print the exact memory address of our string variable
    printf("[*] Secret message string value : \"%s\"\n", secret_message);
    printf("[*] Target Memory address to read: %p\n",(void*)&secret_message);
    printf("[*] Entering endless loop. Attach your debugger now ....\n");

    // Loop infinitely, sleeping for 2 seconds each iteration
    while (1){
        sleep(2);
    }
    return 0;
}