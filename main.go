package main

import (
	"bytes"
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

	// Buffer to accumulate bytes until a newline is found
	buffer := make([]byte, 0, 10)

	// Read the file 8 bytes at a time
	for {
		// Read 8 bytes into a temporary buffer
		tmpBuffer := make([]byte, 8)
		n, err := file.Read(tmpBuffer)
		if err != nil {
			break
		}
		// Append the read bytes to the main buffer
		buffer = append(buffer, tmpBuffer[:n]...)

		// Process the buffer to find and print complete lines
		for {
			// Find the index of the newLine character
			newLineIndex := bytes.IndexByte(buffer, '\n')
			if newLineIndex == -1 {
				break // No newLine found, continue reading
			}

			// Extract the line and print it
			line := buffer[:newLineIndex+1]
			fmt.Printf("read: %s", line)

			// Update the buffer to remove the processed line
			buffer = buffer[newLineIndex+1:]
		}
	}
}
