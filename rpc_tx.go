package main

import "context"

type RpcTx struct {
	TXID          string `json:"txid"`
	Address       string `json:"address"`
	Amount        int    `json:"amount"`
	Confirmations int    `json:"confirmations"`
	Height        int    `json:"height"`
	Timestamp     int    `json:"timestamp"`
	UnlockTime    int    `json:"unlock_time"`
	Type          string `json:"type"`
}

func (t *RpcTx) IsIncoming() bool {
	validTypes := map[string]int{"in": 1, "pool": 1}
	_, ok := validTypes[t.Type]
	return ok
}

type RpcResultTransfers struct {
	Transfers []RpcTx `json:"transfers"`
}

type GetTransferParams struct {
	TXID string `json:"txid"`
}

func NewGetTransferPayload(txid string) RPCRequestPayload {
	return RPCRequestPayload{
		ID:      "0",
		JSONRPC: "2.0",
		Method:  "get_transfer_by_txid",
		Params: GetTransferParams{
			TXID: txid,
		},
	}
}

func (c *RPCClient) GetTransferByTxid(ctx context.Context, txid string) ([]RpcTx, error) {
	rpcReq := NewGetTransferPayload(txid)
	result := RpcResultTransfers{}
	err := c.MakeRequest(ctx, rpcReq, &result)
	if err != nil {
		return nil, err
	}
	return result.Transfers, nil
}
