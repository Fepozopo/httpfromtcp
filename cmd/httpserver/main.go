package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

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
		// We return here so that we don't write another response to the client.
		return
	}

	// If the request is for "/myproblem", we handle it specially with handler500.
	if req.RequestLine.RequestTarget == "/myproblem" {
		handler500(w, req)
		// We return here so that we don't write another response to the client.
		return
	}

	// If the request is for any other URL, we handle it with handler200.
	// This will write a 200 status line to the client, indicating a successful request.
	// It will also write a simple HTML body to the client, telling the user that
	// their request was successful.
	handler200(w, req)
}

// handler400 is a special handler for "/yourproblem" requests.
// It writes a 400 status line (Bad Request) to the client.
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
	// We use the default headers, but override the Content-Type header
	// to be "text/html" so that the client knows to render the HTML.
	h := response.GetDefaultHeaders(len(body))
	h.Override("Content-Type", "text/html")

	// Write the headers to the client
	w.WriteHeaders(h)

	// Write the body to the client
	w.WriteBody(body)
}

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
	// Get the default headers and override the Content-Type header
	// to "text/html" to notify the client to render the HTML content
	h := response.GetDefaultHeaders(len(body))
	h.Override("Content-Type", "text/html")

	// Write the headers to the client
	w.WriteHeaders(h)

	// Write the HTML body to the client
	w.WriteBody(body)
}

// handler200 is a special handler for successful requests.
// It writes a 200 status line (OK) to the client.
// It also writes a simple HTML body to the client, telling the user that
// their request was awesome.
func handler200(w *response.Writer, _ *request.Request) {
	// Write a 200 status line to the client, indicating that everything was
	// good with the request.
	w.WriteStatusLine(response.StatusCodeSuccess)

	// Define the HTML body content for the response
	// This HTML just says "Success!" and tells the user that their request
	// was good.
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
	// We use the default headers, but override the Content-Type header
	// to be "text/html" so that the client knows to render the HTML.
	h := response.GetDefaultHeaders(len(body))
	h.Override("Content-Type", "text/html")

	// Write the headers to the client
	w.WriteHeaders(h)

	// Write the HTML body to the client
	w.WriteBody(body)
}
