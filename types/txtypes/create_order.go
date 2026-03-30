package txtypes

import (
	g "github.com/elliottech/poseidon_crypto/field/goldilocks"
	p2 "github.com/elliottech/poseidon_crypto/hash/poseidon2_goldilocks_plonky2"
)

var _ TxInfo = (*L2CreateOrderTxInfo)(nil)

type L2CreateOrderTxInfo struct {
	AccountIndex int64
	ApiKeyIndex  uint8

	*OrderInfo

	ExpiredAt  int64
	Nonce      int64
	Sig        []byte
	SignedHash string `json:"-"`

	L2TxAttributes
}

func (txInfo *L2CreateOrderTxInfo) GetTxType() uint8 {
	return TxTypeL2CreateOrder
}

func (txInfo *L2CreateOrderTxInfo) GetTxInfo() (string, error) {
	return getTxInfo(txInfo)
}

func (txInfo *L2CreateOrderTxInfo) GetTxHash() string {
	return txInfo.SignedHash
}

func (txInfo *L2CreateOrderTxInfo) Validate() error {
	if err := txInfo.L2TxAttributes.Validate(); err != nil {
		return err
	}

	// AccountIndex
	if txInfo.AccountIndex < MinAccountIndex {
		return ErrAccountIndexTooLow
	}
	if txInfo.AccountIndex > MaxAccountIndex {
		return ErrAccountIndexTooHigh
	}
	// ApiKeyIndex
	if txInfo.ApiKeyIndex < MinApiKeyIndex {
		return ErrApiKeyIndexTooLow
	}
	if txInfo.ApiKeyIndex > MaxApiKeyIndex {
		return ErrApiKeyIndexTooHigh
	}

	// MarketIndex
	isSpotMarket := txInfo.MarketIndex >= MinSpotMarketIndex && txInfo.MarketIndex <= MaxSpotMarketIndex
	isPerpsMarket := txInfo.MarketIndex >= MinPerpsMarketIndex && txInfo.MarketIndex <= MaxPerpsMarketIndex
	if !isSpotMarket && !isPerpsMarket {
		return ErrInvalidMarketIndex
	}

	// ClientOrderIndex
	if txInfo.ClientOrderIndex != NilClientOrderIndex {
		if txInfo.ClientOrderIndex < MinClientOrderIndex {
			return ErrClientOrderIndexTooLow
		}
		if txInfo.ClientOrderIndex > MaxClientOrderIndex {
			return ErrClientOrderIndexTooHigh
		}
	}

	// BaseAmount
	if txInfo.ReduceOnly != 1 && txInfo.BaseAmount == NilOrderBaseAmount {
		return ErrBaseAmountTooLow
	}
	if txInfo.BaseAmount != NilOrderBaseAmount && txInfo.BaseAmount < MinOrderBaseAmount {
		return ErrBaseAmountTooLow
	}
	if txInfo.BaseAmount > MaxOrderBaseAmount {
		return ErrBaseAmountTooHigh
	}

	// Price
	if txInfo.Price < MinOrderPrice {
		return ErrPriceTooLow
	}
	if txInfo.Price > MaxOrderPrice {
		return ErrPriceTooHigh
	}

	// IsAsk
	if txInfo.IsAsk != 0 && txInfo.IsAsk != 1 {
		return ErrIsAskInvalid
	}

	// Type
	if txInfo.TimeInForce != ImmediateOrCancel && txInfo.TimeInForce != GoodTillTime && txInfo.TimeInForce != PostOnly {
		return ErrOrderTimeInForceInvalid
	}

	// ReduceOnly
	if (txInfo.ReduceOnly != 0 && txInfo.ReduceOnly != 1) || (isSpotMarket && txInfo.ReduceOnly == 1) {
		return ErrOrderReduceOnlyInvalid
	}

	// OrderExpiry
	if (txInfo.OrderExpiry < MinOrderExpiry || txInfo.OrderExpiry > MaxOrderExpiry) && txInfo.OrderExpiry != NilOrderExpiry {
		return ErrOrderExpiryInvalid
	}

	switch txInfo.Type {
	case MarketOrder:
		if txInfo.TimeInForce != ImmediateOrCancel {
			return ErrOrderTimeInForceInvalid
		} else if txInfo.OrderExpiry != NilOrderExpiry {
			return ErrOrderExpiryInvalid
		} else if txInfo.TriggerPrice != NilOrderTriggerPrice {
			return ErrOrderTriggerPriceInvalid
		}
	case LimitOrder:
		if txInfo.TriggerPrice != NilOrderTriggerPrice {
			return ErrOrderTriggerPriceInvalid
		} else if txInfo.TimeInForce == ImmediateOrCancel && txInfo.OrderExpiry != NilOrderExpiry {
			return ErrOrderExpiryInvalid
		} else if txInfo.TimeInForce != ImmediateOrCancel && txInfo.OrderExpiry == NilOrderExpiry {
			return ErrOrderExpiryInvalid
		}
	case StopLossOrder, TakeProfitOrder:
		if !isPerpsMarket {
			return ErrOrderTypeInvalid
		} else if txInfo.TimeInForce != ImmediateOrCancel {
			return ErrOrderTimeInForceInvalid
		} else if txInfo.TriggerPrice == NilOrderTriggerPrice {
			return ErrOrderTriggerPriceInvalid
		} else if txInfo.OrderExpiry == NilOrderExpiry {
			return ErrOrderExpiryInvalid
		}
	case StopLossLimitOrder, TakeProfitLimitOrder:
		if !isPerpsMarket {
			return ErrOrderTypeInvalid
		} else if txInfo.TriggerPrice == NilOrderTriggerPrice {
			return ErrOrderTriggerPriceInvalid
		} else if txInfo.OrderExpiry == NilOrderExpiry {
			return ErrOrderExpiryInvalid
		}
	case TWAPOrder:
		if txInfo.TimeInForce != GoodTillTime {
			return ErrOrderTimeInForceInvalid
		} else if txInfo.TriggerPrice != NilOrderTriggerPrice {
			return ErrOrderTriggerPriceInvalid
		} else if txInfo.OrderExpiry == NilOrderExpiry {
			return ErrOrderExpiryInvalid
		}
	default:
		return ErrOrderTypeInvalid
	}

	// TriggerPrice
	if (txInfo.TriggerPrice < MinOrderTriggerPrice || txInfo.TriggerPrice > MaxOrderTriggerPrice) && txInfo.TriggerPrice != NilOrderTriggerPrice {
		return ErrOrderTriggerPriceInvalid
	}

	// Nonce
	if txInfo.Nonce < MinNonce {
		return ErrNonceTooLow
	}

	if txInfo.ExpiredAt < 0 || txInfo.ExpiredAt > MaxTimestamp {
		return ErrExpiredAtInvalid
	}

	return nil
}

