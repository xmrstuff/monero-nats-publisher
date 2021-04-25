package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type RPCRequestPayload struct {
	ID      string      `json:"id"`
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

type RpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type RpcResponse struct {
	ID      string      `json:"id"`
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result"`
	Error   *RpcError   `json:"error"`
}
type RPCClient struct {
	HTTPClient *http.Client
	Host       string
	BasePath   string
}

func (c *RPCClient) BaseURL() string {
	return fmt.Sprintf("%s/%s", c.Host, c.BasePath)
}

func (c *RPCClient) MakeRequest(ctx context.Context, rpcReq interface{}, result interface{}) error {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(rpcReq); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL(), buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	rawResp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer rawResp.Body.Close()

	if rawResp.StatusCode != 200 {
		// RPC returns 200 unless something went really wrong
		return fmt.Errorf("Unknown Error. Code %d", rawResp.StatusCode)
	}

	resp := RpcResponse{
		Result: result,
	}
	if err := json.NewDecoder(rawResp.Body).Decode(&resp); err != nil {
		return err
	}

	if resp.Result == nil && resp.Error == nil {
		return fmt.Errorf("Unable to parse RPC response: %+v", resp)
	}

	if resp.Error != nil {
		return fmt.Errorf("RPC Error. %+v", resp.Error)
	}
	return nil
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
