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

	parsedPayload := Event{Data: Tx{}}
	assert.Nil(t, json.Unmarshal(dp.PayloadPassed, &parsedPayload))

	assert.NotNil(t, parsedPayload.Version)
	assert.Equal(t, txCreated, parsedPayload.Type)
	assert.Equal(t, tx.TXID, parsedPayload.Data.TXID)
}

type DummyFailingPublisher struct{}

func (p *DummyFailingPublisher) Publish(payload []byte, channel string) error {
	return fmt.Errorf("")
}

func TestPushTxEventFailure(t *testing.T) {
	dp := DummyFailingPublisher{}
	p := EventPublishing{Publisher: &dp}

	tx := Tx{}
	assert.Error(t, p.PushTxEvent(tx))
}
