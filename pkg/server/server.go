package server

import (
	"bytes"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
)

const (
	HOST = "0.0.0.0"
	PORT = "9999"
)

type HandlerFunc func(conn net.Conn)

type Server struct {
	addr     string
	mu       sync.RWMutex
	handlers map[string]HandlerFunc
}

func NewServer(add string) *Server {
	return &Server{
		addr:     add,
		handlers: make(map[string]HandlerFunc),
	}
}

func (s *Server) Register(path string, handler HandlerFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[path] = handler
}

func (s *Server) Start() error {
	listner, err := net.Listen("tcp", s.addr)
	if err != nil {
		log.Print(err)
		return err
	}

	for {
		conn, err := listner.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			log.Println(closeErr)
		}
	}()

	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err == io.EOF {
		log.Printf("%s", buf[:n])
	}

	data := buf[:n]
	requestLineDelim := []byte{'\r', '\n'}
	requestLineEnd := bytes.Index(data, requestLineDelim)
	if requestLineEnd == -1 {
		log.Print("requestLineEndErr: ", requestLineEnd)
	}

	requestLine := string(data[:requestLineEnd])
	parts := strings.Split(requestLine, " ")
	if len(parts) != 3 {
		log.Print("partsErr: ", parts)
	}

	s.mu.RLock()
	if handler, ok := s.handlers[parts[1]]; ok {
		s.mu.RUnlock()
		handler(conn)
	}
	return
}

func (s *Server) RouteHandler(body string) func(conn net.Conn) {
	return func(conn net.Conn) {
		_, err := conn.Write([]byte(s.Response(body)))
		if err != nil {
			log.Print(err)
		}
	}
}

func (s *Server) Response(body string) string {
	return "HTTP/1.1 200 OK\r\n" +
		"Content-Length: " + strconv.Itoa(len(body)) + "\r\n" +
		"Content-Type: text/html\r\n" +
		"Connection: close\r\n" +
		"\r\n" + body
}
