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

const (
	WalletRPCURLEnvVar = "WALLET_RPC_URL"
	NATSURLEnvVar      = "NATS_URL"
)

func main() {
	if len(os.Args) != 2 {
		provided := len(os.Args) - 1
		fmt.Printf("Expected 1 argument. %d provided\n", provided)
		os.Exit(1)
	}

	arg := os.Args[1]

	if _, present := helpArgs[arg]; present {
		fmt.Print(helpText)
		os.Exit(0)
	}

	fmt.Printf("Invoked with txid %s\n", arg)

	wu := os.Getenv(WalletRPCURLEnvVar)
	txGetter := NewRPCClient(wu)

	nu := os.Getenv(NATSURLEnvVar)
	evPublisher := NewNatsPublishingClient(nu)

	if err := ProcessTxid(arg, txGetter, evPublisher); err != nil {
		fmt.Printf("%+v\n", err)
		os.Exit(1)
	}
}

type TxGetter interface {
	GetTransferByTxid(context.Context, string) ([]RpcResponseTransaction, error)
}

type NatsTxEventPublisher interface {
	PushTxEvent(Tx) error
}

// ProcessTxid fetches extra context about the Monero Transaction from
// Monero Wallet RPC. Then publishes a NATS event about the Transaction.
func ProcessTxid(txid string, rc TxGetter, nc NatsTxEventPublisher) error {
	ctx := context.Background()
	transfers, err := rc.GetTransferByTxid(ctx, txid)
	if err != nil {
		return err
	}

	tx, err := RpcTransfersToTx(transfers)
	if err != nil {
		return err
	}

	return nc.PushTxEvent(*tx)
}
