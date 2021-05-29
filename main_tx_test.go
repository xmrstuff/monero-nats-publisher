package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockedGetTxByTxidReturn struct {
	Txs []RpcTx
	E   error
}
type MockedTxGetter struct {
	CallsCount int
	TxidArgs   []string
	Returns    []MockedGetTxByTxidReturn
}

func (g *MockedTxGetter) GetTransferByTxid(c context.Context, t string) ([]RpcTx, error) {
	g.CallsCount++

	g.TxidArgs = append(g.TxidArgs, t)

	result := g.Returns[0]
	if len(g.Returns) > 1 {
		g.Returns = g.Returns[1:]
	} else {
		g.Returns = []MockedGetTxByTxidReturn{}
	}

	return result.Txs, result.E
}

type MockedTxPublisher struct {
	CallsCount int
	TxArgs     []Tx
	Returns    []error
}

func (g *MockedTxPublisher) PushTxEvent(tx Tx) error {
	g.CallsCount++

	g.TxArgs = append(g.TxArgs, tx)

	result := g.Returns[0]
	if len(g.Returns) > 1 {
		g.Returns = g.Returns[1:]
	} else {
		g.Returns = []error{}
	}

	return result
}

func TestProcessTxid(t *testing.T) {

	t.Run("Success", func(t *testing.T) {
		txid := "dummy tx"
		txHeight := 3
		txGetter := MockedTxGetter{
			Returns: []MockedGetTxByTxidReturn{
				{
					E:   nil,
					Txs: []RpcTx{{TXID: txid, Type: "in", Height: txHeight}},
				},
			},
		}
		evPublisher := MockedTxPublisher{Returns: []error{nil}}

		ignoreBelowHeight := 0 // Don't ignore any Tx
		err := ProcessTxid(txid, ignoreBelowHeight, &txGetter, &evPublisher)
		assert.Nil(t, err)

		assert.Equal(t, 1, evPublisher.CallsCount)
		assert.Equal(t, txid, evPublisher.TxArgs[0].TXID)
	})

	t.Run("Success, Tx below ignoring height", func(t *testing.T) {
		txid := "dummy tx"
		txHeight := 3
		txGetter := MockedTxGetter{
			Returns: []MockedGetTxByTxidReturn{
				{
					E:   nil,
					Txs: []RpcTx{{TXID: txid, Type: "in", Height: txHeight}},
				},
			},
		}
		evPublisher := MockedTxPublisher{Returns: []error{nil}}

		ignoreBelowHeight := txHeight + 2 // The Tx will be ignored
		err := ProcessTxid(txid, ignoreBelowHeight, &txGetter, &evPublisher)
		assert.Nil(t, err)

		assert.Equal(t, 1, txGetter.CallsCount)

		// The Tx was not pushed to NATS, because it is below ignoring height
		assert.Equal(t, 0, evPublisher.CallsCount)
	})

	t.Run("RPC Error", func(t *testing.T) {
		txid := "dummy tx"
		txGetter := MockedTxGetter{
			Returns: []MockedGetTxByTxidReturn{
				{
					E:   fmt.Errorf("Dummy Error"),
					Txs: nil,
				},
			},
		}
		evPublisher := MockedTxPublisher{
			Returns: []error{nil},
		}

		ignoreBelowHeight := 0
		err := ProcessTxid(txid, ignoreBelowHeight, &txGetter, &evPublisher)
		assert.Error(t, err)

		assert.Equal(t, 1, txGetter.CallsCount)
		assert.Equal(t, txid, txGetter.TxidArgs[0])

		assert.Equal(t, 0, evPublisher.CallsCount)
	})

	t.Run("Event publishing fails", func(t *testing.T) {
		txid := "dummy tx"
		txHeight := 3
		txGetter := MockedTxGetter{
			Returns: []MockedGetTxByTxidReturn{
				{
					E:   nil,
					Txs: []RpcTx{{TXID: txid, Type: "in", Height: txHeight}},
				},
			},
		}
		evPublisher := MockedTxPublisher{
			Returns: []error{fmt.Errorf("Dummy NATS Error")},
		}

		ignoreBelowHeight := 0
		err := ProcessTxid(txid, ignoreBelowHeight, &txGetter, &evPublisher)
		assert.Error(t, err)

		assert.Equal(t, 1, txGetter.CallsCount)
		assert.Equal(t, txid, txGetter.TxidArgs[0])

		assert.Equal(t, 1, evPublisher.CallsCount)
		assert.Equal(t, txid, evPublisher.TxArgs[0].TXID)
	})
}
