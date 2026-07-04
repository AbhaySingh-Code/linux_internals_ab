#include <stdio.h>
#include <unistd.h>

char secret_message[] = "PEEKABOO";

int main(){

    // Get the current process pid and print it 
    printf("[+] PID of the target process is pid : %d", getpid());

    //Print address of string and string
    printf("\n[+] Secret String is : %s, Address is %p\n", secret_message, (void*)&secret_message);
    printf("[+]Entering a endless loop attach your debugger now ..... \n");

    while (1){
        volatile int dummy = 0 ;
        dummy++;
    }
}