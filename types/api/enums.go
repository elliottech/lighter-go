package api

// OrderType represents the type of order
type OrderType uint8

const (
	OrderTypeLimitOrder         OrderType = 0
	OrderTypeMarketOrder        OrderType = 1
	OrderTypeStopLossOrder      OrderType = 2
	OrderTypeStopLossLimitOrder OrderType = 3
	OrderTypeTakeProfitOrder    OrderType = 4
	OrderTypeTakeProfitLimit    OrderType = 5
	OrderTypeTWAPOrder          OrderType = 6
)

// String returns the string representation of OrderType
func (o OrderType) String() string {
	switch o {
	case OrderTypeLimitOrder:
		return "limit"
	case OrderTypeMarketOrder:
		return "market"
	case OrderTypeStopLossOrder:
		return "stop_loss"
	case OrderTypeStopLossLimitOrder:
		return "stop_loss_limit"
	case OrderTypeTakeProfitOrder:
		return "take_profit"
	case OrderTypeTakeProfitLimit:
		return "take_profit_limit"
	case OrderTypeTWAPOrder:
		return "twap"
	default:
		return "unknown"
	}
}

// OrderSide represents the side of an order
type OrderSide uint8

const (
	OrderSideBid OrderSide = 0 // Buy
	OrderSideAsk OrderSide = 1 // Sell
)

// String returns the string representation of OrderSide
func (o OrderSide) String() string {
	if o == OrderSideBid {
		return "bid"
	}
	return "ask"
}

// IsBuy returns true if this is a buy order
func (o OrderSide) IsBuy() bool {
	return o == OrderSideBid
}

// TimeInForce represents order time-in-force options
type TimeInForce uint8

const (
	TimeInForceIOC      TimeInForce = 0 // Immediate or Cancel
	TimeInForceGTT      TimeInForce = 1 // Good Till Time
	TimeInForcePostOnly TimeInForce = 2 // Post Only (maker only)
)

// String returns the string representation of TimeInForce
func (t TimeInForce) String() string {
	switch t {
	case TimeInForceIOC:
		return "ioc"
	case TimeInForceGTT:
		return "gtt"
	case TimeInForcePostOnly:
		return "post_only"
	default:
		return "unknown"
	}
}

// MarginMode represents position margin mode
type MarginMode uint8

const (
	MarginModeCross    MarginMode = 0
	MarginModeIsolated MarginMode = 1
)

// String returns the string representation of MarginMode
func (m MarginMode) String() string {
	if m == MarginModeCross {
		return "cross"
	}
	return "isolated"
}

// MarginDirection represents margin update direction
type MarginDirection uint8

const (
	MarginDirectionRemove MarginDirection = 0
	MarginDirectionAdd    MarginDirection = 1
)

// AssetRouteType represents the asset route type
type AssetRouteType uint8

const (
	AssetRouteTypePerps AssetRouteType = 0
	AssetRouteTypeSpot  AssetRouteType = 1
)

// String returns the string representation of AssetRouteType
func (a AssetRouteType) String() string {
	if a == AssetRouteTypePerps {
		return "perps"
	}
	return "spot"
}

// GroupingType represents order grouping type
type GroupingType uint8

const (
	GroupingTypeNone   GroupingType = 0
	GroupingTypeOTO    GroupingType = 1 // One Triggers Other
	GroupingTypeOCO    GroupingType = 2 // One Cancels Other
	GroupingTypeOTOCO  GroupingType = 3 // One Triggers a One Cancels Other
)

// String returns the string representation of GroupingType
func (g GroupingType) String() string {
	switch g {
	case GroupingTypeNone:
		return "none"
	case GroupingTypeOTO:
		return "oto"
	case GroupingTypeOCO:
		return "oco"
	case GroupingTypeOTOCO:
		return "otoco"
	default:
		return "unknown"
	}
}

// CancelAllMode represents cancel all orders mode
type CancelAllMode uint8

const (
	CancelAllModeImmediate CancelAllMode = 0
	CancelAllModeScheduled CancelAllMode = 1
	CancelAllModeAbort     CancelAllMode = 2
)

// TxType represents transaction types
type TxType uint8

const (
	TxTypeL1Deposit           TxType = 1
	TxTypeL1ChangePubKey      TxType = 2
	TxTypeL1CreateMarket      TxType = 3
	TxTypeL1UpdateMarket      TxType = 4
	TxTypeL1CancelAllOrders   TxType = 5
	TxTypeL1Withdraw          TxType = 6
	TxTypeL1CreateOrder       TxType = 7
	TxTypeL2ChangePubKey      TxType = 8
	TxTypeL2CreateSubAccount  TxType = 9
	TxTypeL2CreatePublicPool  TxType = 10
	TxTypeL2UpdatePublicPool  TxType = 11
	TxTypeL2Transfer          TxType = 12
	TxTypeL2Withdraw          TxType = 13
	TxTypeL2CreateOrder       TxType = 14
	TxTypeL2CancelOrder       TxType = 15
	TxTypeL2CancelAllOrders   TxType = 16
	TxTypeL2ModifyOrder       TxType = 17
	TxTypeL2MintShares        TxType = 18
	TxTypeL2BurnShares        TxType = 19
	TxTypeL2UpdateLeverage    TxType = 20
	TxTypeL2CreateGrouped     TxType = 28
	TxTypeL2UpdateMargin      TxType = 29
)

// DepositStatus represents deposit status
type DepositStatus string

const (
	DepositStatusPending   DepositStatus = "pending"
	DepositStatusConfirmed DepositStatus = "confirmed"
	DepositStatusFailed    DepositStatus = "failed"
)

// WithdrawStatus represents withdrawal status
type WithdrawStatus string

const (
	WithdrawStatusPending   WithdrawStatus = "pending"
	WithdrawStatusConfirmed WithdrawStatus = "confirmed"
	WithdrawStatusFailed    WithdrawStatus = "failed"
)

// CandlestickResolution represents candlestick resolution
type CandlestickResolution string

const (
	Resolution1m  CandlestickResolution = "1"
	Resolution5m  CandlestickResolution = "5"
	Resolution15m CandlestickResolution = "15"
	Resolution30m CandlestickResolution = "30"
	Resolution1h  CandlestickResolution = "60"
	Resolution4h  CandlestickResolution = "240"
	Resolution1d  CandlestickResolution = "1D"
	Resolution1w  CandlestickResolution = "1W"
)

// FundingResolution represents funding resolution
type FundingResolution string

const (
	FundingResolution1h FundingResolution = "1h"
	FundingResolution1d FundingResolution = "1d"
)

// PositionSide represents position side filter
type PositionSide string

const (
	PositionSideLong  PositionSide = "long"
	PositionSideShort PositionSide = "short"
)

// ExportType represents export data type
type ExportType string

const (
	ExportTypeFunding ExportType = "funding"
	ExportTypeTrade   ExportType = "trade"
)
