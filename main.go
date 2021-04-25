package main

import (
	"context"
	"fmt"
	"os"
)

// TODO: Adopt a logging library

func handleHelp(firstArg string) int {
	helpArgs := map[string]bool{"help": true, "-help": true, "--help": true}

	helpText := "Publisher: Takes a Tx ID, fetches extra context about it from " +
		"Monero Wallet RPC, and pushes an event about the Tx to NATS. \n" +
		"It takes 3 non-optional arguments: <rpcURL> <natsURL> <txID> \n" +
		"* rpcURL: URL to the Monero Wallet RPC server\n" +
		"* natsURL: URL to the NATS Streaming Server\n" +
		"* txID: the ID of the Monero Transaction\n"

	if _, present := helpArgs[firstArg]; present {
		fmt.Print(helpText)
		return 0
	}
	fmt.Printf("Unknown argument %s", firstArg)
	return 1
}

func main() {
	argCount := len(os.Args)
	if argCount != 4 && argCount != 2 {
		provided := argCount - 1
		fmt.Printf("Expected 3 argument. %d provided\n", provided)
		os.Exit(1)
	}

	if argCount == 2 {
		os.Exit(handleHelp(os.Args[1]))
	}

	rpcURL := os.Args[1]
	natsURL := os.Args[2]
	txID := os.Args[3]

	fmt.Printf("Invoked with txid %s\n", txID)

	txGetter := NewRPCClient(rpcURL)

	evPublisher := NewNatsPublishingClient(natsURL)

	if err := ProcessTxid(txID, txGetter, evPublisher); err != nil {
		fmt.Printf("%+v\n", err)
		os.Exit(1)
	}
}

type TxGetter interface {
	GetTransferByTxid(context.Context, string) ([]RpcTx, error)
}

type TxEventPublisher interface {
	PushTxEvent(Tx) error
}

// ProcessTxid fetches extra context about the Monero Transaction from
// Monero Wallet RPC. Then publishes a NATS event about the Transaction.
func ProcessTxid(txid string, rc TxGetter, nc TxEventPublisher) error {
	ctx := context.Background()
	transfers, err := rc.GetTransferByTxid(ctx, txid)
	if err != nil {
		return err
	}

	tx, err := RpcTxToTx(transfers)
	if err != nil {
		return err
	}

	return nc.PushTxEvent(*tx)
}

type BlockGetter interface {
	GetBlockByHash(context.Context, string) (*RpcBlock, error)
}

type BlockEventPublisher interface {
	PushBlockEvent(Block) error
}

func ProcessBlockHash(blockHash string, rc BlockGetter, nc BlockEventPublisher) error {
	ctx := context.Background()
	rpcBlock, err := rc.GetBlockByHash(ctx, blockHash)
	if err != nil {
		return err
	}

	blk := RpcBlockToBlock(*rpcBlock)
	return nc.PushBlockEvent(blk)
}
