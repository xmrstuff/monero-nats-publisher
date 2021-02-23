package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func makeServer(t *testing.T, expectedURL string, expectedMethod string, expectedTXID string, respStatus int, respBody string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Make sure we're using the expected URL and HTTP method
		assert.Equal(t, expectedURL, req.URL.String())
		assert.Equal(t, expectedMethod, req.Method)

		// Make sure we're passing the expected request body
		reqData := rpcRequest{}
		err := json.NewDecoder(req.Body).Decode(&reqData)
		assert.Nil(t, err)
		assert.Equal(t, expectedTXID, reqData.Params.TXID)

		rw.WriteHeader(respStatus)
		rw.Header().Set("Content-Type", "application/json")
		rw.Write([]byte(respBody))
	}))
}

func TestUnknownError(t *testing.T) {
	txid := "invalid"
	server := makeServer(t, "/json_rpc", "POST", txid, 500, "")
	defer server.Close()

	client := newClient(server.URL)
	client.HTTPClient = server.Client()

	ctx := context.Background()
	tx, err := client.GetTransferByTxid(ctx, txid)
	assert.Nil(t, tx)
	assert.Error(t, err)
}

func TestRPCError(t *testing.T) {
	txid := "stuff"
	server := makeServer(t, "/json_rpc", "POST", txid, 200, `{"error": {"code": -8, "message": "some RPC error"}}`)
	defer server.Close()

	client := newClient(server.URL)
	client.HTTPClient = server.Client()

	ctx := context.Background()
	tx, err := client.GetTransferByTxid(ctx, txid)
	assert.Nil(t, tx)
	assert.Error(t, err)
}

func TestSuccess(t *testing.T) {
	txid := "3c05afeedd910877a9f23e25de2204fdb85b4b26c0d7da74fdea1a8ff25bddf3"
	jsonResp := fmt.Sprintf(`
		{
			"result": {
				"txid": "%s",
				"height": 300,
				"timestamp": 1535918400,
				"destinations": [
					{
						"address": "addr1",
						"amount": 1
					},
					{
						"address": "addr2",
						"amount": 2
					}
				]
				
			}
		}
	`, txid)
	server := makeServer(t, "/json_rpc", "POST", txid, 200, jsonResp)
	defer server.Close()

	client := newClient(server.URL)
	client.HTTPClient = server.Client()

	ctx := context.Background()
	tx, err := client.GetTransferByTxid(ctx, txid)
	assert.Nil(t, err)
	assert.Equal(t, txid, tx.TXID)
	assert.Equal(t, "addr1", tx.Destinations[0].Address)
	assert.Equal(t, 1, tx.Destinations[0].Amount)
	assert.Equal(t, "addr2", tx.Destinations[1].Address)
	assert.Equal(t, 2, tx.Destinations[1].Amount)
}
