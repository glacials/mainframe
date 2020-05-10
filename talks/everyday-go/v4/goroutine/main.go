package main

import (
	"fmt"
	"sync"
	"time"
)

// main.go
var wg = sync.WaitGroup{}

func main() {
	fmt.Println("Sending email...")
	wg.Add(1)

	sendEmail()

	fmt.Println("Ready to handle more requests!")
	wg.Wait()
}

func sendEmail() {
	defer wg.Done()

	time.Sleep(3 * time.Second)
	fmt.Println("Email sent!")
}
