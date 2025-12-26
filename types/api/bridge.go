package api

// Bridge represents a bridge transaction
type Bridge struct {
	BridgeIndex    int64  `json:"bridge_index"`
	L1Address      string `json:"l1_address"`
	Direction      string `json:"direction"` // "deposit" or "withdraw"
	AssetIndex     int16  `json:"asset_index"`
	AssetSymbol    string `json:"asset_symbol,omitempty"`
	Amount         string `json:"amount"`
	Fee            string `json:"fee,omitempty"`
	L1TxHash       string `json:"l1_tx_hash,omitempty"`
	L2TxHash       string `json:"l2_tx_hash,omitempty"`
	Status         string `json:"status"` // "pending", "confirmed", "failed"
	IsFastBridge   bool   `json:"is_fast_bridge"`
	CreatedAt      int64  `json:"created_at"`
	ConfirmedAt    int64  `json:"confirmed_at,omitempty"`
}

// RespGetBridgesByL1Addr is the response for bridges by L1 address
type RespGetBridgesByL1Addr struct {
	BaseResponse
	Bridges []Bridge `json:"bridges"`
	Cursor  Cursor   `json:"cursor,omitempty"`
}

// RespGetIsNextBridgeFast is the response for fast bridge eligibility check
type RespGetIsNextBridgeFast struct {
	BaseResponse
	IsFast       bool   `json:"is_fast"`
	Reason       string `json:"reason,omitempty"`
	NextFastTime int64  `json:"next_fast_time,omitempty"`
}

// RespGetFastBridgeInfo is the response for fast bridge info
type RespGetFastBridgeInfo struct {
	BaseResponse
	Enabled           bool   `json:"enabled"`
	MinAmount         string `json:"min_amount"`
	MaxAmount         string `json:"max_amount"`
	Fee               string `json:"fee"`
	FeeRate           string `json:"fee_rate"`
	AvailableLiquidity string `json:"available_liquidity"`
	EstimatedTime     int64  `json:"estimated_time"` // In seconds
}

// BridgeSupportedNetwork represents a supported bridge network
type BridgeSupportedNetwork struct {
	ChainId          int64  `json:"chain_id"`
	Name             string `json:"name"`
	RpcUrl           string `json:"rpc_url,omitempty"`
	ExplorerUrl      string `json:"explorer_url,omitempty"`
	ContractAddress  string `json:"contract_address"`
	IsMainnet        bool   `json:"is_mainnet"`
	MinConfirmations int    `json:"min_confirmations"`
}

// BridgeNetworks is the response for supported networks query
type BridgeNetworks struct {
	BaseResponse
	Networks []BridgeSupportedNetwork `json:"networks"`
}
