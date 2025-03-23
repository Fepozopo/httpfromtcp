package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	// Resolve the remote UDP address and port
	addr, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		log.Fatalf("error trying to resolve UDP address: %v", err)
	}

	// Dial the UDP connection, specifying nil for the local address and the remote address
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatalf("error preparing a UDP connection: %v", err)
	}
	defer conn.Close()

	// Create a new bufio.Reader that reads from standard input
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print(">") // Prompt for input

		// Read a line of input, terminated by a newLine byte
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("error reading input: %v", err)
			continue
		}

		// Send the line to the remote address through the UDP connection
		_, err = conn.Write([]byte(line))
		if err != nil {
			log.Printf("error sending message: %v", err)
		}
	}
}
