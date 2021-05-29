package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockedGetBlockReturn struct {
	b *RpcBlock
	e error
}

type MockedGetBlocksRangeReturn struct {
	b []RpcBlockHeader
	e error
}

type MockedGetBlocksRangeArg struct {
	Start int
	End   int
}

type MockedBlockGetter struct {
	GetBlockCallsCount       int
	HashArgs                 []string
	GetBlockReturns          []MockedGetBlockReturn
	GetBlocksRangeCallsCount int
	GetBlocksRangeArgs       []MockedGetBlocksRangeArg
	GetBlocksRangeReturns    []MockedGetBlocksRangeReturn
}

func (g *MockedBlockGetter) GetBlockByHash(c context.Context, h string) (*RpcBlock, error) {
	g.HashArgs = append(g.HashArgs, h)

	g.GetBlockCallsCount++

	// pop first entry
	result := g.GetBlockReturns[0]
	if len(g.GetBlockReturns) > 1 {
		g.GetBlockReturns = g.GetBlockReturns[1:]
	} else {
		g.GetBlockReturns = []MockedGetBlockReturn{}
	}

	return result.b, result.e
}

func (g *MockedBlockGetter) GetBlockHeadersRange(c context.Context, start, end int) ([]RpcBlockHeader, error) {
	arg := MockedGetBlocksRangeArg{start, end}
	g.GetBlocksRangeArgs = append(g.GetBlocksRangeArgs, arg)

	g.GetBlocksRangeCallsCount++

	// pop first entry
	result := g.GetBlocksRangeReturns[0]
	if len(g.GetBlocksRangeReturns) > 1 {
		g.GetBlocksRangeReturns = g.GetBlocksRangeReturns[1:]
	} else {
		g.GetBlocksRangeReturns = []MockedGetBlocksRangeReturn{}
	}

	return result.b, result.e
}

type MockedBlockEventPublisher struct {
	CallsCount   int
	PassedBlocks []Block
	Returns      []error
}

func (p *MockedBlockEventPublisher) PushBlockEvent(b Block) error {
	p.CallsCount++

	p.PassedBlocks = append(p.PassedBlocks, b)

	// pop first entry
	result := p.Returns[0]
	if len(p.Returns) > 1 {
		p.Returns = p.Returns[1:]
	} else {
		p.Returns = []error{}
	}

	return result
}

