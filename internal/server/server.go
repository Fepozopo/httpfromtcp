package server

import (
	"fmt"
	"net"
	"sync/atomic"
)

type Server struct {
	listener net.Listener
	closed   atomic.Bool
}

func Serve(port int) (*Server, error) {
	// Attempt to create a TCP listener on the specified port.
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	// Initialize a new Server instance with the created listener.
	s := &Server{listener: listener}
	// Start listening for incoming connections in a separate goroutine.
	go s.listen()
	// Return the initialized Server instance.
	return s, nil
}

// Close marks the server as closed and closes the underlying listener.
// If the server is already closed, this method will return nil.
// Otherwise, it will return the error from closing the listener.
func (s *Server) Close() error {
	// Atomically swap the value of the closed field to true.
	// If the value was already true, the method will return nil.
	// Otherwise, it will return the error from closing the listener.
	if s.closed.Swap(true) {
		return nil
	}
	// Close the underlying listener.
	// This will cause the Accept method to return an error,
	// which will cause the goroutine in the listen method to exit.
	return s.listener.Close()
}

// listen will continuously accept incoming connections in a loop
// until the server is closed.
func (s *Server) listen() {
	// Loop until the server is closed.
	for !s.closed.Load() {
		// Attempt to accept an incoming connection.
		conn, err := s.listener.Accept()
		// If the server is closed or an error occurred, continue to the next iteration.
		if s.closed.Load() || err != nil {
			continue
		}
		// Start a new goroutine to handle the connection.
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	// Ensure that the connection is closed when the function exits
	defer conn.Close()

	// Write a simple HTTP response back to the client
	// The response includes the HTTP version, status code, headers, and body
	_, err := conn.Write([]byte("HTTP/1.1 200 OK\r\n" + // HTTP version and status code
		"Content-Type: text/plain\r\n\r\n" + // Header specifying the content type
		"Hello World!")) // Body of the response
	if err != nil {
		fmt.Println(err)
	}
}
