package main 

import (
	"os/exec"
	"fmt"
	"log"
)

func main(){
	//.1 Prepare the command (This doesn't run it yet)
	cmd := exec.Command("ls", "-l","/tmp")

	//.2 Runt he command and capture the combined stdout asnd stderr
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("OS stderr: %s", string(output))
		log.Fatalf("Command failed: %s", err)
	}

	//3. Print the result
	fmt.Printf("Command Output: ")
	fmt.Println(string(output))
}