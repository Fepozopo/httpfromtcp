package request

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
)

const bufferSize = 8

type Request struct {
	RequestLine RequestLine
	state       int
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

const (
	initialized = iota
	done
)

func RequestFromReader(reader io.Reader) (*Request, error) {
	buffer := make([]byte, bufferSize)
	readToIndex := 0
	r := &Request{state: initialized}

	for r.state != done {
		// Grow the buffer if it's full
		if readToIndex == len(buffer) {
			newBuffer := make([]byte, len(buffer)*2)
			copy(newBuffer, buffer)
			buffer = newBuffer
		}

		// Read from the reader
		n, err := reader.Read(buffer[readToIndex:])
		if err != nil {
			if err == io.EOF {
				r.state = done
				break
			}
			return nil, fmt.Errorf("error reading from reader: %v", err)
		}
		readToIndex += n

		// Parse the data
		consumed, err := r.parse(buffer[:readToIndex])
		if err != nil {
			return nil, err
		}

		// Remove the parsed data from the buffer
		if consumed > 0 {
			copy(buffer, buffer[consumed:]) // Shift remaining data to the front
			readToIndex -= consumed         // Update the readToIndex
		}
	}

	if r.state == done {
		return r, nil
	}
	return nil, errors.New("error: request not fully parsed")
}

func (r *Request) parse(data []byte) (int, error) {
	if r.state == initialized {
		consumed, err := r.parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if consumed == 0 {
			return 0, nil // Need more data
		}
		r.state = done
		return consumed, nil
	}
	if r.state == done {
		return 0, errors.New("error: trying to read data in a done state")
	}
	return 0, errors.New("error: unknown state")
}

func (r *Request) parseRequestLine(data []byte) (int, error) {
	// Look for the \r\n sequence
	endIndex := strings.Index(string(data), "\r\n")
	if endIndex == -1 {
		return 0, nil // Need more data
	}

	// Parse the request line
	firstLine := data[:endIndex]
	parts := strings.Split(string(firstLine), " ")
	if len(parts) != 3 {
		return 0, errors.New("error: invalid request line")
	}
	httpParts := strings.Split(parts[2], "/")
	if len(httpParts) != 2 {
		return 0, errors.New("error: invalid http version")
	}

	r.RequestLine = RequestLine{
		HttpVersion:   httpParts[1],
		RequestTarget: parts[1],
		Method:        parts[0],
	}

	// Define the regular expression for uppercase letters and verify the method
	re := regexp.MustCompile(`^[A-Z]+$`)
	if !re.MatchString(r.RequestLine.Method) {
		return 0, errors.New("error: unsupported method")
	}

	// Verify the http version is 1.1
	if r.RequestLine.HttpVersion != "1.1" {
		return 0, errors.New("error: unsupported http version")
	}

	return endIndex + 2, nil // Return the number of bytes consumed (including \r\n)
}
