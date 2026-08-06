package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/elliottech/lighter-go/signer"
	"github.com/elliottech/lighter-go/types/txtypes"
	g "github.com/elliottech/poseidon_crypto/field/goldilocks"
	p2 "github.com/elliottech/poseidon_crypto/hash/poseidon2_goldilocks"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

const (
	defaultExpireTime = time.Minute*10 - time.Second
)

type LightClient struct {
	chainId      uint32
	accountIndex *int64
	apiKeyIndex  uint8
	keyManager   signer.KeyManager
	skipNonce    bool
}

func NewLightClient(seed string, chainId uint32) (*LightClient, error) {
	keyManager, err := signer.NewSeedKeyManager(seed)
	if err != nil {
		return nil, err
	}

	return &LightClient{
		chainId:    chainId,
		keyManager: keyManager,
	}, nil
}

func NewLightClientByPrv(prv string, chainId uint32) (*LightClient, error) {
	prv = strings.TrimPrefix(prv, "0x")
	prvBytes, err := hex.DecodeString(prv)
	if err != nil {
		return nil, fmt.Errorf("failed to decode private key hex string: %w", err)
	}

	keyManager, err := signer.NewKeyManager(prvBytes)
	if err != nil {
		return nil, err
	}

	return &LightClient{
		chainId:    chainId,
		keyManager: keyManager,
	}, nil
}

func (c *LightClient) SetAccountIndex(accountIndex int64) {
	c.accountIndex = &accountIndex
}

func (c *LightClient) SetApiKeyIndex(apiKeyIndex uint8) {
	c.apiKeyIndex = apiKeyIndex
}

func (c *LightClient) SetSkipNonce(skipNonce bool) {
	c.skipNonce = skipNonce
}

func (c *LightClient) KeyManager() signer.KeyManager {
	return c.keyManager
}

func (c *LightClient) ChainId() uint32 {
	return c.chainId
}

func (c *LightClient) FullFillDefaultOps(ops *TransactOpts) (*TransactOpts, error) {
	if ops == nil {
		ops = new(TransactOpts)
	}
	if ops.ExpiredAt == 0 {
		ops.ExpiredAt = time.Now().Add(defaultExpireTime).UnixMilli()
	}
	if ops.FromAccountIndex == nil {
		if c.accountIndex != nil {
			accountIndex := *c.accountIndex
			ops.FromAccountIndex = &accountIndex
		}
	}
	if ops.ApiKeyIndex == nil {
		ops.ApiKeyIndex = &c.apiKeyIndex
	}
	if c.skipNonce {
		ops.SkipNonce = true
	}

	return ops, nil
}

func (c *LightClient) GenerateChangePubKeySignBody(txData *ChangePubKeyReq, ops *TransactOpts) (string, error) {
	ops, err := c.FullFillDefaultOps(ops)
	if err != nil {
		return "", err
	}

	txInfo, err := ConstructChangePubKeyTx(c.keyManager, c.chainId, txData, ops)
	if err != nil {
		return "", err
	}
	return txInfo.GetL1SignatureBody(), nil
}

func (c *LightClient) GenerateTransferSignBody(txData *TransferTxReq, ops *TransactOpts) (string, error) {
	ops, err := c.FullFillDefaultOps(ops)
	if err != nil {
		return "", err
	}

	txInfo, err := ConstructTransferTx(c.keyManager, c.chainId, txData, ops)
	if err != nil {
		return "", err
	}
	return txInfo.GetL1SignatureBody(c.chainId), nil
}

