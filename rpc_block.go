package main

import (
	"context"
)

type RpcBlock struct {
	BlockHeader struct {
		Hash      string `json:"hash"`
		Height    int    `json:"Height"`
		Timestamp int    `json:"timestamp"`
		PrevHash  string `json:"prev_hash"`
	} `json:"block_header"`
	TxHashes []string `json:"tx_hashes"`
}

type GetBlockParams struct {
	Hash string `json:"hash"`
}

func NewGetBlockPayload(hash string) RPCRequestPayload {
	return RPCRequestPayload{
		ID:      "0",
		JSONRPC: "2.0",
		Method:  "get_block",
		Params: GetBlockParams{
			Hash: hash,
		},
	}
}

func (c *RPCClient) GetBlockByHash(ctx context.Context, hash string) (*RpcBlock, error) {
	rpcReq := NewGetBlockPayload(hash)
	rpcBlock := RpcBlock{}
	err := c.MakeRequest(ctx, rpcReq, &rpcBlock)
	if err != nil {
		return nil, err
	}
	return &rpcBlock, nil
}
