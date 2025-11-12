package request

import (
	"fmt"
	"io"
	"strings"
	"unicode"

	"github.com/CheeseFizz/httpfromtcp/internal/headers"
)

const bufferSize int = 8

type requestState int

const (
	stateInitializing requestState = iota
	stateDone
)

type Request struct {
	RequestLine RequestLine
	state       requestState
	Headers     headers.Headers
}

func (r *Request) parse(data []byte) (int, error) {

	switch r.state {
	case 0:
		b, rline, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		} else if b == 0 {
			return 0, nil
		} else {
			r.RequestLine = *rline
			r.state = 1
			return b, nil
		}

	case 1:
		return 0, fmt.Errorf("error: trying to read data in 'done' state")

	default:
		return 0, fmt.Errorf("error: unknown state")

	}
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func isUpper(s string) bool {
	for _, r := range s {
		if !unicode.IsUpper(r) && unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

func parseRequestLine(reqbytes []byte) (int, *RequestLine, error) {
	requestLine := RequestLine{}

	reqstr := string(reqbytes)
	rlines := (strings.Split(reqstr, "\r\n"))

	// full request not in yet
	if len(rlines) == 1 {
		return 0, &requestLine, nil
	}

	// split request line into parts for validation
	chunks := strings.Split(rlines[0], " ")

	if len(chunks) > 3 {
		return 0, &requestLine, fmt.Errorf("invalid http request line")
	}

	if !isUpper(chunks[0]) {
		return 0, &requestLine, fmt.Errorf("invalid http method: %s", chunks[0])
	}

	if chunks[2] != "HTTP/1.1" {
		return 0, &requestLine, fmt.Errorf("unsupported http version: %s", chunks[2])
	}

	requestLine.HttpVersion = "1.1"
	requestLine.Method = chunks[0]
	requestLine.RequestTarget = chunks[1]

	return len([]byte(rlines[0])), &requestLine, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	empty := Request{}
	request := Request{
		state: stateInitializing,
	}

	buf := make([]byte, bufferSize)
	readToIndex := 0

	for request.state != stateDone {
		if len(buf) == readToIndex {
			newbuf := make([]byte, 2*len(buf))
			copy(newbuf, buf)
			buf = newbuf
		}

		// get more data from reader
		n, err := reader.Read(buf[readToIndex:])
		if err == io.EOF {
			request.state = stateDone
			break
		}
		readToIndex += n

		// try parsing; if successful remove old data from buffer
		n, err = request.parse(buf)
		if err != nil {
			return &empty, err
		}
		if n > 0 {
			newbuf := make([]byte, len(buf))
			copy(newbuf, buf[n:readToIndex])
			buf = newbuf
			readToIndex -= n
		}

	}

	return &request, nil
}