func (c *LightClient) GetChangePubKeyTransaction(tx *ChangePubKeyReq, ops *TransactOpts, signatureList ...string) (*txtypes.L2ChangePubKeyTxInfo, error) {
	if c.keyManager == nil {
		return nil, fmt.Errorf("key manager is nil")
	}

	ops, err := c.FullFillDefaultOps(ops)
	if err != nil {
		return nil, err
	}

	txInfo, err := ConstructChangePubKeyTx(c.keyManager, c.chainId, tx, ops)
	if err != nil {
		return nil, err
	}

	signature := ""
	if len(signatureList) == 0 {
		return nil, errors.New("signature is expected to be passed")
	} else if len(signatureList) == 1 {
		signature = signatureList[0]
	} else {
		return nil, errors.New("multiple signatures provided")
	}

	txInfo.L1Sig = signature
	return txInfo, nil
}

func (c *LightClient) GetCreateSubAccountTransaction(ops *TransactOpts) (*txtypes.L2CreateSubAccountTxInfo, error) {
	ops, err := c.FullFillDefaultOps(ops)
	if err != nil {
		return nil, err
	}
	return ConstructCreateSubAccountTx(c.keyManager, c.chainId, ops)
}

func (c *LightClient) GetCreatePublicPoolTransaction(tx *CreatePublicPoolTxReq, ops *TransactOpts) (*txtypes.L2CreatePublicPoolTxInfo, error) {
	ops, err := c.FullFillDefaultOps(ops)
	if err != nil {
		return nil, err
	}
	return ConstructCreatePublicPoolTx(c.keyManager, c.chainId, tx, ops)
}

func (c *LightClient) GetUpdatePublicPoolTransaction(tx *UpdatePublicPoolTxReq, ops *TransactOpts) (*txtypes.L2UpdatePublicPoolTxInfo, error) {
	ops, err := c.FullFillDefaultOps(ops)
	if err != nil {
		return nil, err
	}
	return ConstructUpdatePublicPoolTx(c.keyManager, c.chainId, tx, ops)
}

func (c *LightClient) GetCreateOrderTransaction(tx *CreateOrderTxReq, ops *TransactOpts) (*txtypes.L2CreateOrderTxInfo, error) {
	ops, err := c.FullFillDefaultOps(ops)
	if err != nil {
		return nil, err
	}
	return ConstructCreateOrderTx(c.keyManager, c.chainId, tx, ops)
}

func (c *LightClient) GetCancelOrderTransaction(tx *CancelOrderTxReq, ops *TransactOpts) (*txtypes.L2CancelOrderTxInfo, error) {
	ops, err := c.FullFillDefaultOps(ops)
	if err != nil {
		return nil, err
	}
	return ConstructL2CancelOrderTx(c.keyManager, c.chainId, tx, ops)
}

func (c *LightClient) GetWithdrawTransaction(tx *WithdrawTxReq, ops *TransactOpts) (*txtypes.L2WithdrawTxInfo, error) {
	ops, err := c.FullFillDefaultOps(ops)
	if err != nil {
		return nil, err
	}
	return ConstructWithdrawTx(c.keyManager, c.chainId, tx, ops)
}

func (c *LightClient) GetCancelAllOrdersTransaction(tx *CancelAllOrdersTxReq, ops *TransactOpts) (*txtypes.L2CancelAllOrdersTxInfo, error) {
	ops, err := c.FullFillDefaultOps(ops)
	if err != nil {
		return nil, err
	}
	return ConstructL2CancelAllOrdersTx(c.keyManager, c.chainId, tx, ops)
}

func (c *LightClient) GetModifyOrderTransaction(tx *ModifyOrderTxReq, ops *TransactOpts) (*txtypes.L2ModifyOrderTxInfo, error) {
	ops, err := c.FullFillDefaultOps(ops)
	if err != nil {
		return nil, err
	}
	return ConstructL2ModifyOrderTx(c.keyManager, c.chainId, tx, ops)
}

