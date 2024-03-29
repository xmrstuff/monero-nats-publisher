package main

import "fmt"

type Destination struct {
	Amount  int    `json:"amount"`
	Address string `json:"address"`
}

type Tx struct {
	TXID          string        `json:"txid"`
	Destinations  []Destination `json:"destinations"`
	Height        int           `json:"height"`
	Timestamp     int           `json:"timestamp"`
	UnlockTime    int           `json:"unlock_time"`
	Confirmations int           `json:"confirmations"`
}

// RpcTxToTx converts the Monero Transaction representation
// returned by the RPC, into the representation that we intend to
// push through NATS
func RpcTxToTx(rpcTxs []RpcTx) (*Tx, error) {
	tx := Tx{}
	for _, rpcTx := range rpcTxs {
		if !rpcTx.IsIncoming() {
			continue
		}

		tx.TXID = rpcTx.TXID
		tx.Height = rpcTx.Height
		tx.Timestamp = rpcTx.Timestamp
		tx.UnlockTime = rpcTx.UnlockTime
		tx.Confirmations = rpcTx.Confirmations

		dest := Destination{
			Amount:  rpcTx.Amount,
			Address: rpcTx.Address,
		}

		tx.Destinations = append(tx.Destinations, dest)
	}

	if tx.TXID == "" || len(tx.Destinations) == 0 {
		return nil, fmt.Errorf("Unable to turn RPC result into TX: %+v", rpcTxs)
	}

	return &tx, nil
}

type Block struct {
	Hash       string   `json:"hash"`
	Height     int      `json:"height"`
	Timestamp  int      `json:"timestamp"`
	PrevHashes []string `json:"prev_hashes"`
	TxHashes   []string `json:"tx_hashes"`
}

func RpcBlockToBlock(b RpcBlock) Block {
	prevHashes := []string{}
	if b.BlockHeader.PrevHash != "" {
		prevHashes = append(prevHashes, b.BlockHeader.PrevHash)
	}

	return Block{
		Hash:       b.BlockHeader.Hash,
		Height:     b.BlockHeader.Height,
		Timestamp:  b.BlockHeader.Timestamp,
		PrevHashes: prevHashes,
		TxHashes:   b.TxHashes,
	}
}
