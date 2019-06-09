package test

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net"
	"testing"
	"unity/message-delivery-system/internal/client"
	"unity/message-delivery-system/internal/server"
)

const serverPort = 50000

func TestIntegration(t *testing.T) {
	srv := server.New()

	serverAddr := net.TCPAddr{Port: serverPort}
	require.NoError(t, srv.Start(&serverAddr))
	defer assertDoesNotError(t, srv.Stop)

	// Create clients
	client1, client1ID := createClientAndFetchID(t)
	defer assertDoesNotError(t, client1.Close)
	client1Ch := make(chan client.IncomingMessage)
	defer close(client1Ch)

	client2, client2ID := createClientAndFetchID(t)
	defer assertDoesNotError(t, client2.Close)
	client2Ch := make(chan client.IncomingMessage)
	defer close(client2Ch)

	client3, client3ID := createClientAndFetchID(t)
	defer assertDoesNotError(t, client3.Close)
	client3Ch := make(chan client.IncomingMessage)
	defer close(client3Ch)

	t.Run("List other clients from each client", func(t *testing.T) {
		ids, err := client1.ListClientIDs()
		assert.NoError(t, err)
		assert.ElementsMatch(t, []uint64{client2ID, client3ID}, ids)

		ids, err = client2.ListClientIDs()
		assert.NoError(t, err)
		assert.ElementsMatch(t, []uint64{client1ID, client3ID}, ids)

		ids, err = client3.ListClientIDs()
		assert.NoError(t, err)
		assert.ElementsMatch(t, []uint64{client1ID, client2ID}, ids)
	})

	t.Run("Send message from the first client to the two other clients", func(t *testing.T) {
		body := []byte("Hello world!")
		assert.Equal(t, nil, client1.SendMsg([]uint64{client2ID, client3ID}, body))

		go client2.HandleIncomingMessages(client2Ch)
		incomingMessage := <-client2Ch
		assert.Equal(t, body, incomingMessage.Body)
		assert.Equal(t, uint64(client1ID), incomingMessage.SenderID)

		go client3.HandleIncomingMessages(client3Ch)
		incomingMessage = <-client3Ch
		assert.Equal(t, body, incomingMessage.Body)
		assert.Equal(t, uint64(client1ID), incomingMessage.SenderID)
	})
}

func assertDoesNotError(tb testing.TB, fn func() error) {
	assert.NoError(tb, fn())
}

func createClientAndFetchID(t *testing.T) (*client.Client, uint64) {
	cli := client.New()
	serverAddr := net.TCPAddr{Port: serverPort}
	require.NoError(t, cli.Connect(&serverAddr))
	id, err := cli.WhoAmI()
	assert.NoError(t, err)
	return cli, id
}
