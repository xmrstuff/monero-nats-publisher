package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGetBlockPayload(t *testing.T) {
	block_hash := "dummy hash"
	req := NewGetBlockPayload(block_hash)
	assert.Equal(t, "0", req.ID)
	assert.Equal(t, "2.0", req.JSONRPC)
	assert.Equal(t, "get_block", req.Method)
	params, ok := req.Params.(GetBlockParams)
	assert.True(t, ok)
	assert.Equal(t, block_hash, params.Hash)
}

func TestGetBlockByHashSuccess(t *testing.T) {
	blockHash := "hash of the block"
	height := 300
	prevHash := "hash of the previous block"
	txHashes := []string{"hash1", "hash2", "hash3"}
	jsonResp := fmt.Sprintf(`
		{
			"result": {
				"block_header": {
					"hash": "%s",
					"height": %d,
					"timestamp": 1535918400,
					"prev_hash": "%s"
				},
				"tx_hashes": ["%s", "%s", "%s"]
			}
		}
	`, blockHash, height, prevHash, txHashes[0], txHashes[1], txHashes[2])
	server := makeServer(t, "/json_rpc", "POST", blockHash, 200, jsonResp)
	defer server.Close()

	client := NewRPCClient(server.URL)
	client.HTTPClient = server.Client()

	ctx := context.Background()
	block, err := client.GetBlockByHash(ctx, blockHash)
	assert.Nil(t, err)
	assert.Equal(t, blockHash, block.BlockHeader.Hash)
	assert.Equal(t, height, block.BlockHeader.Height)
	assert.Equal(t, prevHash, block.BlockHeader.PrevHash)
	assert.Equal(t, txHashes[0], block.TxHashes[0])
	assert.Equal(t, txHashes[1], block.TxHashes[1])
	assert.Equal(t, txHashes[2], block.TxHashes[2])
}

func TestGetBlockByHashErrors(t *testing.T) {
	errorCases := []struct {
		Description string
		RespCode    int
		JSONResp    string
	}{
		{"Unexpected HTTP error", 500, ""},
		{"Malformed response payload", 200, "[]"},
		{"RPC Error", 200, `{"error": {"code": -8, "message": "some RPC error"}}`},
	}
	for _, c := range errorCases {
		t.Run(c.Description, func(t *testing.T) {
			blockHash := "invalid"
			server := makeServer(t, "/json_rpc", "POST", blockHash, c.RespCode, c.JSONResp)
			defer server.Close()

			client := NewRPCClient(server.URL)
			client.HTTPClient = server.Client()

			ctx := context.Background()
			tx, err := client.GetBlockByHash(ctx, blockHash)
			assert.Nil(t, tx)
			assert.Error(t, err)
		})
	}
}

func TestNewGetsBlocksRangePayload(t *testing.T) {
	start, end := 200, 30
	p := NewGetBlocksRangePayload(start, end)

	assert.Equal(t, "0", p.ID)
	assert.Equal(t, "2.0", p.JSONRPC)
	assert.Equal(t, "get_block_headers_range", p.Method)

	par, ok := p.Params.(GetBlocksRangeParams)
	assert.True(t, ok)
	assert.Equal(t, par.StartHeight, start)
	assert.Equal(t, par.EndHeight, end)
}

func TestGetBlockHeadersRangeSuccess(t *testing.T) {
	jsonResp := `
		{
			"result": {
				"headers": [
					{
						"hash": "block5",
						"height": 5,
						"timestamp": 1535918400,
						"prev_hash": "block4"
					},
					{
						"hash": "block4",
						"height": 4,
						"timestamp": 1535916400,
						"prev_hash": "block3"
					},
					{
						"hash": "block3",
						"height": 3,
						"timestamp": 1535914400,
						"prev_hash": "block2"
					}
				]
			}
		}
	`
	blocksRange := "" // TODO: This is not used by makeServer anymore. remove
	server := makeServer(t, "/json_rpc", "POST", blocksRange, 200, jsonResp)
	defer server.Close()

	client := NewRPCClient(server.URL)
	client.HTTPClient = server.Client()

	ctx := context.Background()
	blocks, err := client.GetBlockHeadersRange(ctx, 3, 5)

	assert.Nil(t, err)
	assert.Equal(t, len(blocks), 3)
	for _, block := range blocks {
		assert.NotNil(t, block.Hash)
		assert.NotNil(t, block.Height)
		assert.NotNil(t, block.Timestamp)
		assert.NotNil(t, block.PrevHash)
	}
}

func TestGetBlockHeadersRangeErrors(t *testing.T) {
	errorCases := []struct {
		Description string
		RespCode    int
		JSONResp    string
	}{
		{"Unexpected HTTP error", 500, ""},
		{"Malformed response payload", 200, "[]"},
		{"RPC Error", 200, `{"error": {"code": -8, "message": "some RPC error"}}`},
	}
	for _, c := range errorCases {
		t.Run(c.Description, func(t *testing.T) {
			server := makeServer(t, "/json_rpc", "POST", "", c.RespCode, c.JSONResp)
			defer server.Close()

			client := NewRPCClient(server.URL)
			client.HTTPClient = server.Client()

			ctx := context.Background()
			blocks, err := client.GetBlockHeadersRange(ctx, 3, 5)
			assert.Nil(t, blocks)
			assert.Error(t, err)
		})
	}
}