func (c *LightClient) GetTransferTransaction(tx *TransferTxReq, ops *TransactOpts, signatureList ...string) (*txtypes.L2TransferTxInfo, error) {
	ops, err := c.FullFillDefaultOps(ops)
	if err != nil {
		return nil, err
	}

	txInfo, err := ConstructTransferTx(c.keyManager, c.chainId, tx, ops)
	if err != nil {
		return nil, err
	}

	signature := ""
	if len(signatureList) == 0 {
		return nil, errors.New("signature is expected to be passed")
	} else if len(signatureList) == 1 {
		signature = signatureList[0]
	} else {
		return nil, errors.New("multiple signatures provided")
	}

	if signature != "" && signature != "0x" {
		txInfo.L1Sig = signature
	}

	return txInfo, nil
}

func (c *LightClient) GetMintSharesTransaction(tx *MintSharesTxReq, ops *TransactOpts) (*txtypes.L2MintSharesTxInfo, error) {
	ops, err := c.FullFillDefaultOps(ops)
	if err != nil {
		return nil, err
	}
	return ConstructMintSharesTx(c.keyManager, c.chainId, tx, ops)
}

func (c *LightClient) GetBurnSharesTransaction(tx *BurnSharesTxReq, ops *TransactOpts) (*txtypes.L2BurnSharesTxInfo, error) {
	ops, err := c.FullFillDefaultOps(ops)
	if err != nil {
		return nil, err
	}
	return ConstructBurnSharesTx(c.keyManager, c.chainId, tx, ops)
}

func (c *LightClient) GetStakeAssetsTransaction(tx *StakeAssetsTxReq, ops *TransactOpts) (*txtypes.L2StakeAssetsTxInfo, error) {
	ops, err := c.FullFillDefaultOps(ops)
	if err != nil {
		return nil, err
	}
	return ConstructStakeAssetsTx(c.keyManager, c.chainId, tx, ops)
}

func (c *LightClient) GetUnstakeAssetsTransaction(tx *UnstakeAssetsTxReq, ops *TransactOpts) (*txtypes.L2UnstakeAssetsTxInfo, error) {
	ops, err := c.FullFillDefaultOps(ops)
	if err != nil {
		return nil, err
	}
	return ConstructUnstakeAssetsTx(c.keyManager, c.chainId, tx, ops)
}

func (c *LightClient) GetUpdateLeverageTransaction(tx *UpdateLeverageTxReq, ops *TransactOpts) (*txtypes.L2UpdateLeverageTxInfo, error) {
	ops, err := c.FullFillDefaultOps(ops)
	if err != nil {
		return nil, err
	}
	return ConstructUpdateLeverageTx(c.keyManager, c.chainId, tx, ops)
}

func (c *LightClient) GetUpdateAccountConfigTransaction(tx *UpdateAccountConfigTxReq, ops *TransactOpts) (*txtypes.L2UpdateAccountConfigTxInfo, error) {
	ops, err := c.FullFillDefaultOps(ops)
	if err != nil {
		return nil, err
	}
	return ConstructUpdateAccountConfigTx(c.keyManager, c.chainId, tx, ops)
}

func (c *LightClient) GetUpdateAccountAssetConfigTransaction(tx *UpdateAccountAssetConfigTxReq, ops *TransactOpts) (*txtypes.L2UpdateAccountAssetConfigTxInfo, error) {
	ops, err := c.FullFillDefaultOps(ops)
	if err != nil {
		return nil, err
	}
	return ConstructUpdateAccountAssetConfigTx(c.keyManager, c.chainId, tx, ops)
}

func (c *LightClient) GetUpdateMarginTransaction(tx *UpdateMarginTxReq, ops *TransactOpts) (*txtypes.L2UpdateMarginTxInfo, error) {
	ops, err := c.FullFillDefaultOps(ops)
	if err != nil {
		return nil, err
	}
	return ConstructUpdateMarginTx(c.keyManager, c.chainId, tx, ops)
}

func (c *LightClient) GetCreateGroupedOrdersTransaction(tx *CreateGroupedOrdersTxReq, ops *TransactOpts) (*txtypes.L2CreateGroupedOrdersTxInfo, error) {
	ops, err := c.FullFillDefaultOps(ops)
	if err != nil {
		return nil, err
	}
	return ConstructL2CreateGroupedOrdersTx(c.keyManager, c.chainId, tx, ops)
}

