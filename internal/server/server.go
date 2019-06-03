package server

import (
	"fmt"
	"github.com/hashicorp/go-multierror"
	"net"
)

const PORT = "9000"

type Server struct {
	listener    net.Listener
	connections map[string]net.Conn
}

func New() *Server {
	listener, err := net.Listen("tcp4", PORT)
	if err != nil {
		fmt.Errorf(err.Error())
		return nil
	}

	return &Server{listener: listener, connections: make(map[string]net.Conn)}
}


func (server *Server) Stop() error {
	var allErrors *multierror.Error

	for userID, connection := range server.connections {
		err := connection.Close()
		if err != nil {
			fmt.Errorf("Error closing connection for client with user_id %s: %s", userID, err.Error())
			allErrors = multierror.Append(allErrors, err)
		}
	}

	err := server.listener.Close()
	if err != nil {
		fmt.Errorf("Error closing server: %s", err.Error())
		allErrors = multierror.Append(allErrors, err)
	}

	return allErrors.ErrorOrNil()
}