func TestProcessBlockHash(t *testing.T) {

	t.Run("Success, ignoring ancestors", func(t *testing.T) {
		blockHash := "block X"
		rpcClient := MockedBlockGetter{
			GetBlockReturns: []MockedGetBlockReturn{
				{
					b: &RpcBlock{BlockHeader: RpcBlockHeader{Hash: blockHash}},
					e: nil,
				},
				{
					b: &RpcBlock{BlockHeader: RpcBlockHeader{Hash: "Block X-1"}},
					e: nil,
				},
			},
		}
		evPublisher := MockedBlockEventPublisher{
			Returns: []error{nil},
		}

		maxAncestors := 0      // Ignoring extra ancestors
		ignoreBelowHeight := 0 // Not ignoring any height
		err := ProcessBlockHash(blockHash, maxAncestors, ignoreBelowHeight, &rpcClient, &evPublisher)
		assert.Nil(t, err)

		assert.Equal(t, 1, rpcClient.GetBlockCallsCount)
		assert.Equal(t, blockHash, rpcClient.HashArgs[0])

		assert.Equal(t, 1, evPublisher.CallsCount)
	})

	t.Run("Success, more ancestors than requested", func(t *testing.T) {
		hashes := []string{"block 2", "block 3", "block 4", "block 5"}
		heights := []int{2, 3, 4, 5}

		rpcClient := MockedBlockGetter{
			GetBlockReturns: []MockedGetBlockReturn{
				{
					b: &RpcBlock{
						BlockHeader: RpcBlockHeader{Hash: hashes[3], Height: heights[3], PrevHash: hashes[2]},
					},
					e: nil,
				},
			},
			GetBlocksRangeReturns: []MockedGetBlocksRangeReturn{
				{
					e: nil,
					b: []RpcBlockHeader{
						{Hash: hashes[2], Height: heights[2]},
						{Hash: hashes[1], Height: heights[1]},
					},
				},
			},
		}
		evPublisher := MockedBlockEventPublisher{
			Returns: []error{nil},
		}

		maxAncestors := 2
		ignoreBelowHeight := 0 // Not ignoring any height
		err := ProcessBlockHash(hashes[3], maxAncestors, ignoreBelowHeight, &rpcClient, &evPublisher)
		assert.Nil(t, err)

		assert.Equal(t, 1, rpcClient.GetBlockCallsCount)
		assert.Equal(t, []string{hashes[3]}, rpcClient.HashArgs)

		assert.Equal(t, 1, rpcClient.GetBlocksRangeCallsCount)
		assert.Equal(t, heights[1], rpcClient.GetBlocksRangeArgs[0].Start)
		assert.Equal(t, heights[2], rpcClient.GetBlocksRangeArgs[0].End)

		assert.Equal(t, 1, evPublisher.CallsCount)
		assert.Equal(t, hashes[3], evPublisher.PassedBlocks[0].Hash)
		assert.Equal(t, []string{hashes[2], hashes[1]}, evPublisher.PassedBlocks[0].PrevHashes)
	})

	t.Run("Success, less ancestors than requested", func(t *testing.T) {
		hashes := []string{"block genesis", "block 1", "block 2", "block 3"}
		heights := []int{0, 1, 2, 3}

		rpcClient := MockedBlockGetter{
			GetBlockReturns: []MockedGetBlockReturn{
				{
					b: &RpcBlock{
						BlockHeader: RpcBlockHeader{Hash: hashes[3], Height: heights[3], PrevHash: hashes[2]},
					},
					e: nil,
				},
			},
			GetBlocksRangeReturns: []MockedGetBlocksRangeReturn{
				{
					e: nil,
					b: []RpcBlockHeader{
						{Hash: hashes[2], Height: heights[2]},
						{Hash: hashes[1], Height: heights[1]},
						{Hash: hashes[0], Height: heights[0]},
					},
				},
			},
		}
		evPublisher := MockedBlockEventPublisher{
			Returns: []error{nil},
		}

		maxAncestors := 5
		ignoreBelowHeight := 0 // Not ignoring any height
		err := ProcessBlockHash(hashes[3], maxAncestors, ignoreBelowHeight, &rpcClient, &evPublisher)
		assert.Nil(t, err)

		assert.Equal(t, 1, rpcClient.GetBlockCallsCount)
		assert.Equal(t, []string{hashes[3]}, rpcClient.HashArgs)

		assert.Equal(t, 1, rpcClient.GetBlocksRangeCallsCount)
		assert.Equal(t, heights[0], rpcClient.GetBlocksRangeArgs[0].Start)
		assert.Equal(t, heights[2], rpcClient.GetBlocksRangeArgs[0].End)

		assert.Equal(t, 1, evPublisher.CallsCount)
		assert.Equal(t, hashes[3], evPublisher.PassedBlocks[0].Hash)
		assert.Equal(t, []string{hashes[2], hashes[1], hashes[0]}, evPublisher.PassedBlocks[0].PrevHashes)
	})

	t.Run("Success, block below ignoring height", func(t *testing.T) {
		hashes := []string{"block 2", "block 3", "block 4", "block 5"}
		heights := []int{2, 3, 4, 5}

		rpcClient := MockedBlockGetter{
			GetBlockReturns: []MockedGetBlockReturn{
				{
					b: &RpcBlock{
						BlockHeader: RpcBlockHeader{Hash: hashes[3], Height: heights[3], PrevHash: hashes[2]},
					},
					e: nil,
				},
			},
			GetBlocksRangeReturns: []MockedGetBlocksRangeReturn{
				{
					e: nil,
					b: []RpcBlockHeader{
						{Hash: hashes[2], Height: heights[2]},
						{Hash: hashes[1], Height: heights[1]},
					},
				},
			},
		}
		evPublisher := MockedBlockEventPublisher{
			Returns: []error{nil},
		}

		maxAncestors := 2
		ignoreBelowHeight := heights[3] + 1 // the block will be ignored
		err := ProcessBlockHash(hashes[3], maxAncestors, ignoreBelowHeight, &rpcClient, &evPublisher)
		assert.Nil(t, err)

		assert.Equal(t, 1, rpcClient.GetBlockCallsCount)
		assert.Equal(t, []string{hashes[3]}, rpcClient.HashArgs)

		// The ancestors were not fetched, and the block was not published, because
		// it's below ignoring height
		assert.Equal(t, 0, rpcClient.GetBlocksRangeCallsCount)
		assert.Equal(t, 0, evPublisher.CallsCount)
	})

	t.Run("GetBlockByHash RPC call fails", func(t *testing.T) {
		blockHash := "dummy hash"
		rpcClient := MockedBlockGetter{
			GetBlockReturns: []MockedGetBlockReturn{
				{
					b: nil,
					e: fmt.Errorf("Dummy error"),
				},
			},
		}

		evPublisher := MockedBlockEventPublisher{
			Returns: []error{nil},
		}

		err := ProcessBlockHash(blockHash, 0, 0, &rpcClient, &evPublisher)
		assert.Error(t, err)

		assert.Equal(t, 1, rpcClient.GetBlockCallsCount)
		assert.Equal(t, blockHash, rpcClient.HashArgs[0])

		assert.Equal(t, 0, rpcClient.GetBlocksRangeCallsCount)

		assert.Equal(t, 0, evPublisher.CallsCount)
	})

	t.Run("GetBlockHeadersRange RPC call fails", func(t *testing.T) {
		blockHash := "dummy hash"
		rpcClient := MockedBlockGetter{
			GetBlockReturns: []MockedGetBlockReturn{
				{
					b: &RpcBlock{
						BlockHeader: RpcBlockHeader{Hash: "block 5", Height: 5, PrevHash: "block 4"},
					},
					e: nil,
				},
			},
			GetBlocksRangeReturns: []MockedGetBlocksRangeReturn{
				{
					e: fmt.Errorf("Dummy error"),
					b: nil,
				},
			},
		}

		evPublisher := MockedBlockEventPublisher{
			Returns: []error{nil},
		}

		err := ProcessBlockHash(blockHash, 1, 0, &rpcClient, &evPublisher)
		assert.Error(t, err)

		assert.Equal(t, 1, rpcClient.GetBlockCallsCount)
		assert.Equal(t, blockHash, rpcClient.HashArgs[0])

		assert.Equal(t, 1, rpcClient.GetBlocksRangeCallsCount)
		assert.Equal(t, 4, rpcClient.GetBlocksRangeArgs[0].Start)
		assert.Equal(t, 4, rpcClient.GetBlocksRangeArgs[0].End)

		assert.Equal(t, 0, evPublisher.CallsCount)
	})

	t.Run("Event publishing fails", func(t *testing.T) {
		blockHash := "dummy hash"
		rpcClient := MockedBlockGetter{
			GetBlockReturns: []MockedGetBlockReturn{
				{
					b: &RpcBlock{BlockHeader: RpcBlockHeader{Hash: blockHash}},
					e: nil,
				},
			},
		}

		evPublisher := MockedBlockEventPublisher{
			Returns: []error{fmt.Errorf("Dummy Error")},
		}

		err := ProcessBlockHash(blockHash, 0, 0, &rpcClient, &evPublisher)
		assert.Error(t, err)

		assert.Equal(t, 1, rpcClient.GetBlockCallsCount)
		assert.Equal(t, blockHash, rpcClient.HashArgs[0])

		assert.Equal(t, 1, evPublisher.CallsCount)
		assert.Equal(t, blockHash, evPublisher.PassedBlocks[0].Hash)
	})
}
