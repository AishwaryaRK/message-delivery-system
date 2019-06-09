package client

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"net"
	"sync"
)

type IncomingMessage struct {
	SenderID uint64
	Body     []byte
}

type Client struct {
	connection net.Conn
	mutex      sync.RWMutex
}

func New() *Client {
	return &Client{connection: nil, mutex: sync.RWMutex{}}
}

func (client *Client) Connect(serverAddr *net.TCPAddr) error {
	connection, err := net.Dial("tcp", serverAddr.String())
	if err != nil {
		fmt.Errorf(err.Error())
		return err
	}

	client.mutex.Lock()
	client.connection = connection
	client.mutex.Unlock()

	return nil
}

func (client *Client) Close() error {
	err := client.connection.Close()
	if err != nil {
		fmt.Errorf("Error closing connection for client: %s", err.Error())
	}

	return err
}

func (client *Client) WhoAmI() (uint64, error) {
	var userID uint64
	messageType := "who_am_i"
	messageTypeLength := len(messageType)
	_, err := client.connection.Write([]byte{byte(messageTypeLength)})
	if err != nil {
		fmt.Errorf("Error sending `who_am_i` request to server: %s", err.Error())
		return userID, err
	}

	_, err = client.connection.Write([]byte(messageType))
	if err != nil {
		fmt.Errorf("Error sending `who_am_i` request to server: %s", err.Error())
		return userID, err
	}

	userIDBuffer := make([]byte, 8)
	client.mutex.RLock()
	_, err = client.connection.Read(userIDBuffer)
	client.mutex.RUnlock()
	if err != nil {
		fmt.Errorf("Error reading userID from server: %s", err.Error())
		return userID, err
	}
	userID = binary.LittleEndian.Uint64(userIDBuffer)

	return userID, nil
}

func (client *Client) ListClientIDs() ([]uint64, error) {
	var userIDs []uint64
	messageType := "who_is_here"
	messageTypeLength := len(messageType)

	_, err := client.connection.Write([]byte{byte(messageTypeLength)})
	if err != nil {
		fmt.Errorf("Error sending `who_is_here` request to server: %s", err.Error())
		return userIDs, err
	}

	_, err = client.connection.Write([]byte(messageType))
	if err != nil {
		fmt.Errorf("Error sending `who_is_here` request to server: %s", err.Error())
		return userIDs, err
	}

	userIDsLengthBuffer := make([]byte, 1)
	client.mutex.RLock()
	_, err = client.connection.Read(userIDsLengthBuffer)
	client.mutex.RUnlock()
	if err != nil {
		fmt.Errorf("Error reading `who_is_here` response from server: %s", err.Error())
		return userIDs, err
	}
	userIDsLength, err := binary.ReadUvarint(bytes.NewBuffer(userIDsLengthBuffer))
	if err != nil {
		fmt.Errorf("Incorrect `who_is_here` response from server: %s", err.Error())
		return userIDs, err
	}

	userIDsBuffer := make([]byte, userIDsLength)
	client.mutex.RLock()
	_, err = client.connection.Read(userIDsBuffer)
	client.mutex.RUnlock()
	if err != nil {
		fmt.Errorf("Error in `relay` reading receivers list: %s", err.Error())
		return userIDs, err
	}

	gobBuffer := gob.NewDecoder(bytes.NewBuffer(userIDsBuffer))
	err = gobBuffer.Decode(&userIDs)
	if err != nil {
		fmt.Errorf("Error in `who_is_here` response decoding userIDs: %s", err.Error())
		return userIDs, err
	}

	return userIDs, nil
}

func (client *Client) SendMsg(recipients []uint64, body []byte) error {
	var buffer bytes.Buffer
	gobBuffer := gob.NewEncoder(&buffer)
	err := gobBuffer.Encode(recipients)
	if err != nil {
		fmt.Errorf("Error encoding recipients: %s", err.Error())
		return err
	}

	messageType := "relay"
	messageTypeLength := len(messageType)

	_, err = client.connection.Write([]byte{byte(messageTypeLength)})
	if err != nil {
		fmt.Errorf("Error sending `relay` request to server: %s", err.Error())
		return err
	}

	_, err = client.connection.Write([]byte(messageType))
	if err != nil {
		fmt.Errorf("Error sending `relay` request to server: %s", err.Error())
		return err
	}

	recipientsLength := len(buffer.Bytes())
	_, err = client.connection.Write([]byte{byte(recipientsLength)})
	if err != nil {
		fmt.Errorf("Error sending `relay` request to server: %s", err.Error())
		return err
	}

	_, err = client.connection.Write(buffer.Bytes())
	if err != nil {
		fmt.Errorf("Error sending `who_is_here` response to client: %s", err.Error())
		return err
	}

	var messageLength uint32
	messageLength = uint32(len(body))
	msgLengthBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(msgLengthBytes, messageLength)
	_, err = client.connection.Write(msgLengthBytes)
	if err != nil {
		fmt.Errorf("Error sending `relay` request to server: %s", err.Error())
		return err
	}

	_, err = client.connection.Write(body)
	if err != nil {
		fmt.Errorf("Error sending `relay` request to server: %s", err.Error())
		return err
	}

	return nil
}

func (client *Client) HandleIncomingMessages(writeCh chan<- IncomingMessage) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("recovered from writing to a closed incoming messages channel")
			return
		}
	}()

	for {
		senderIDBuffer := make([]byte, 8)
		client.mutex.RLock()
		_, err := client.connection.Read(senderIDBuffer)
		client.mutex.RUnlock()
		if err != nil {
			fmt.Errorf("Error reading senderID from server: %s", err.Error())
			return
		}
		senderID := binary.LittleEndian.Uint64(senderIDBuffer)

		messageLengthBuffer := make([]byte, 4)
		client.mutex.RLock()
		_, err = client.connection.Read(messageLengthBuffer)
		client.mutex.RUnlock()
		if err != nil {
			fmt.Errorf("Error in `incoming_message` reading message length: %s", err.Error())
			return
		}
		var messageLength uint32
		messageLength = binary.LittleEndian.Uint32(messageLengthBuffer)
		messageBuffer := make([]byte, messageLength)
		client.mutex.RLock()
		_, err = client.connection.Read(messageBuffer)
		client.mutex.RUnlock()
		if err != nil {
			fmt.Errorf("Error in `incoming_message` reading message: %s", err.Error())
			return
		}

		incomingMessage := IncomingMessage{SenderID: senderID, Body: messageBuffer}

		writeCh <- incomingMessage
	}
}
