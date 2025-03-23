package request

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	message, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("error reading request: %v", err)
	}

	// Get the request-line from the message
	lines := strings.Split(string(message), "\r\n")
	firstLine := lines[0]

	// Parse the request-line and map to a RequestLine struct
	parts := strings.Split(firstLine, " ")
	if len(parts) != 3 {
		return nil, errors.New("error: invalid request line")
	}
	httpParts := strings.Split(parts[2], "/")
	if len(httpParts) != 2 {
		return nil, errors.New("error: invalid http version")
	}
	requestLine := RequestLine{
		HttpVersion:   httpParts[1],
		RequestTarget: parts[1],
		Method:        parts[0],
	}

	// Define the regular expression for uppercase letters and verify the method
	re := regexp.MustCompile(`^[A-Z]+$`)
	if !re.MatchString(requestLine.Method) {
		return nil, errors.New("error: unsupported method")
	}

	// Verify the http version is HTTP/1.1
	if requestLine.HttpVersion != "1.1" {
		return nil, errors.New("error: unsupported http version")
	}

	request := new(Request)
	request.RequestLine = requestLine

	return request, nil
}
