package main

import (
	"context"
)

type RpcBlockHeader struct {
	Hash      string `json:"hash"`
	Height    int    `json:"Height"`
	Timestamp int    `json:"timestamp"`
	PrevHash  string `json:"prev_hash"`
}

type RpcBlockHeaders struct {
	Headers []RpcBlockHeader `json:"headers"`
}

type RpcBlock struct {
	BlockHeader RpcBlockHeader `json:"block_header"`
	TxHashes    []string       `json:"tx_hashes"`
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

type GetBlocksRangeParams struct {
	StartHeight int `json:"start_height"`
	EndHeight   int `json:"end_height"`
}

func NewGetBlocksRangePayload(start, end int) RPCRequestPayload {
	return RPCRequestPayload{
		ID:      "0",
		JSONRPC: "2.0",
		Method:  "get_block_headers_range",
		Params: GetBlocksRangeParams{
			StartHeight: start,
			EndHeight:   end,
		},
	}
}

func (c *RPCClient) GetBlockHeadersRange(ctx context.Context, start, end int) ([]RpcBlockHeader, error) {
	rpcReq := NewGetBlocksRangePayload(start, end)
	rpcBlocks := RpcBlockHeaders{}
	if err := c.MakeRequest(ctx, rpcReq, &rpcBlocks); err != nil {
		return nil, err
	}
	return rpcBlocks.Headers, nil
}
