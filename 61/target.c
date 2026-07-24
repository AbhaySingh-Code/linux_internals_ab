#include <stdio.h>
#include <unistd.h>

int main(){
    printf("[+] Target running. Process ID: %d\n",getpid());
    while(1) {
        puts("Normal system behaviour");
        sleep(2);
    }
    return 0;
}