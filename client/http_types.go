package client

const (
	CodeOK = 200
)

type ResultCode struct {
	Code    int32  `json:"code,example=200"`
	Message string `json:"message,omitempty"`
}

type NextNonce struct {
	ResultCode
	Nonce int64 `json:"nonce,example=722"`
}

type AccountByL1Address struct {
	ResultCode
	L1Address   string                  `json:"l1_address"`
	SubAccounts []SubAccountByL1Address `json:"sub_accounts"`
}

type SubAccountByL1Address struct {
	// The code field is also returned in the SubAccounts structure, but with code = 0
	// SubAccounts结构里也会返回code字段，但是code=0
	ResultCode

	// Lighter API users can operate under a Standard or Premium accounts.
	// The Standard account is fee-less.
	// Premium accounts pay 0.2 bps maker and 2 bps taker fees. Find out more in Account Types.
	AccountType             uint8  `json:"account_type"`
	Index                   int64  `json:"index"`
	L1Address               string `json:"l1_address"`
	CancelAllTime           int64  `json:"cancel_all_time"`
	TotalOrderCount         int64  `json:"total_order_count"`
	TotalIsolatedOrderCount int64  `json:"total_isolated_order_count"`
	PendingOrderCount       int64  `json:"pending_order_count"`
	AvailableBalance        string `json:"available_balance"`
	Status                  uint8  `json:"status"`     // 1 is active, 0 is inactive.
	Collateral              string `json:"collateral"` // The amount of collateral in the account.
}

type ApiKey struct {
	AccountIndex int64  `json:"account_index,example=3"`
	ApiKeyIndex  uint8  `json:"api_key_index,example=0"`
	Nonce        int64  `json:"nonce,example=722"`
	PublicKey    string `json:"public_key"`
}

type AccountApiKeys struct {
	ResultCode
	ApiKeys []*ApiKey `json:"api_keys"`
}

type TxHash struct {
	ResultCode
	TxHash string `json:"tx_hash,example=0x70997970C51812dc3A010C7d01b50e0d17dc79C8"`
}

type TransferFeeInfo struct {
	ResultCode
	TransferFee int64 `json:"transfer_fee_usdc"`
}
