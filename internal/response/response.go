package response

import (
	"fmt"
	"io"
	"log"

	"github.com/CheeseFizz/httpfromtcp/internal/headers"
)

type StatusCode int

const (
	status200 StatusCode = 200
	status400 StatusCode = 400
	status500 StatusCode = 500
)

func GetDefaultHeaders(contentLen int) headers.Headers {
	new_headers := headers.NewHeaders()
	hbytes := []byte(fmt.Sprintf("Content-Length: %d\r\n", contentLen) +
		"Connection: close\r\n" +
		"Content-Type: text/plain\r\n" +
		"\r\n")
	_, _, err := new_headers.Parse(hbytes)
	if err != nil {
		log.Fatalf("parsing default headers failed: %v\n", err)
	}

	return new_headers
}

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	switch statusCode {
	case status200:
		_, err := w.Write([]byte("HTTP/1.1 200 OK\r\n"))
		if err != nil {
			return err
		}
	case status400:
		_, err := w.Write([]byte("HTTP/1.1 400 Bad Request\r\n"))
		if err != nil {
			return err
		}
	case status500:
		_, err := w.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n"))
		if err != nil {
			return err
		}
	default:
		_, err := fmt.Fprintf(w, "HTTP/1.1 %d \r\n", statusCode)
		if err != nil {
			return err
		}
	}
	return nil
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for key, value := range headers {
		_, err := fmt.Fprintf(w, "%s: %s\r\n", key, value)
		if err != nil {
			return err
		}
	}
	_, err := fmt.Fprint(w, "\r\n")
	if err != nil {
		return err
	}

	return nil
}
