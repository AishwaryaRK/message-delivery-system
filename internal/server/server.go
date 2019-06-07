package server

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"github.com/hashicorp/go-multierror"
	"net"
	"sync"
	"unity/message-delivery-system/internal/utility"
)

var MESSAGE_TYPES = map[string]func(server *Server, c net.Conn){
	"who_am_i":    handleWhoAmIRequest,
	"who_is_here": handleListClientIDsRequest,
	"relay":       handleRelayRequest,
}

type Server struct {
	listener    net.Listener
	connections map[uint64]net.Conn
	mutex       sync.Mutex
}

func New() *Server {
	return &Server{listener: nil, connections: make(map[uint64]net.Conn), mutex: sync.Mutex{}}
}

func (server *Server) Start(laddr *net.TCPAddr) error {
	listener, err := net.Listen("tcp", laddr.String())
	if err != nil {
		fmt.Errorf(err.Error())
		return err
	}

	server.listener = listener

	go func() {
		for {
			connection, err := server.listener.Accept()
			if err != nil {
				fmt.Errorf("Error accepting a client connection: %s", err.Error())
				continue
			}

			server.mutex.Lock()
			userID := utility.GenerateID()
			server.connections[userID] = connection

			fmt.Printf("Start handling client connection with userID: %d\n", userID)
			go server.handleConnection(connection)
			server.mutex.Unlock()
		}
	}()

	return nil
}

func (server *Server) Stop() error {
	var allErrors *multierror.Error

	for userID, connection := range server.connections {
		err := connection.Close()
		if err != nil {
			fmt.Errorf("Error closing connection for client with user_id %d: %s", userID, err.Error())
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
		if request, ok := MESSAGE_TYPES[messageType]; ok {
			request(server, connection)
		} else {
			fmt.Errorf("Incorrect message type: %s", messageType)
			continue
		}
	}
}

var handleWhoAmIRequest = func(server *Server, connection net.Conn) {
	for userID, conn := range server.connections {
		if conn == connection {
			userIDBytes := make([]byte, 8)
			binary.LittleEndian.PutUint64(userIDBytes, userID)
			_, err := connection.Write(userIDBytes)
			if err != nil {
				fmt.Errorf("Error sending `who_am_i` response to client with user_id %d: %s", userID, err.Error())
				return
			}
		}
	}
}

var handleListClientIDsRequest = func(server *Server, connection net.Conn) {
	var userIDs []uint64

	for userID, conn := range server.connections {
		if conn != connection {
			userIDs = append(userIDs, userID)
		}
	}

	var buffer bytes.Buffer
	gobBuffer := gob.NewEncoder(&buffer)
	err := gobBuffer.Encode(userIDs)
	if err != nil {
		fmt.Errorf("Error sending `who_is_here` response to client: %s", err.Error())
		return
	}

	_, err = connection.Write([]byte{byte(len(buffer.Bytes()))})
	if err != nil {
		fmt.Errorf("Error sending `who_is_here` response to client: %s", err.Error())
		return
	}

	_, err = connection.Write(buffer.Bytes())
	if err != nil {
		fmt.Errorf("Error sending `who_is_here` response to client: %s", err.Error())
		return
	}
}

var handleRelayRequest = func(server *Server, connection net.Conn) {
	receiverListLengthBuffer := make([]byte, 1)
	_, err := connection.Read(receiverListLengthBuffer)
	if err != nil {
		fmt.Errorf("Error in `relay` reading receiver list length: %s", err.Error())
		return
	}

	receiverListLength, err := binary.ReadUvarint(bytes.NewBuffer(receiverListLengthBuffer))
	if err != nil {
		fmt.Errorf("Error in `relay` incorrect receiver list length: %s", err.Error())
		return
	}
	receiversBuffer := make([]byte, receiverListLength)
	_, err = connection.Read(receiversBuffer)
	if err != nil {
		fmt.Errorf("Error in `relay` reading receivers list: %s", err.Error())
		return
	}

	var receivers []uint64
	gobBuffer := gob.NewDecoder(bytes.NewBuffer(receiversBuffer))
	err = gobBuffer.Decode(&receivers)
	if err != nil {
		fmt.Errorf("Error in `relay` decoding receivers: %s", err.Error())
		return
	}

	messageLengthBuffer := make([]byte, 4)
	_, err = connection.Read(messageLengthBuffer)
	if err != nil {
		fmt.Errorf("Error in `relay` reading message length: %s", err.Error())
		return
	}
	messageLength, err := binary.ReadUvarint(bytes.NewBuffer(messageLengthBuffer))
	if err != nil {
		fmt.Errorf("Error in `relay` incorrect message length: %s", err.Error())
		return
	}
	messageBuffer := make([]byte, messageLength)
	_, err = connection.Read(messageBuffer)
	if err != nil {
		fmt.Errorf("Error in `relay` reading message: %s", err.Error())
		return
	}

	var senderID uint64
	for userID, conn := range server.connections {
		if conn == connection {
			senderID = userID
			break
		}
	}

	for _, receiver := range receivers {
		if conn, ok := server.connections[receiver]; ok {
			senderIDBytes := make([]byte, 8)
			binary.LittleEndian.PutUint64(senderIDBytes, senderID)
			_, err := conn.Write(senderIDBytes)
			if err != nil {
				fmt.Errorf("Error relaying message to receiver %d: %s", receiver, err.Error())
				return
			}

			_, err = conn.Write(messageLengthBuffer)
			if err != nil {
				fmt.Errorf("Error relaying message to receiver %d: %s", receiver, err.Error())
				return
			}

			_, err = conn.Write(messageBuffer)
			if err != nil {
				fmt.Errorf("Error relaying message to receiver %d: %s", receiver, err.Error())
				return
			}
		}
	}
}
