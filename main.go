package main

import (
	"context"
	"fmt"
	"log"
	"os"

	cli "github.com/urfave/cli/v2"
)

// TODO: Adopt a logging library

func main() {
	var natsURL, walletURL, daemonURL string
	var maxExtraAncestors, ignoreBelowHeight int
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "nats-url",
				Aliases:     []string{"nats", "n"},
				Value:       "http://localhost:4222",
				Usage:       "URL to the NATS Streaming Server",
				Destination: &natsURL,
			},
			&cli.IntFlag{
				Name:        "ignore-below-height",
				Aliases:     []string{"i"},
				Value:       0,
				Usage:       "Ignores Blocks and Transactions which height lower than this value",
				Destination: &ignoreBelowHeight,
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "ping",
				Usage: "Pings the NATS server, to verify that connection is configured properly",
				Action: func(c *cli.Context) error {
					evPublisher := NewNatsPublishingClient(natsURL)
					if !evPublisher.IsConnected() {
						return fmt.Errorf("failed to ping NATS at %s", natsURL)
					}
					return nil
				},
			},
			{
				Name:    "transaction",
				Aliases: []string{"tx"},
				Usage:   "Gather extra context about a Monero Tx and publish it through NATS",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "monero-wallet-rpc-url",
						Aliases:     []string{"wallet", "w"},
						Value:       "http://localhost:38083",
						Usage:       "URL to the RPC server of the Monero Wallet",
						Destination: &walletURL,
					},
				},
				Action: func(c *cli.Context) error {
					txid := c.Args().First()
					if txid == "" {
						return fmt.Errorf("tx command requires a txid argument")
					}

					rpcClient := NewRPCClient(walletURL)
					evPublisher := NewNatsPublishingClient(natsURL)
					return ProcessTxid(txid, ignoreBelowHeight, rpcClient, evPublisher)
				},
			},
			{
				Name:    "block",
				Aliases: []string{"blk"},
				Usage:   "Gather extra context about a Monero Block and publish it through NATS",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "monero-daemon-rpc-url",
						Aliases:     []string{"daemon", "d"},
						Value:       "http://localhost:38081",
						Usage:       "URL to the RPC server of the Monero Daemon",
						Destination: &daemonURL,
					},
					&cli.IntFlag{
						Name:        "max-extra-ancestor-blocks",
						Aliases:     []string{"extra-ancestors", "ea"},
						Value:       0,
						Usage:       "Max number of extra ancestor blocks to include with each published block",
						Destination: &maxExtraAncestors,
					},
				},
				Action: func(c *cli.Context) error {
					blockHash := c.Args().First()
					if blockHash == "" {
						return fmt.Errorf("block command requires a blockHash argument")
					}

					rpcClient := NewRPCClient(daemonURL)
					evPublisher := NewNatsPublishingClient(natsURL)
					return ProcessBlockHash(blockHash, maxExtraAncestors, ignoreBelowHeight, rpcClient, evPublisher)
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
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
func ProcessTxid(txid string, ignoreBelowheight int, rc TxGetter, nc TxEventPublisher) error {
	ctx := context.Background()
	transfers, err := rc.GetTransferByTxid(ctx, txid)
	if err != nil {
		return err
	}

	tx, err := RpcTxToTx(transfers)
	if err != nil {
		return err
	}

	if tx.Height < ignoreBelowheight {
		// The Tx is below ignoring height. It won't be
		// published to NATS
		return nil
	}

	return nc.PushTxEvent(*tx)
}

type BlockGetter interface {
	GetBlockByHash(context.Context, string) (*RpcBlock, error)
	GetBlockHeadersRange(context.Context, int, int) ([]RpcBlockHeader, error)
}

type BlockEventPublisher interface {
	PushBlockEvent(Block) error
}

func ProcessBlockHash(blockHash string, maxExtraAncestors, ignoreBelowHeight int, bg BlockGetter, nc BlockEventPublisher) error {
	ctx := context.Background()
	rpcBlock, err := bg.GetBlockByHash(ctx, blockHash)
	if err != nil {
		return err
	}

	blk := RpcBlockToBlock(*rpcBlock)

	if blk.Height < ignoreBelowHeight {
		// Block is below ignoring height. Its ancestors won't be fetched
		// and it won't be published to NATS
		return nil
	}

	if blk.Height == 0 {
		// Blocks is Genesis Block. It has no ancestors
		return nc.PushBlockEvent(blk)
	}

	end := blk.Height - 1
	start := blk.Height - maxExtraAncestors
	if start < 0 {
		start = 0
	}

	blocks, err := bg.GetBlockHeadersRange(ctx, start, end)
	if err != nil {
		return err
	}

	prevHashes := []string{}
	for _, b := range blocks {
		prevHashes = append(prevHashes, b.Hash)
	}
	blk.PrevHashes = prevHashes

	return nc.PushBlockEvent(blk)
}
