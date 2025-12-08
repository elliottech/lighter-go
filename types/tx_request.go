package types

import (
	"fmt"
	"time"

	"github.com/elliottech/lighter-go/signer"
	"github.com/elliottech/lighter-go/types/txtypes"
	g "github.com/elliottech/poseidon_crypto/field/goldilocks"
	gFp5 "github.com/elliottech/poseidon_crypto/field/goldilocks_quintic_extension"
	p2 "github.com/elliottech/poseidon_crypto/hash/poseidon2_goldilocks"
	ethCommon "github.com/ethereum/go-ethereum/common"
)

type TransactOpts struct {
	FromAccountIndex *int64
	ApiKeyIndex      *uint8
	ExpiredAt        int64
	Nonce            *int64
	DryRun           bool
}

type PublicKey = gFp5.Element

type ChangePubKeyReq struct {
	PubKey [40]byte
}

type TransferTxReq struct {
	ToAccountIndex int64
	USDCAmount     int64
	Fee            int64
	Memo           [32]byte
}

type WithdrawTxReq struct {
	USDCAmount uint64
}

type CreateOrderTxReq struct {
	MarketIndex      int16
	ClientOrderIndex int64
	BaseAmount       int64
	Price            uint32
	IsAsk            uint8
	Type             uint8
	TimeInForce      uint8
	ReduceOnly       uint8
	TriggerPrice     uint32
	OrderExpiry      int64
}

type CreateGroupedOrdersTxReq struct {
	GroupingType uint8
	Orders       []*CreateOrderTxReq
}

type ModifyOrderTxReq struct {
	MarketIndex  int16
	Index        int64
	BaseAmount   int64
	Price        uint32
	TriggerPrice uint32
}

type CancelOrderTxReq struct {
	MarketIndex int16
	Index       int64
}

type CancelAllOrdersTxReq struct {
	TimeInForce uint8
	Time        int64
}

type CreatePublicPoolTxReq struct {
	OperatorFee          int64
	InitialTotalShares   int64
	MinOperatorShareRate int64
}

type UpdatePublicPoolTxReq struct {
	PublicPoolIndex      int64
	Status               uint8
	OperatorFee          int64
	MinOperatorShareRate int64
}

type MintSharesTxReq struct {
	PublicPoolIndex int64
	ShareAmount     int64
}

type BurnSharesTxReq struct {
	PublicPoolIndex int64
	ShareAmount     int64
}

type UpdateLeverageTxReq struct {
	MarketIndex           int16
	InitialMarginFraction uint16
	MarginMode            uint8
}

type UpdateMarginTxReq struct {
	MarketIndex int16
	USDCAmount  int64
	Direction   uint8
}

func ConstructAuthToken(key signer.Signer, deadline time.Time, ops *TransactOpts) (string, error) {
	if ops.FromAccountIndex == nil {
		return "", fmt.Errorf("missing FromAccountIndex")
	}
	if ops.ApiKeyIndex == nil {
		return "", fmt.Errorf("missing ApiKeyIndex")
	}
	message := fmt.Sprintf("%v:%v:%v", deadline.Unix(), *ops.FromAccountIndex, *ops.ApiKeyIndex)

	msgInField, err := g.ArrayFromCanonicalLittleEndianBytes([]byte(message))
	if err != nil {
		return "", fmt.Errorf("failed to convert bytes to field element. message: %s, error: %w", message, err)
	}

	msgHash := p2.HashToQuinticExtension(msgInField).ToLittleEndianBytes()

	signatureBytes, err := key.Sign(msgHash, p2.NewPoseidon2())
	if err != nil {
		return "", err
	}
	signature := ethCommon.Bytes2Hex(signatureBytes)

	return fmt.Sprintf("%v:%v", message, signature), err
}

func ConstructCreateOrderTx(key signer.Signer, lighterChainId uint32, tx *CreateOrderTxReq, ops *TransactOpts) (*txtypes.L2CreateOrderTxInfo, error) {
	convertedTx := ConvertCreateOrderTx(tx, ops)
	err := convertedTx.Validate()
	if err != nil {
		return nil, err
	}

	msgHash, err := convertedTx.Hash(lighterChainId)
	if err != nil {
		return nil, err
	}

	signature, err := key.Sign(msgHash, p2.NewPoseidon2())
	if err != nil {
		return nil, err
	}

	convertedTx.SignedHash = ethCommon.Bytes2Hex(msgHash)
	convertedTx.Sig = signature
	return convertedTx, nil
}

func ConstructL2CancelOrderTx(key signer.Signer, lighterChainId uint32, tx *CancelOrderTxReq, ops *TransactOpts) (*txtypes.L2CancelOrderTxInfo, error) {
	convertedTx := ConvertCancelOrderTx(tx, ops)
	err := convertedTx.Validate()
	if err != nil {
		return nil, err
	}

	msgHash, err := convertedTx.Hash(lighterChainId)
	if err != nil {
		return nil, err
	}

	signature, err := key.Sign(msgHash, p2.NewPoseidon2())
	if err != nil {
		return nil, err
	}

	convertedTx.SignedHash = ethCommon.Bytes2Hex(msgHash)
	convertedTx.Sig = signature
	return convertedTx, nil
}

