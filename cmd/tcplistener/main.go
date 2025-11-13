package main

import (
	"fmt"
	"net"

	"github.com/CheeseFizz/httpfromtcp/internal/request"
)

// func getLinesChannel(f io.ReadCloser) <-chan string {

// 	linech := make(chan string)

// 	go func(f io.Reader, lch chan string) {
// 		b := make([]byte, 8)
// 		var line string
// 		for {
// 			i, err := f.Read(b)
// 			if err != nil {
// 				if err == io.EOF {
// 					lch <- line
// 					close(lch)
// 					return
// 				} else {
// 					fmt.Println(err)
// 					return
// 				}
// 			}

// 			multilines := strings.Split(string(b[0:i]), "\n")
// 			line = line + multilines[0]

// 			if len(multilines) > 1 {
// 				for _, l := range multilines[1:] {
// 					lch <- line
// 					line = l
// 				}
// 			}
// 		}
// 	}(f, linech)

// 	return linech
// }

func main() {
	address := ":42069"

	listener, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Printf("Error starting listener on %s: %v\n", address, err)
		return
	}
	defer listener.Close()

	fmt.Printf("Started listening on %s\n", address)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Error on connection: %v\n", err)
		}
		fmt.Printf("Accepted connection from %s\n", conn.RemoteAddr().String())

		//linech := getLinesChannel(conn)
		// for item := range linech {
		// 	fmt.Println(item)
		// }

		request, err := request.RequestFromReader(conn)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}

		fmt.Printf(
			"Request line:\n- Method: %s\n- Target: %s\n- Version: %s\n",
			request.RequestLine.Method,
			request.RequestLine.RequestTarget,
			request.RequestLine.HttpVersion,
		)

		fmt.Println("Headers:")
		for key, val := range request.Headers {
			fmt.Printf("- %s: %s\n", key, val)
		}

		fmt.Printf("Body:\n%s\n", string(request.Body))

		fmt.Printf("Connection from %s closed\n", conn.RemoteAddr().String())
	}
}
