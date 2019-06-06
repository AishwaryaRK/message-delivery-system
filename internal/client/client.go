package client

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"net"
)

type Client struct {
	connection net.Conn
}

func New() *Client {
	return &Client{connection: nil}
}

func (client *Client) Connect(serverAddr *net.TCPAddr) error {
	connection, err := net.Dial("tcp", serverAddr.String())
	if err != nil {
		fmt.Errorf(err.Error())
		return err
	}

	client.connection = connection
	return nil
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
	_, err = client.connection.Read(userIDBuffer)
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
	_, err = client.connection.Read(userIDsLengthBuffer)
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
	_, err = client.connection.Read(userIDsBuffer)
	if err != nil {
		fmt.Errorf("Error in `relay` reading receivers list: %s", err.Error())
		return userIDs, err
	}

	var buffer bytes.Buffer
	gobBuffer := gob.NewDecoder(&buffer)
	err = gobBuffer.Decode(userIDs)
	if err != nil {
		fmt.Errorf("Error in `who_is_here` response decoding userIDs: %s", err.Error())
		return userIDs, err
	}

	return userIDs, nil
}
