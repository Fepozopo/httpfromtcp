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

// NewWriter creates a new Writer that writes to the provided io.Writer.
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

// WriteChunkedBody writes a chunk of the body of the HTTP response to the Writer.
// It must be called only when the Writer is in the writerStateBody
// state. If the Writer is in any other state, WriteChunkedBody will return
// an error.
//
// The body is written directly to the Writer, and the number of bytes
// written is returned.
func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	// If the Writer is not in the writerStateBody state, we cannot write the body.
	if w.writerState != writerStateBody {
		return 0, fmt.Errorf("cannot write body in state %d", w.writerState)
	}
	// Write the chunk size in hexadecimal, followed by "\r\n", and then the chunk data.
	chunkSize := fmt.Sprintf("%x\r\n", len(p))
	_, err := w.writer.Write([]byte(chunkSize))
	if err != nil {
		return 0, err
	}
	n, err := w.writer.Write(p)
	if err != nil {
		return 0, err
	}
	_, err = w.writer.Write([]byte("\r\n"))
	return n, err
}

// WriteChunkedBodyDone writes the final chunk of the body of the HTTP response to the Writer.
// It must be called only when the Writer is in the writerStateBody
// state. If the Writer is in any other state, WriteChunkedBodyDone will return
// an error.
//
// The body is written directly to the Writer, and the number of bytes
// written is returned.
func (w *Writer) WriteChunkedBodyDone() (int, error) {
	// If the Writer is not in the writerStateBody state, we cannot write the body.
	if w.writerState != writerStateBody {
		return 0, fmt.Errorf("cannot write body in state %d", w.writerState)
	}
	// Write "0\r\n\r\n" to indicate the end of the body.
	_, err := w.writer.Write([]byte("0\r\n\r\n"))
	return 5, err
}
