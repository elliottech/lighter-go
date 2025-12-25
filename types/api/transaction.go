package api

// Tx represents a transaction
type Tx struct {
	Hash           string `json:"hash"`
	Type           TxType `json:"type"`
	TypeName       string `json:"type_name,omitempty"`
	AccountIndex   int64  `json:"account_index"`
	ApiKeyIndex    uint8  `json:"api_key_index,omitempty"`
	Nonce          int64  `json:"nonce"`
	Status         string `json:"status"` // "pending", "confirmed", "failed"
	SequenceIndex  int64  `json:"sequence_index,omitempty"`
	BlockHeight    int64  `json:"block_height,omitempty"`
	Timestamp      int64  `json:"timestamp"`
	Data           string `json:"data,omitempty"` // JSON-encoded tx info
}

// EnrichedTx includes parsed transaction details
type EnrichedTx struct {
	BaseResponse
	Tx
	ParsedData interface{} `json:"parsed_data,omitempty"`
	Events     []TxEvent   `json:"events,omitempty"`
}

// TxEvent represents an event from a transaction
type TxEvent struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp"`
}

// Txs is the response for transaction queries
type Txs struct {
	BaseResponse
	Txs    []Tx   `json:"txs"`
	Cursor Cursor `json:"cursor,omitempty"`
}

// TxHash represents a transaction hash response
type TxHash struct {
	BaseResponse
	Hash string `json:"hash"`
}

// TxStatus represents transaction status
type TxStatus struct {
	BaseResponse
	Hash        string `json:"hash"`
	Status      string `json:"status"` // "pending", "confirmed", "failed"
	BlockHeight int64  `json:"block_height,omitempty"`
	Confirmations int64 `json:"confirmations,omitempty"`
	Error       string `json:"error,omitempty"`
}

// RespSendTx is the response for sending a transaction
type RespSendTx struct {
	BaseResponse
	TxHash        string `json:"tx_hash"`
	SequenceIndex int64  `json:"sequence_index,omitempty"`
}

// RespSendTxBatch is the response for sending batch transactions
type RespSendTxBatch struct {
	BaseResponse
	TxHashes []string `json:"tx_hashes"`
	Errors   []string `json:"errors,omitempty"`
}

// SendTxRequest is the request body for sending a transaction
type SendTxRequest struct {
	TxType          uint8  `json:"tx_type"`
	TxInfo          string `json:"tx_info"` // JSON-encoded transaction info
	PriceProtection *PriceProtection `json:"price_protection,omitempty"`
}

// SendTxBatchRequest is the request body for sending batch transactions
type SendTxBatchRequest struct {
	TxTypes []uint8  `json:"tx_types"`
	TxInfos []string `json:"tx_infos"` // JSON-encoded transaction infos
}

// PriceProtection represents price protection parameters
type PriceProtection struct {
	MaxPrice string `json:"max_price,omitempty"`
	MinPrice string `json:"min_price,omitempty"`
}

// NextNonce represents the next nonce response
type NextNonce struct {
	BaseResponse
	Nonce int64 `json:"nonce"`
}

// RespWithdrawalDelay is the response for withdrawal delay query
type RespWithdrawalDelay struct {
	BaseResponse
	NormalDelaySeconds int64 `json:"normal_delay_seconds"`
	FastDelaySeconds   int64 `json:"fast_delay_seconds"`
}
