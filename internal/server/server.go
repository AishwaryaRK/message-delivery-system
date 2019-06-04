package server

import (
	"../utility"
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/hashicorp/go-multierror"
	"net"
	"sync"
)

const PORT = "9000"

type Server struct {
	listener    net.Listener
	connections map[string]net.Conn
	mutex       sync.Mutex
}

func New() *Server {
	listener, err := net.Listen("tcp4", PORT)
	if err != nil {
		fmt.Errorf(err.Error())
		return nil
	}

	return &Server{listener: listener, connections: make(map[string]net.Conn), mutex: sync.Mutex{}}
}

func (server *Server) Start(laddr *net.TCPAddr) error {
	connection, err := server.listener.Accept()
	if err != nil {
		fmt.Errorf("Error accepting a client connection: %s", err.Error())
		return err
	}

	server.mutex.Lock()
	userID, err := utility.GenerateUUID()
	if err != nil {
		fmt.Errorf("Error generating userID: %s", err.Error())
		server.mutex.Unlock()
		return err
	}
	server.connections[userID] = connection
	server.mutex.Unlock()

	fmt.Println("Start handling client connection with userID: %s", userID)
	go server.handleConnection(connection)
	return nil
}

func (server *Server) handleConnection(connection net.Conn) {
	for {
		messageTypeLengthBuffer := make([]byte, 1)
		_, err := connection.Read(messageTypeLengthBuffer)
		if err != nil {
			fmt.Errorf("Error reading message type length: %s", err.Error())
			continue
		}

		messageLength, err := binary.ReadUvarint(bytes.NewBuffer(messageTypeLengthBuffer))
		if err != nil {
			fmt.Errorf("Incorrect message type length: %s", err.Error())
			continue
		}

		messageTypeBuffer := make([]byte, messageLength)
		_, err = connection.Read(messageTypeBuffer)
		if err != nil {
			fmt.Errorf("Error reading message type: %s", err.Error())
			continue
		}

		messageType := string(messageTypeBuffer)
		print(messageType)
	}
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
