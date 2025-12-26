package api

// Block represents a block on the chain
type Block struct {
	Height        int64  `json:"height"`
	Hash          string `json:"hash"`
	ParentHash    string `json:"parent_hash,omitempty"`
	StateRoot     string `json:"state_root,omitempty"`
	TxCount       int    `json:"tx_count"`
	Timestamp     int64  `json:"timestamp"`
	Commitment    string `json:"commitment,omitempty"`
	ProposerIndex int64  `json:"proposer_index,omitempty"`
}

// Blocks is the response for block queries
type Blocks struct {
	BaseResponse
	Blocks []Block `json:"blocks"`
}

// CurrentHeight is the response for current height query
type CurrentHeight struct {
	BaseResponse
	Height    int64 `json:"height"`
	Timestamp int64 `json:"timestamp"`
}

// BlockTxs is the response for block transactions query
type BlockTxs struct {
	BaseResponse
	BlockHeight int64 `json:"block_height"`
	Txs         []Tx  `json:"txs"`
}

// BlockInfo includes block with additional details
type BlockInfo struct {
	Block
	FirstTxIndex int64 `json:"first_tx_index,omitempty"`
	LastTxIndex  int64 `json:"last_tx_index,omitempty"`
}
