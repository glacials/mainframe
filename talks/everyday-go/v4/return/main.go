// return/main.go

package main

import (
	"io"
	"log"
	"os"
)

func main() {
	file, err := os.Open("return/sample.txt")
	if err != nil {
		log.Fatal(err)
	}

	io.Copy(os.Stdout, file)
}
