package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Fepozopo/httpfromtcp/internal/headers"
	"github.com/Fepozopo/httpfromtcp/internal/request"
	"github.com/Fepozopo/httpfromtcp/internal/response"
	"github.com/Fepozopo/httpfromtcp/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

// handler is the main handler function for our server.
// It takes a Writer and a Request, and writes a response to the client.
// The response depends on the RequestTarget of the request.
func handler(w *response.Writer, req *request.Request) {
	// If the request is for "/yourproblem", we handle it specially with handler400.
	if req.RequestLine.RequestTarget == "/yourproblem" {
		handler400(w, req)
		return
	}

	// If the request is for "/myproblem", we handle it specially with handler500.
	if req.RequestLine.RequestTarget == "/myproblem" {
		handler500(w, req)
		return
	}

	// Check if the request is for the proxy endpoint
	if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/") {
		proxyHandler(w, req)
		return
	}

	// If the request is for any other URL, we handle it with handler200.
	handler200(w, req)
}

// handler400 writes a 400 status line (Bad Request) to the client.
// It also writes a simple HTML body to the client, telling the user that
// their request was bad.
func handler400(w *response.Writer, _ *request.Request) {
	// Write a 400 status line to the client
	w.WriteStatusLine(response.StatusCodeBadRequest)

	// Write a simple HTML body to the client
	body := []byte(`<html>
	<head>
	<title>400 Bad Request</title>
	</head>
	<body>
	<h1>Bad Request</h1>
	<p>Your request honestly kinda sucked.</p>
	</body>
	</html>
	`)

	// Set up the headers for the response
	h := response.GetDefaultHeaders(len(body))
	h.Override("Content-Type", "text/html")

	// Write the headers to the client
	w.WriteHeaders(h)

	// Write the body to the client
	w.WriteBody(body)
}

// handler500 writes a 500 status line (Internal Server Error) to the client.
// It also writes a simple HTML body to the client, indicating that an internal
// server error occurred. The response includes a title and message acknowledging
// the server fault. The headers are set to indicate the content type as HTML.
func handler500(w *response.Writer, _ *request.Request) {
	// Write a 500 status line (Internal Server Error) to the client
	w.WriteStatusLine(response.StatusCodeInternalServerError)

	// Define the HTML body content for the response
	body := []byte(`<html>
	<head>
	<title>500 Internal Server Error</title>
	</head>
	<body>
	<h1>Internal Server Error</h1>
	<p>Okay, you know what? This one is on me.</p>
	</body>
	</html>
	`)

	// Set up the headers for the response
	h := response.GetDefaultHeaders(len(body))
	h.Override("Content-Type", "text/html")

	// Write the headers to the client
	w.WriteHeaders(h)

	// Write the HTML body to the client
	w.WriteBody(body)
}

// handler200 writes a 200 status line (OK) to the client.
// It also writes a simple HTML body to the client, telling the user that
// their request was awesome.
func handler200(w *response.Writer, _ *request.Request) {
	// Write a 200 status line to the client, indicating that everything was
	// good with the request.
	w.WriteStatusLine(response.StatusCodeSuccess)

	// Define the HTML body content for the response
	body := []byte(`<html>
	<head>
	<title>200 OK</title>
	</head>
	<body>
	<h1>Success!</h1>
	<p>Your request was an absolute banger.</p>
	</body>
	</html>
	`)

	// Set up the headers for the response
	h := response.GetDefaultHeaders(len(body))
	h.Override("Content-Type", "text/html")

	// Write the headers to the client
	w.WriteHeaders(h)

	// Write the HTML body to the client
	w.WriteBody(body)
}

// proxyHandler handles requests to /httpbin/* by proxying to httpbin.org
func proxyHandler(w *response.Writer, req *request.Request) {
	// Extract the path after /httpbin/
	path := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin/")

	// Make the request to httpbin.org
	resp, err := http.Get("https://httpbin.org/" + path)
	if err != nil {
		// If there's an error, return a 500 status
		w.WriteStatusLine(response.StatusCodeInternalServerError)

		body := []byte(fmt.Sprintf("Internal Server Error: %v", err))
		w.WriteHeaders(response.GetDefaultHeaders(len(body)))
		w.WriteBody(body)
		return
	}
	defer resp.Body.Close()

	// Set status code from the response
	w.WriteStatusLine(response.StatusCode(resp.StatusCode))

	// Read the headers from the response
	headers := headers.NewHeaders()
	for key, value := range resp.Header {
		for _, v := range value {
			headers.Set(key, v)
		}
	}

	// Remove Content-Length and set Transfer-Encoding
	headers.Override("Content-Length", "0")
	headers.Override("Transfer-Encoding", "chunked")

	// Write the headers to the client
	w.WriteHeaders(headers)

	// Read the response body in chunks and write them immediately
	buffer := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buffer)
		if n == 0 && err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error reading response: %v", err)
			break
		}

		// Write the chunk
		w.WriteChunkedBody(buffer[:n])

		// Log the chunk size
		log.Printf("Read chunk of size %d", n)
	}

	// Signal end of chunked body
	w.WriteChunkedBodyDone()
}
