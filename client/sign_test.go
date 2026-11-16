package client

import (
	"encoding/hex"
	"errors"
	"strings"
	"testing"

	"github.com/elliottech/lighter-go/types"
	"github.com/elliottech/lighter-go/types/txtypes"
	schnorr "github.com/elliottech/poseidon_crypto/signature/schnorr"
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

func TestGetChangePubKeyTransactionNilApiClient(t *testing.T) {
	priv, _, errr := GenerateAPIKey()
	if errr != nil {
		t.Fatalf("GenerateAPIKey error: %v", errr)
	}
	c := newTestClient(t, priv)

	nonce := testNonce
	_, err := c.GetChangePubKeyTransaction(
		&types.ChangePubKeyReq{PubKey: c.keyManager.PubKeyBytes()},
		&types.TransactOpts{Nonce: &nonce},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// Hash and owner message from Core TestSendToEvmHashGoldenVector.
func TestSignSendToEvm(t *testing.T) {
	priv, _, err := GenerateAPIKey()
	if err != nil {
		t.Fatal(err)
	}
	c, err := NewTxClient(nil, priv, 5, 2, 300)
	if err != nil {
		t.Fatal(err)
	}
	nonce := int64(3)
	req := &types.SendToEvmTxReq{
		AssetIndex: 7, EvmRecipient: "0x70997970C51812dc3A010C7d01b50e0d17dc79C8",
		RouteType: 1, BaseAmount: 1000000, USDCFee: 100,
	}
	tx, err := c.GetSendToEvmTransaction(req, &types.TransactOpts{Nonce: &nonce, ExpiredAt: 1700000000000})
	if err != nil {
		t.Fatal(err)
	}
	if tx.GetTxType() != txtypes.TxTypeL2SendToEvm {
		t.Errorf("unexpected txType: %d", tx.GetTxType())
	}
	const expectedHash = "759cb61dc1fb4c056a366a6e9243b4ddf08266f976548971dcc8abdae2fc65f0ff972cc7f302e20d"
	if tx.GetTxHash() != expectedHash {
		t.Fatalf("hash = %s, want %s", tx.GetTxHash(), expectedHash)
	}
	const expectedMessage = "Send to EVM\n\nnonce: 0x0000000000000003\nfrom: 0x0000000000000005 (route 0x0000000000000001)\napi key: 0x0000000000000002\nto evm: 0x70997970C51812dc3A010C7d01b50e0d17dc79C8\nasset: 0x0000000000000007\namount: 0x00000000000f4240\nfee: 0x0000000000000064\nchainId: 0x000000000000012c\nOnly sign this message for a trusted client!"
	if tx.GetOwnerSignatureBody(300) != expectedMessage {
		t.Error("owner message differs from Core")
	}
	hash, err := hex.DecodeString(expectedHash)
	if err != nil {
		t.Fatal(err)
	}
	pub := c.GetKeyManager().PubKeyBytes()
	if err := schnorr.Validate(pub[:], hash, tx.Sig); err != nil {
		t.Fatalf("Core hash rejects signature: %v", err)
	}
	assertSkipNonce(t, "SignSendToEvm", tx.L2TxAttributes, false)
	tx.OwnerSig = "0x" + strings.Repeat("11", 65)
	hash, err = tx.Hash(300)
	if err != nil || hex.EncodeToString(hash) != expectedHash {
		t.Fatal("OwnerSig must not change the API-key hash")
	}

	// Core TestSendToEvmHashGoldenVector fixture with nonzero upper amount and
	// fee limbs. The expected hash was computed by Core.
	largeReq := *req
	largeReq.BaseAmount = 4294967303
	largeReq.USDCFee = 8589934601
	tx, err = c.GetSendToEvmTransaction(&largeReq, &types.TransactOpts{Nonce: &nonce, ExpiredAt: 1700000000000})
	if err != nil {
		t.Fatal(err)
	}
	const expectedLargeHash = "281d3dd0560230803f8b29c7c21df7d1fb94ba72675f69bdaee08a861bc5d6a591abf3a64d8fd477"
	if tx.GetTxHash() != expectedLargeHash {
		t.Fatalf("hash = %s, want %s", tx.GetTxHash(), expectedLargeHash)
	}
	hash, err = hex.DecodeString(expectedLargeHash)
	if err != nil {
		t.Fatal(err)
	}
	if err := schnorr.Validate(pub[:], hash, tx.Sig); err != nil {
		t.Fatalf("Core hash rejects signature: %v", err)
	}

	// Also match backend TestSignerProtocol.
	c, err = NewTxClient(nil, priv, 100, 0, 300)
	if err != nil {
		t.Fatal(err)
	}
	nonce = 0
	tx, err = c.GetSendToEvmTransaction(&types.SendToEvmTxReq{
		AssetIndex: 1, EvmRecipient: "0x1234567890123456789012345678901234567890",
		RouteType: 1, BaseAmount: 100000,
	}, &types.TransactOpts{Nonce: &nonce, ExpiredAt: 1800000000000})
	if err != nil {
		t.Fatal(err)
	}
	if tx.GetTxHash() != "8855342f8fe15d1483ec91c41896bb32c7fd5db391e08f95348bb861417070e2afdecc9c33cd3cbd" {
		t.Error("hash differs from backend TestSignerProtocol")
	}
	tx.AccountIndex = -1
	if err := tx.Validate(); err != txtypes.ErrInteropAccountIndexTooLow {
		t.Errorf("negative account index accepted: %v", err)
	}
	for _, recipient := range []string{"0x0000000000000000000000000000000000000000", "0x1234", "invalid"} {
		req.EvmRecipient = recipient
		if _, err := c.GetSendToEvmTransaction(req, &types.TransactOpts{Nonce: &nonce}); err != txtypes.ErrEvmRecipientInvalid {
			t.Errorf("invalid recipient %q accepted: %v", recipient, err)
		}
	}
}

// Hash and owner message from Core TestWithdrawToL1AddressHashGoldenVector.
func TestSignWithdrawToL1Address(t *testing.T) {
	priv, _, err := GenerateAPIKey()
	if err != nil {
		t.Fatal(err)
	}
	c, err := NewTxClient(nil, priv, 5, 2, 300)
	if err != nil {
		t.Fatal(err)
	}
	nonce := int64(3)
	req := &types.WithdrawToL1AddressTxReq{
		AssetIndex: 7, L1Recipient: "0x70997970C51812dc3A010C7d01b50e0d17dc79C8",
		RouteType: 1, BaseAmount: 1000000,
	}
	tx, err := c.GetWithdrawToL1AddressTransaction(req, &types.TransactOpts{Nonce: &nonce, ExpiredAt: 1700000000000})
	if err != nil {
		t.Fatal(err)
	}
	if tx.GetTxType() != txtypes.TxTypeL2WithdrawToL1Address {
		t.Errorf("unexpected txType: %d", tx.GetTxType())
	}
	const expectedHash = "ea4057a2fdf7062c681b4460713b8b0c1dcbe11ef0866a16f4486bfe77d55ef006379c76944f75e4"
	if tx.GetTxHash() != expectedHash {
		t.Fatalf("hash = %s, want %s", tx.GetTxHash(), expectedHash)
	}
	const expectedMessage = "Withdraw to L1 Address\n\nnonce: 0x0000000000000003\nfrom: 0x0000000000000005 (route 0x0000000000000001)\napi key: 0x0000000000000002\nto l1: 0x70997970C51812dc3A010C7d01b50e0d17dc79C8\nasset: 0x0000000000000007\namount: 0x00000000000f4240\nchainId: 0x000000000000012c\nOnly sign this message for a trusted client!"
	if tx.GetOwnerSignatureBody(300) != expectedMessage {
		t.Error("owner message differs from Core")
	}
	hash, err := hex.DecodeString(expectedHash)
	if err != nil {
		t.Fatal(err)
	}
	pub := c.GetKeyManager().PubKeyBytes()
	if err := schnorr.Validate(pub[:], hash, tx.Sig); err != nil {
		t.Fatalf("Core hash rejects signature: %v", err)
	}
	assertSkipNonce(t, "SignWithdrawToL1Address", tx.L2TxAttributes, false)
	tx.OwnerSig = "0x" + strings.Repeat("11", 65)
	hash, err = tx.Hash(300)
	if err != nil || hex.EncodeToString(hash) != expectedHash {
		t.Fatal("OwnerSig must not change the API-key hash")
	}

	// Also match backend TestWithdrawToL1Address, exercising all address limbs
	// and the upper amount limb.
	c, err = NewTxClient(nil, priv, 100, 2, 300)
	if err != nil {
		t.Fatal(err)
	}
	tx, err = c.GetWithdrawToL1AddressTransaction(&types.WithdrawToL1AddressTxReq{
		AssetIndex: 1, L1Recipient: "0x0000000100000002000000030000000400000005",
		RouteType: 1, BaseAmount: 4294967303,
	}, &types.TransactOpts{Nonce: &nonce, ExpiredAt: 1800000000000})
	if err != nil {
		t.Fatal(err)
	}
	if tx.GetTxHash() != "dda2a9e648b3df8cee9449aebe359744fe475e5cd53e94448a111617add31f3658988de6c9b8e2ff" {
		t.Error("hash differs from backend TestWithdrawToL1Address")
	}
	tx.AccountIndex = -1
	if err := tx.Validate(); err != txtypes.ErrInteropAccountIndexTooLow {
		t.Errorf("negative account index accepted: %v", err)
	}
	for _, recipient := range []string{"0x0000000000000000000000000000000000000000", "0x1234", "invalid"} {
		req.L1Recipient = recipient
		if _, err := c.GetWithdrawToL1AddressTransaction(req, &types.TransactOpts{Nonce: &nonce}); err != txtypes.ErrL1RecipientInvalid {
			t.Errorf("invalid recipient %q accepted: %v", recipient, err)
		}
	}
}

func TestCoreInteropAttributes(t *testing.T) {
	priv, _, err := GenerateAPIKey()
	if err != nil {
		t.Fatal(err)
	}
	c := newTestClient(t, priv)
	const recipient = "0x1234567890abcdef1234567890abcdef12345678"
	send := &types.SendToEvmTxReq{AssetIndex: 3, EvmRecipient: recipient, BaseAmount: 1}
	withdraw := &types.WithdrawToL1AddressTxReq{AssetIndex: 3, L1Recipient: recipient, BaseAmount: 1}
	nonce := testNonce
	for skip, expectedErr := range map[uint8]error{0: nil, 1: txtypes.ErrInvalidAttributeType, 2: txtypes.ErrNonceSkipAttributeInvalid} {
		ops := &types.TransactOpts{Nonce: &nonce, TxAttributes: &types.L2TxAttributes{SkipNonce: &skip}}
		if _, err := c.GetSendToEvmTransaction(send, ops); err != expectedErr {
			t.Errorf("SignSendToEvm(skipNonce=%d): error = %v, want %v", skip, err, expectedErr)
		}
		if _, err := c.GetWithdrawToL1AddressTransaction(withdraw, ops); err != expectedErr {
			t.Errorf("SignWithdrawToL1Address(skipNonce=%d): error = %v, want %v", skip, err, expectedErr)
		}
	}
	attrs := txtypes.L2TxAttributes{txtypes.MaxAttributeType + 1: 0}
	if err := attrs.Validate(); !errors.Is(err, txtypes.ErrInvalidAttributeType) {
		t.Errorf("unknown attribute: error = %v, want %v", err, txtypes.ErrInvalidAttributeType)
	}
}
