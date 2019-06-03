package server

import (
	"fmt"
	"net"
)

const PORT = "9000"

type Server struct {
	listener net.Listener
}

func New() *Server {
	listener, err := net.Listen("tcp4", PORT)
	if err != nil {
		fmt.Errorf(err.Error())
		return nil
	}

	return &Server{listener: listener}
}



func (server *Server) Stop() error {
	server.listener.Close()
	return nil
}
