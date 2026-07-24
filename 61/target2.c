#include <stdio.h>
#include <unistd.h>

int main(){
    printf("[+] Target running. PID: %d\n", getpid());
    while (1){
        puts("whoami");
        sleep(2);
    }
    return 0;
}