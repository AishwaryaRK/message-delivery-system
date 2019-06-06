package client

import (
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


