package main

import (
	"fmt"
	"time"
)

//easyjson:json
type TransactionBlock struct {
	Jsonrpc string  `json:"jsonrpc"`
	ID      string  `json:"id"`
	Error   *Error  `json:"error,omitempty"`
	Result  *Result `json:"result"`
}

//easyjson:json
type Result struct {
	Hash         string        `json:"hash"`
	Height       string        `json:"height"`
	Time         time.Time     `json:"time"`
	NumTxs       string        `json:"num_txs"`
	TotalTxs     string        `json:"total_txs"`
	Transactions []Transaction `json:"transactions"`
	BlockReward  string        `json:"block_reward"`
	Size         string        `json:"size"`
	Proposer     string        `json:"proposer"`
	Validators   []Validator   `json:"validators"`
}

//easyjson:json
type Transaction struct {
	Hash        string    `json:"hash" db:"hash"`
	RawTx       string    `json:"raw_tx" db:"raw_tx"`
	From        string    `json:"from" db:"from_ad"`
	Nonce       string    `json:"nonce" db:"nonce"`
	GasPrice    int       `json:"gas_price" db:"gas_price"`
	TxTime      time.Time `json:"tx_time" db:"tx_data"`
	Type        int       `json:"type" db:"type"`
	Data        *Data     `json:"data"`
	Coin        string    `json:"coin" db:"coin"`
	To          string    `json:"to" db:"to_ad"`
	Value       float64   `json:"value" db:"value"`
	Payload     string    `json:"payload" db:"payload"`
	ServiceData string    `json:"service_data" db:"service_data"`
	Gas         string    `json:"gas" db:"gas"`
	GasCoin     string    `json:"gas_coin" db:"gas_coin"`
	CreatedAt   time.Time `db:"created_at"`
	//Tags        *Tag      `json:"tags" db:"tags"`
}

//easyjson:json
type Data struct {
	Coin  string `json:"coin"`
	To    string `json:"to"`
	Value string `json:"value"`
}

//easyjson:json
type Tag struct {
	TxType string `json:"tx.type"`
	TxFrom string `json:"tx.from"`
	TxTo   string `json:"tx.to"`
	TxCoin string `json:"tx.coin"`
}

//easyjson:json
type Validator struct {
	PubKey string `json:"pub_key"`
	Signed bool   `json:"signed"`
}

//easyjson:json
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("error from minter api: code:%d, msg: %s, data: %s", e.Code, e.Message, e.Data)
}
