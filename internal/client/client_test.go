package client

import (
	"encoding/binary"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"net"
	"testing"
	"time"
)

type ServerTestSuite struct {
	suite.Suite
	client *Client
}

func (s *ServerTestSuite) SetupSuite() {
	s.client = New()
}

func (s *ServerTestSuite) TestWhoAmIRequest() {
	serverPort := 9002
	serverAddr := net.TCPAddr{Port: serverPort}
	listener, err := net.Listen("tcp", serverAddr.String())
	assert.NoError(s.T(), err, "should not return error while creating server")

	expectedUserID := uint64(2134567)

	var connection net.Conn
	go func() {
		for {
			connection, err = listener.Accept()
			assert.NoError(s.T(), err, "should not return error while accepting client connection")
		}
	}()

	time.Sleep(1 * time.Second)

	s.client.Connect(&serverAddr)

	go func() {
		for {
			messageTypeLengthBuffer := make([]byte, 1)
			_, err = connection.Read(messageTypeLengthBuffer)
			assert.NoError(s.T(), err, "should not return error while connecting to server")

			messageTypeBuffer := make([]byte, 8)
			_, err = connection.Read(messageTypeBuffer)
			if err != nil {
				fmt.Errorf("Error reading message type: %s", err.Error())
				continue
			}

			messageType := string(messageTypeBuffer)
			assert.Equal(s.T(), "who_am_i", messageType)

			userIDBytes := make([]byte, 8)
			binary.LittleEndian.PutUint64(userIDBytes, expectedUserID)
			_, err = connection.Write(userIDBytes)
			assert.NoError(s.T(), err, "should not return error while sending userID to client")
		}
	}()

	userID, err := s.client.WhoAmI()
	assert.Equal(s.T(), expectedUserID, userID)
}

func (s *ServerTestSuite) TearDownSuite() {
	require.NoError(s.T(), s.client.Close())
}

func TestServerTestSuite(t *testing.T) {
	suite.Run(t, new(ServerTestSuite))
}
