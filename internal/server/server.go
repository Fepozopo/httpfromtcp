package server

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"sync/atomic"

	"github.com/Fepozopo/httpfromtcp/internal/request"
	"github.com/Fepozopo/httpfromtcp/internal/response"
)

type Server struct {
	listener net.Listener
	closed   atomic.Bool
	handler  Handler
}

type Handler func(w io.Writer, req *request.Request) *HandlerError

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

// Serve creates a new Server and starts it listening on the specified port.
// It returns the server and any error encountered while creating the listener.
// The server is started in a goroutine, so it will immediately return.
// The handler is called in a separate goroutine for each connection.
func Serve(port int, handler Handler) (*Server, error) {
	// Create a listener on the specified port.
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	// Create a new Server instance.
	// The handler is the function that will be called for each connection.
	// The listener is the listener we just created.
	// The closed flag is initially set to false.
	s := &Server{
		handler:  handler,
		listener: listener,
	}
	// Start the server listening in a goroutine.
	// This will call the listen method in a separate goroutine.
	go s.listen()
	// Return the server and no error.
	return s, nil
}

// listen is the main loop for the server.
// It continually accepts new connections, and then handles each connection in a goroutine.
// If there's an error accepting a connection, it will log the error and continue.
// If the server is closed, it will return from the function.
func (s *Server) listen() {
	// Loop indefinitely.
	// This will continually accept new connections and handle them in goroutines.
	for {
		// Accept a new connection.
		// This will block until a new connection is ready.
		conn, err := s.listener.Accept()
		if err != nil {
			// If there's an error, check if the server is closed.
			// If it is, return from the function.
			// If not, log the error and continue.
			if s.closed.Load() {
				return
			}
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		// Handle the connection in a new goroutine.
		// This will call the handle method with the accepted connection.
		go s.handle(conn)
	}
}

// Close marks the server as closed and closes the underlying listener.
func (s *Server) Close() error {
	if s.closed.CompareAndSwap(false, true) {
		return nil
	}
	return s.listener.Close()
}

// handle is called once for each new connection.
// It takes a net.Conn, closes it when done, and uses it to read an HTTP request.
// If there's an error reading the request, it will write a 400 Bad Request response
// with the error message as the body.
// If there's no error, it will call the handler with the request and a new bytes.Buffer.
// The handler should write the response to the bytes.Buffer.
// If the handler returns a *HandlerError, it will write the response to the net.Conn.
// If the handler returns nil, it will write a 200 OK response with the bytes.Buffer as the body.
// The response will have a Content-Length header set to the length of the bytes.Buffer,
// and a Content-Type header set to text/plain.
func (s *Server) handle(conn net.Conn) {
	defer conn.Close()                          // Close the connection when we're done with it.
	req, err := request.RequestFromReader(conn) // Read an HTTP request from the connection.
	if err != nil {
		// If there's an error, write a 400 Bad Request response with the error message as the body.
		hErr := &HandlerError{
			StatusCode: response.StatusBadRequest,
			Message:    err.Error(),
		}
		hErr.Write(conn)
		return
	}
	buf := bytes.NewBuffer([]byte{}) // Create a new bytes.Buffer for the handler to write to.
	hErr := s.handler(buf, req)      // Call the handler with the request and the bytes.Buffer.
	if hErr != nil {
		// If the handler returns a *HandlerError, write the response to the net.Conn.
		hErr.Write(conn)
		return
	}
	b := buf.Bytes()                                  // Get the bytes from the bytes.Buffer.
	response.WriteStatusLine(conn, response.StatusOK) // Write a 200 OK response.
	headers := response.GetDefaultHeaders(len(b))     // Get the default headers for a response with the given body.
	response.WriteHeaders(conn, headers)              // Write the headers to the connection.
	conn.Write(b)                                     // Write the body to the connection.
	return
}

func (hErr *HandlerError) Write(w io.Writer) {
	// Convert the error message into bytes to get its length
	body := []byte(hErr.Message)
	// Write the status line
	response.WriteStatusLine(w, hErr.StatusCode)
	// Write the headers
	headers := response.GetDefaultHeaders(len(body))
	response.WriteHeaders(w, headers)
	// Write the body
	w.Write(body)
}
