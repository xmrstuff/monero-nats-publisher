package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGetTransferPayload(t *testing.T) {
	txid := "dummy txid"
	req := NewGetTransferPayload(txid)
	assert.Equal(t, "0", req.ID)
	assert.Equal(t, "2.0", req.JSONRPC)
	assert.Equal(t, "get_transfer_by_txid", req.Method)
	params, ok := req.Params.(GetTransferParams)
	assert.True(t, ok)
	assert.Equal(t, txid, params.TXID)
}

func TestGetTransferByTxidSuccess(t *testing.T) {
	txid := "3c05afeedd910877a9f23e25de2204fdb85b4b26c0d7da74fdea1a8ff25bddf3"
	jsonResp := fmt.Sprintf(`
		{
			"result": {
				"transfers": [
					{
						"txid": "%s",
						"height": 300,
						"timestamp": 1535918400,
						"address": "addr1",
						"amount": 1,
						"confirmations": 20
					},
					{
						"txid": "%s",
						"height": 300,
						"timestamp": 1535918400,
						"address": "addr2",
						"amount": 2,
						"confirmations": 20
					}
				]
			}
		}
	`, txid, txid)
	server := makeServer(t, "/json_rpc", "POST", txid, 200, jsonResp)
	defer server.Close()

	client := NewRPCClient(server.URL)
	client.HTTPClient = server.Client()

	ctx := context.Background()
	transfers, err := client.GetTransferByTxid(ctx, txid)
	assert.Nil(t, err)
	assert.Equal(t, txid, transfers[0].TXID)
	assert.Equal(t, txid, transfers[1].TXID)
	assert.Equal(t, "addr1", transfers[0].Address)
	assert.Equal(t, "addr2", transfers[1].Address)
	assert.Equal(t, 1, transfers[0].Amount)
	assert.Equal(t, 2, transfers[1].Amount)
}

func TestGetTransferByTxidErrors(t *testing.T) {
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
			txid := "invalid"
			server := makeServer(t, "/json_rpc", "POST", txid, c.RespCode, c.JSONResp)
			defer server.Close()

			client := NewRPCClient(server.URL)
			client.HTTPClient = server.Client()

			ctx := context.Background()
			tx, err := client.GetTransferByTxid(ctx, txid)
			assert.Nil(t, tx)
			assert.Error(t, err)
		})
	}
}
