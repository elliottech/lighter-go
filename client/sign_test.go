package client

import (
	"testing"

	"github.com/elliottech/lighter-go/types"
	"github.com/elliottech/lighter-go/types/txtypes"
)

const (
	testChainID      uint32 = 304
	testAccountIndex int64  = 1
	testAPIKeyIndex  uint8  = 0
	testNonce        int64  = 42
)

func newTestClient(t *testing.T, privateKey string) *TxClient {
	t.Helper()
	c, err := NewTxClient(nil, privateKey, testAccountIndex, testAPIKeyIndex, testChainID)
	if err != nil {
		t.Fatalf("NewTxClient failed: %v", err)
	}
	return c
}

func opsWithSkipNonce(skipNonce uint8, nonce int64) *types.TransactOpts {
	attr := &types.L2TxAttributes{}
	if skipNonce == 1 {
		sn := skipNonce
		attr.SkipNonce = &sn
	}
	n := nonce
	return &types.TransactOpts{
		Nonce:        &n,
		TxAttributes: attr,
	}
}

func assertSkipNonce(t *testing.T, name string, attrs txtypes.L2TxAttributes, expectedSet bool) {
	t.Helper()
	if expectedSet {
		if attrs == nil {
			t.Errorf("%s: attributes map should be non-nil when skipNonce=1", name)
			return
		}
		att, ok := attrs[txtypes.AttributeTypeSkipTxNonce]
		if !ok {
			t.Errorf("%s: SkipTxNonce key missing", name)
			return
		}
		if att != 1 {
			t.Errorf("%s: SkipTxNonce = %d, want 1", name, att)
		}
		return
	}
	if attrs != nil {
		t.Errorf("%s: attributes map should be nil when skipNonce=0, got %v", name, attrs)
	}
}

func TestGenerateAPIKey(t *testing.T) {
	priv, pub, err := GenerateAPIKey()
	if err != nil {
		t.Fatalf("GenerateAPIKey error: %v", err)
	}
	if len(priv) < 3 || priv[:2] != "0x" {
		t.Errorf("privateKey should be a 0x-prefixed hex string, got %q", priv)
	}
	if len(pub) < 3 || pub[:2] != "0x" {
		t.Errorf("publicKey should be a 0x-prefixed hex string, got %q", pub)
	}
}

func TestCreateClient(t *testing.T) {
	priv, _, err := GenerateAPIKey()
	if err != nil {
		t.Fatalf("GenerateAPIKey error: %v", err)
	}
	c := newTestClient(t, priv)
	if c.GetChainId() != testChainID {
		t.Errorf("chainId = %d, want %d", c.GetChainId(), testChainID)
	}
	if c.GetAccountIndex() != testAccountIndex {
		t.Errorf("accountIndex = %d, want %d", c.GetAccountIndex(), testAccountIndex)
	}
	if c.GetApiKeyIndex() != testAPIKeyIndex {
		t.Errorf("apiKeyIndex = %d, want %d", c.GetApiKeyIndex(), testAPIKeyIndex)
	}
}

func TestSignCancelOrder(t *testing.T) {
	priv, _, err := GenerateAPIKey()
	if err != nil {
		t.Fatalf("GenerateAPIKey error: %v", err)
	}
	c := newTestClient(t, priv)
	req := &types.CancelOrderTxReq{MarketIndex: 0, Index: 12345}

	// skipNonce = 1
	tx1, err := c.GetCancelOrderTransaction(req, opsWithSkipNonce(1, testNonce))
	if err != nil {
		t.Fatalf("GetCancelOrderTransaction (skipNonce=1) failed: %v", err)
	}
	if tx1.GetTxType() != txtypes.TxTypeL2CancelOrder {
		t.Errorf("txType = %d, want %d", tx1.GetTxType(), txtypes.TxTypeL2CancelOrder)
	}
	if tx1.AccountIndex != testAccountIndex {
		t.Errorf("AccountIndex = %d, want %d", tx1.AccountIndex, testAccountIndex)
	}
	if tx1.ApiKeyIndex != testAPIKeyIndex {
		t.Errorf("ApiKeyIndex = %d, want %d", tx1.ApiKeyIndex, testAPIKeyIndex)
	}
	if tx1.MarketIndex != 0 {
		t.Errorf("MarketIndex = %d, want 0", tx1.MarketIndex)
	}
	if tx1.Index != 12345 {
		t.Errorf("Index = %d, want 12345", tx1.Index)
	}
	if tx1.Nonce != testNonce {
		t.Errorf("Nonce = %d, want %d", tx1.Nonce, testNonce)
	}
	if len(tx1.Sig) == 0 {
		t.Error("Sig should not be empty")
	}
	assertSkipNonce(t, "SignCancelOrder(skipNonce=1)", tx1.L2TxAttributes, true)

	tx2, err := c.GetCancelOrderTransaction(req, opsWithSkipNonce(0, testNonce))
	if err != nil {
		t.Fatalf("GetCancelOrderTransaction (skipNonce=0) failed: %v", err)
	}
	assertSkipNonce(t, "SignCancelOrder(skipNonce=0)", tx2.L2TxAttributes, false)

	if tx1.SignedHash == tx2.SignedHash {
		t.Error("skipNonce flag should affect the signed tx hash, but hashes are identical")
	}
}

