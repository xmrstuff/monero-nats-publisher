package main

type Destination struct {
	Amount  int    `json:"amount"`
	Address string `json:"address"`
}

type Tx struct {
	TXID         string        `json:"txid"`
	Destinations []Destination `json:"destinations"`
	Height       int           `json:"height"`
	Timestamp    int           `json:"timestamp"`
	UnlockTime   int           `json:"unlock_time"`
}
