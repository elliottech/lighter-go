package txtypes

import (
	g "github.com/elliottech/poseidon_crypto/field/goldilocks"
	p2 "github.com/elliottech/poseidon_crypto/hash/poseidon2_goldilocks_plonky2"
)

var _ TxInfo = (*L2CreateStakingPoolTxInfo)(nil)

type L2CreateStakingPoolTxInfo struct {
	AccountIndex int64 // Master account index
	ApiKeyIndex  uint8

	InitialTotalShares   int64
	MinOperatorShareRate uint16

	ExpiredAt  int64
	Nonce      int64
	Sig        []byte
	SignedHash string `json:"-"`

	L2TxAttributes
}

func (txInfo *L2CreateStakingPoolTxInfo) GetTxType() uint8 {
	return TxTypeL2CreateStakingPool
}

func (txInfo *L2CreateStakingPoolTxInfo) GetTxInfo() (string, error) {
	return getTxInfo(txInfo)
}

func (txInfo *L2CreateStakingPoolTxInfo) GetTxHash() string {
	return txInfo.SignedHash
}

func (txInfo *L2CreateStakingPoolTxInfo) Validate() error {
	if err := txInfo.L2TxAttributes.Validate(); err != nil {
		return err
	}

	// AccountIndex
	if txInfo.AccountIndex != TreasuryAccountIndex {
		return ErrAccountIndexMustBeTreasury
	}

	// ApiKeyIndex
	if txInfo.ApiKeyIndex < MinApiKeyIndex {
		return ErrApiKeyIndexTooLow
	}
	if txInfo.ApiKeyIndex > MaxApiKeyIndex {
		return ErrApiKeyIndexTooHigh
	}

	// InitialTotalShares
	if txInfo.InitialTotalShares < MinInitialTotalStakingShares {
		return ErrStakingPoolInitialTotalSharesTooLow
	}
	if txInfo.InitialTotalShares > MaxInitialTotalStakingShares {
		return ErrStakingPoolInitialTotalSharesTooHigh
	}

	// MinOperatorShareRate
	if txInfo.MinOperatorShareRate > ShareTick {
		return ErrPoolMinOperatorShareRateTooHigh
	}

	// MinOperatorShareRate
	if txInfo.MinOperatorShareRate > ShareTick {
		return ErrPoolMinOperatorShareRateTooHigh
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

func (txInfo *L2CreateStakingPoolTxInfo) Hash(lighterChainId uint32) (msgHash []byte, err error) {
	elems := make([]g.GoldilocksField, 0, 8)

	elems = append(elems, g.GoldilocksField(lighterChainId))
	elems = append(elems, g.GoldilocksField(TxTypeL2CreateStakingPool))
	elems = append(elems, g.GoldilocksField(txInfo.Nonce))
	elems = append(elems, g.GoldilocksField(txInfo.ExpiredAt))

	elems = append(elems, g.GoldilocksField(txInfo.AccountIndex))
	elems = append(elems, g.GoldilocksField(txInfo.ApiKeyIndex))
	elems = append(elems, g.GoldilocksField(txInfo.InitialTotalShares))
	elems = append(elems, g.GoldilocksField(txInfo.MinOperatorShareRate))

	txHash := p2.HashToQuinticExtension(elems)
	return txInfo.L2TxAttributes.AggregateTxHash(txHash)
}
