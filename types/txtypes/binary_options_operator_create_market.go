package txtypes

import (
	g "github.com/elliottech/poseidon_crypto/field/goldilocks"
	p2 "github.com/elliottech/poseidon_crypto/hash/poseidon2_goldilocks_plonky2"
)

var _ TxInfo = (*BinaryOptionsOperatorCreateMarketTxInfo)(nil)

// BinaryOptionsOperatorCreateMarketTxInfo creates a binary options order book on a market slot.
// The sender must be the market operator assigned to the slot.
type BinaryOptionsOperatorCreateMarketTxInfo struct {
	AccountIndex int64
	ApiKeyIndex  uint8

	// MarketSlot is the slot the order book lives on; serialized as MarketIndex for the server.
	MarketSlot int16 `json:"MarketIndex"`
	// MarketType is the type of market created on the slot. Only MarketTypeBinaryOptions is supported.
	MarketType       uint8
	BaseAssetId      int16
	QuoteAssetId     int16
	StartTimestampMs int64 `json:"StartTimestamp"`
	EndTimestampMs   int64 `json:"EndTimestamp"`

	SizeExtensionMultiplier  int64
	QuoteExtensionMultiplier int64
	// SettlementCap is the maximum settlement value of a share in price ticks, so that
	// SettlementCap * QuoteExtensionMultiplier == SizeExtensionMultiplier.
	SettlementCap uint32
	// SettlementType is BinaryOptionsSettlementTypeDiscrete (settles at 0 or SettlementCap) or
	// BinaryOptionsSettlementTypeContinuous (settles at any price in [0, SettlementCap]).
	SettlementType uint8
	// DefaultPrice is the settlement price used when the market is refunded, in price ticks.
	// It is immutable after market creation.
	DefaultPrice      uint32
	TakerFee          uint32
	MakerFee          uint32
	MinBaseAmount     int64
	MinQuoteAmount    int64
	OrderQuoteLimit   int64
	OpenInterestLimit int64
	// IsFrozen creates the order book frozen (1): no orders can be placed, modified or matched on it.
	IsFrozen uint8

	Symbol        [32]byte
	SizeDecimals  uint8
	PriceDecimals uint8

	ExpiredAt  int64
	Nonce      int64
	Sig        []byte
	SignedHash string `json:"-"`

	L2TxAttributes
}

func (txInfo *BinaryOptionsOperatorCreateMarketTxInfo) GetTxType() uint8 {
	return TxTypeBinaryOptionsOperatorCreateMarket
}

func (txInfo *BinaryOptionsOperatorCreateMarketTxInfo) GetTxInfo() (string, error) {
	return getTxInfo(txInfo)
}

func (txInfo *BinaryOptionsOperatorCreateMarketTxInfo) GetTxHash() string {
	return txInfo.SignedHash
}

