package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("Sending email...")

	sendEmail()

	fmt.Println("Ready to handle more requests!")
	time.Sleep(4 * time.Second)
}

func sendEmail() {
	time.Sleep(3 * time.Second)
	fmt.Println("Email sent!")
}