func (c *LightClient) SignMessage(message string) (string, error) {
	msgInField, err := g.ArrayFromCanonicalLittleEndianBytes([]byte(message))
	if err != nil {
		return "", fmt.Errorf("failed to convert bytes to field element. message: %s, error: %w", message, err)
	}

	msgHash := p2.HashToQuinticExtension(msgInField).ToLittleEndianBytes()

	signature, err := c.keyManager.Sign(msgHash, p2.NewPoseidon2())
	if err != nil {
		return "", err
	}
	return common.Bytes2Hex(signature), nil
}

func errToJson(err error) map[string]any {
	return map[string]any{
		"error": err.Error(),
	}
}

func getTxResponse(tx interface{}, lighterChainId uint32) map[string]any {
	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return errToJson(err)
	}

	var hash []byte
	switch t := tx.(type) {
	case *txtypes.L2ChangePubKeyTxInfo:
		hash, err = t.Hash(lighterChainId)
	case *txtypes.L2CreateSubAccountTxInfo:
		hash, err = t.Hash(lighterChainId)
	case *txtypes.L2CreatePublicPoolTxInfo:
		hash, err = t.Hash(lighterChainId)
	case *txtypes.L2UpdatePublicPoolTxInfo:
		hash, err = t.Hash(lighterChainId)
	case *txtypes.L2CreateOrderTxInfo:
		hash, err = t.Hash(lighterChainId)
	case *txtypes.L2CancelOrderTxInfo:
		hash, err = t.Hash(lighterChainId)
	case *txtypes.L2CancelAllOrdersTxInfo:
		hash, err = t.Hash(lighterChainId)
	case *txtypes.L2ModifyOrderTxInfo:
		hash, err = t.Hash(lighterChainId)
	case *txtypes.L2TransferTxInfo:
		hash, err = t.Hash(lighterChainId)
	case *txtypes.L2MintSharesTxInfo:
		hash, err = t.Hash(lighterChainId)
	case *txtypes.L2BurnSharesTxInfo:
		hash, err = t.Hash(lighterChainId)
	case *txtypes.L2UpdateLeverageTxInfo:
		hash, err = t.Hash(lighterChainId)
	case *txtypes.L2UpdateAccountConfigTxInfo:
		hash, err = t.Hash(lighterChainId)
	case *txtypes.L2UpdateAccountAssetConfigTxInfo:
		hash, err = t.Hash(lighterChainId)
	case *txtypes.L2UpdateMarginTxInfo:
		hash, err = t.Hash(lighterChainId)
	case *txtypes.L2WithdrawTxInfo:
		hash, err = t.Hash(lighterChainId)
	case *txtypes.L2CreateGroupedOrdersTxInfo:
		hash, err = t.Hash(lighterChainId)
	case *txtypes.L2UnstakeAssetsTxInfo:
		hash, err = t.Hash(lighterChainId)
	case *txtypes.L2StakeAssetsTxInfo:
		hash, err = t.Hash(lighterChainId)
	default:
		return errToJson(fmt.Errorf("unknown tx type"))
	}

	if err != nil {
		return errToJson(err)
	}

	return map[string]any{
		"txHash": common.Bytes2Hex(hash),
		"txInfo": string(txInfoBytes),
	}
}

func getHex10FromUint64(value uint64) string {
	v := hexutil.EncodeUint64(value)
	v = strings.Replace(v, "0x", "", 1)

	// Make sure result has fixed bytes
	vBytes := []byte(v)
	if len(vBytes) < 16 {
		toAppend := make([]byte, 16-len(vBytes))
		for i := range toAppend {
			toAppend[i] = 48
		}
		vBytes = append(toAppend, vBytes...)
	}

	return fmt.Sprintf("0x%s", string(vBytes))
}
