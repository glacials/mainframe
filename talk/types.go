package main

import (
	"io"
	"log"
	"os"
)

const filename = "types.go"

func main() {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}

	io.Copy(os.Stdout, file)
}
