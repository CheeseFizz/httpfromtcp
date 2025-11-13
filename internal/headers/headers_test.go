package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeaders(t *testing.T) {
	// Test: Valid single header
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, len(data)-2, n)
	assert.False(t, done)

	// Test: Invalid single header
	headers = NewHeaders()
	data = []byte("H@st: localhost:42069\r\n\r\n")
	_, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.False(t, done)

	// Test: Valid single header with extra white space
	headers = NewHeaders()
	data = []byte("    Host:   localhost:42069   \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, len(data)-2, n)
	assert.False(t, done)

	// Test: Valid 2 headers
	headers = NewHeaders()
	data = []byte("Host: localhost:42069\r\n Content-Type: application/json\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, 2, len(headers))
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, "application/json", headers["content-type"])
	assert.Equal(t, len(data)-2, n)
	assert.False(t, done)

	// Test: Valid multiple values
	headers = NewHeaders()
	data = []byte("Host: localhost:42069\r\n Host: test.com:5555\r\n\r\n")
	_, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069,test.com:5555", headers["host"])
	assert.False(t, done)

	// Test: Valid done
	headers = NewHeaders()
	data = []byte("\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, len(data)-2, n)
	assert.True(t, done)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Invalid no colon
	headers = NewHeaders()
	data = []byte("Host localhost:42069\r\n\r\n")
	_, _, err = headers.Parse(data)
	require.Error(t, err)

	// Test: Invalid missing end of headers with content
	headers = NewHeaders()
	data = []byte("Host: localhost:42069\r\nStarting the body \r\n")
	_, _, err = headers.Parse(data)
	require.Error(t, err)

	// Test: Invalid missing end of headers without content
	headers = NewHeaders()
	data = []byte("Host: localhost:42069\r\n")
	_, done, _ = headers.Parse(data)
	assert.False(t, done)
}
