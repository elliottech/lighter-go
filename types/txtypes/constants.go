package txtypes

import (
	"math"

	curve "github.com/elliottech/poseidon_crypto/curve/ecgfp5"
	schnorr "github.com/elliottech/poseidon_crypto/signature/schnorr"
)

type (
	Signature  = schnorr.Signature
	PrivateKey = curve.ECgFp5Scalar
)

const (
	NilApiKeyIndex       = MaxApiKeyIndex + 1
	TreasuryAccountIndex = int64(0)
)

// Keep this in sync with sequencer types/tx.go
const (
	TxTypeEmpty             = 0
	TxTypeL1Deposit         = 1
	TxTypeL1ChangePubKey    = 2
	TxTypeL1CreateMarket    = 3
	TxTypeL1UpdateMarket    = 4
	TxTypeL1CancelAllOrders = 5
	TxTypeL1Withdraw        = 6
	TxTypeL1CreateOrder     = 7

	TxTypeL2ChangePubKey       = 8
	TxTypeL2CreateSubAccount   = 9
	TxTypeL2CreatePublicPool   = 10
	TxTypeL2UpdatePublicPool   = 11
	TxTypeL2Transfer           = 12
	TxTypeL2Withdraw           = 13
	TxTypeL2CreateOrder        = 14
	TxTypeL2CancelOrder        = 15
	TxTypeL2CancelAllOrders    = 16
	TxTypeL2ModifyOrder        = 17
	TxTypeL2MintShares         = 18
	TxTypeL2BurnShares         = 19
	TxTypeL2UpdateLeverage     = 20
	TxTypeL2ForceBurnShares    = 40
	TxTypeL2StrategyTransfer   = 43
	TxTypeL2UpdateMarketConfig = 44
	TxTypeL2ApproveIntegrator  = 45

	TxTypeInternalClaimOrder        = 21
	TxTypeInternalCancelOrder       = 22
	TxTypeInternalDeleverage        = 23
	TxTypeInternalExitPosition      = 24
	TxTypeInternalCancelAllOrders   = 25
	TxTypeInternalLiquidatePosition = 26
	TxTypeInternalCreateOrder       = 27

	TxTypeL2CreateGroupedOrders = 28
	TxTypeL2UpdateMargin        = 29
	TxTypeL2UpdateAccountConfig = 41

	TxTypeL1BurnShares    = 30
	TxTypeL1RegisterAsset = 31
	TxTypeL1UpdateAsset   = 32

	TxTypeL2CreateStakingPool = 33
	// TxTypeL2UpdateStakingPool = 34
	TxTypeL2StakeAssets     = 35
	TxTypeL2UnstakeAssets   = 36
	TxTypeL1UnstakeAssets   = 37
	TxTypeL1SetSystemConfig = 38
)

// Order Type
const (
	// User set order types
	LimitOrder           = iota
	MarketOrder          = 1
	StopLossOrder        = 2
	StopLossLimitOrder   = 3
	TakeProfitOrder      = 4
	TakeProfitLimitOrder = 5
	TWAPOrder            = 6

	// Internal order types
	TWAPSubOrder     = 7
	LiquidationOrder = 8

	ApiMaxOrderType = TWAPOrder
)

// Order Time-In-Force
const (
	ImmediateOrCancel = iota
	GoodTillTime      = 1
	PostOnly          = 2
)

// Grouping Type
const (
	GroupingType                                = 0
	GroupingType_OneTriggersTheOther            = 1
	GroupingType_OneCancelsTheOther             = 2
	GroupingType_OneTriggersAOneCancelsTheOther = 3
)

// Cancel All Orders Time-In-Force
const (
	ImmediateCancelAll      = iota
	ScheduledCancelAll      = 1
	AbortScheduledCancelAll = 2
)

// Asset Margin Mode
const (
	AssetMarginMode_Disabled = 0
	AssetMarginMode_Enabled  = 1
	AssetMarginMode_Max      = AssetMarginMode_Enabled
)

// Asset Route Type
const (
	AssetRouteType_Perps = 0
	AssetRouteType_Spot  = 1
)

// Position Margin Mode
const (
	CrossMargin    = iota
	IsolatedMargin = 1
)

// Margin Update Direction
const (
	RemoveFromIsolatedMargin = iota
	AddToIsolatedMargin      = 1
)

