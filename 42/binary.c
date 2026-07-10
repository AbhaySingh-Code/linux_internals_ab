#include <stdio.h>
#include <unistd.h>

char secret_message[] = "PEEKABOO";

int main(){
    printf("[+] Current process id is : %d\n", getpid());
    printf("[+] Secret Message is : %s\n",secret_message);
    printf("[+] Address of secret message is : %p\n", (void*)&secret_message);
    printf("[+] Entering endless loop. Watch the message change ..... \n");
    
    while(1){
        printf("[+] Current message is : %s\n", secret_message);
        sleep(5);
    }
}