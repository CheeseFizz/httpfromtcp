package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/CheeseFizz/httpfromtcp/internal/request"
	"github.com/CheeseFizz/httpfromtcp/internal/server"
)

const port = 42069

func MyHandler(w io.Writer, req *request.Request) *server.HandlerError {
	herr := server.HandlerError{}
	switch req.RequestLine.RequestTarget {
	case "/yourproblem":
		herr.StatusCode = 400
		herr.Message = "Your problem is not my problem\n"

	case "/myproblem":
		herr.StatusCode = 500
		herr.Message = "Woopsie, my bad\n"

	default:
		herr.StatusCode = 200
		w.Write([]byte("All good, frfr\n"))
	}
	return &herr
}

func main() {
	server, err := server.Serve(port, MyHandler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
