package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type rpcResponse struct {
	ID      string    `json:"id"`
	JSONRPC string    `json:"jsonrpc"`
	Result  *Tx       `json:"result"`
	Error   *rpcError `json:"error"`
}

type getTransferParams struct {
	TXID string `json:"txid"`
}

type rpcRequest struct {
	ID      string            `json:"id"`
	JSONRPC string            `json:"jsonrpc"`
	Method  string            `json:"method"`
	Params  getTransferParams `json:"params"`
}

func newRPCRequest(txid string) rpcRequest {
	return rpcRequest{
		Params: getTransferParams{
			TXID: txid,
		},
	}
}

type rpcClient struct {
	HTTPClient *http.Client
	Host       string
	BasePath   string
}

func (c *rpcClient) BaseURL() string {
	return fmt.Sprintf("%s/%s", c.Host, c.BasePath)
}

func (c *rpcClient) GetTransferByTxid(ctx context.Context, txid string) (*Tx, error) {
	reqData := newRPCRequest(txid)
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(reqData); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL(), buf)
	if err != nil {
		return nil, err
	}

	rawResp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer rawResp.Body.Close()

	if rawResp.StatusCode != 200 {
		// RPC returns 200 unless something went really wrong
		return nil, fmt.Errorf("Unknown Error. Code %d", rawResp.StatusCode)
	}

	resp := rpcResponse{}
	if err := json.NewDecoder(rawResp.Body).Decode(&resp); err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("RPC Error. %v", resp.Error)
	}

	return resp.Result, nil
}

func newClient(host string) rpcClient {
	return rpcClient{
		Host:     host,
		BasePath: "json_rpc",
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}
