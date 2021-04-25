package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRpcTransfersToTxSuccess(t *testing.T) {
	transfers := []RpcTx{
		{
			TXID:          "dummy txid",
			Height:        20,
			Timestamp:     2000,
			UnlockTime:    100,
			Confirmations: 1,
			Amount:        333333,
			Address:       "addr1",
			Type:          "in",
		},
		{
			TXID:          "dummy txid",
			Height:        20,
			Timestamp:     2000,
			UnlockTime:    100,
			Confirmations: 1,
			Amount:        666666,
			Address:       "addr2",
			Type:          "pool",
		},
		{
			TXID:          "dummy txid",
			Height:        20,
			Timestamp:     2000,
			UnlockTime:    100,
			Confirmations: 1,
			Amount:        99999,
			Address:       "addr2",
			Type:          "out",
		},
	}
	tx, err := RpcTxToTx(transfers)
	assert.Nil(t, err)

	// The Transfer with type "out" was ignored
	assert.Equal(t, 2, len(tx.Destinations))

	for idx, _ := range tx.Destinations {
		assert.Equal(t, transfers[idx].TXID, tx.TXID)
		assert.Equal(t, transfers[idx].Height, tx.Height)
		assert.Equal(t, transfers[idx].Timestamp, tx.Timestamp)
		assert.Equal(t, transfers[idx].UnlockTime, tx.UnlockTime)
		assert.Equal(t, transfers[idx].Confirmations, tx.Confirmations)

		assert.Equal(t, transfers[idx].Address, tx.Destinations[idx].Address)
		assert.Equal(t, transfers[idx].Amount, tx.Destinations[idx].Amount)
	}
}

func TestRpcTransfersToTxFailure(t *testing.T) {
	transfers := []RpcTx{
		{
			TXID:          "dummy txid",
			Height:        20,
			Timestamp:     2000,
			UnlockTime:    100,
			Confirmations: 1,
			Amount:        99999,
			Address:       "addr2",
			Type:          "out",
		},
	}
	tx, err := RpcTxToTx(transfers)
	// No TX was created, because the only Transfer
	// has type "out"
	assert.Error(t, err)
	assert.Nil(t, tx)
}

func TestRpcBlockToBlock(t *testing.T) {
	rb := RpcBlock{
		BlockHeader: RpcBlockHeader{
			Hash:      "some hash",
			Height:    100,
			Timestamp: 300,
			PrevHash:  "hash of prev block",
		},
		TxHashes: []string{"hash1", "hash2"},
	}

	b := RpcBlockToBlock(rb)

	assert.Equal(t, rb.BlockHeader.Hash, b.Hash)
	assert.Equal(t, rb.BlockHeader.Height, b.Height)
	assert.Equal(t, rb.BlockHeader.Timestamp, b.Timestamp)
	assert.Equal(t, rb.BlockHeader.PrevHash, b.PrevHash)
	assert.Equal(t, rb.TxHashes, b.TxHashes)
}
