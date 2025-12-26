package http

const (
	CodeOK = 200
)

type ResultCode struct {
	Code    int32  `json:"code"`
	Message string `json:"message,omitempty"`
}

type NextNonce struct {
	ResultCode
	Nonce int64 `json:"nonce"`
}

type ApiKey struct {
	AccountIndex int64  `json:"account_index"`
	ApiKeyIndex  uint8  `json:"api_key_index"`
	Nonce        int64  `json:"nonce"`
	PublicKey    string `json:"public_key"`
}

type AccountApiKeys struct {
	ResultCode
	ApiKeys []*ApiKey `json:"api_keys"`
}
