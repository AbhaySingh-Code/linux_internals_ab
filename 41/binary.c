#include <stdio.h>
#include <unistd.h>

char secret_message[] = "PEEKABOO";

int main(){
    printf("[+] Current process pid is : %d\n",getpid());
    printf("[+] Secret message is : %s\n",secret_message);
    printf("[+] Address of secret message is : %p\n",(void*)&secret_message);
    printf("[+] Entering a endless loop attaching your debugger now ..... \n");

    while (1){
        volatile int dummy = 0;
        dummy++; 
    }
}