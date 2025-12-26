package api

// DepositEntry represents a single deposit
type DepositEntry struct {
	DepositIndex   int64  `json:"deposit_index"`
	AccountIndex   int64  `json:"account_index"`
	L1Address      string `json:"l1_address"`
	L1TxHash       string `json:"l1_tx_hash"`
	L2TxHash       string `json:"l2_tx_hash,omitempty"`
	AssetIndex     int16  `json:"asset_index"`
	AssetSymbol    string `json:"asset_symbol,omitempty"`
	Amount         string `json:"amount"`
	Status         string `json:"status"` // "pending", "confirmed", "failed"
	L1BlockNumber  int64  `json:"l1_block_number,omitempty"`
	L2BlockHeight  int64  `json:"l2_block_height,omitempty"`
	CreatedAt      int64  `json:"created_at"`
	ConfirmedAt    int64  `json:"confirmed_at,omitempty"`
}

// DepositHistory is the response for deposit history queries
type DepositHistory struct {
	BaseResponse
	Deposits []DepositEntry `json:"deposits"`
	Cursor   Cursor         `json:"cursor,omitempty"`
}

// WithdrawEntry represents a single withdrawal
type WithdrawEntry struct {
	WithdrawIndex  int64  `json:"withdraw_index"`
	AccountIndex   int64  `json:"account_index"`
	ToAddress      string `json:"to_address"`
	L2TxHash       string `json:"l2_tx_hash"`
	L1TxHash       string `json:"l1_tx_hash,omitempty"`
	AssetIndex     int16  `json:"asset_index"`
	AssetSymbol    string `json:"asset_symbol,omitempty"`
	Amount         string `json:"amount"`
	Fee            string `json:"fee,omitempty"`
	Status         string `json:"status"` // "pending", "processing", "confirmed", "failed"
	IsFastWithdraw bool   `json:"is_fast_withdraw"`
	L2BlockHeight  int64  `json:"l2_block_height,omitempty"`
	L1BlockNumber  int64  `json:"l1_block_number,omitempty"`
	CreatedAt      int64  `json:"created_at"`
	ConfirmedAt    int64  `json:"confirmed_at,omitempty"`
}

// WithdrawHistory is the response for withdrawal history queries
type WithdrawHistory struct {
	BaseResponse
	Withdrawals []WithdrawEntry `json:"withdrawals"`
	Cursor      Cursor          `json:"cursor,omitempty"`
}

// TransferEntry represents a single transfer
type TransferEntry struct {
	TransferIndex     int64  `json:"transfer_index"`
	FromAccountIndex  int64  `json:"from_account_index"`
	ToAccountIndex    int64  `json:"to_account_index"`
	AssetIndex        int16  `json:"asset_index"`
	AssetSymbol       string `json:"asset_symbol,omitempty"`
	Amount            string `json:"amount"`
	Fee               string `json:"fee,omitempty"`
	TxHash            string `json:"tx_hash"`
	BlockHeight       int64  `json:"block_height,omitempty"`
	Timestamp         int64  `json:"timestamp"`
}

// TransferHistory is the response for transfer history queries
type TransferHistory struct {
	BaseResponse
	Transfers []TransferEntry `json:"transfers"`
	Cursor    Cursor          `json:"cursor,omitempty"`
}

// ExportData is the response for data export
type ExportData struct {
	BaseResponse
	Type       string `json:"type"` // "funding" or "trade"
	Data       string `json:"data"` // CSV data
	StartTime  int64  `json:"start_time"`
	EndTime    int64  `json:"end_time"`
}
