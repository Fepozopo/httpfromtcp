package headers

import (
	"errors"
	"regexp"
	"strings"
)

type Headers map[string]string

// Parse method for Headers
func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	// Convert data to string for easier manipulation
	strData := string(data)

	// Look for the CRLF sequence
	crlfIndex := strings.Index(strData, "\r\n")
	if crlfIndex == -1 {
		// No CRLF found, not enough data
		return 0, false, nil
	}

	// If CRLF is at the start, we've reached the end of headers
	if crlfIndex == 0 {
		return 2, true, nil
	}

	// Extract the header line before the CRLF
	headerLine := strData[:crlfIndex]
	n = crlfIndex + 2 // Number of bytes consumed (including CRLF)

	// Check for spaces before the colon in the original header line
	if strings.Contains(headerLine, " :") {
		return 0, false, errors.New("invalid header format: spaces in key")
	}

	// Split the header line into key and value
	parts := strings.SplitN(headerLine, ":", 2)
	if len(parts) != 2 {
		return 0, false, errors.New("invalid header format")
	}

	// Trim whitespace from key and value
	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])

	// Return an error if the key contains an invalid character
	re := regexp.MustCompile(`^[a-zA-Z0-9!#$%&'*+.\-^_` + "`|~]*$")
	if !re.MatchString(key) {
		return 0, false, errors.New("invalid header key")
	}

	// Always lowercase the key
	key = strings.ToLower(key)

	// Check if the key already exists in the map
	if existingValue, exists := h[key]; exists {
		// Append the new value to the existing value, separated by a comma
		h[key] = existingValue + ", " + value
	} else {
		// Add the key-value pair to the Headers map
		h[key] = value
	}

	// Return the number of bytes consumed, done=false since we haven't hit the end yet
	return n, false, nil
}

// ParseAll method for Headers
func (h Headers) ParseAll(data []byte) (n int, err error) {
	// Parse the headers until the end
	for {
		// Parse the header line
		consumed, done, err := h.Parse(data)
		if err != nil {
			return 0, err
		}

		// Add the number of bytes consumed to the total number
		n += consumed

		// If we've reached the end of headers, return the total number of bytes
		if done {
			return n, nil
		}

		// Remove the parsed data from the buffer
		data = data[consumed:]
	}
}
