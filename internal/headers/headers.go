package headers

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

type Headers map[string]string

func (h Headers) Parse(data []byte) (n int, done bool, err error) {

	// match invalid field name characters
	namefilter, err := regexp.Compile(`[^A-Za-z0-9\x60!#$%\^&*_\-+\.\|~]`)

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

	for line := range strings.SplitSeq(strdata, "\r\n") {
		if len(line) == 0 {
			continue
		}
		field := strings.SplitN(line, ":", 2)
		if len(field) < 2 {
			return 0, false, fmt.Errorf("no 'key: value' in header")
		}

		rfname := []rune(field[0])
		if unicode.IsSpace(rfname[len(rfname)-1]) {
			return 0, false, fmt.Errorf("bad request: header")
		}
		fname := strings.TrimSpace(field[0])
		fname = strings.ToLower(fname)
		if namefilter.FindStringIndex(fname) != nil {
			return 0, false, fmt.Errorf("bad request: header")
		}

		fvalue := strings.TrimSpace(field[1])

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
