// Package main is an example application.
package main

import (
	"fmt"
	"time"
)

func main() {
	//nolint:forbidigo // Example code uses fmt.Println.
	fmt.Println("🚀 Hello from the 'hello-world' example!")
	//nolint:forbidigo // Example code uses fmt.Println.
	fmt.Println("This is a simple test application.")
	time.Sleep(1 * time.Second)
	//nolint:forbidigo // Example code uses fmt.Println.
	fmt.Println("✅ Example finished successfully.")
}
