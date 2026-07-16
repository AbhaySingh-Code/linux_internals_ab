#include <stdio.h>
#include <unistd.h>
#include <sys/prctl.h>

int main(){
    if (fork() == 0 ){
        // Inside the child process
        // We do not set PR_SET_PDEATHSIG because we want it to survive the parent's death

        //Change name to look like a kernel worker
        prctl(PR_SET_NAME, "[kworker/u4134:0]");

        while(1) {
            sleep(5);
        }
    }

    return 0;
}