func ConstructL2ModifyOrderTx(key signer.Signer, lighterChainId uint32, tx *ModifyOrderTxReq, ops *TransactOpts) (*txtypes.L2ModifyOrderTxInfo, error) {
	convertedTx := ConvertModifyOrderTx(tx, ops)
	err := convertedTx.Validate()
	if err != nil {
		return nil, err
	}

	msgHash, err := convertedTx.Hash(lighterChainId)
	if err != nil {
		return nil, err
	}

	signature, err := key.Sign(msgHash, p2.NewPoseidon2())
	if err != nil {
		return nil, err
	}

	convertedTx.SignedHash = ethCommon.Bytes2Hex(msgHash)
	convertedTx.Sig = signature
	return convertedTx, nil
}

func ConstructUpdateLeverageTx(key signer.Signer, lighterChainId uint32, tx *UpdateLeverageTxReq, ops *TransactOpts) (*txtypes.L2UpdateLeverageTxInfo, error) {
	convertedTx := ConvertUpdateLeverageTx(tx, ops)
	err := convertedTx.Validate()
	if err != nil {
		return nil, err
	}

	msgHash, err := convertedTx.Hash(lighterChainId)
	if err != nil {
		return nil, err
	}

	signature, err := key.Sign(msgHash, p2.NewPoseidon2())
	if err != nil {
		return nil, err
	}

	convertedTx.SignedHash = ethCommon.Bytes2Hex(msgHash)
	convertedTx.Sig = signature
	return convertedTx, nil
}

func ConstructUpdateMarginTx(key signer.Signer, lighterChainId uint32, tx *UpdateMarginTxReq, ops *TransactOpts) (*txtypes.L2UpdateMarginTxInfo, error) {
	convertedTx := ConvertUpdateMarginTx(tx, ops)
	err := convertedTx.Validate()
	if err != nil {
		return nil, err
	}

	msgHash, err := convertedTx.Hash(lighterChainId)
	if err != nil {
		return nil, err
	}

	signature, err := key.Sign(msgHash, p2.NewPoseidon2())
	if err != nil {
		return nil, err
	}

	convertedTx.SignedHash = ethCommon.Bytes2Hex(msgHash)
	convertedTx.Sig = signature
	return convertedTx, nil
}

func ConvertCreateOrderTx(tx *CreateOrderTxReq, ops *TransactOpts) *txtypes.L2CreateOrderTxInfo {
	return &txtypes.L2CreateOrderTxInfo{
		AccountIndex: *ops.FromAccountIndex,
		ApiKeyIndex:  *ops.ApiKeyIndex,
		OrderInfo: &txtypes.OrderInfo{MarketIndex: tx.MarketIndex,
			ClientOrderIndex: tx.ClientOrderIndex,
			BaseAmount:       tx.BaseAmount,
			Price:            tx.Price,
			IsAsk:            tx.IsAsk,
			Type:             tx.Type,
			TimeInForce:      tx.TimeInForce,
			ReduceOnly:       tx.ReduceOnly,
			TriggerPrice:     tx.TriggerPrice,
			OrderExpiry:      tx.OrderExpiry,
		},
		ExpiredAt: ops.ExpiredAt,
		Nonce:     *ops.Nonce,
	}
}

func ConvertCancelOrderTx(tx *CancelOrderTxReq, ops *TransactOpts) *txtypes.L2CancelOrderTxInfo {
	return &txtypes.L2CancelOrderTxInfo{
		AccountIndex: *ops.FromAccountIndex,
		ApiKeyIndex:  *ops.ApiKeyIndex,
		MarketIndex:  tx.MarketIndex,
		Index:        tx.Index,
		ExpiredAt:    ops.ExpiredAt,
		Nonce:        *ops.Nonce,
	}
}

func ConvertModifyOrderTx(tx *ModifyOrderTxReq, ops *TransactOpts) *txtypes.L2ModifyOrderTxInfo {
	return &txtypes.L2ModifyOrderTxInfo{
		AccountIndex: *ops.FromAccountIndex,
		ApiKeyIndex:  *ops.ApiKeyIndex,
		MarketIndex:  tx.MarketIndex,
		Index:        tx.Index,
		BaseAmount:   tx.BaseAmount,
		Price:        tx.Price,
		TriggerPrice: tx.TriggerPrice,
		ExpiredAt:    ops.ExpiredAt,
		Nonce:        *ops.Nonce,
	}
}

func ConvertUpdateLeverageTx(tx *UpdateLeverageTxReq, ops *TransactOpts) *txtypes.L2UpdateLeverageTxInfo {
	return &txtypes.L2UpdateLeverageTxInfo{
		AccountIndex:          *ops.FromAccountIndex,
		ApiKeyIndex:           *ops.ApiKeyIndex,
		MarketIndex:           tx.MarketIndex,
		InitialMarginFraction: tx.InitialMarginFraction,
		MarginMode:            tx.MarginMode,
		ExpiredAt:             ops.ExpiredAt,
		Nonce:                 *ops.Nonce,
	}
}

func ConvertUpdateMarginTx(tx *UpdateMarginTxReq, ops *TransactOpts) *txtypes.L2UpdateMarginTxInfo {
	return &txtypes.L2UpdateMarginTxInfo{
		AccountIndex: *ops.FromAccountIndex,
		ApiKeyIndex:  *ops.ApiKeyIndex,
		MarketIndex:  tx.MarketIndex,
		USDCAmount:   tx.USDCAmount,
		Direction:    tx.Direction,
		ExpiredAt:    ops.ExpiredAt,
		Nonce:        *ops.Nonce,
	}
}
