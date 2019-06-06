package server

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net"
	"reflect"
	"testing"
)

const serverPort = 9001

func TestWhoAmIRequest(t *testing.T) {
	server := New()
	serverAddr := net.TCPAddr{Port: serverPort}
	require.NoError(t, server.Start(&serverAddr), "should not return error on server start")

	clientConnection, err := net.Dial("tcp", serverAddr.String())
	assert.NoError(t, err, "should not return error while connecting to server")

	messageType := "who_am_i"
	messageTypeLength := len(messageType)

	_, err = clientConnection.Write([]byte{byte(messageTypeLength)})
	assert.NoError(t, err, "should not return error while writing messageTypeLength to server")

	_, err = clientConnection.Write([]byte(messageType))
	assert.NoError(t, err, "should not return error while writing messageType to server")

	expectedUserIDBytes := 8
	userIDBuffer := make([]byte, 1024)
	userIDBytes, err := clientConnection.Read(userIDBuffer)
	assert.NoError(t, err, "should not return error while reading userID from server")

	userID := binary.LittleEndian.Uint64(userIDBuffer)

	assert.Equal(t, expectedUserIDBytes, userIDBytes)
	assert.Equal(t, reflect.Uint64, reflect.TypeOf(userID).Kind())

	defer require.NoError(t, server.Stop())
	defer require.NoError(t, clientConnection.Close())
}

func TestListClientIDsRequest(t *testing.T) {
	server := New()
	serverAddr := net.TCPAddr{Port: serverPort}
	require.NoError(t, server.Start(&serverAddr), "should not return error on server start")

	clientConnection1, err := net.Dial("tcp", serverAddr.String())
	assert.NoError(t, err, "should not return error while connecting to server")

	clientConnection2, err := net.Dial("tcp", serverAddr.String())
	assert.NoError(t, err, "should not return error while connecting to server")

	clientConnection3, err := net.Dial("tcp", serverAddr.String())
	assert.NoError(t, err, "should not return error while connecting to server")

	messageType := "who_is_here"
	messageTypeLength := len(messageType)

	_, err = clientConnection1.Write([]byte{byte(messageTypeLength)})
	assert.NoError(t, err, "should not return error while writing messageTypeLength to server")

	_, err = clientConnection1.Write([]byte(messageType))
	assert.NoError(t, err, "should not return error while writing messageType to server")

	expectedUserIDsLength  := 2
	userIDsLengthBuffer := make([]byte, 1)
	_, err = clientConnection1.Read(userIDsLengthBuffer)
	assert.NoError(t, err, "should not return error while reading UserIdsLength from server")

	userIDsLength, err := binary.ReadUvarint(bytes.NewBuffer(userIDsLengthBuffer))
	assert.NoError(t, err, "should not return error while decoding UserIdsLength")

	userIDsBuffer := make([]byte, userIDsLength)
	_, err = clientConnection1.Read(userIDsBuffer)
	assert.NoError(t, err, "should not return error while reading UserIDs from server")

	var userIDs []uint64
	gobBuffer := gob.NewDecoder(bytes.NewBuffer(userIDsBuffer))
	err = gobBuffer.Decode(&userIDs)
	assert.NoError(t, err, "should not return error while decoding UserIDs")

	assert.Equal(t, expectedUserIDsLength , len(userIDs))

	defer require.NoError(t, server.Stop())
	defer require.NoError(t, clientConnection1.Close())
	defer require.NoError(t, clientConnection2.Close())
	defer require.NoError(t, clientConnection3.Close())
}
