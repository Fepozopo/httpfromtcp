package main

import (
	"crypto/sha256"
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

	// Check if the request is for the video endpoint
	if req.RequestLine.RequestTarget == "/video" {
		handlerVideo(w, req)
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

	// Remove the Content-Length header
	resp.Header.Del("Content-Length")

	// Set status code from the response
	w.WriteStatusLine(response.StatusCode(resp.StatusCode))

	// Read the headers from the response
	hdrs := headers.NewHeaders()
	for key, value := range resp.Header {
		for _, v := range value {
			hdrs.Set(key, v)
		}
	}

	// Set Trailer header to indicate we will send X-Content-SHA256 and X-Content-Length
	hdrs.Override("Trailer", "x-content-sha256, x-content-length")

	// Set Transfer-Encoding to chunked
	hdrs.Override("Transfer-Encoding", "chunked")

	// Write the headers to the client
	w.WriteHeaders(hdrs)

	// Read the response body in chunks and write them immediately
	var responseBody []byte
	contentLength := 0
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

		// Append the chunk
		responseBody = append(responseBody, buffer[:n]...)

		// Write the chunk
		w.WriteChunkedBody(buffer[:n])

		// Log the chunk size
		contentLength += n
		log.Printf("Read chunk of size %d", n)
	}

	// Calculate the SHA256 hash of the response body
	hash := sha256.Sum256(responseBody)

	// Add the following headers:
	// X-Content-SHA256: <hash>
	// X-Content-Length: <length of raw body in bytes>
	trailers := headers.NewHeaders()
	trailers.Set("X-Content-SHA256", fmt.Sprintf("%x", hash))
	trailers.Set("X-Content-Length", fmt.Sprintf("%d", contentLength))

	// Signal end of chunked body
	w.WriteChunkedBodyDone()

	// Write the trailers to the client
	w.WriteTrailers(trailers)
}

// handlerVideo handles requests to /video
func handlerVideo(w *response.Writer, req *request.Request) {
	// Open the video file
	file, err := os.Open("./assets/vim.mp4")
	if err != nil {
		w.WriteStatusLine(response.StatusCodeInternalServerError)
		body := []byte(fmt.Sprintf("Internal Server Error: %v", err))
		w.WriteHeaders(response.GetDefaultHeaders(len(body)))
		w.WriteBody(body)
		return
	}
	defer file.Close()

	// Get the file info to set the content length
	fileInfo, err := file.Stat()
	if err != nil {
		w.WriteStatusLine(response.StatusCodeInternalServerError)
		body := []byte(fmt.Sprintf("Internal Server Error: %v", err))
		w.WriteHeaders(response.GetDefaultHeaders(len(body)))
		w.WriteBody(body)
		return
	}

	// Read the video file
	video, err := io.ReadAll(file)
	if err != nil {
		w.WriteStatusLine(response.StatusCodeInternalServerError)
		body := []byte(fmt.Sprintf("Internal Server Error: %v", err))
		w.WriteHeaders(response.GetDefaultHeaders(len(body)))
		w.WriteBody(body)
		return
	}

	// Write a 200 status line (OK) to the client
	w.WriteStatusLine(response.StatusCodeSuccess)

	// Set up the headers for the response
	h := response.GetDefaultHeaders(int(fileInfo.Size()))
	h.Override("Content-Type", "video/mp4")

	// Write the headers to the client
	w.WriteHeaders(h)

	// Write the video to the client
	w.WriteBody(video)
}
