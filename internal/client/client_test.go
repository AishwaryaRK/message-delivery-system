package client

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"net"
	"sync"
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

	var wg sync.WaitGroup
	wg.Add(1)
	var connection net.Conn
	go func() {
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
		wg.Done()
	}()

	time.Sleep(1 * time.Second)

	err1 := s.client.Connect(&serverAddr)
	assert.NoError(s.T(), err1, "should not return error while creating client")

	userID, err := s.client.WhoAmI()
	assert.Equal(s.T(), expectedUserID, userID)
	wg.Wait()
}

func (s *ServerTestSuite) TestListClientIDsRequest() {
	serverPort := 9003
	serverAddr := net.TCPAddr{Port: serverPort}
	listener, err := net.Listen("tcp", serverAddr.String())
	assert.NoError(s.T(), err, "should not return error while creating server")

	expectedUserIDOne := uint64(11765426)
	expectedUserIDTwo := uint64(326578899)
	expecteduserIDs := []uint64{expectedUserIDOne, expectedUserIDTwo}

	var wg sync.WaitGroup
	wg.Add(1)
	var connection net.Conn
	go func() {
		var err2 error
		connection, err2 = listener.Accept()
		assert.NoError(s.T(), err2, "should not return error while accepting client connection")

		messageTypeLengthBuffer := make([]byte, 1)
		_, err2 = connection.Read(messageTypeLengthBuffer)
		assert.NoError(s.T(), err2, "should not return error while reading messageType length from client")

		messageTypeLength, err2 := binary.ReadUvarint(bytes.NewBuffer(messageTypeLengthBuffer))
		assert.NoError(s.T(), err2, "messageType length should not be incorrect from")

		messageTypeBuffer := make([]byte, messageTypeLength)
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
		wg.Done()
	}()

	time.Sleep(1 * time.Second)

	err1 := s.client.Connect(&serverAddr)
	assert.NoError(s.T(), err1, "should not return error while creating client")

	userIDs, err := s.client.ListClientIDs()
	assert.ElementsMatch(s.T(), expecteduserIDs, userIDs)
	wg.Wait()
}

func (s *ServerTestSuite) TestSendMsgRequest() {
	serverPort := 9004
	serverAddr := net.TCPAddr{Port: serverPort}
	listener, err := net.Listen("tcp", serverAddr.String())
	assert.NoError(s.T(), err, "should not return error while creating server")

	expectedMessage := "Hello Server!"
	expectedUserIDOne := uint64(11765426)
	expectedUserIDTwo := uint64(326578899)
	expecteduserIDs := []uint64{expectedUserIDOne, expectedUserIDTwo}

	var wg sync.WaitGroup
	wg.Add(1)
	var connection net.Conn
	go func() {
		var err2 error
		connection, err2 = listener.Accept()
		assert.NoError(s.T(), err2, "should not return error while accepting client connection")

		messageTypeLengthBuffer := make([]byte, 1)
		_, err2 = connection.Read(messageTypeLengthBuffer)
		assert.NoError(s.T(), err2, "should not return error while reading messageType length from client")

		messageTypeLength, err2 := binary.ReadUvarint(bytes.NewBuffer(messageTypeLengthBuffer))
		assert.NoError(s.T(), err2, "messageType length should not be incorrect from")

		messageTypeBuffer := make([]byte, messageTypeLength)
		_, err2 = connection.Read(messageTypeBuffer)
		assert.NoError(s.T(), err2, "should not return error while reading messageType from client")

		messageType := string(messageTypeBuffer)
		assert.Equal(s.T(), "relay", messageType)

		receiverListLengthBuffer := make([]byte, 1)
		_, err2 = connection.Read(receiverListLengthBuffer)
		assert.NoError(s.T(), err2, "should not return error while reading receivers length from client")

		receiverListLength, err2 := binary.ReadUvarint(bytes.NewBuffer(receiverListLengthBuffer))
		assert.NoError(s.T(), err2, "should not get incorrect receivers length from client")

		receiversBuffer := make([]byte, receiverListLength)
		_, err2 = connection.Read(receiversBuffer)
		assert.NoError(s.T(), err2, "should not return error while reading receivers from client")

		var receivers []uint64
		gobBuffer := gob.NewDecoder(bytes.NewBuffer(receiversBuffer))
		err2 = gobBuffer.Decode(&receivers)
		assert.NoError(s.T(), err2, "should not return error while decoding receivers from client")

		assert.ElementsMatch(s.T(), expecteduserIDs, receivers)

		messageLengthBuffer := make([]byte, 4)
		_, err2 = connection.Read(messageLengthBuffer)
		assert.NoError(s.T(), err2, "should not return error while reading message length from client")

		messageLength, err2 := binary.ReadUvarint(bytes.NewBuffer(messageLengthBuffer))
		assert.NoError(s.T(), err2, "should not get incorrect message length from client")

		messageBuffer := make([]byte, messageLength)
		_, err2 = connection.Read(messageBuffer)
		assert.NoError(s.T(), err2, "should not return error while reading message from client")

		assert.Equal(s.T(), expectedMessage, string(messageBuffer))
		wg.Done()
	}()

	time.Sleep(1 * time.Second)

	err1 := s.client.Connect(&serverAddr)
	assert.NoError(s.T(), err1, "should not return error while creating client")

	err1 = s.client.SendMsg(expecteduserIDs, []byte(expectedMessage))
	assert.NoError(s.T(), err1, "should not return error while sending message to peer clients")

	wg.Wait()
}

func (s *ServerTestSuite) TestHandleIncomingMessages() {
	serverPort := 9005
	serverAddr := net.TCPAddr{Port: serverPort}
	listener, err := net.Listen("tcp", serverAddr.String())
	assert.NoError(s.T(), err, "should not return error while creating server")

	expectedMessage := "Hello Receiver!"
	expectedSenderID := uint64(764354876876673)

	var wg sync.WaitGroup
	wg.Add(1)
	var connection net.Conn
	go func() {
		var err2 error
		connection, err2 = listener.Accept()
		assert.NoError(s.T(), err2, "should not return error while accepting client connection")

		senderIDBytes := make([]byte, 8)
		binary.LittleEndian.PutUint64(senderIDBytes, expectedSenderID)
		_, err2 = connection.Write(senderIDBytes)
		assert.NoError(s.T(), err2, "should not return error while sending senderID to client")

		var messageLength uint32
		messageLength = uint32(len(expectedMessage))
		msgLengthBytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(msgLengthBytes, messageLength)
		_, err = connection.Write(msgLengthBytes)
		assert.NoError(s.T(), err2, "should not return error while sending message length to client")

		_, err = connection.Write([]byte(expectedMessage))
		assert.NoError(s.T(), err2, "should not return error while sending message to client")

		wg.Done()
	}()

	time.Sleep(1 * time.Second)

	err1 := s.client.Connect(&serverAddr)
	assert.NoError(s.T(), err1, "should not return error while creating client")

	writeCh := make(chan IncomingMessage, 1)
	s.client.HandleIncomingMessages(writeCh)
	assert.NoError(s.T(), err1, "should not return error while sending message to peer clients")

	incomingMessage := <-writeCh
	assert.Equal(s.T(), expectedSenderID, incomingMessage.SenderID)
	assert.Equal(s.T(), expectedMessage, string(incomingMessage.Body))
	wg.Wait()
}

func (s *ServerTestSuite) TearDownSuite() {
	require.NoError(s.T(), s.client.Close())
}

func TestServerTestSuite(t *testing.T) {
	suite.Run(t, new(ServerTestSuite))
}
