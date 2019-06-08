package client

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
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
			var err2 error
			connection, err2 = listener.Accept()
			assert.NoError(s.T(), err2, "should not return error while accepting client connection")

			messageTypeLengthBuffer := make([]byte, 1)
			_, err2 = connection.Read(messageTypeLengthBuffer)
			assert.NoError(s.T(), err2, "should not return error while reading messageType length from client")

			messageTypeBuffer := make([]byte, 8)
			_, err2 = connection.Read(messageTypeBuffer)
			assert.NoError(s.T(), err2, "should not return error while reading messageType from client")

			messageType := string(messageTypeBuffer)
			assert.Equal(s.T(), "who_am_i", messageType)

			userIDBytes := make([]byte, 8)
			binary.LittleEndian.PutUint64(userIDBytes, expectedUserID)
			_, err2 = connection.Write(userIDBytes)
			assert.NoError(s.T(), err2, "should not return error while sending userID to client")
		}
	}()

	time.Sleep(1 * time.Second)

	err1 := s.client.Connect(&serverAddr)
	assert.NoError(s.T(), err1, "should not return error while creating client")

	userID, err := s.client.WhoAmI()
	assert.Equal(s.T(), expectedUserID, userID)
}

func (s *ServerTestSuite) TestListClientIDsRequest() {
	serverPort := 9003
	serverAddr := net.TCPAddr{Port: serverPort}
	listener, err := net.Listen("tcp", serverAddr.String())
	assert.NoError(s.T(), err, "should not return error while creating server")

	expectedUserIDOne := uint64(11765426)
	expectedUserIDTwo := uint64(326578899)
	expecteduserIDs := []uint64{expectedUserIDOne, expectedUserIDTwo}

	var connection net.Conn
	go func() {
		for {
			var err2 error
			connection, err2 = listener.Accept()
			assert.NoError(s.T(), err2, "should not return error while accepting client connection")

			messageTypeLengthBuffer := make([]byte, 1)
			_, err2 = connection.Read(messageTypeLengthBuffer)
			assert.NoError(s.T(), err2, "should not return error while reading messageType length from client")

			messageLength, err2 := binary.ReadUvarint(bytes.NewBuffer(messageTypeLengthBuffer))
			assert.NoError(s.T(), err2, "messageType length should not be incorrect from")

			messageTypeBuffer := make([]byte, messageLength)
			_, err2 = connection.Read(messageTypeBuffer)
			assert.NoError(s.T(), err2, "should not return error while reading messageType from client")

			messageType := string(messageTypeBuffer)
			assert.Equal(s.T(), "who_is_here", messageType)

			var buffer bytes.Buffer
			gobBuffer := gob.NewEncoder(&buffer)
			err2 = gobBuffer.Encode(expecteduserIDs)
			assert.NoError(s.T(), err2, "should not return error while encoding userIDs")

			_, err2 = connection.Write([]byte{byte(len(buffer.Bytes()))})
			assert.NoError(s.T(), err2, "should not return error while sending userIDs length to client")

			_, err2 = connection.Write(buffer.Bytes())
			assert.NoError(s.T(), err2, "should not return error while sending userIDs to client")
		}
	}()

	time.Sleep(1 * time.Second)

	err1 := s.client.Connect(&serverAddr)
	assert.NoError(s.T(), err1, "should not return error while creating client")

	userIDs, err := s.client.ListClientIDs()
	assert.Equal(s.T(), expecteduserIDs, userIDs)
}

func (s *ServerTestSuite) TearDownSuite() {
	require.NoError(s.T(), s.client.Close())
}

func TestServerTestSuite(t *testing.T) {
	suite.Run(t, new(ServerTestSuite))
}
