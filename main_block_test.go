package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockedBlockGetterReturn struct {
	b *RpcBlock
	e error
}
type MockedBlockGetter struct {
	CallsCount int
	HashArgs   []string
	Returns    []MockedBlockGetterReturn
}

func (g *MockedBlockGetter) GetBlockByHash(c context.Context, h string) (*RpcBlock, error) {
	g.HashArgs = append(g.HashArgs, h)
	g.CallsCount++
	result := g.Returns[0]
	g.Returns = g.Returns[1:]
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
	result := p.Returns[0]
	p.Returns = p.Returns[1:]
	return result
}

func TestProcessBlockHash(t *testing.T) {

	t.Run("Success, ignoring ancestors", func(t *testing.T) {
		blockHash := "block X"
		rpcClient := MockedBlockGetter{
			Returns: []MockedBlockGetterReturn{
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

		maxAncestors := 0 // Ignoring ancestors
		err := ProcessBlockHash(blockHash, maxAncestors, &rpcClient, &evPublisher)
		assert.Nil(t, err)

		assert.Equal(t, 1, rpcClient.CallsCount)
		assert.Equal(t, blockHash, rpcClient.HashArgs[0])

		assert.Equal(t, 1, evPublisher.CallsCount)
	})

	t.Run("Success, more ancestors than requested", func(t *testing.T) {
		hash0, hash1, hash2, hash3 := "block X-3", "block X-2", "block X-1", "block X"
		rpcClient := MockedBlockGetter{
			Returns: []MockedBlockGetterReturn{
				{
					b: &RpcBlock{
						BlockHeader: RpcBlockHeader{Hash: hash3, PrevHash: hash2},
					},
					e: nil,
				},
				{
					b: &RpcBlock{
						BlockHeader: RpcBlockHeader{Hash: hash2, PrevHash: hash1},
					},
					e: nil,
				},
				{
					b: &RpcBlock{
						BlockHeader: RpcBlockHeader{Hash: hash1, PrevHash: hash0},
					},
					e: nil,
				},
				{
					b: &RpcBlock{
						BlockHeader: RpcBlockHeader{Hash: hash0},
					},
					e: nil,
				},
			},
		}
		evPublisher := MockedBlockEventPublisher{
			Returns: []error{nil},
		}

		maxAncestors := 1
		err := ProcessBlockHash(hash3, maxAncestors, &rpcClient, &evPublisher)
		assert.Nil(t, err)

		assert.Equal(t, 2, rpcClient.CallsCount)
		assert.Equal(t, []string{hash3, hash2}, rpcClient.HashArgs)

		assert.Equal(t, 1, evPublisher.CallsCount)
		assert.Equal(t, hash3, evPublisher.PassedBlocks[0].Hash)
		assert.Equal(t, []string{hash2, hash1}, evPublisher.PassedBlocks[0].PrevHashes)
	})

	t.Run("Success, less ancestors than requested", func(t *testing.T) {
		hash1, hash2 := "block X-1", "block X"
		rpcClient := MockedBlockGetter{
			Returns: []MockedBlockGetterReturn{
				{
					b: &RpcBlock{
						BlockHeader: RpcBlockHeader{Hash: hash2, PrevHash: hash1},
					},
					e: nil,
				},
				{
					b: &RpcBlock{
						BlockHeader: RpcBlockHeader{Hash: hash1},
					},
					e: nil,
				},
			},
		}
		evPublisher := MockedBlockEventPublisher{
			Returns: []error{nil},
		}

		maxAncestors := 5
		err := ProcessBlockHash(hash2, maxAncestors, &rpcClient, &evPublisher)
		assert.Nil(t, err)

		assert.Equal(t, 2, rpcClient.CallsCount)
		assert.Equal(t, []string{hash2, hash1}, rpcClient.HashArgs)

		assert.Equal(t, 1, evPublisher.CallsCount)
		assert.Equal(t, hash2, evPublisher.PassedBlocks[0].Hash)
		assert.Equal(t, []string{hash1}, evPublisher.PassedBlocks[0].PrevHashes)
	})

	t.Run("RPC call fails", func(t *testing.T) {
		blockHash := "dummy hash"
		rpcClient := MockedBlockGetter{
			Returns: []MockedBlockGetterReturn{
				{
					b: nil,
					e: fmt.Errorf("Dummy error"),
				},
			},
		}

		evPublisher := MockedBlockEventPublisher{
			Returns: []error{nil},
		}

		err := ProcessBlockHash(blockHash, 0, &rpcClient, &evPublisher)
		assert.Error(t, err)

		assert.Equal(t, 1, rpcClient.CallsCount)
		assert.Equal(t, blockHash, rpcClient.HashArgs[0])

		assert.Equal(t, 0, evPublisher.CallsCount)
	})

	t.Run("Event publishing fails", func(t *testing.T) {
		blockHash := "dummy hash"
		rpcClient := MockedBlockGetter{
			Returns: []MockedBlockGetterReturn{
				{
					b: &RpcBlock{BlockHeader: RpcBlockHeader{Hash: blockHash}},
					e: nil,
				},
			},
		}

		evPublisher := MockedBlockEventPublisher{
			Returns: []error{fmt.Errorf("Dummy Error")},
		}

		err := ProcessBlockHash(blockHash, 0, &rpcClient, &evPublisher)
		assert.Error(t, err)

		assert.Equal(t, 1, rpcClient.CallsCount)
		assert.Equal(t, blockHash, rpcClient.HashArgs[0])

		assert.Equal(t, 1, evPublisher.CallsCount)
		assert.Equal(t, blockHash, evPublisher.PassedBlocks[0].Hash)
	})
}
