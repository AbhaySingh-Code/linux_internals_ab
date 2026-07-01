package main 

import (
	"os"
	"os/exec"
)

func main(){
	//1. Define the command you want to run 
	cmd := exec.Command("ping","-c","3","8.8.8.8")
	//Redirect the command stdout and stderr to the main process terminal
	//Capitalized = exported public
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	//start and wait 
	_ = cmd.Run()
}