func (txInfo *BinaryOptionsOperatorCreateMarketTxInfo) Validate() error {
	if err := txInfo.L2TxAttributes.Validate(); err != nil {
		return err
	}

	if txInfo.AccountIndex < MinAccountIndex {
		return ErrFromAccountIndexTooLow
	}
	if txInfo.AccountIndex > MaxAccountIndex {
		return ErrFromAccountIndexTooHigh
	}

	// ApiKeyIndex
	if txInfo.ApiKeyIndex < MinApiKeyIndex {
		return ErrApiKeyIndexTooLow
	}
	if txInfo.ApiKeyIndex > MaxApiKeyIndex {
		return ErrApiKeyIndexTooHigh
	}

	if txInfo.MarketType != MarketTypeBinaryOptions {
		return ErrInvalidMarketType
	}

	if txInfo.BaseAssetId != int16(USDCAssetIndex) || txInfo.QuoteAssetId != int16(USDCAssetIndex) {
		return ErrInvalidAssetIndex
	}

	// MarketSlot
	if txInfo.MarketSlot < MinBinaryOptionsMarketSlot || txInfo.MarketSlot > MaxBinaryOptionsMarketSlot {
		return ErrInvalidMarketIndex
	}

	if txInfo.StartTimestampMs < 0 || txInfo.EndTimestampMs > MaxTimestamp || txInfo.StartTimestampMs >= txInfo.EndTimestampMs {
		return ErrInvalidMarketTimestamps
	}

	if txInfo.SizeExtensionMultiplier == 0 || txInfo.SizeExtensionMultiplier > MaxExtensionMultiplier || txInfo.SizeExtensionMultiplier%FeeTick != 0 {
		return ErrInvalidExtensionMultiplier
	}
	if txInfo.QuoteExtensionMultiplier == 0 || txInfo.QuoteExtensionMultiplier > MaxExtensionMultiplier || txInfo.QuoteExtensionMultiplier%FeeTick != 0 {
		return ErrInvalidExtensionMultiplier
	}

	if txInfo.SettlementCap == 0 ||
		txInfo.SizeExtensionMultiplier%txInfo.QuoteExtensionMultiplier != 0 ||
		txInfo.SizeExtensionMultiplier/txInfo.QuoteExtensionMultiplier != int64(txInfo.SettlementCap) {
		return ErrInvalidSettlementCap
	}

	if txInfo.SettlementType != BinaryOptionsSettlementTypeDiscrete && txInfo.SettlementType != BinaryOptionsSettlementTypeContinuous {
		return ErrInvalidSettlementType
	}

	if txInfo.DefaultPrice > txInfo.SettlementCap {
		return ErrInvalidDefaultPrice
	}

	if int64(txInfo.TakerFee) > FeeTick || int64(txInfo.MakerFee) > FeeTick {
		return ErrFeeTooHigh
	}

	if txInfo.MinBaseAmount == 0 || txInfo.MinBaseAmount > MaxOrderBaseAmount {
		return ErrInvalidMinBaseAmount
	}
	if txInfo.MinQuoteAmount == 0 || txInfo.MinQuoteAmount > MaxOrderQuoteAmount {
		return ErrInvalidMinQuoteAmount
	}
	if txInfo.OrderQuoteLimit > MaxOrderQuoteAmount || txInfo.OrderQuoteLimit < txInfo.MinQuoteAmount {
		return ErrOrderQuoteLimitTooHigh
	}
	if txInfo.OpenInterestLimit < txInfo.OrderQuoteLimit || txInfo.OpenInterestLimit > MaxOpenInterest {
		return ErrOpenInterestLimitTooLow
	}

	if txInfo.IsFrozen > 1 {
		return ErrInvalidIsFrozen
	}

	if txInfo.Nonce < MinNonce {
		return ErrNonceTooLow
	}

	if txInfo.ExpiredAt < 0 || txInfo.ExpiredAt > MaxTimestamp {
		return ErrExpiredAtInvalid
	}

	return nil
}

func (txInfo *BinaryOptionsOperatorCreateMarketTxInfo) Hash(lighterChainId uint32) (msgHash []byte, err error) {
	elems := make([]g.GoldilocksField, 0, 24)

	elems = append(elems, g.GoldilocksField(lighterChainId))
	elems = append(elems, g.GoldilocksField(TxTypeBinaryOptionsOperatorCreateMarket))
	elems = append(elems, g.GoldilocksField(txInfo.Nonce))
	elems = append(elems, g.GoldilocksField(txInfo.ExpiredAt))
	elems = append(elems, g.GoldilocksField(txInfo.AccountIndex))
	elems = append(elems, g.GoldilocksField(txInfo.ApiKeyIndex))
	elems = append(elems, g.GoldilocksField(txInfo.MarketSlot))
	elems = append(elems, g.GoldilocksField(txInfo.MarketType))
	elems = append(elems, g.GoldilocksField(txInfo.BaseAssetId))
	elems = append(elems, g.GoldilocksField(txInfo.QuoteAssetId))
	elems = append(elems, g.GoldilocksField(txInfo.StartTimestampMs))
	elems = append(elems, g.GoldilocksField(txInfo.EndTimestampMs))
	elems = append(elems, g.GoldilocksField(txInfo.SizeExtensionMultiplier))
	elems = append(elems, g.GoldilocksField(txInfo.QuoteExtensionMultiplier))
	elems = append(elems, g.GoldilocksField(txInfo.SettlementCap))
	elems = append(elems, g.GoldilocksField(txInfo.SettlementType))
	elems = append(elems, g.GoldilocksField(txInfo.DefaultPrice))
	elems = append(elems, g.GoldilocksField(txInfo.TakerFee))
	elems = append(elems, g.GoldilocksField(txInfo.MakerFee))
	elems = append(elems, g.GoldilocksField(txInfo.MinBaseAmount))
	elems = append(elems, g.GoldilocksField(txInfo.MinQuoteAmount))
	elems = append(elems, g.GoldilocksField(txInfo.OrderQuoteLimit))
	elems = append(elems, g.GoldilocksField(txInfo.OpenInterestLimit))
	elems = append(elems, g.GoldilocksField(txInfo.IsFrozen))

	txHash := p2.HashToQuinticExtension(elems)
	return txInfo.L2TxAttributes.AggregateTxHash(txHash)
}
