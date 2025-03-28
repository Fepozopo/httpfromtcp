package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/Fepozopo/httpfromtcp/internal/headers"
)

// Request represents a parsed HTTP request.
type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	state       requestState
	Body        []byte
}

// RequestLine contains details parsed from the start-line of the HTTP request.
type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

// requestState represents different stages in processing a request.
type requestState int

const (
	requestStateInitialized requestState = iota
	requestStateParsingHeaders
	requestStateParsingBody
	requestStateDone
)

const (
	crlf       = "\r\n"
	bufferSize = 8
)

// RequestFromReader reads data from the provided io.Reader, parses it as an HTTP request,
// and returns a pointer to the Request structure.
func RequestFromReader(reader io.Reader) (*Request, error) {
	// Create an initial buffer for reading data.
	buf := make([]byte, bufferSize)
	readToIndex := 0

	// Initialize the Request structure with the initial state.
	req := &Request{
		state:   requestStateInitialized,
		Headers: headers.NewHeaders(),
	}

	// Loop until the whole HTTP request is parsed (state becomes requestStateDone).
	for req.state != requestStateDone {
		// If our buffer is full, double its size to accommodate more data.
		if readToIndex >= len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		// Read data into the buffer starting at the current index.
		numBytesRead, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				// If we get an EOF and the request is still incomplete we return an error.
				if req.state != requestStateDone {
					return nil, fmt.Errorf("incomplete request, in state: %d, read n bytes on EOF: %d", req.state, numBytesRead)
				}
				break
			}
			// Return any other error encountered during reading.
			return nil, err
		}
		// Increase index by the number of newly read bytes.
		readToIndex += numBytesRead

		// Parse the data currently in the buffer.
		numBytesParsed, err := req.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}

		// Shift any unparsed data to the beginning of the buffer for the next iteration.
		copy(buf, buf[numBytesParsed:])
		readToIndex -= numBytesParsed
	}
	return req, nil
}

// parseRequestLine searches for the CRLF indicating end of the request-line,
// then parses and returns the RequestLine object.
func parseRequestLine(data []byte) (*RequestLine, int, error) {
	// Find the position of CRLF which indicates the end of the request-line.
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		// CRLF not found, meaning the request-line is not complete yet.
		return nil, 0, nil
	}

	// Convert the request-line to a string.
	requestLineText := string(data[:idx])
	// Parse the request-line into its parts.
	requestLine, err := requestLineFromString(requestLineText)
	if err != nil {
		return nil, 0, err
	}

	// Return the parsed RequestLine and the total number of bytes consumed (including CRLF).
	return requestLine, idx + 2, nil
}

// requestLineFromString splits the request-line string and validates its format.
func requestLineFromString(str string) (*RequestLine, error) {
	parts := strings.Split(str, " ")
	if len(parts) != 3 {
		return nil, fmt.Errorf("poorly formatted request-line: %s", str)
	}

	// Validate that the HTTP method is uppercase.
	method := parts[0]
	for _, c := range method {
		if c < 'A' || c > 'Z' {
			return nil, fmt.Errorf("invalid method: %s", method)
		}
	}

	requestTarget := parts[1]

	// Split the HTTP version (it should be in the form "HTTP/1.1").
	versionParts := strings.Split(parts[2], "/")
	if len(versionParts) != 2 {
		return nil, fmt.Errorf("malformed start-line: %s", str)
	}

	httpPart := versionParts[0]
	if httpPart != "HTTP" {
		return nil, fmt.Errorf("unrecognized HTTP-version: %s", httpPart)
	}
	version := versionParts[1]
	if version != "1.1" {
		return nil, fmt.Errorf("unrecognized HTTP-version: %s", version)
	}

	// Return the constructed RequestLine structure.
	return &RequestLine{
		Method:        method,
		RequestTarget: requestTarget,
		HttpVersion:   versionParts[1],
	}, nil
}

// parse iteratively calls parseSingle until no more bytes can be parsed in the current state.
func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0
	// Continue parsing data until legacy protocol state is done.
	for r.state != requestStateDone {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return 0, err
		}
		totalBytesParsed += n
		// If no progress was made, it means we need more data.
		if n == 0 {
			break
		}
	}
	return totalBytesParsed, nil
}

// parseSingle parses a single section of the request based on the current state.
func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.state {
	case requestStateInitialized:
		// When state is initialized, parse the request-line.
		requestLine, n, err := parseRequestLine(data)
		if err != nil {
			// Return an error if one occurred during parsing.
			return 0, err
		}
		if n == 0 {
			// Need more data since we haven't received the full request-line.
			return 0, nil
		}
		// Save the parsed request-line and move to header parsing.
		r.RequestLine = *requestLine
		r.state = requestStateParsingHeaders
		return n, nil

	case requestStateParsingHeaders:
		// Parse headers using the helper from headers package.
		n, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}
		// When done parsing all headers, update the state.
		if done {
			r.state = requestStateParsingBody
		}
		return n, nil

	case requestStateParsingBody:
		// If there is no Content-Length header, we're done.
		if _, ok := r.Headers["content-length"]; !ok {
			r.state = requestStateDone
			return len(data), nil
		}
		// Append all the data to the requests .Body field.
		r.Body = append(r.Body, data...)
		// If the length of the body is greater than the Content-Length header, return an error.
		contentLength, err := strconv.Atoi(r.Headers["content-length"])
		if err != nil {
			return 0, fmt.Errorf("invalid content-length header: %w", err)
		}
		if len(r.Body) > contentLength {
			return 0, fmt.Errorf("error: body length greater than Content-Length")
		}
		// If the length of the body is equal to the Content-Length header, move to the done state.
		if len(r.Body) == contentLength {
			r.state = requestStateDone
		}
		// Report that you've consumed the entire length of the data you were given.
		return len(data), nil

	case requestStateDone:
		// If parsing is already complete, any additional data is unexpected.
		return 0, fmt.Errorf("error: trying to read data in a done state")

	default:
		// Return error if the state is unknown.
		return 0, fmt.Errorf("unknown state")
	}
}
