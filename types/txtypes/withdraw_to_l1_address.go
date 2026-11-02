package txtypes

import (
	"fmt"

	g "github.com/elliottech/poseidon_crypto/field/goldilocks"
	p2 "github.com/elliottech/poseidon_crypto/hash/poseidon2_goldilocks_plonky2"
	"github.com/ethereum/go-ethereum/common"
)

var _ TxInfo = (*L2WithdrawToL1AddressTxInfo)(nil)

type L2WithdrawToL1AddressTxInfo struct {
	AccountIndex int64
	ApiKeyIndex  uint8
	AssetIndex   uint32
	L1Recipient  string
	RouteType    uint8
	BaseAmount   uint64

	ExpiredAt  int64
	Nonce      int64
	Sig        []byte
	OwnerSig   string
	SignedHash string `json:"-"`

	L2TxAttributes
}

func (txInfo *L2WithdrawToL1AddressTxInfo) Validate() error {
	if err := txInfo.L2TxAttributes.Validate(); err != nil {
		return err
	}
	if !txInfo.L2TxAttributes.IsEmpty() {
		return ErrInvalidAttributeType
	}

	// Core's minimum is zero; -1 is only a client-selection sentinel.
	if txInfo.AccountIndex < 0 {
		return ErrInteropAccountIndexTooLow
	}
	if txInfo.AccountIndex > MaxAccountIndex {
		return ErrAccountIndexTooHigh
	}

	if txInfo.ApiKeyIndex > MaxApiKeyIndex {
		return ErrApiKeyIndexTooHigh
	}

	if txInfo.AssetIndex < MinAssetIndex {
		return ErrAssetIndexTooLow
	}
	if txInfo.AssetIndex > MaxAssetIndex {
		return ErrAssetIndexTooHigh
	}

	if !common.IsHexAddress(txInfo.L1Recipient) || common.HexToAddress(txInfo.L1Recipient) == (common.Address{}) {
		return ErrL1RecipientInvalid
	}

	if txInfo.RouteType != AssetRouteType_Perps && txInfo.RouteType != AssetRouteType_Spot {
		return ErrRouteTypeInvalid
	}

	if txInfo.BaseAmount == 0 {
		return ErrWithdrawalAmountTooLow
	}
	if txInfo.BaseAmount > MaxWithdrawalAmount {
		return ErrWithdrawalAmountTooHigh
	}

	if txInfo.Nonce < MinNonce {
		return ErrNonceTooLow
	}

	if txInfo.ExpiredAt < 0 || txInfo.ExpiredAt > MaxTimestamp {
		return ErrExpiredAtInvalid
	}

	return nil
}

func (txInfo *L2WithdrawToL1AddressTxInfo) GetTxType() uint8 {
	return TxTypeL2WithdrawToL1Address
}

func (txInfo *L2WithdrawToL1AddressTxInfo) GetTxHash() string {
	return txInfo.SignedHash
}

func (txInfo *L2WithdrawToL1AddressTxInfo) GetTxInfo() (string, error) {
	return getTxInfo(txInfo)
}

// GetOwnerSignatureBody returns the EIP-191 personal-sign message, not its digest.
func (txInfo *L2WithdrawToL1AddressTxInfo) GetOwnerSignatureBody(chainId uint32) string {
	signatureBody := fmt.Sprintf(
		TemplateWithdrawToL1Address,
		getHex10FromUint64(uint64(txInfo.Nonce)),
		getHex10FromUint64(uint64(txInfo.AccountIndex)),
		getHex10FromUint64(uint64(txInfo.RouteType)),
		getHex10FromUint64(uint64(txInfo.ApiKeyIndex)),
		common.HexToAddress(txInfo.L1Recipient).Hex(),
		getHex10FromUint64(uint64(txInfo.AssetIndex)),
		getHex10FromUint64(txInfo.BaseAmount),
		getHex10FromUint64(uint64(chainId)),
	)
	return signatureBody
}

func (txInfo *L2WithdrawToL1AddressTxInfo) Hash(lighterChainId uint32) (msgHash []byte, err error) {
	elems := make([]g.GoldilocksField, 0, 15)

	elems = append(elems, g.GoldilocksField(lighterChainId))
	elems = append(elems, g.GoldilocksField(TxTypeL2WithdrawToL1Address))
	elems = append(elems, g.GoldilocksField(txInfo.Nonce))
	elems = append(elems, g.GoldilocksField(txInfo.ExpiredAt))

	elems = append(elems, g.GoldilocksField(txInfo.AccountIndex))
	elems = append(elems, g.GoldilocksField(txInfo.ApiKeyIndex))
	elems = append(elems, g.GoldilocksField(txInfo.AssetIndex))
	elems = appendAddressLimbs(elems, txInfo.L1Recipient)
	elems = append(elems, g.GoldilocksField(txInfo.RouteType))
	elems = append(elems, g.GoldilocksField(txInfo.BaseAmount&0xFFFFFFFF))
	elems = append(elems, g.GoldilocksField(txInfo.BaseAmount>>32))

	txHash := p2.HashToQuinticExtension(elems)
	return txInfo.L2TxAttributes.AggregateTxHash(txHash)
}
