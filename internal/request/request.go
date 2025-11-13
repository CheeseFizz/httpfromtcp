package request

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"

	"github.com/CheeseFizz/httpfromtcp/internal/headers"
)

const bufferSize int = 8

type requestState int

const (
	requestStateInitializing requestState = iota
	requestStateParsingHeaders
	requestStateParsingBody
	requestStateDone
)

type Request struct {
	RequestLine RequestLine
	state       requestState
	Headers     headers.Headers
	Body        []byte
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.state {
	case requestStateInitializing:
		b, rline, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		} else if b == 0 {
			return 0, nil
		} else {
			r.RequestLine = *rline
			r.state = requestStateParsingHeaders
			return b, nil
		}

	case requestStateParsingHeaders:
		b, done, err := r.Headers.Parse(data)
		if err != nil {
			return b, err
		}
		if done {
			r.state = requestStateParsingBody
		}
		return b, nil

	case requestStateParsingBody:
		content_length_str, ok := r.Headers.Get("Content-Length")
		if !ok || content_length_str == string(rune(0)) {
			// no content-length == no body to process
			r.state = requestStateDone
			return 0, nil
		}

		content_length, err := strconv.Atoi(content_length_str)
		if err != nil {
			return 0, err
		}

		r.Body = bytes.TrimRight(append(r.Body, data...), "\x00")

		if len(r.Body) == content_length {
			r.state = requestStateDone
		}

		if len(r.Body) > content_length {
			return 0, fmt.Errorf("content greater than stated length:\n%v\nlength: %v\ncontent-length: %v", r.Body, len(r.Body), content_length)
		}

		return len(bytes.TrimRight(data, "\x00")), nil

	case requestStateDone:
		return 0, fmt.Errorf("error: trying to read data in 'done' state")

	default:
		return 0, fmt.Errorf("error: unknown state")

	}
}

func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0
	for r.state != requestStateDone {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return 0, err
		}
		if n == 0 {
			break
		}
		totalBytesParsed += n
	}
	return totalBytesParsed, nil
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
	lencrlf := len([]byte("\r\n"))
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

	return len([]byte(rlines[0])) + lencrlf, &requestLine, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	empty := Request{}
	request := Request{
		state: requestStateInitializing,
	}

	request.Headers = headers.NewHeaders()
	request.Body = make([]byte, 0)

	buf := make([]byte, bufferSize)
	readToIndex := 0

	for request.state != requestStateDone {
		if len(buf) == readToIndex {
			newbuf := make([]byte, 2*len(buf))
			copy(newbuf, buf)
			buf = newbuf
		}

		// get more data from reader
		n, err := reader.Read(buf[readToIndex:])
		if err == io.EOF {
			if request.state != requestStateDone {
				return &empty, fmt.Errorf("reached end of reader without reaching end of request")
			}
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
