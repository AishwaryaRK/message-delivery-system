package server

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"net"
	"reflect"
	"testing"
)

type ServerTestSuite struct {
	suite.Suite
	server                *Server
	clientConnectionOne   net.Conn
	clientConnectionTwo   net.Conn
	clientConnectionThree net.Conn
	userIDOne             uint64
	userIDTwo             uint64
	userIDThree           uint64
}

func (s *ServerTestSuite) SetupSuite() {
	serverPort := 9001
	s.server = New()
	serverAddr := net.TCPAddr{Port: serverPort}
	require.NoError(s.T(), s.server.Start(&serverAddr), "should not return error on server start")

	var err error
	s.clientConnectionOne, err = net.Dial("tcp", serverAddr.String())
	assert.NoError(s.T(), err, "should not return error while connecting to server")

	s.clientConnectionTwo, err = net.Dial("tcp", serverAddr.String())
	assert.NoError(s.T(), err, "should not return error while connecting to server")

	s.clientConnectionThree, err = net.Dial("tcp", serverAddr.String())
	assert.NoError(s.T(), err, "should not return error while connecting to server")

	s.userIDOne, err = getUserID(s.clientConnectionOne)
	assert.NoError(s.T(), err, "should not return error while getting userID from server")
	s.userIDTwo, err = getUserID(s.clientConnectionTwo)
	assert.NoError(s.T(), err, "should not return error while getting userID from server")
	s.userIDThree, err = getUserID(s.clientConnectionThree)
	assert.NoError(s.T(), err, "should not return error while getting userID from server")
}

func getUserID(clientConnection net.Conn) (uint64, error) {
	var userID uint64
	messageType := "who_am_i"
	messageTypeLength := len(messageType)

	_, err := clientConnection.Write([]byte{byte(messageTypeLength)})
	if err != nil {
		return userID, err
	}

	_, err = clientConnection.Write([]byte(messageType))
	if err != nil {
		return userID, err
	}

	userIDBuffer := make([]byte, 8)
	_, err = clientConnection.Read(userIDBuffer)
	if err != nil {
		return userID, err
	}

	userID = binary.LittleEndian.Uint64(userIDBuffer)
	return userID, nil
}

func (s *ServerTestSuite) TestWhoAmIRequest() {
	messageType := "who_am_i"
	messageTypeLength := len(messageType)

	_, err := s.clientConnectionOne.Write([]byte{byte(messageTypeLength)})
	assert.NoError(s.T(), err, "should not return error while writing messageTypeLength to server")

	_, err = s.clientConnectionOne.Write([]byte(messageType))
	assert.NoError(s.T(), err, "should not return error while writing messageType to server")

	expectedUserIDBytes := 8
	userIDBuffer := make([]byte, 1024)
	userIDBytes, err := s.clientConnectionOne.Read(userIDBuffer)
	assert.NoError(s.T(), err, "should not return error while reading userID from server")

	userID := binary.LittleEndian.Uint64(userIDBuffer)

	assert.Equal(s.T(), expectedUserIDBytes, userIDBytes)
	assert.Equal(s.T(), reflect.Uint64, reflect.TypeOf(userID).Kind())
}

