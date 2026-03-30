package txtypes

import (
	g "github.com/elliottech/poseidon_crypto/field/goldilocks"
	p2 "github.com/elliottech/poseidon_crypto/hash/poseidon2_goldilocks_plonky2"
)

var _ TxInfo = (*L2UpdatePublicPoolTxInfo)(nil)

type L2UpdatePublicPoolTxInfo struct {
	AccountIndex int64 // Master account index
	ApiKeyIndex  uint8

	PublicPoolIndex int64

	Status               uint8
	OperatorFee          int64
	MinOperatorShareRate uint16

	ExpiredAt  int64
	Nonce      int64
	Sig        []byte
	SignedHash string `json:"-"`

	L2TxAttributes
}

func (txInfo *L2UpdatePublicPoolTxInfo) GetTxType() uint8 {
	return TxTypeL2UpdatePublicPool
}

func (txInfo *L2UpdatePublicPoolTxInfo) GetTxInfo() (string, error) {
	return getTxInfo(txInfo)
}

func (txInfo *L2UpdatePublicPoolTxInfo) GetTxHash() string {
	return txInfo.SignedHash
}

func (txInfo *L2UpdatePublicPoolTxInfo) Validate() error {
	if err := txInfo.L2TxAttributes.Validate(); err != nil {
		return err
	}

	// AccountIndex
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

	// PublicPoolIndex
	if txInfo.PublicPoolIndex < MinAccountIndex {
		return ErrPublicPoolIndexTooLow
	}
	if txInfo.PublicPoolIndex > MaxAccountIndex {
		return ErrPublicPoolIndexTooHigh
	}

	// Status
	if txInfo.Status != 0 && txInfo.Status != 1 {
		return ErrInvalidPoolStatus
	}

	// OperatorFee
	if txInfo.OperatorFee < 0 || txInfo.OperatorFee > FeeTick {
		return ErrInvalidPoolOperatorFee
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

func (txInfo *L2UpdatePublicPoolTxInfo) Hash(lighterChainId uint32) (msgHash []byte, err error) {
	elems := make([]g.GoldilocksField, 0, 10)

	elems = append(elems, g.GoldilocksField(lighterChainId))
	elems = append(elems, g.GoldilocksField(TxTypeL2UpdatePublicPool))
	elems = append(elems, g.GoldilocksField(txInfo.Nonce))
	elems = append(elems, g.GoldilocksField(txInfo.ExpiredAt))

	elems = append(elems, g.GoldilocksField(txInfo.AccountIndex))
	elems = append(elems, g.GoldilocksField(txInfo.ApiKeyIndex))
	elems = append(elems, g.GoldilocksField(txInfo.PublicPoolIndex))
	elems = append(elems, g.GoldilocksField(txInfo.Status))
	elems = append(elems, g.GoldilocksField(txInfo.OperatorFee))
	elems = append(elems, g.GoldilocksField(txInfo.MinOperatorShareRate))

	txHash := p2.HashToQuinticExtension(elems)
	return txInfo.L2TxAttributes.AggregateTxHash(txHash)
}
