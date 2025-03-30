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
	writerStateTrailers
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
	defer func() { w.writerState = writerStateHeaders }()

	_, err := w.writer.Write(getStatusLine(statusCode))
	return err
}

// WriteHeaders writes the headers of the HTTP response to the Writer.
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
//
// The body is written directly to the Writer, and the number of bytes
// written is returned.
func (w *Writer) WriteBody(p []byte) (int, error) {
	// If the Writer is not in the writerStateBody state, we cannot write the body.
	if w.writerState != writerStateBody {
		return 0, fmt.Errorf("cannot write body in state %d", w.writerState)
	}
	defer func() { w.writerState = writerStateTrailers }()

	// Write the body to the Writer and return the number of bytes written.
	return w.writer.Write(p)
}

// WriteTrailers writes the trailers of the HTTP response to the Writer.
//
// The trailers are written in the following format:
//   - The key-value pairs are written in the format "key: value\r\n"
//   - The final trailer is followed by a blank line ("\r\n") to
//     indicate the end of the trailers.
func (w *Writer) WriteTrailers(h headers.Headers) error {
	if w.writerState != writerStateTrailers {
		return fmt.Errorf("cannot write trailers in state %d", w.writerState)
	}
	for k, v := range h {
		// Write each trailer in the format "key: value\r\n"
		_, err := w.writer.Write([]byte(fmt.Sprintf("%s: %s\r\n", k, v)))
		if err != nil {
			return err
		}
	}

	// Write a blank line to indicate the end of the trailers
	_, err := w.writer.Write([]byte("\r\n"))
	return err
}

// WriteChunkedBody writes a chunk of the body of the HTTP response to the Writer.
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
//
// The body is written directly to the Writer, and the number of bytes
// written is returned.
func (w *Writer) WriteChunkedBodyDone() (int, error) {
	// If the Writer is not in the writerStateBody state, we cannot write the body.
	if w.writerState != writerStateBody {
		return 0, fmt.Errorf("cannot write body in state %d", w.writerState)
	}
	defer func() { w.writerState = writerStateTrailers }()

	// Write "0\r\n" to indicate the end of the body and the start of the trailers
	_, err := w.writer.Write([]byte("0\r\n"))
	return 3, err
}