func (s *ServerTestSuite) TestListClientIDsRequest() {
	messageType := "who_is_here"
	messageTypeLength := len(messageType)

	_, err := s.clientConnectionOne.Write([]byte{byte(messageTypeLength)})
	assert.NoError(s.T(), err, "should not return error while writing messageTypeLength to server")

	_, err = s.clientConnectionOne.Write([]byte(messageType))
	assert.NoError(s.T(), err, "should not return error while writing messageType to server")

	expectedUserIDsLength := 2
	userIDsLengthBuffer := make([]byte, 1)
	_, err = s.clientConnectionOne.Read(userIDsLengthBuffer)
	assert.NoError(s.T(), err, "should not return error while reading UserIdsLength from server")

	userIDsLength, err := binary.ReadUvarint(bytes.NewBuffer(userIDsLengthBuffer))
	assert.NoError(s.T(), err, "should not return error while decoding UserIdsLength")

	userIDsBuffer := make([]byte, userIDsLength)
	_, err = s.clientConnectionOne.Read(userIDsBuffer)
	assert.NoError(s.T(), err, "should not return error while reading UserIDs from server")

	var userIDs []uint64
	gobBuffer := gob.NewDecoder(bytes.NewBuffer(userIDsBuffer))
	err = gobBuffer.Decode(&userIDs)
	assert.NoError(s.T(), err, "should not return error while decoding UserIDs")

	assert.Equal(s.T(), expectedUserIDsLength, len(userIDs))
	assert.ElementsMatch(s.T(), []uint64{s.userIDTwo, s.userIDThree}, userIDs)
}

func (s *ServerTestSuite) TestRelayRequest() {
	messageType := "relay"
	messageTypeLength := len(messageType)

	_, err := s.clientConnectionOne.Write([]byte{byte(messageTypeLength)})
	assert.NoError(s.T(), err, "should not return error while writing messageTypeLength to server")

	_, err = s.clientConnectionOne.Write([]byte(messageType))
	assert.NoError(s.T(), err, "should not return error while writing messageType to server")

	recipients := []uint64{s.userIDTwo}
	var recipientsBuffer bytes.Buffer
	gobBuffer := gob.NewEncoder(&recipientsBuffer)
	require.NoError(s.T(), gobBuffer.Encode(recipients), "should not return error while encoding recipients")

	_, err = s.clientConnectionOne.Write([]byte{byte(len(recipientsBuffer.Bytes()))})
	assert.NoError(s.T(), err, "should not return error while writing number of recipients to server")

	_, err = s.clientConnectionOne.Write(recipientsBuffer.Bytes())
	assert.NoError(s.T(), err, "should not return error while writing recipients to server")

	message := "Hello recipient!"
	body := []byte(message)
	var messageLength uint32
	messageLength = uint32(len(body))

	msgLengthBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(msgLengthBytes, messageLength)
	_, err = s.clientConnectionOne.Write(msgLengthBytes)
	assert.NoError(s.T(), err, "should not return error while writing messageLength to server")

	_, err = s.clientConnectionOne.Write(body)
	assert.NoError(s.T(), err, "should not return error while writing message to server")

	senderIDBuffer := make([]byte, 8)
	_, err = s.clientConnectionTwo.Read(senderIDBuffer)
	assert.NoError(s.T(), err, "should not return error while reading recipient message - senderID from server")

	senderID := binary.LittleEndian.Uint64(senderIDBuffer)
	assert.Equal(s.T(), s.userIDOne, senderID)

	messageLengthBuffer := make([]byte, 4)
	_, err = s.clientConnectionTwo.Read(messageLengthBuffer)
	assert.NoError(s.T(), err, "should not return error while reading recipient message - length from server")

	msgLength, err := binary.ReadUvarint(bytes.NewBuffer(messageLengthBuffer))
	assert.NoError(s.T(), err, "should not return error while reading recipient message from server")

	messageBuffer := make([]byte, msgLength)
	_, err = s.clientConnectionTwo.Read(messageBuffer)
	assert.NoError(s.T(), err, "should not return error while reading recipient message from server")

	assert.Equal(s.T(), message, string(messageBuffer))
}

func (s *ServerTestSuite) TearDownSuite() {
	require.NoError(s.T(), s.server.Stop())
	require.NoError(s.T(), s.clientConnectionOne.Close())
	require.NoError(s.T(), s.clientConnectionTwo.Close())
	require.NoError(s.T(), s.clientConnectionThree.Close())
}

func TestServerTestSuite(t *testing.T) {
	suite.Run(t, new(ServerTestSuite))
}
