package main

// --- Common response types ---

type SignedTxResp struct {
	TxType        uint8  `json:"txType"`
	TxInfo        string `json:"txInfo,omitempty"`
	TxHash        string `json:"txHash,omitempty"`
	MessageToSign string `json:"messageToSign,omitempty"`
	Error         string `json:"error,omitempty"`
}

type StrOrErrResp struct {
	Result string `json:"result,omitempty"`
	Error  string `json:"error,omitempty"`
}

type ApiKeyResp struct {
	PrivateKey string `json:"privateKey,omitempty"`
	PublicKey  string `json:"publicKey,omitempty"`
	Error      string `json:"error,omitempty"`
}

type ErrorResp struct {
	Error string `json:"error"`
}

// --- Common fields embedded in most signing requests ---

type TransactFields struct {
	SkipNonce   uint8 `json:"skipNonce"`
	Nonce       int64 `json:"nonce"`
	ApiKeyIndex int   `json:"apiKeyIndex"`
	AccountIndex int64 `json:"accountIndex"`
}

type IntegratorFields struct {
	IntegratorAccountIndex int64  `json:"integratorAccountIndex"`
	IntegratorTakerFee     uint32 `json:"integratorTakerFee"`
	IntegratorMakerFee     uint32 `json:"integratorMakerFee"`
}

// --- Request types ---

type CreateClientReq struct {
	URL          string `json:"url"`
	PrivateKey   string `json:"privateKey"`
	ChainID      int    `json:"chainId"`
	ApiKeyIndex  int    `json:"apiKeyIndex"`
	AccountIndex int64  `json:"accountIndex"`
}

type CheckClientReq struct {
	ApiKeyIndex  int   `json:"apiKeyIndex"`
	AccountIndex int64 `json:"accountIndex"`
}

type ChangePubKeyReq struct {
	PubKey string `json:"pubKey"`
	TransactFields
}

type CreateOrderReq struct {
	MarketIndex      int16  `json:"marketIndex"`
	ClientOrderIndex int64  `json:"clientOrderIndex"`
	BaseAmount       int64  `json:"baseAmount"`
	Price            uint32 `json:"price"`
	IsAsk            uint8  `json:"isAsk"`
	OrderType        uint8  `json:"orderType"`
	TimeInForce      uint8  `json:"timeInForce"`
	ReduceOnly       uint8  `json:"reduceOnly"`
	TriggerPrice     uint32 `json:"triggerPrice"`
	OrderExpiry      int64  `json:"orderExpiry"`
	IntegratorFields
	TransactFields
}

type OrderInGroup struct {
	MarketIndex      int16  `json:"marketIndex"`
	ClientOrderIndex int64  `json:"clientOrderIndex"`
	BaseAmount       int64  `json:"baseAmount"`
	Price            uint32 `json:"price"`
	IsAsk            uint8  `json:"isAsk"`
	Type             uint8  `json:"type"`
	TimeInForce      uint8  `json:"timeInForce"`
	ReduceOnly       uint8  `json:"reduceOnly"`
	TriggerPrice     uint32 `json:"triggerPrice"`
	OrderExpiry      int64  `json:"orderExpiry"`
}

type CreateGroupedOrdersReq struct {
	GroupingType uint8          `json:"groupingType"`
	Orders       []OrderInGroup `json:"orders"`
	IntegratorFields
	TransactFields
}

type CancelOrderReq struct {
	MarketIndex int16 `json:"marketIndex"`
	OrderIndex  int64 `json:"orderIndex"`
	TransactFields
}

type WithdrawReq struct {
	AssetIndex int16  `json:"assetIndex"`
	RouteType  uint8  `json:"routeType"`
	Amount     uint64 `json:"amount"`
	TransactFields
}

type CreateSubAccountReq struct {
	TransactFields
}

type CancelAllOrdersReq struct {
	TimeInForce uint8 `json:"timeInForce"`
	Time        int64 `json:"time"`
	TransactFields
}

type ModifyOrderReq struct {
	MarketIndex  int16  `json:"marketIndex"`
	Index        int64  `json:"index"`
	BaseAmount   int64  `json:"baseAmount"`
	Price        uint32 `json:"price"`
	TriggerPrice uint32 `json:"triggerPrice"`
	IntegratorFields
	TransactFields
}

type TransferReq struct {
	ToAccountIndex int64  `json:"toAccountIndex"`
	AssetIndex     int16  `json:"assetIndex"`
	FromRouteType  uint8  `json:"fromRouteType"`
	ToRouteType    uint8  `json:"toRouteType"`
	Amount         int64  `json:"amount"`
	USDCFee        int64  `json:"usdcFee"`
	Memo           string `json:"memo"`
	TransactFields
}

type CreatePublicPoolReq struct {
	OperatorFee          int64  `json:"operatorFee"`
	InitialTotalShares   int64  `json:"initialTotalShares"`
	MinOperatorShareRate uint16 `json:"minOperatorShareRate"`
	TransactFields
}

type UpdatePublicPoolReq struct {
	PublicPoolIndex      int64  `json:"publicPoolIndex"`
	Status               uint8  `json:"status"`
	OperatorFee          int64  `json:"operatorFee"`
	MinOperatorShareRate uint16 `json:"minOperatorShareRate"`
	TransactFields
}

type MintSharesReq struct {
	PublicPoolIndex int64 `json:"publicPoolIndex"`
	ShareAmount     int64 `json:"shareAmount"`
	TransactFields
}

type BurnSharesReq struct {
	PublicPoolIndex int64 `json:"publicPoolIndex"`
	ShareAmount     int64 `json:"shareAmount"`
	TransactFields
}

type UpdateLeverageReq struct {
	MarketIndex           int16  `json:"marketIndex"`
	InitialMarginFraction uint16 `json:"initialMarginFraction"`
	MarginMode            uint8  `json:"marginMode"`
	TransactFields
}

type CreateAuthTokenReq struct {
	Deadline     int64 `json:"deadline"`
	ApiKeyIndex  int   `json:"apiKeyIndex"`
	AccountIndex int64 `json:"accountIndex"`
}

type UpdateMarginReq struct {
	MarketIndex int16 `json:"marketIndex"`
	USDCAmount  int64 `json:"usdcAmount"`
	Direction   uint8 `json:"direction"`
	TransactFields
}

type StakeAssetsReq struct {
	StakingPoolIndex int64 `json:"stakingPoolIndex"`
	ShareAmount      int64 `json:"shareAmount"`
	TransactFields
}

type UnstakeAssetsReq struct {
	StakingPoolIndex int64 `json:"stakingPoolIndex"`
	ShareAmount      int64 `json:"shareAmount"`
	TransactFields
}

type ApproveIntegratorReq struct {
	IntegratorIndex  int64  `json:"integratorIndex"`
	MaxPerpsTakerFee uint32 `json:"maxPerpsTakerFee"`
	MaxPerpsMakerFee uint32 `json:"maxPerpsMakerFee"`
	MaxSpotTakerFee  uint32 `json:"maxSpotTakerFee"`
	MaxSpotMakerFee  uint32 `json:"maxSpotMakerFee"`
	ApprovalExpiry   int64  `json:"approvalExpiry"`
	TransactFields
}

type UpdateAccountConfigReq struct {
	AccountTradingMode uint8 `json:"accountTradingMode"`
	TransactFields
}

type UpdateAccountAssetConfigReq struct {
	AssetIndex      int16 `json:"assetIndex"`
	AssetMarginMode uint8 `json:"assetMarginMode"`
	TransactFields
}
