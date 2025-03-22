package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
)

func main() {
	// Create a listener
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatalf("error creating listener: %v", err)
	}
	defer listener.Close()

	fmt.Println("Server is listening on :42069")

	for {
		// Accept a connection
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("error creating connection: %v", err)
			continue
		}
		fmt.Println("A connection has been accepted")

		// Handle the connection in a new goroutine
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	lines := getLinesChannel(conn)

	// Loop over the returned channel
	for line := range lines {
		fmt.Printf("%s", line)
	}
	fmt.Print("\n") // Print a final terminating newline

	fmt.Println("The connection has been closed")
}

func getLinesChannel(f io.ReadCloser) <-chan string {
	// Channel for each line
	ch := make(chan string)

	go func() {
		defer f.Close()
		defer close(ch)

		buffer := make([]byte, 0, 64) // Buffer to accumulate bytes until a newline is found
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
		// Send any remaining data as the last line
		if len(buffer) > 0 {
			line := string(buffer)
			ch <- line
		}
	}()
	return ch
}
