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
	var maxBlockAncestors int
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "nats-url",
				Aliases:     []string{"nats", "n"},
				Value:       "http://localhost:4222",
				Usage:       "URL to the NATS Streaming Server",
				Destination: &natsURL,
			},
		},
		Commands: []*cli.Command{
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
					return ProcessTxid(txid, rpcClient, evPublisher)
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
						Name:        "max-block-ancestors",
						Aliases:     []string{"ancestors", "ma"},
						Value:       2,
						Usage:       "Max number of ancestor blocks to fetch with each block",
						Destination: &maxBlockAncestors,
					},
				},
				Action: func(c *cli.Context) error {
					blockHash := c.Args().First()
					if blockHash == "" {
						return fmt.Errorf("block command requires a blockHash argument")
					}

					rpcClient := NewRPCClient(daemonURL)
					evPublisher := NewNatsPublishingClient(natsURL)
					return ProcessBlockHash(blockHash, maxBlockAncestors, rpcClient, evPublisher)
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

func ProcessBlockHash(blockHash string, maxAncestors int, rc BlockGetter, nc BlockEventPublisher) error {
	ctx := context.Background()
	rpcBlock, err := rc.GetBlockByHash(ctx, blockHash)
	if err != nil {
		return err
	}

	blk := RpcBlockToBlock(*rpcBlock)

	if len(blk.PrevHashes) > 0 {
		ancestorHash := blk.PrevHashes[0]
		for i := 0; i < maxAncestors; i++ {
			ancestor, err := rc.GetBlockByHash(ctx, ancestorHash)
			if err != nil {
				return err
			}
			ancestorHash = ancestor.BlockHeader.PrevHash
			if ancestorHash == "" {
				break
			}
			blk.PrevHashes = append(blk.PrevHashes, ancestorHash)
		}
	}

	return nc.PushBlockEvent(blk)
}