func TestSignCancelAllOrders(t *testing.T) {
	priv, _, err := GenerateAPIKey()
	if err != nil {
		t.Fatalf("GenerateAPIKey error: %v", err)
	}
	c := newTestClient(t, priv)
	req := &types.CancelAllOrdersTxReq{TimeInForce: 0, Time: 0}

	tx, err := c.GetCancelAllOrdersTransaction(req, opsWithSkipNonce(1, testNonce))
	if err != nil {
		t.Fatalf("GetCancelAllOrdersTransaction failed: %v", err)
	}
	if tx.GetTxType() != txtypes.TxTypeL2CancelAllOrders {
		t.Errorf("txType = %d, want %d", tx.GetTxType(), txtypes.TxTypeL2CancelAllOrders)
	}
	if tx.TimeInForce != 0 {
		t.Errorf("TimeInForce = %d, want 0", tx.TimeInForce)
	}
	if tx.Time != 0 {
		t.Errorf("Time = %d, want 0", tx.Time)
	}
	assertSkipNonce(t, "SignCancelAllOrders(skipNonce=1)", tx.L2TxAttributes, true)
}

func TestSignCreateOrder(t *testing.T) {
	priv, _, err := GenerateAPIKey()
	if err != nil {
		t.Fatalf("GenerateAPIKey error: %v", err)
	}
	c := newTestClient(t, priv)

	req := &types.CreateOrderTxReq{
		MarketIndex:      0,
		ClientOrderIndex: 1,
		BaseAmount:       1000,
		Price:            50000,
		IsAsk:            0,
		Type:             0,
		TimeInForce:      0,
		ReduceOnly:       0,
		TriggerPrice:     0,
		OrderExpiry:      0,
	}
	tx, err := c.GetCreateOrderTransaction(req, opsWithSkipNonce(1, testNonce))
	if err != nil {
		t.Fatalf("GetCreateOrderTransaction failed: %v", err)
	}
	if tx.GetTxType() != txtypes.TxTypeL2CreateOrder {
		t.Errorf("txType = %d, want %d", tx.GetTxType(), txtypes.TxTypeL2CreateOrder)
	}
	if tx.MarketIndex != 0 {
		t.Errorf("MarketIndex = %d, want 0", tx.MarketIndex)
	}
	if tx.ClientOrderIndex != 1 {
		t.Errorf("ClientOrderIndex = %d, want 1", tx.ClientOrderIndex)
	}
	if tx.BaseAmount != 1000 {
		t.Errorf("BaseAmount = %d, want 1000", tx.BaseAmount)
	}
	if tx.Price != 50000 {
		t.Errorf("Price = %d, want 50000", tx.Price)
	}
	assertSkipNonce(t, "SignCreateOrder(skipNonce=1)", tx.L2TxAttributes, true)
}

