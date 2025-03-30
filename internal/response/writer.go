package response

import (
	"fmt"
	"io"

	"github.com/Fepozopo/httpfromtcp/internal/headers"
)

type writerState int

const (
	writerStateStatusLine writerState = iota
	writerStateHeaders
	writerStateBody
)

type Writer struct {
	writerState writerState
	writer      io.Writer
}

// NewWriter creates a new Writer for the given io.Writer, starting in the
// writerStateStatusLine state.
//
// A Writer is an object that is used to write an HTTP response to an
// io.Writer. It keeps track of the current state of the response (status line,
// headers, or body) and ensures that the response is written in the correct order.
//
// The Writer is created in the writerStateStatusLine state, which means that the
// first call to WriteStatusLine will write the status line of the response to the
// underlying io.Writer. After that, the Writer transitions to the writerStateHeaders
// state, which means that the next call to WriteHeaders will write the headers of
// the response. Finally, the Writer transitions to the writerStateBody state,
// which means that all subsequent calls to WriteBody will write the body of the
// response.
func NewWriter(w io.Writer) *Writer {
	return &Writer{
		writerState: writerStateStatusLine,
		writer:      w,
	}
}

// WriteStatusLine writes the status line of the HTTP response to the Writer.
// It must be called only once, and only when the Writer is in the
// writerStateStatusLine state. If the Writer is in any other state, WriteStatusLine
// will return an error.
//
// The status line is written using the provided StatusCode, which must be one of
// the StatusCode constants defined in this package.
//
// After writing the status line, the Writer transitions to the writerStateHeaders
// state, so that the next call to WriteHeaders will write the headers of the
// response.
func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.writerState != writerStateStatusLine {
		return fmt.Errorf("cannot write status line in state %d", w.writerState)
	}
	// We defer this function call so that it will be called after the write to
	// the Writer has completed.
	defer func() { w.writerState = writerStateHeaders }()
	_, err := w.writer.Write(getStatusLine(statusCode))
	return err
}

// WriteHeaders writes the headers of the HTTP response to the Writer.
// It must be called only once, and only when the Writer is in the
// writerStateHeaders state. If the Writer is in any other state,
// WriteHeaders will return an error.
//
// The headers are written using the provided Headers map, which should
// contain all of the headers desired for the response.
//
// The headers are written in the following format:
//   - The key-value pairs are written in the format "key: value\r\n"
//   - The final header is followed by a blank line ("\r\n") to
//     indicate the end of the headers.
//
// After writing the headers, the Writer transitions to the writerStateBody
// state, so that the next call to WriteBody will write the body of the
// response.
func (w *Writer) WriteHeaders(h headers.Headers) error {
	if w.writerState != writerStateHeaders {
		return fmt.Errorf("cannot write headers in state %d", w.writerState)
	}
	// We defer this function call so that it will be called after the write to
	// the Writer has completed.
	defer func() { w.writerState = writerStateBody }()
	for k, v := range h {
		// Write each header in the format "key: value\r\n"
		_, err := w.writer.Write([]byte(fmt.Sprintf("%s: %s\r\n", k, v)))
		if err != nil {
			return err
		}
	}
	// Write a blank line to indicate the end of the headers
	_, err := w.writer.Write([]byte("\r\n"))
	return err
}

// WriteBody writes the body of the HTTP response to the Writer.
// It must be called only when the Writer is in the writerStateBody
// state. If the Writer is in any other state, WriteBody will return
// an error.
//
// The body is written directly to the Writer, and the number of bytes
// written is returned.
func (w *Writer) WriteBody(p []byte) (int, error) {
	// If the Writer is not in the writerStateBody state, we cannot write the body.
	if w.writerState != writerStateBody {
		return 0, fmt.Errorf("cannot write body in state %d", w.writerState)
	}
	// Write the body to the Writer and return the number of bytes written.
	return w.writer.Write(p)
}
