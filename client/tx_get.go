package client

import (
	"fmt"

	"github.com/elliottech/lighter-go/types"
	"github.com/elliottech/lighter-go/types/txtypes"
)

func (c *TxClient) GetCreateOrderTransaction(tx *types.CreateOrderTxReq, ops *types.TransactOpts) (*txtypes.L2CreateOrderTxInfo, error) {
	ops, err := c.FullFillDefaultOps(ops)
	if err != nil {
		return nil, err
	}
	txInfo, err := types.ConstructCreateOrderTx(c.keyManager, c.chainId, tx, ops)
	if err != nil {
		return nil, err
	}
	return txInfo, nil
}

func (c *TxClient) GetCancelOrderTransaction(tx *types.CancelOrderTxReq, ops *types.TransactOpts) (*txtypes.L2CancelOrderTxInfo, error) {
	ops, err := c.FullFillDefaultOps(ops)
	if err != nil {
		return nil, err
	}
	txInfo, err := types.ConstructL2CancelOrderTx(c.keyManager, c.chainId, tx, ops)
	if err != nil {
		return nil, err
	}
	return txInfo, nil
}

func (c *TxClient) GetModifyOrderTransaction(tx *types.ModifyOrderTxReq, ops *types.TransactOpts) (*txtypes.L2ModifyOrderTxInfo, error) {
	ops, err := c.FullFillDefaultOps(ops)
	if err != nil {
		return nil, err
	}

	txInfo, err := types.ConstructL2ModifyOrderTx(c.keyManager, c.chainId, tx, ops)
	if err != nil {
		return nil, err
	}

	return txInfo, nil
}

func (c *TxClient) GetUpdateLeverageTransaction(tx *types.UpdateLeverageTxReq, ops *types.TransactOpts) (*txtypes.L2UpdateLeverageTxInfo, error) {
	ops, err := c.FullFillDefaultOps(ops)
	if err != nil {
		return nil, err
	}
	txInfo, err := types.ConstructUpdateLeverageTx(c.keyManager, c.chainId, tx, ops)
	if err != nil {
		return nil, err
	}
	return txInfo, nil
}

func (c *TxClient) GetUpdateMarginTransaction(tx *types.UpdateMarginTxReq, ops *types.TransactOpts) (*txtypes.L2UpdateMarginTxInfo, error) {
	if c.keyManager == nil {
		return nil, fmt.Errorf("key manager is nil")
	}

	if ops == nil {
		ops = new(types.TransactOpts)
	}

	txInfo, err := types.ConstructUpdateMarginTx(c.keyManager, c.chainId, tx, ops)
	if err != nil {
		return nil, err
	}
	return txInfo, nil
}
