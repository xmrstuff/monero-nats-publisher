package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type BlockGetterBroken struct {
	CallsCount      int
	PassedBlockHash string
}

func (g *BlockGetterBroken) GetBlockByHash(c context.Context, h string) (*RpcBlock, error) {
	g.PassedBlockHash = h
	g.CallsCount++
	return nil, fmt.Errorf("Dummy Error")
}

type BlockEventPublisherRecording struct {
	CallsCount  int
	PassedBlock *Block
}

func (p *BlockEventPublisherRecording) PushBlockEvent(b Block) error {
	p.PassedBlock = &b
	p.CallsCount++
	return nil
}

type BlockGetterRecording struct {
	CallsCount      int
	PassedBlockHash string
}

func (g *BlockGetterRecording) GetBlockByHash(c context.Context, h string) (*RpcBlock, error) {
	g.PassedBlockHash = h
	g.CallsCount++
	return &RpcBlock{BlockHeader: RpcBlockHeader{Hash: h}}, nil
}

type BlockEvPublisherBreaking struct {
	CallsCount  int
	PassedBlock *Block
}

func (p *BlockEvPublisherBreaking) PushBlockEvent(b Block) error {
	p.CallsCount++
	p.PassedBlock = &b
	return fmt.Errorf("Dummy Error")
}

func TestProcessBlockHash(t *testing.T) {

	t.Run("Success", func(t *testing.T) {
		blockHash := "dummy hash"
		rpcClient := BlockGetterRecording{}
		evPublisher := BlockEventPublisherRecording{}

		err := ProcessBlockHash(blockHash, &rpcClient, &evPublisher)
		assert.Nil(t, err)
	})

	t.Run("RPC call fails", func(t *testing.T) {
		blockHash := "dummy hash"
		rpcClient := BlockGetterBroken{}
		evPublisher := BlockEventPublisherRecording{}

		err := ProcessBlockHash(blockHash, &rpcClient, &evPublisher)
		assert.Error(t, err)

		assert.Equal(t, 1, rpcClient.CallsCount)
		assert.Equal(t, blockHash, rpcClient.PassedBlockHash)

		assert.Equal(t, 0, evPublisher.CallsCount)
	})

	t.Run("Event publishing fails", func(t *testing.T) {
		blockHash := "dummy hash"
		rpcClient := BlockGetterRecording{}
		evPublisher := BlockEvPublisherBreaking{}

		err := ProcessBlockHash(blockHash, &rpcClient, &evPublisher)
		assert.Error(t, err)

		assert.Equal(t, 1, rpcClient.CallsCount)
		assert.Equal(t, blockHash, rpcClient.PassedBlockHash)

		assert.Equal(t, 1, evPublisher.CallsCount)
		assert.Equal(t, blockHash, evPublisher.PassedBlock.Hash)
	})
}
