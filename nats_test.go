package main

import (
	"testing"

	"github.com/nats-io/nats-streaming-server/server"
	"github.com/stretchr/testify/assert"
)

// func TestNATSPublishtSuccess(t *testing.T) {
// 	tx := Tx{
// 		TXID:   "some id",
// 		Height: 400,
// 		Destinations: []destination{
// 			{Amount: 2, Address: "addr1"},
// 			{Amount: 3, Address: "addr2"},
// 		},
// 	}

// 	clusterID := "test-cluster"
// 	channelID := "monero"

// 	ss, err := server.RunServer(clusterID)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, ss)
// 	defer ss.Shutdown()

// 	publisher := publisherClient{
// 		NATSHost:  ss.ClientURL(),
// 		ClusterID: clusterID,
// 		ChannelID: channelID,
// 		ClientID:  "random-client",
// 	}

// 	err = publisher.PushTxEvent(tx)
// 	assert.Nil(t, err)
// }

func TestNATSPublishtSuccess(t *testing.T) {
	payload := []byte("testing")

	channelID := "monero"

	ss, err := server.RunServer(ClusterID)
	assert.Nil(t, err)
	assert.NotNil(t, ss)
	defer ss.Shutdown()

	publisher := NewNATSClient(ss.ClientURL())
	err = publisher.Publish(payload, channelID)
	assert.Nil(t, err)
}

func TestPushTxEventFailedToConnect(t *testing.T) {
	payload := []byte("testing")

	channelID := "monero"

	publisher := NewNATSClient("nats://127.0.0.1:4222")
	err := publisher.Publish(payload, channelID)
	assert.Error(t, err)
}
