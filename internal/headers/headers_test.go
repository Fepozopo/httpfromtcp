package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// NewHeaders creates and returns a new instance of Headers.
func NewHeaders() Headers {
	return make(Headers)
}

func TestParseAll(t *testing.T) {
	// Test: Valid single header
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, err := headers.ParseAll(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["Host"])
	assert.Equal(t, 25, n)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, err = headers.ParseAll(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)

	// Test: Valid single header with extra whitespace
	headers = NewHeaders()
	data = []byte("         Host:       localhost:42069         \r\n\r\n")
	n, err = headers.ParseAll(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["Host"])
	assert.Equal(t, 49, n)

	// Test: Valid 2 headers with existing headers
	headers = NewHeaders()
	data = []byte("Host: localhost:42069\r\nUser-Agent: curl/7.81.0\r\n\r\n")
	n, err = headers.ParseAll(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["Host"])
	assert.Equal(t, "curl/7.81.0", headers["User-Agent"])
	assert.Equal(t, 50, n)

	// Test: Valid done
	headers = NewHeaders()
	data = []byte("\r\n")
	n, err = headers.ParseAll(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, 2, n)
}
