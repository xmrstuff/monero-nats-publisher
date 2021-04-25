package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TxGetterBroken struct {
	CallsCount int
	PassedTXID string
}

func (g *TxGetterBroken) GetTransferByTxid(c context.Context, t string) ([]RpcTx, error) {
	g.PassedTXID = t
	g.CallsCount++
	return nil, fmt.Errorf("Dummy Error")
}

type EventPublisherRecording struct {
	CallsCount int
	PassedTX   *Tx
}

func (p *EventPublisherRecording) PushTxEvent(t Tx) error {
	p.PassedTX = &t
	p.CallsCount++
	return nil
}

func TestRPCFails(t *testing.T) {
	txid := "dummy tx"
	txGetter := TxGetterBroken{}
	evPublisher := EventPublisherRecording{}

	err := ProcessTxid(txid, &txGetter, &evPublisher)
	assert.Error(t, err)

	assert.Equal(t, 1, txGetter.CallsCount)
	assert.Equal(t, txid, txGetter.PassedTXID)

	assert.Equal(t, 0, evPublisher.CallsCount)
}

type TxGetterRecording struct {
	CallsCount int
	PassedTXID string
}

func (g *TxGetterRecording) GetTransferByTxid(c context.Context, t string) ([]RpcTx, error) {
	g.PassedTXID = t
	g.CallsCount++
	return []RpcTx{{TXID: t, Type: "in"}}, nil
}

type EvPublisherBreaking struct {
	CallsCount int
	PassedTX   *Tx
}

func (p *EvPublisherBreaking) PushTxEvent(tx Tx) error {
	p.CallsCount++
	p.PassedTX = &tx
	return fmt.Errorf("Dummy Error")
}

func TestPublishingFails(t *testing.T) {
	txid := "dummy tx"
	txGetter := TxGetterRecording{}
	evPublisher := EvPublisherBreaking{}

	err := ProcessTxid(txid, &txGetter, &evPublisher)
	assert.Error(t, err)

	assert.Equal(t, 1, txGetter.CallsCount)
	assert.Equal(t, txid, txGetter.PassedTXID)

	assert.Equal(t, 1, evPublisher.CallsCount)
	assert.Equal(t, txid, evPublisher.PassedTX.TXID)
}

func TestProcessTxidSucceeds(t *testing.T) {
	txid := "dummy tx"
	txGetter := TxGetterRecording{}
	evPublisher := EventPublisherRecording{}

	err := ProcessTxid(txid, &txGetter, &evPublisher)
	assert.Nil(t, err)
}
