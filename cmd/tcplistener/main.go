package main

import (
	"fmt"
	"log"
	"net"

	"github.com/Fepozopo/httpfromtcp/internal/request"
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

	requestLine, err := request.RequestFromReader(conn)
	if err != nil {
		log.Printf("error creating request: %v", err)
		return
	}

	fmt.Printf("Request line:\n- Method: %s\n- Target: %s\n- Version: %s\n", requestLine.RequestLine.Method, requestLine.RequestLine.RequestTarget, requestLine.RequestLine.HttpVersion)
}
