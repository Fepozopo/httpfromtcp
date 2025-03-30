package response

import (
	"fmt"

	"github.com/Fepozopo/httpfromtcp/internal/headers"
)

// GetDefaultHeaders creates and returns a default headers map that can be used
// as the starting point for constructing an HTTP response. The map will contain
// the following default headers:
//   - Content-Length: the length of the content in the response body, which is
//     passed as an argument to this function
//   - Connection: "close", indicating that the connection should be closed
//     after the response is sent
//   - Content-Type: "text/plain", indicating that the response body contains
//     plain text
func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()
	h.Set("Content-Length", fmt.Sprintf("%d", contentLen))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")

	return h
}