func TestSignCreateSubAccount(t *testing.T) {
	priv, _, err := GenerateAPIKey()
	if err != nil {
		t.Fatalf("GenerateAPIKey error: %v", err)
	}
	c := newTestClient(t, priv)

	tx, err := c.GetCreateSubAccountTransaction(opsWithSkipNonce(1, testNonce))
	if err != nil {
		t.Fatalf("GetCreateSubAccountTransaction failed: %v", err)
	}
	if tx.GetTxType() != txtypes.TxTypeL2CreateSubAccount {
		t.Errorf("txType = %d, want %d", tx.GetTxType(), txtypes.TxTypeL2CreateSubAccount)
	}
	if tx.AccountIndex != testAccountIndex {
		t.Errorf("AccountIndex = %d, want %d", tx.AccountIndex, testAccountIndex)
	}
	if tx.Nonce != testNonce {
		t.Errorf("Nonce = %d, want %d", tx.Nonce, testNonce)
	}
	assertSkipNonce(t, "SignCreateSubAccount(skipNonce=1)", tx.L2TxAttributes, true)
}

func TestSignUpdateLeverage(t *testing.T) {
	priv, _, err := GenerateAPIKey()
	if err != nil {
		t.Fatalf("GenerateAPIKey error: %v", err)
	}
	c := newTestClient(t, priv)

	req := &types.UpdateLeverageTxReq{
		MarketIndex:           0,
		InitialMarginFraction: 100,
		MarginMode:            0,
	}
	tx, err := c.GetUpdateLeverageTransaction(req, opsWithSkipNonce(1, testNonce))
	if err != nil {
		t.Fatalf("GetUpdateLeverageTransaction failed: %v", err)
	}
	if tx.GetTxType() != txtypes.TxTypeL2UpdateLeverage {
		t.Errorf("txType = %d, want %d", tx.GetTxType(), txtypes.TxTypeL2UpdateLeverage)
	}
	if tx.MarketIndex != 0 {
		t.Errorf("MarketIndex = %d, want 0", tx.MarketIndex)
	}
	if tx.InitialMarginFraction != 100 {
		t.Errorf("InitialMarginFraction = %d, want 100", tx.InitialMarginFraction)
	}
	assertSkipNonce(t, "SignUpdateLeverage(skipNonce=1)", tx.L2TxAttributes, true)
}

func TestSignCreateGroupedOrders(t *testing.T) {
	priv, _, err := GenerateAPIKey()
	if err != nil {
		t.Fatalf("GenerateAPIKey error: %v", err)
	}
	c := newTestClient(t, priv)

	const expiry int64 = 7 * 24 * 60 * 60 * 1000

	req := &types.CreateGroupedOrdersTxReq{
		GroupingType: 1,
		Orders: []*types.CreateOrderTxReq{
			{
				MarketIndex: 0, ClientOrderIndex: 0, BaseAmount: 1000, Price: 50000,
				IsAsk: 0, Type: 0, TimeInForce: 1, ReduceOnly: 0, TriggerPrice: 0, OrderExpiry: expiry,
			},
			{
				MarketIndex: 0, ClientOrderIndex: 0, BaseAmount: 0, Price: 51000,
				IsAsk: 1, Type: 4, TimeInForce: 0, ReduceOnly: 1, TriggerPrice: 49000, OrderExpiry: expiry,
			},
		},
	}

	tx, err := c.GetCreateGroupedOrdersTransaction(req, opsWithSkipNonce(1, testNonce))
	if err != nil {
		t.Fatalf("GetCreateGroupedOrdersTransaction failed: %v", err)
	}
	if tx.GetTxType() != txtypes.TxTypeL2CreateGroupedOrders {
		t.Errorf("txType = %d, want %d", tx.GetTxType(), txtypes.TxTypeL2CreateGroupedOrders)
	}
	if tx.GroupingType != 1 {
		t.Errorf("GroupingType = %d, want 1", tx.GroupingType)
	}
	if len(tx.Orders) != 2 {
		t.Fatalf("Orders len = %d, want 2", len(tx.Orders))
	}
	if tx.Orders[0].TimeInForce != 1 {
		t.Errorf("Orders[0].TimeInForce = %d, want 1", tx.Orders[0].TimeInForce)
	}
	if tx.Orders[1].Type != 4 {
		t.Errorf("Orders[1].Type = %d, want 4", tx.Orders[1].Type)
	}
	if tx.Orders[1].TriggerPrice != 49000 {
		t.Errorf("Orders[1].TriggerPrice = %d, want 49000", tx.Orders[1].TriggerPrice)
	}
	assertSkipNonce(t, "SignCreateGroupedOrders(skipNonce=1)", tx.L2TxAttributes, true)
}
