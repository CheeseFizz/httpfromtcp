package server

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"sync/atomic"

	"github.com/CheeseFizz/httpfromtcp/internal/request"
	"github.com/CheeseFizz/httpfromtcp/internal/response"
)

type ServerState int

const (
	serverStateInitializing ServerState = iota
	serverStateStarted
	serverStateStopped
)

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

func (herr *HandlerError) WriteError(conn net.Conn) error {
	h := response.GetDefaultHeaders(len(herr.Message))
	response.WriteStatusLine(conn, herr.StatusCode)
	response.WriteHeaders(conn, h)
	_, err := fmt.Fprintf(conn, "%s\r\n", herr.Message)
	if err != nil {
		return err
	}
	return nil
}

type Handler func(w io.Writer, req *request.Request) *HandlerError

type Server struct {
	Port     int
	Handler  Handler
	state    ServerState
	listener net.Listener
	closed   atomic.Bool
}

func (s *Server) Close() error {
	s.closed.Store(true)
	s.listener.Close()

	s.state = serverStateStopped

	return nil
}

func (s *Server) handle(conn net.Conn) {

	log.Printf("Connection to %s", conn.RemoteAddr().String())

	req, err := request.RequestFromReader(conn)
	if err != nil {
		log.Println(err)
	}
	log.Printf("%s requested: %s", conn.RemoteAddr().String(), req.RequestLine.RequestTarget)
	buf := bytes.NewBuffer([]byte(""))
	herr := s.Handler(buf, req)
	if (herr.StatusCode < 200) || (herr.StatusCode >= 300) {
		err = herr.WriteError(conn)
		if err != nil {
			log.Println(err)
		}
	} else {
		h := response.GetDefaultHeaders(len(buf.Bytes()))
		err = response.WriteStatusLine(conn, 200)
		if err != nil {
			log.Println(err)
		}
		err = response.WriteHeaders(conn, h)
		if err != nil {
			log.Println(err)
		}
		_, err = fmt.Fprintf(conn, "%s\r\n", buf.String())
		if err != nil {
			log.Println(err)
		}
	}

	err = conn.Close()
	if err != nil {
		log.Println(err)
	}

}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil && !s.closed.Load() {
			log.Println(err)
			continue
		}
		if s.closed.Load() {
			return
		}
		go s.handle(conn)
	}
}

func Serve(port int, handler Handler) (server *Server, err error) {
	server = &Server{
		state:   serverStateInitializing,
		Port:    port,
		Handler: handler,
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return server, err
	}
	server.listener = listener

	go server.listen()

	server.state = serverStateStarted

	return server, nil
}
