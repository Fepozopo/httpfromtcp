package headers

import (
	"bytes"
	"fmt"
	"strings"
)

const crlf = "\r\n" // Constant for CRLF (Carriage Return + Line Feed) used in HTTP headers

// Headers is a map that stores HTTP header key-value pairs
type Headers map[string]string

// NewHeaders creates and returns a new Headers map
func NewHeaders() Headers {
	return map[string]string{}
}

// Parse processes the provided byte slice to extract headers
// It returns the number of bytes consumed, whether the headers are done, and any error encountered
func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	// Find the index of the first CRLF in the data
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		// No CRLF found, meaning we can't parse any headers yet
		return 0, false, nil
	}
	if idx == 0 {
		// An empty line indicates the end of headers; consume the CRLF
		return 2, true, nil
	}

	// Split the header line into key and value at the first colon
	parts := bytes.SplitN(data[:idx], []byte(":"), 2)
	key := strings.ToLower(string(parts[0])) // Convert the key to lowercase

	// Check for invalid header name (trailing spaces)
	if key != strings.TrimRight(key, " ") {
		return 0, false, fmt.Errorf("invalid header name: %s", key)
	}

	// Trim whitespace from the value and validate the key
	value := bytes.TrimSpace(parts[1])
	key = strings.TrimSpace(key)
	if !validTokens([]byte(key)) {
		// Validate that the key contains only valid token characters
		return 0, false, fmt.Errorf("invalid header token found: %s", key)
	}

	// Set the header in the map
	h.Set(key, string(value))
	return idx + 2, false, nil // Return the number of bytes consumed and indicate that headers are not done
}

// Set adds or updates a header in the Headers map
// If the key already exists, it appends the new value to the existing value
func (h Headers) Set(key, value string) {
	key = strings.ToLower(key) // Ensure the key is lowercase
	v, ok := h[key]            // Check if the key already exists
	if ok {
		// If it exists, join the existing value with the new value
		value = strings.Join([]string{
			v,
			value,
		}, ", ")
	}
	h[key] = value // Set the key-value pair in the map
}

// Get retrieves the value for the given key, keeping case insensitivity in mind
func (h Headers) Get(key string) string {
	key = strings.ToLower(key) // Ensure the key is lowercase
	return h[key]
}

// tokenChars contains valid characters for HTTP header tokens
var tokenChars = []byte{'!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~'}

// validTokens checks if the data contains only valid tokens
// or characters that are allowed in a token
func validTokens(data []byte) bool {
	// Iterate through each character in the data
	for _, c := range data {
		// Check if the character is a letter, number, or one of the allowed characters
		if !(('A' <= c && c <= 'Z') || ('a' <= c && c <= 'z') || ('0' <= c && c <= '9') || bytes.Contains(tokenChars, []byte{c})) {
			return false // Invalid character found
		}
	}
	return true // All characters are valid tokens
}
