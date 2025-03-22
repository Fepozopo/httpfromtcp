package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	// Open messages.txt
	file, err := os.Open("messages.txt")
	if err != nil {
		log.Printf("error opening file: %v", err)
		return
	}
	defer file.Close()

	// Create an empty byte slice
	buffer := make([]byte, 8)

	// Read the file 8 bytes at a time and print to stdout
	for {
		n, err := file.Read(buffer)
		if err != nil {
			break
		}
		bufferString := string(buffer[:n])
		fmt.Printf("read: %s\n", bufferString)
	}
}
