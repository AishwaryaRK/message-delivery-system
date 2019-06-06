package client

import (
	"encoding/binary"
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

	userIDBuffer := make([]byte, 8)
	_, err = client.connection.Read(userIDBuffer)
	if err != nil {
		fmt.Errorf("Error reading userID from server: %s", err.Error())
		return userID, err
	}

	userID = binary.LittleEndian.Uint64(userIDBuffer)
	return userID, nil
}
