package main

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"math"
	"strings"
	"testing"

	"github.com/elliottech/lighter-go/client"
	"github.com/elliottech/lighter-go/types"
	curve "github.com/elliottech/poseidon_crypto/curve/ecgfp5"
)

const (
	testBatchAccountIndex = int64(424_242)
	testBatchAPIKeyIndex  = uint8(3)
	defaultAPIKeyIndex    = uint8(255)
)

func registerBatchTestClient(t *testing.T) {
	t.Helper()
	privateKey := "0x" + hex.EncodeToString(curve.ONE.ToLittleEndianBytes())
	if _, err := client.CreateClient(
		nil, privateKey, 304, testBatchAPIKeyIndex, testBatchAccountIndex,
	); err != nil {
		t.Fatalf("create test client: %v", err)
	}
}

func validBatchOrder(index int) types.CreateOrderTxReq {
	return types.CreateOrderTxReq{
		MarketIndex:      0,
		ClientOrderIndex: int64(index + 1),
		BaseAmount:       1_000,
		Price:            50_000,
		Type:             0,
		TimeInForce:      0,
		OrderExpiry:      0,
	}
}

func defaultBatchOptions(firstNonce int64) createOrderBatchOptions {
	return createOrderBatchOptions{
		firstNonce:             firstNonce,
		requestedAPIKeyIndex:   defaultAPIKeyIndex,
		requestedAccountIndex:  testBatchAccountIndex,
		integratorAccountIndex: 0,
	}
}

func TestSignCreateOrdersBatchResolvesDefaultClient(t *testing.T) {
	registerBatchTestClient(t)
	results, err := signCreateOrdersBatch(
		[]types.CreateOrderTxReq{validBatchOrder(0), validBatchOrder(1)},
		defaultBatchOptions(700),
	)
	if err != nil {
		t.Fatalf("sign batch: %v", err)
	}
	for i, result := range results {
		if result.err != "" {
			t.Fatalf("result %d failed: %s", i, result.err)
		}
		var tx struct {
			AccountIndex int64 `json:"AccountIndex"`
			APIKeyIndex  uint8 `json:"ApiKeyIndex"`
			Nonce        int64 `json:"Nonce"`
		}
		if err := json.Unmarshal([]byte(result.txInfo), &tx); err != nil {
			t.Fatalf("decode result %d: %v", i, err)
		}
		if tx.AccountIndex != testBatchAccountIndex || tx.APIKeyIndex != testBatchAPIKeyIndex {
			t.Fatalf("result %d used account/key %d/%d", i, tx.AccountIndex, tx.APIKeyIndex)
		}
		if want := int64(700 + i); tx.Nonce != want {
			t.Fatalf("result %d nonce = %d, want %d", i, tx.Nonce, want)
		}
	}
}

func TestSignCreateOrdersBatchRequiresExplicitNonce(t *testing.T) {
	registerBatchTestClient(t)
	_, err := signCreateOrdersBatch(
		[]types.CreateOrderTxReq{validBatchOrder(0), validBatchOrder(1)},
		defaultBatchOptions(-1),
	)
	if err == nil || !strings.Contains(err.Error(), "explicit non-negative first nonce") {
		t.Fatalf("first nonce -1 error = %v", err)
	}
}

func TestSignCreateOrdersBatchMixedPerItemErrors(t *testing.T) {
	registerBatchTestClient(t)
	orders := []types.CreateOrderTxReq{
		validBatchOrder(0),
		validBatchOrder(1),
		validBatchOrder(2),
	}
	orders[1].BaseAmount = 0
	results, err := signCreateOrdersBatch(orders, defaultBatchOptions(800))
	if err != nil {
		t.Fatalf("sign batch: %v", err)
	}
	if results[0].err != "" || results[2].err != "" {
		t.Fatalf("valid results failed: %q, %q", results[0].err, results[2].err)
	}
	if !strings.Contains(results[1].err, "BaseAmount") {
		t.Fatalf("invalid result error = %q", results[1].err)
	}
}

func TestSignCreateOrdersBatchBounds(t *testing.T) {
	registerBatchTestClient(t)
	if _, err := signCreateOrdersBatch(nil, defaultBatchOptions(0)); err == nil {
		t.Fatal("empty batch succeeded")
	}
	if _, err := signCreateOrdersBatch(
		make([]types.CreateOrderTxReq, maxCreateOrderBatch+1), defaultBatchOptions(0),
	); err == nil {
		t.Fatal("oversized batch succeeded")
	}
	if _, err := signCreateOrdersBatch(
		[]types.CreateOrderTxReq{validBatchOrder(0), validBatchOrder(1)},
		defaultBatchOptions(math.MaxInt64),
	); err == nil || !strings.Contains(err.Error(), "overflows") {
		t.Fatalf("overflowing nonce range error = %v", err)
	}
}

func TestSignCreateOrdersBatchParallel(t *testing.T) {
	registerBatchTestClient(t)
	const count = 2_000
	orders := make([]types.CreateOrderTxReq, count)
	for i := range orders {
		orders[i] = validBatchOrder(i)
	}
	results, err := signCreateOrdersBatch(orders, defaultBatchOptions(1_000))
	if err != nil {
		t.Fatalf("sign parallel batch: %v", err)
	}
	hashes := make(map[string]struct{}, count)
	for i, result := range results {
		if result.err != "" {
			t.Fatalf("result %d failed: %s", i, result.err)
		}
		if _, exists := hashes[result.txHash]; exists {
			t.Fatalf("duplicate signed hash at result %d", i)
		}
		hashes[result.txHash] = struct{}{}
	}
}

func TestMarshalSignedTxResponses(t *testing.T) {
	packed := marshalSignedTxResponses([]packedSignedTxResponse{
		{txType: 14, txInfo: "tx", txHash: "hash"},
		{err: "bad order"},
	})
	if got := binary.LittleEndian.Uint32(packed[:4]); got != 2 {
		t.Fatalf("packed count = %d, want 2", got)
	}
	if len(packed) != 4+2*packedSignedTxHeaderSize+2+4+len("bad order") {
		t.Fatalf("packed length = %d", len(packed))
	}
}

func TestFastEnableStackBoundCacheIsIdempotent(t *testing.T) {
	first := FastEnableStackBoundCache()
	if second := FastEnableStackBoundCache(); second != first {
		t.Fatalf("cache enable result changed from %d to %d", first, second)
	}
}
