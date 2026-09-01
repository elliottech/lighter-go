package types

import (
	"hash"
	"testing"
	"time"

	"github.com/elliottech/lighter-go/signer"
	curve "github.com/elliottech/poseidon_crypto/curve/ecgfp5"
)

type legacyTestSigner struct {
	called bool
	hash   hash.Hash
}

func (signer *legacyTestSigner) Sign(_ []byte, h hash.Hash) ([]byte, error) {
	signer.called = true
	signer.hash = h
	return []byte("legacy"), nil
}

type directTestSigner struct {
	legacyCalled bool
	directCalled bool
}

func (signer *directTestSigner) Sign(_ []byte, _ hash.Hash) ([]byte, error) {
	signer.legacyCalled = true
	return []byte("legacy"), nil
}

func (signer *directTestSigner) SignHash(_ []byte) ([]byte, error) {
	signer.directCalled = true
	return []byte("direct"), nil
}

func TestSignHashedMessageUsesDirectSigner(t *testing.T) {
	signer := &directTestSigner{}
	signature, err := signHashedMessage(signer, []byte("message"))
	if err != nil {
		t.Fatalf("sign hashed message: %v", err)
	}
	if string(signature) != "direct" || !signer.directCalled || signer.legacyCalled {
		t.Fatalf("direct signer was not selected: %+v", signer)
	}
}

func TestSignHashedMessageSupportsLegacySigner(t *testing.T) {
	signer := &legacyTestSigner{}
	signature, err := signHashedMessage(signer, []byte("message"))
	if err != nil {
		t.Fatalf("sign hashed message: %v", err)
	}
	if string(signature) != "legacy" || !signer.called || signer.hash == nil {
		t.Fatalf("legacy signer fallback was not used: %+v", signer)
	}
}

func benchmarkCreateOrder(b *testing.B) (signer.KeyManager, *CreateOrderTxReq, *TransactOpts) {
	b.Helper()
	key, err := signer.NewKeyManager(curve.ONE.ToLittleEndianBytes())
	if err != nil {
		b.Fatalf("new key manager: %v", err)
	}
	accountIndex := int64(1)
	apiKeyIndex := uint8(0)
	nonce := int64(1)
	return key, &CreateOrderTxReq{
			MarketIndex:      0,
			ClientOrderIndex: 1,
			BaseAmount:       10_000,
			Price:            1_000,
			Type:             0,
			TimeInForce:      1,
			OrderExpiry:      time.Now().Add(24 * time.Hour).UnixMilli(),
		}, &TransactOpts{
			FromAccountIndex: &accountIndex,
			ApiKeyIndex:      &apiKeyIndex,
			ExpiredAt:        time.Now().Add(10 * time.Minute).UnixMilli(),
			Nonce:            &nonce,
		}
}

func BenchmarkCreateOrderStages(b *testing.B) {
	key, order, opts := benchmarkCreateOrder(b)
	converted := ConvertCreateOrderTx(order, opts)
	messageHash, err := converted.Hash(304)
	if err != nil {
		b.Fatalf("hash order: %v", err)
	}
	signed, err := ConstructCreateOrderTx(key, 304, order, opts)
	if err != nil {
		b.Fatalf("construct order: %v", err)
	}

	b.Run("hash", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			if _, err := converted.Hash(304); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("sign", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			if _, err := signHashedMessage(key, messageHash); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("serialize", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			if _, err := signed.GetTxInfo(); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("complete", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			if _, err := ConstructCreateOrderTx(key, 304, order, opts); err != nil {
				b.Fatal(err)
			}
		}
	})
}
