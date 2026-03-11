package txtypes

import (
	g "github.com/elliottech/poseidon_crypto/field/goldilocks"
	p2 "github.com/elliottech/poseidon_crypto/hash/poseidon2_goldilocks"
)

var _ TxInfo = (*L2StakeAssetsTxInfo)(nil)

type L2StakeAssetsTxInfo struct {
	AccountIndex int64
	ApiKeyIndex  uint8

	StakingPoolIndex int64
	ShareAmount      int64

	ExpiredAt  int64
	Nonce      int64
	Sig        []byte
	SignedHash string `json:"-"`
}

func (txInfo *L2StakeAssetsTxInfo) GetTxType() uint8 {
	return TxTypeL2StakeAssets
}

func (txInfo *L2StakeAssetsTxInfo) GetTxInfo() (string, error) {
	return getTxInfo(txInfo)
}

func (txInfo *L2StakeAssetsTxInfo) GetTxHash() string {
	return txInfo.SignedHash
}

func (txInfo *L2StakeAssetsTxInfo) Validate() error {
	if txInfo.AccountIndex < MinAccountIndex {
		return ErrFromAccountIndexTooLow
	}
	if txInfo.AccountIndex > MaxAccountIndex {
		return ErrFromAccountIndexTooHigh
	}

	if txInfo.ApiKeyIndex < MinApiKeyIndex {
		return ErrApiKeyIndexTooLow
	}
	if txInfo.ApiKeyIndex > MaxApiKeyIndex {
		return ErrApiKeyIndexTooHigh
	}

	if txInfo.StakingPoolIndex < MinSubAccountIndex {
		return ErrStakingPoolIndexTooLow
	}
	if txInfo.StakingPoolIndex > MaxAccountIndex {
		return ErrStakingPoolIndexTooHigh
	}

	if txInfo.ShareAmount < MinStakingSharesToMintOrBurn {
		return ErrPoolStakeShareAmountTooLow
	}
	if txInfo.ShareAmount > MaxStakingSharesToMintOrBurn {
		return ErrPoolStakeShareAmountTooHigh
	}

	if txInfo.Nonce < MinNonce {
		return ErrNonceTooLow
	}

	if txInfo.ExpiredAt < 0 || txInfo.ExpiredAt > MaxTimestamp {
		return ErrExpiredAtInvalid
	}

	return nil
}

func (txInfo *L2StakeAssetsTxInfo) Hash(lighterChainId uint32, extra ...g.Element) (msgHash []byte, err error) {
	elems := make([]g.Element, 0, 8)

	elems = append(elems, g.FromUint32(lighterChainId))
	elems = append(elems, g.FromUint32(TxTypeL2StakeAssets))
	elems = append(elems, g.FromInt64(txInfo.Nonce))
	elems = append(elems, g.FromInt64(txInfo.ExpiredAt))

	elems = append(elems, g.FromInt64(txInfo.AccountIndex))
	elems = append(elems, g.FromUint32(uint32(txInfo.ApiKeyIndex)))
	elems = append(elems, g.FromInt64(txInfo.StakingPoolIndex))
	elems = append(elems, g.FromInt64(txInfo.ShareAmount))

	return p2.HashToQuinticExtension(elems).ToLittleEndianBytes(), nil
}
