package response

import (
	"fmt"
	"io"

	"github.com/Fepozopo/httpfromtcp/internal/headers"
)

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

// WriteStatusLine writes a status line to the given io.Writer,
// with the provided StatusCode translated to a human-readable
// reason phrase.
//
// The HTTP status line is always in the form "HTTP/1.1 <status-code> <reason-phrase>\r\n"
// and this function takes care of converting the StatusCode enum to the
// proper status code and reason phrase.
func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	var reasonPhrase string
	// The reason phrase is just a human-readable string
	// that's useful for debugging but not used in the protocol
	// itself.
	switch statusCode {
	case StatusOK:
		reasonPhrase = "200 OK"
	case StatusBadRequest:
		reasonPhrase = "400 Bad Request"
	case StatusInternalServerError:
		reasonPhrase = "500 Internal Server Error"
	default:
		// If the StatusCode is invalid, just use the number
		// for the reason phrase.
		reasonPhrase = fmt.Sprintf("%d", statusCode)
	}
	// Write the status line to the writer.
	_, err := fmt.Fprintf(w, "HTTP/1.1 %s\r\n", reasonPhrase)
	return err
}

// GetDefaultHeaders creates and returns a default headers map, suitable for
// use when responding to a request with a static body.
//
// The returned headers map will contain the following key-value pairs:
//
//   - Content-Length: The length of the body in bytes.
//   - Connection: close. This tells the client that the connection is going
//     to be closed after the response is sent.
//   - Content-Type: text/plain. This specifies the type of content in the
//     response.
func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()
	h.Set("Content-Length", fmt.Sprintf("%d", contentLen))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")
	return h
}

// WriteHeaders takes a writer and a headers map and writes the headers to the writer.
//
// In HTTP, headers are sent as key-value pairs, separated by a colon and a space.
// Each key-value pair is sent on a new line.
// The headers are terminated by a blank line.
func WriteHeaders(w io.Writer, hdrs headers.Headers) error {
	// Iterate over the headers map and write each key-value pair to the writer.
	for key, value := range hdrs {
		// Write the key, a colon, a space, the value, and a newline character.
		_, err := fmt.Fprintf(w, "%s: %s\r\n", key, value)
		if err != nil {
			// If there's an error, return it.
			return err
		}
	}
	// Write a blank line to terminate the headers.
	_, err := io.WriteString(w, "\r\n")
	return err
}
