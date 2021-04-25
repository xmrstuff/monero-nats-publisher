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
