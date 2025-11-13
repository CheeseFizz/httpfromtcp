package headers

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

type Headers map[string]string

func (h Headers) Get(key string) (value string, ok bool) {
	lower_key := strings.ToLower(key)
	val, ok := h[lower_key]
	if !ok {
		return "", false
	}

	return val, true
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	// Note: this function will always return done=false for the first run of valid data, even if there are headers
	// 		The only time this won't be the case is when data starts with CRLF

	// match invalid field name characters
	namefilter, err := regexp.Compile(`[^A-Za-z0-9\x60!#$%\^&*_\-+\.\|~]`)
	if err != nil {
		return 0, false, fmt.Errorf("regexp.Compile failed")
	}

	strdata := string(data)
	lencrlf := len([]byte("\r\n"))

	crlfi := strings.Index(strdata, "\r\n")
	switch crlfi {
	case -1:
		return 0, false, nil
	case 0:
		return lencrlf, true, nil
	}

	n = 0

	// changed from using the iterator SplitSeq to support streaming data handlers
	lines := strings.Split(strdata, "\r\n")

	// range over complete header info (all but last item in slice)
	for _, line := range lines[:len(lines)-1] {
		if len(line) == 0 {
			continue
		}

		field := strings.SplitN(line, ":", 2)
		if len(field) < 2 {
			return 0, false, fmt.Errorf("no 'key: value' in header: '%s'", line)
		}

		rfname := []rune(field[0])
		if unicode.IsSpace(rfname[len(rfname)-1]) {
			return 0, false, fmt.Errorf("bad request: header name has trailing space '%s'", field[0])
		}
		fname := strings.TrimSpace(field[0])
		fname = strings.ToLower(fname)
		if namefilter.FindStringIndex(fname) != nil {
			return 0, false, fmt.Errorf("bad request: invalid header name '%s'", fname)
		}

		fvalue := strings.TrimSpace(field[1])
		if len(fvalue) == 0 {
			return 0, false, fmt.Errorf("bad request: missing value for header %s", fname)
		}

		_, ok := h[fname]
		if ok {
			h[fname] = fmt.Sprintf("%s,%s", h[fname], fvalue)
		} else {
			h[fname] = fvalue
		}

		n += len([]byte(line)) + lencrlf
	}

	return n, false, nil
}

func NewHeaders() Headers {
	headers := make(Headers)
	return headers
}
