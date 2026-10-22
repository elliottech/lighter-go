package txtypes

import (
	g "github.com/elliottech/poseidon_crypto/field/goldilocks"
	p2 "github.com/elliottech/poseidon_crypto/hash/poseidon2_goldilocks_plonky2"
)

var _ TxInfo = (*BinaryOptionsOperatorUpdateMarketTxInfo)(nil)

// MaxUpdateMarketLiquidationFee caps the liquidation fee a BinaryOptionsOperatorUpdateMarket tx can set
// on a perps market at 1%.
const MaxUpdateMarketLiquidationFee int64 = FeeTick / 100

// BinaryOptionsOperatorUpdateMarketTxInfo updates the mutable parameters of an active order book on behalf
// of its market operator. For perps and spot markets, StartTimestampMs, EndTimestampMs and IsFrozen do not
// apply and must be zero; LiquidationFee only applies to perps markets and is capped at 1%.
// Every field of the order book that is not carried here is immutable after creation.
type BinaryOptionsOperatorUpdateMarketTxInfo struct {
	AccountIndex int64
	ApiKeyIndex  uint8

	// MarketSlot is the slot the order book lives on; serialized as MarketIndex for the server.
	MarketSlot       int16 `json:"MarketIndex"`
	MarketIndex      int64 `json:"PublicMarketIndex"`
	StartTimestampMs int64 `json:"StartTimestamp"`
	EndTimestampMs   int64 `json:"EndTimestamp"`

	TakerFee          uint32
	MakerFee          uint32
	LiquidationFee    uint32 // only applies to perps markets, capped at 1%; must be zero for spot and binary options markets
	MinBaseAmount     int64
	MinQuoteAmount    int64
	OrderQuoteLimit   int64
	OpenInterestLimit int64
	// IsFrozen freezes (1) or unfreezes (0) the order book. While frozen no orders can be placed,
	// modified or matched; cancels stay allowed.
	IsFrozen uint8

	ExpiredAt  int64
	Nonce      int64
	Sig        []byte
	SignedHash string `json:"-"`

	L2TxAttributes
}

func (txInfo *BinaryOptionsOperatorUpdateMarketTxInfo) GetTxType() uint8 {
	return TxTypeBinaryOptionsOperatorUpdateMarket
}

func (txInfo *BinaryOptionsOperatorUpdateMarketTxInfo) GetTxInfo() (string, error) {
	return getTxInfo(txInfo)
}

func (txInfo *BinaryOptionsOperatorUpdateMarketTxInfo) GetTxHash() string {
	return txInfo.SignedHash
}

func (txInfo *BinaryOptionsOperatorUpdateMarketTxInfo) Validate() error {
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

	// MarketSlot
	isBinaryOptionsMarket := txInfo.MarketSlot >= MinBinaryOptionsMarketSlot && txInfo.MarketSlot <= MaxBinaryOptionsMarketSlot
	isPerpsMarket := txInfo.MarketSlot >= MinPerpsMarketSlot && txInfo.MarketSlot <= MaxPerpsMarketSlot
	isSpotMarket := txInfo.MarketSlot >= MinSpotMarketSlot && txInfo.MarketSlot <= MaxSpotMarketSlot
	if !isBinaryOptionsMarket && !isPerpsMarket && !isSpotMarket {
		return ErrInvalidMarketIndex
	}

	// MarketIndex
	if txInfo.MarketIndex < MinMarketIndex || txInfo.MarketIndex > MaxMarketIndex ||
		txInfo.MarketIndex == NilMarketIndex {
		return ErrInvalidMarketIndex
	}

	if isBinaryOptionsMarket {
		if txInfo.StartTimestampMs < 0 || txInfo.EndTimestampMs > MaxTimestamp || txInfo.StartTimestampMs >= txInfo.EndTimestampMs {
			return ErrInvalidMarketTimestamps
		}
	} else if txInfo.StartTimestampMs != 0 || txInfo.EndTimestampMs != 0 {
		// StartTimestampMs and EndTimestampMs only apply to binary options markets
		return ErrInvalidMarketTimestamps
	}

	if int64(txInfo.TakerFee) > FeeTick || int64(txInfo.MakerFee) > FeeTick {
		return ErrFeeTooHigh
	}
	// LiquidationFee only applies to perps markets and is capped at 1%
	if isPerpsMarket {
		if int64(txInfo.LiquidationFee) > MaxUpdateMarketLiquidationFee {
			return ErrFeeTooHigh
		}
	} else if txInfo.LiquidationFee != 0 {
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

	// Spot markets do not carry an open interest limit
	if isSpotMarket {
		if txInfo.OpenInterestLimit != 0 {
			return ErrOpenInterestLimitInvalid
		}
	} else if txInfo.OpenInterestLimit < txInfo.OrderQuoteLimit || txInfo.OpenInterestLimit > MaxOpenInterest {
		return ErrOpenInterestLimitTooLow
	}

	if isBinaryOptionsMarket {
		if txInfo.IsFrozen > 1 {
			return ErrInvalidIsFrozen
		}
	} else if txInfo.IsFrozen != 0 {
		// IsFrozen only applies to binary options markets
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

func (txInfo *BinaryOptionsOperatorUpdateMarketTxInfo) Hash(lighterChainId uint32) (msgHash []byte, err error) {
	elems := make([]g.GoldilocksField, 0, 17)

	elems = append(elems, g.GoldilocksField(lighterChainId))
	elems = append(elems, g.GoldilocksField(TxTypeBinaryOptionsOperatorUpdateMarket))
	elems = append(elems, g.GoldilocksField(txInfo.Nonce))
	elems = append(elems, g.GoldilocksField(txInfo.ExpiredAt))
	elems = append(elems, g.GoldilocksField(txInfo.AccountIndex))
	elems = append(elems, g.GoldilocksField(txInfo.ApiKeyIndex))
	elems = append(elems, g.GoldilocksField(txInfo.MarketSlot))
	elems = append(elems, g.GoldilocksField(txInfo.MarketIndex))
	elems = append(elems, g.GoldilocksField(txInfo.StartTimestampMs))
	elems = append(elems, g.GoldilocksField(txInfo.EndTimestampMs))
	elems = append(elems, g.GoldilocksField(txInfo.TakerFee))
	elems = append(elems, g.GoldilocksField(txInfo.MakerFee))
	elems = append(elems, g.GoldilocksField(txInfo.LiquidationFee))
	elems = append(elems, g.GoldilocksField(txInfo.MinBaseAmount))
	elems = append(elems, g.GoldilocksField(txInfo.MinQuoteAmount))
	elems = append(elems, g.GoldilocksField(txInfo.OrderQuoteLimit))
	elems = append(elems, g.GoldilocksField(txInfo.OpenInterestLimit))
	elems = append(elems, g.GoldilocksField(txInfo.IsFrozen))

	txHash := p2.HashToQuinticExtension(elems)
	return txInfo.L2TxAttributes.AggregateTxHash(txHash)
}