func (txInfo *L2CreateOrderTxInfo) Hash(lighterChainId uint32) (msgHash []byte, err error) {
	elems := make([]g.GoldilocksField, 0, 16)

	elems = append(elems, g.GoldilocksField(lighterChainId))
	elems = append(elems, g.GoldilocksField(TxTypeL2CreateOrder))
	elems = append(elems, g.GoldilocksField(txInfo.Nonce))
	elems = append(elems, g.GoldilocksField(txInfo.ExpiredAt))

	elems = append(elems, g.GoldilocksField(txInfo.AccountIndex))
	elems = append(elems, g.GoldilocksField(txInfo.ApiKeyIndex))
	elems = append(elems, g.GoldilocksField(txInfo.MarketIndex))
	elems = append(elems, g.GoldilocksField(txInfo.ClientOrderIndex))
	elems = append(elems, g.GoldilocksField(txInfo.BaseAmount))
	elems = append(elems, g.GoldilocksField(txInfo.Price))
	elems = append(elems, g.GoldilocksField(txInfo.IsAsk))
	elems = append(elems, g.GoldilocksField(txInfo.Type))
	elems = append(elems, g.GoldilocksField(txInfo.TimeInForce))
	elems = append(elems, g.GoldilocksField(txInfo.ReduceOnly))
	elems = append(elems, g.GoldilocksField(txInfo.TriggerPrice))
	elems = append(elems, g.GoldilocksField(txInfo.OrderExpiry))

	txHash := p2.HashToQuinticExtension(elems)
	return txInfo.L2TxAttributes.AggregateTxHash(txHash)
}
