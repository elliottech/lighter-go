package txtypes

import (
	g "github.com/elliottech/poseidon_crypto/field/goldilocks"
	p2 "github.com/elliottech/poseidon_crypto/hash/poseidon2_goldilocks_plonky2"
)

var _ TxInfo = (*BinaryOptionsOperatorUpdateMarketTxInfo)(nil)

// BinaryOptionsOperatorUpdateMarketTxInfo updates the mutable parameters of an active binary options
// order book on behalf of its market operator. Every field of the order book that is not carried here
// is immutable after creation.
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
	LiquidationFee    uint32 // only applies to perps markets; must be zero for binary options markets
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
	if txInfo.MarketSlot < MinBinaryOptionsMarketSlot || txInfo.MarketSlot > MaxBinaryOptionsMarketSlot {
		return ErrInvalidMarketIndex
	}

	// MarketIndex
	if txInfo.MarketIndex < MinMarketIndex || txInfo.MarketIndex > MaxMarketIndex ||
		txInfo.MarketIndex == NilMarketIndex {
		return ErrInvalidMarketIndex
	}

	if txInfo.StartTimestampMs < 0 || txInfo.EndTimestampMs > MaxTimestamp || txInfo.StartTimestampMs >= txInfo.EndTimestampMs {
		return ErrInvalidMarketTimestamps
	}

	if int64(txInfo.TakerFee) > FeeTick || int64(txInfo.MakerFee) > FeeTick {
		return ErrFeeTooHigh
	}
	if txInfo.LiquidationFee != 0 {
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
