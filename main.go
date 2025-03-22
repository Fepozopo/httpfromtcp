package main

import (
	"bytes"
	"fmt"
	"io"
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

	lines := getLinesChannel(file)

	// Loop over the returned channel
	for line := range lines {
		fmt.Printf("read: %s", line)
	}
}

func getLinesChannel(f io.ReadCloser) <-chan string {
	// Channel for each line
	ch := make(chan string)

	go func() {
		defer f.Close()

		buffer := make([]byte, 0, 10) // Buffer to accumulate bytes until a newline is found
		readBuffer := make([]byte, 8) // Buffer to read 8 bytes at a time

		// Read the file
		for {
			n, err := f.Read(readBuffer)
			if err != nil {
				if err != io.EOF {
					log.Printf("error reading: %v", err)
				}
				break
			}
			// Append the read bytes to the main buffer
			buffer = append(buffer, readBuffer[:n]...)

			// Process the buffer to find and print complete lines
			for {
				// Find the index of the newLine character
				newLineIndex := bytes.IndexByte(buffer, '\n')
				if newLineIndex == -1 {
					break // No newLine found, continue reading
				}

				// Extract the line and send it to the channel
				line := string(buffer[:newLineIndex+1])
				ch <- line

				// Update the buffer to remove the processed line
				buffer = buffer[newLineIndex+1:]
			}
		}
		close(ch)
	}()
	return ch
}
