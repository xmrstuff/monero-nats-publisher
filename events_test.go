package main

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type DummySucessfulPublisher struct {
	ChannelPassed string
	PayloadPassed []byte
}

func (p *DummySucessfulPublisher) Publish(payload []byte, channel string) error {
	p.ChannelPassed = channel
	p.PayloadPassed = payload

	return nil
}

type DummyFailingPublisher struct{}

func (p *DummyFailingPublisher) Publish(payload []byte, channel string) error {
	return fmt.Errorf("")
}

func TestPushTxEventSuccess(t *testing.T) {

	dp := DummySucessfulPublisher{}
	p := EventPublishing{Publisher: &dp}

	assert.Zero(t, dp.ChannelPassed)
	assert.Zero(t, dp.PayloadPassed)

	tx := Tx{
		TXID: "some tx id",
		Destinations: []Destination{
			{Amount: 2, Address: "addr1"},
			{Amount: 4, Address: "addr2"},
		},
	}
	assert.Nil(t, p.PushTxEvent(tx))
	assert.Equal(t, moneroNATSChannel, dp.ChannelPassed)

	evTx := Tx{}
	evPayload := Event{Data: &evTx}
	assert.Nil(t, json.Unmarshal(dp.PayloadPassed, &evPayload))

	assert.NotNil(t, evPayload.Version)
	assert.Equal(t, txCreated, evPayload.Type)
	assert.Equal(t, tx.TXID, evTx.TXID)
}

func TestPushTxEventFailure(t *testing.T) {
	dp := DummyFailingPublisher{}
	p := EventPublishing{Publisher: &dp}

	tx := Tx{}
	assert.Error(t, p.PushTxEvent(tx))
}

func TestPushBlockEventSuccess(t *testing.T) {
	dp := DummySucessfulPublisher{}
	p := EventPublishing{Publisher: &dp}

	assert.Zero(t, dp.ChannelPassed)
	assert.Zero(t, dp.PayloadPassed)

	blk := Block{
		Hash:      "some hash",
		Height:    300,
		Timestamp: 9000,
		PrevHash:  "hash of prev block",
		TxHashes:  []string{"tx1", "tx2"},
	}
	assert.Nil(t, p.PushBlockEvent(blk))
	assert.Equal(t, moneroNATSChannel, dp.ChannelPassed)

	evBlk := Block{}
	evPayload := Event{Data: &evBlk}
	assert.Nil(t, json.Unmarshal(dp.PayloadPassed, &evPayload))

	assert.NotNil(t, evPayload.Version)
	assert.Equal(t, blockCreated, evPayload.Type)
	assert.Equal(t, blk.Hash, evBlk.Hash)
}

func TestPushBlockEventFailure(t *testing.T) {
	dp := DummyFailingPublisher{}
	p := EventPublishing{Publisher: &dp}

	blk := Block{}
	assert.Error(t, p.PushBlockEvent(blk))
}