const (
	OneUSDC = 1000000
	OneLIT  = 100_000_000

	FeeTick            int64  = 1_000_000
	MarginFractionTick int64  = 10_000
	ShareTick          uint16 = 10_000

	MinAccountIndex       int64 = -1
	MaxAccountIndex       int64 = 281474976710654 // (1 << 48) - 2
	MaxMasterAccountIndex int64 = 140737488355327 // (1 << 47) - 1
	MinSubAccountIndex    int64 = 140737488355328 // (1 << 47)
	MinApiKeyIndex        uint8 = 0
	MaxApiKeyIndex        uint8 = 254 // (1 << 8) - 2

	MinMarketIndex      int16 = 0
	MinPerpsMarketIndex int16 = 0
	MaxPerpsMarketIndex int16 = 254 // (1 << 8) - 2
	NilMarketIndex      int16 = 255
	MinSpotMarketIndex  int16 = 2048 // (1 << 11)
	MaxSpotMarketIndex  int16 = 4094 // (1 << 12) - 2

	NilIntegratorIndex    = 0
	NilIntegratorTakerFee = 0
	NilIntegratorMakerFee = 0

	NativeAssetIndex = uint16(1)
	USDCAssetIndex   = uint16(3)
	MinAssetIndex    = 1
	MaxAssetIndex    = (1 << 6) - 2
	NilAssetIndex    = 0

	DefaultStrategyIndex uint8 = 0
	MinStrategyIndex     uint8 = 0
	MaxStrategyIndex     uint8 = 7
	NilStrategyIndex     uint8 = 8

	MaxInvestedPublicPoolCount int64 = 16
	InitialPoolShareValue      int64 = 1_000                                             // 0.001 USDC
	MinInitialTotalShares      int64 = 1_000 * (OneUSDC / InitialPoolShareValue)         // 1,000 USDC worth of shares
	MaxInitialTotalShares      int64 = 1_000_000_000 * (OneUSDC / InitialPoolShareValue) // 1,000,000,000 USDC worth of shares
	MaxPoolShares              int64 = (1 << 60) - 1
	MaxBurntShareUSDCValue     int64 = (1 << 60) - 1

	MaxPoolEntryUSDC                = (1 << 56) - 1 // 2^56 - 1 max USDC to invest in a pool
	MinPoolSharesToMintOrBurn int64 = 1
	MaxPoolSharesToMintOrBurn int64 = (1 << 60) - 1

	MinInitialTotalStakingShares int64 = 100_000 * (OneLIT / InitialPoolShareValue)       // 100,000 LIT worth of shares
	MaxInitialTotalStakingShares int64 = 1_000_000_000 * (OneLIT / InitialPoolShareValue) // 1,000,000,000 LIT worth of shares
	MinStakingSharesToMintOrBurn int64 = 1
	MaxStakingSharesToMintOrBurn int64 = (1 << 60) - 1
	MaxStakingPoolShares         int64 = (1 << 60) - 1

	NbAttributesPerTx = 4

	MinNonce int64 = 0

	MinOrderNonce int64 = 0
	MaxOrderNonce int64 = (1 << 48) - 1

	NilClientOrderIndex int64 = 0
	NilOrderIndex       int64 = 0

	MinClientOrderIndex int64 = 1
	MaxClientOrderIndex int64 = (1 << 48) - 1

	MinOrderIndex int64 = MaxClientOrderIndex + 1
	MaxOrderIndex int64 = (1 << 60) - 1

	MinOrderBaseAmount int64 = 1
	MaxOrderBaseAmount int64 = (1 << 48) - 1
	NilOrderBaseAmount int64 = 0

	NilOrderPrice uint32 = 0
	MinOrderPrice uint32 = 1
	MaxOrderPrice uint32 = (1 << 32) - 1

	MinOrderCancelAllPeriod int64 = 1000 * 60 * 5            // 5 minutes
	MaxOrderCancelAllPeriod int64 = 1000 * 60 * 60 * 24 * 15 // 15 days

	NilOrderExpiry int64 = 0
	MinOrderExpiry int64 = 1
	MaxOrderExpiry int64 = math.MaxInt64

	MinOrderExpiryPeriod int64 = 1000 * 60 * 5            // 5 minutes
	MaxOrderExpiryPeriod int64 = 1000 * 60 * 60 * 24 * 30 // 30 days

	NilOrderTriggerPrice uint32 = 0
	MinOrderTriggerPrice uint32 = 1
	MaxOrderTriggerPrice uint32 = (1 << 32) - 1

	MaxGroupedOrderCount int64 = 3

	MaxTimestamp = (1 << 48) - 1

	InsuranceFundOperatorAccountIndex = 1
)

const (
	MaxExchangeUSDC = (1 << 60) - 1

	MinTransferAmount int64 = 1
	MaxTransferAmount int64 = MaxExchangeUSDC

	MinWithdrawalAmount uint64 = 1
	MaxWithdrawalAmount uint64 = MaxExchangeUSDC
)
