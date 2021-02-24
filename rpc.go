package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type RpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type RpcResponse struct {
	ID      string    `json:"id"`
	JSONRPC string    `json:"jsonrpc"`
	Result  *Tx       `json:"result"`
	Error   *RpcError `json:"error"`
}

type GetTransferParams struct {
	TXID string `json:"txid"`
}

type RpcRequest struct {
	ID      string            `json:"id"`
	JSONRPC string            `json:"jsonrpc"`
	Method  string            `json:"method"`
	Params  GetTransferParams `json:"params"`
}

func NewRPCRequest(txid string) RpcRequest {
	return RpcRequest{
		ID:      "0",
		JSONRPC: "2.0",
		Method:  "get_transfer_by_txid",
		Params: GetTransferParams{
			TXID: txid,
		},
	}
}

type RPCClient struct {
	HTTPClient *http.Client
	Host       string
	BasePath   string
}

func (c *RPCClient) BaseURL() string {
	return fmt.Sprintf("%s/%s", c.Host, c.BasePath)
}

func (c *RPCClient) GetTransferByTxid(ctx context.Context, txid string) (*Tx, error) {
	reqData := NewRPCRequest(txid)
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(reqData); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL(), buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	rawResp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer rawResp.Body.Close()

	if rawResp.StatusCode != 200 {
		// RPC returns 200 unless something went really wrong
		return nil, fmt.Errorf("Unknown Error. Code %d", rawResp.StatusCode)
	}

	resp := RpcResponse{}
	if err := json.NewDecoder(rawResp.Body).Decode(&resp); err != nil {
		return nil, err
	}

	if resp.Result == nil && resp.Error == nil {
		return nil, fmt.Errorf("Unable to parse RPC response: %+v", resp)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("RPC Error. %+v", resp.Error)
	}

	return resp.Result, nil
}

func NewRPCClient(host string) *RPCClient {
	return &RPCClient{
		Host:     host,
		BasePath: "json_rpc",
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}
