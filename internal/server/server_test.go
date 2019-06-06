package server

import (
	"encoding/binary"
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
