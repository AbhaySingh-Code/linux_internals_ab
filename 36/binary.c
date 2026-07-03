#include <stdio.h>
#include <unistd.h>

char secret_message[] = "PEEKABOOSECOND";

int main(){
    
    // Print our own process id
    printf("[+] Target process running pid: %d\n", getpid());

    //Print the exact memory address of our string variable
    printf("[+] Secret message string value : %s\n", secret_message);
    printf("[+] Target memory address to read: %p\n", (void*)&secret_message);
    printf("[+] Entering endless loop. Attach your debugger now ....\n");

    //Loop infinite, sleeping for 2 second for each iteration
    while (1){
        volatile int dummy = 0;
        dummy++;
    }
    return 0;
}