#include <stdio.h>
#include <unistd.h>

char secret_message[] = "PEEKABOO";

int main(){

    //Print your own process id 
    printf("[+] Target process pid: %d", getpid());

    //Print the exact memory address of our string variable
    printf("[+] Secret message is : %s\n", secret_message);
    printf("[+] Address of the secret message is: %p\n", (void*)&secret_message);
    printf("[+] Entering endless loop. Press cntl + C to stop. Attach your debugger now \n");

    while (1) {
        volatile int dummy = 0;
        dummy++;
    }
}