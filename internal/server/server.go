package server

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"github.com/Fepozopo/httpfromtcp/internal/request"
	"github.com/Fepozopo/httpfromtcp/internal/response"
)

type Handler func(w *response.Writer, req *request.Request)

// Server is an HTTP 1.1 server
type Server struct {
	handler  Handler
	listener net.Listener
	closed   atomic.Bool
}

// Serve initializes and starts a new HTTP server on the specified port using
// the provided handler function. It returns a pointer to the Server instance
// and any error encountered during the setup.
func Serve(port int, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	// Instantiate a new Server object with the provided handler and the created listener.
	s := &Server{
		handler:  handler,
		listener: listener,
	}

	// Start the server's listener in a new goroutine to handle incoming connections
	// concurrently. This allows the Serve function to return immediately, while the
	// server continues to operate in the background.
	go s.listen()

	return s, nil
}

// Close will shut down the server gracefully. It will close the underlying
// listener so that no new connections can be made, and then wait for all
// existing connections to be closed. This ensures that the server is not
// immediately terminated in the middle of a request, which would cause the
// client to see a connection reset error.
//
// It is safe to call Close on a server that has already been closed.
func (s *Server) Close() error {
	s.closed.Store(true)

	if s.listener != nil {
		return s.listener.Close()
	}

	return nil
}

// listen is the main loop for the server. It runs in a goroutine when the
// server is started. It is responsible for accepting new connections and
// starting a new goroutine to handle each one.
//
// The loop runs indefinitely until the server is closed with the Close
// method. When the server is closed, the "closed" flag is set, and the loop
// will return immediately when a connection error occurs.
func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}

			// If the server is not closed, then the error is unexpected, so
			// we log it and continue.
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		// Once a connection is accepted, we start a new goroutine to handle
		// the connection. This allows the server to handle multiple
		// connections concurrently.
		go s.handle(conn)
	}
}

// handle is the main entry point for handling incoming connections on the
// server. It will read and parse an HTTP request from the connection, and then
// invoke the server's handler with the parsed request and a response writer for
// the connection. If there's an error parsing the request, it will write a 400
// Bad Request response to the connection.
func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	// Create a new response writer for the connection
	w := response.NewWriter(conn)

	// Attempt to read and parse an HTTP request from the connection
	req, err := request.RequestFromReader(conn)
	if err != nil {
		w.WriteStatusLine(response.StatusCodeBadRequest)

		body := []byte(fmt.Sprintf("Error parsing request: %v", err))

		w.WriteHeaders(response.GetDefaultHeaders(len(body)))

		w.WriteBody(body)

		return
	}

	// If the request is successfully parsed, invoke the server's handler
	// with the response writer and the parsed request
	s.handler(w, req)
}
