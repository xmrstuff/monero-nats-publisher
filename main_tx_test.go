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

type TxEventPublisherRecording struct {
	CallsCount int
	PassedTX   *Tx
}

func (p *TxEventPublisherRecording) PushTxEvent(t Tx) error {
	p.PassedTX = &t
	p.CallsCount++
	return nil
}

type TxGetterRecording struct {
	CallsCount int
	PassedTXID string
}

func (g *TxGetterRecording) GetTransferByTxid(c context.Context, t string) ([]RpcTx, error) {
	g.PassedTXID = t
	g.CallsCount++
	return []RpcTx{{TXID: t, Type: "in", Height: 3}}, nil
}

type TxEvPublisherBreaking struct {
	CallsCount int
	PassedTX   *Tx
}

func (p *TxEvPublisherBreaking) PushTxEvent(tx Tx) error {
	p.CallsCount++
	p.PassedTX = &tx
	return fmt.Errorf("Dummy Error")
}

func TestProcessTxid(t *testing.T) {

	t.Run("Success", func(t *testing.T) {
		txid := "dummy tx"
		txGetter := TxGetterRecording{}
		evPublisher := TxEventPublisherRecording{}

		ignoreBelowHeight := 0 // Don't ignore any Tx
		err := ProcessTxid(txid, ignoreBelowHeight, &txGetter, &evPublisher)
		assert.Nil(t, err)

		assert.Equal(t, 1, evPublisher.CallsCount)
		assert.Equal(t, txid, evPublisher.PassedTX.TXID)
	})

	t.Run("Success, Tx below ignoring height", func(t *testing.T) {
		txid := "dummy tx"
		txGetter := TxGetterRecording{}
		evPublisher := TxEventPublisherRecording{}

		ignoreBelowHeight := 5 // The Tx will be ignored
		err := ProcessTxid(txid, ignoreBelowHeight, &txGetter, &evPublisher)
		assert.Nil(t, err)

		// The Tx was not pushed to NATS, because it is below ignoring height
		assert.Equal(t, 0, evPublisher.CallsCount)
	})

	t.Run("RPC Error", func(t *testing.T) {
		txid := "dummy tx"
		txGetter := TxGetterBroken{}
		evPublisher := TxEventPublisherRecording{}

		ignoreBelowHeight := 0
		err := ProcessTxid(txid, ignoreBelowHeight, &txGetter, &evPublisher)
		assert.Error(t, err)

		assert.Equal(t, 1, txGetter.CallsCount)
		assert.Equal(t, txid, txGetter.PassedTXID)

		assert.Equal(t, 0, evPublisher.CallsCount)
	})

	t.Run("Event publishing fails", func(t *testing.T) {
		txid := "dummy tx"
		txGetter := TxGetterRecording{}
		evPublisher := TxEvPublisherBreaking{}

		ignoreBelowHeight := 0
		err := ProcessTxid(txid, ignoreBelowHeight, &txGetter, &evPublisher)
		assert.Error(t, err)

		assert.Equal(t, 1, txGetter.CallsCount)
		assert.Equal(t, txid, txGetter.PassedTXID)

		assert.Equal(t, 1, evPublisher.CallsCount)
		assert.Equal(t, txid, evPublisher.PassedTX.TXID)
	})
}
