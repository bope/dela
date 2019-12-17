package delafs

import (
	"net"
	"net/http"
)

type Server struct {
	listener net.Listener
	Port     int
	path     string
}

func NewServer(path string) (*Server, error) {
	var err error
	server := &Server{
		path: path,
	}

	if server.listener, err = net.Listen("tcp", ":0"); err != nil {
		return nil, err
	}

	server.Port = server.listener.Addr().(*net.TCPAddr).Port
	return server, nil
}

func (s *Server) Start() error {
	fs := FileServer(http.Dir(s.path))
	return http.Serve(s.listener, fs)
}
