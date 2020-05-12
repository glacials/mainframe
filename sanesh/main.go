package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

func main() {
	if err := zsh(); err != nil {
		log.Fatalf(err.Error())
	}
}

func zsh() error {
	fmt.Println("sanesh started")

	zsh := exec.Cmd{
		Path:   "/bin/zsh",
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
	if err := zsh.Run(); err != nil {
		return err
	}

	fmt.Println("sanesh exited")
	return nil
}
