package main

import (
	"context"
	"fmt"
	"os"
)

var helpArgs = map[string]bool{"help": true, "-help": true, "--help": true}

const helpText = `
	Takes a single string argument <txid>.

	Calls the get_transfer_by_txid method of Monero Wallet RPC, passing
	the value of <txid>; to obtain extra context about the transaction.

	If successful, pushes a "transaction.created" event to NATS with all
	the transaction context.
`

func main() {
	if len(os.Args) != 2 {
		provided := len(os.Args) - 1
		fmt.Printf("Expected 1 argument. %d provided\n", provided)
		return
	}

	arg := os.Args[1]

	if _, pres := helpArgs[arg]; pres {
		fmt.Print(helpText)
		return
	}

	fmt.Printf("Invoked with txid %s\n", arg)

	txGetter := newClient("wallet-rpc-url")
	evPublisher := NewNatsPublishingClient("nats-host")

	ProcessTxid(arg, txGetter, evPublisher)
}

type TxGetter interface {
	GetTransferByTxid(context.Context, string) (*Tx, error)
}

type NatsTxEventPublisher interface {
	PushTxEvent(Tx) error
}

// ProcessTxid fetches extra context about the Monero Transaction from
// Monero Wallet RPC. Then publishes a NATS event about the Transaction.
func ProcessTxid(txid string, rc TxGetter, nc NatsTxEventPublisher) error {
	ctx := context.Background()
	tx, err := rc.GetTransferByTxid(ctx, txid)
	if err != nil {
		return err
	}
	return nc.PushTxEvent(*tx)
}
