# Binary options operator transactions

Operators assigned to a binary options market slot can create, update and settle the order book on that slot with
`BinaryOptionsOperatorCreateMarket`, `BinaryOptionsOperatorUpdateMarket` and `BinaryOptionsOperatorSettleMarket`.
The requests live in `types/tx_request.go`, the constants in `types/txtypes/constants.go`, and the C ABI structs in
`sharedlib/main.go` (`BinaryOptionsOperator*TxReq`).

## Slots vs public market index

- `MarketSlot` (`int16`) is the internal slot the order book is created on. Binary options slots are
  `1000..2000` (`MinBinaryOptionsMarketSlot`..`MaxBinaryOptionsMarketSlot`). The slot is serialized as `MarketIndex`
  in the signed tx info, which is the name the server expects.
- `MarketIndex` (`int64`) is the market index users trade with (`CreateOrder` etc). It is assigned when the
  market is created and is required by update and settle.

## Fixed values

The SDK fills in the fields that have only one accepted value for binary options:

| Field | Value |
| --- | --- |
| `MarketType` | `MarketTypeBinaryOptions` (2) |
| `BaseAssetId`, `QuoteAssetId` | `USDCAssetIndex` (3) |
| `QuoteExtensionMultiplier` | `10^(12 - SizeDecimals - PriceDecimals)` |
| `SizeExtensionMultiplier` | `SettlementCap * QuoteExtensionMultiplier` |
| `LiquidationFee` (update) | 0 |

The `12` is USDC decimals (6) plus the collateral multiplier decimals (6), so `SizeDecimals + PriceDecimals <= 6`.

## CreateMarket

| Field | Type | Accepted values |
| --- | --- | --- |
| `MarketSlot` | int16 | `1000..2000` |
| `StartTimestampMs` | int64 | unix ms, `>= 0`, `< EndTimestampMs` |
| `EndTimestampMs` | int64 | unix ms, `<= MaxTimestamp` |
| `SettlementCap` | uint32 | `> 0`, max settlement value of one share in price ticks |
| `SettlementType` | uint8 | `BinaryOptionsSettlementTypeDiscrete` (0): settles at 0 or `SettlementCap`; `BinaryOptionsSettlementTypeContinuous` (1): settles anywhere in `[0, SettlementCap]` |
| `DefaultPrice` | uint32 | `<= SettlementCap`, price used when the market is refunded; immutable |
| `TakerFee`, `MakerFee` | uint32 | `<= FeeTick` (10^6 = 100%) |
| `MinBaseAmount` | int64 | `1..MaxOrderBaseAmount`, in size ticks |
| `MinQuoteAmount` | int64 | `1..MaxOrderQuoteAmount`, in USDC ticks |
| `OrderQuoteLimit` | int64 | `MinQuoteAmount..MaxOrderQuoteAmount` |
| `OpenInterestLimit` | int64 | `OrderQuoteLimit..MaxOpenInterest` |
| `IsFrozen` | uint8 | `0` or `1`; a frozen book accepts no orders |
| `Symbol` | string | valid UTF-8, `< 32` bytes |
| `SizeDecimals`, `PriceDecimals` | uint8 | `SizeDecimals + PriceDecimals <= 6` |

### Example: 1 USDC payout per share

```go
&types.BinaryOptionsOperatorCreateMarketTxReq{
	MarketSlot:        1001,
	StartTimestampMs:  time.Now().UnixMilli(),
	EndTimestampMs:    time.Now().Add(24 * time.Hour).UnixMilli(),
	SettlementCap:     100, // 100 price ticks of 0.01 = 1 USDC
	SettlementType:    txtypes.BinaryOptionsSettlementTypeDiscrete,
	DefaultPrice:      50, // 0.50 USDC on refund
	TakerFee:          1000, // 0.1%
	MakerFee:          0,
	MinBaseAmount:     10_000, // 1 share
	MinQuoteAmount:    1_000_000, // 1 USDC
	OrderQuoteLimit:   1_000_000_000,
	OpenInterestLimit: 10_000_000_000,
	Symbol:            "BTC-100K-YES",
	SizeDecimals:      4,
	PriceDecimals:     2,
}
```

`QuoteExtensionMultiplier = 10^(12-4-2) = 1_000_000`, `SizeExtensionMultiplier = 100 * 1_000_000 = 100_000_000`.

## UpdateMarket

| Field | Type | Accepted values |
| --- | --- | --- |
| `MarketSlot` | int16 | `1000..2000` |
| `MarketIndex` | int64 | the market's public index |
| `StartTimestampMs`, `EndTimestampMs` | int64 | unix ms, same rules as create |
| `TakerFee`, `MakerFee` | uint32 | `<= FeeTick` |
| `MinBaseAmount`, `MinQuoteAmount`, `OrderQuoteLimit`, `OpenInterestLimit` | int64 | same rules as create |
| `IsFrozen` | uint8 | `0` or `1` |

## SettleMarket

| Field | Type | Accepted values |
| --- | --- | --- |
| `MarketSlot` | int16 | `1000..2000` |
| `MarketIndex` | int64 | the market's public index |
| `SettlementPrice` | uint32 | `0..SettlementCap`; must be `0` when `IsRefund` is set |
| `IsRefund` | uint8 | `0` or `1`; `1` settles every position at `DefaultPrice` |

## C ABI

The shared library exports take a pointer to the request struct plus the usual nonce/account arguments:

```c
SignedTxResponse SignBinaryOptionsOperatorCreateMarket(BinaryOptionsOperatorCreateMarketTxReq* req, uint8_t skipNonce, long long nonce, int apiKeyIndex, long long accountIndex);
SignedTxResponse SignBinaryOptionsOperatorUpdateMarket(BinaryOptionsOperatorUpdateMarketTxReq* req, uint8_t skipNonce, long long nonce, int apiKeyIndex, long long accountIndex);
SignedTxResponse SignBinaryOptionsOperatorSettleMarket(BinaryOptionsOperatorSettleMarketTxReq* req, uint8_t skipNonce, long long nonce, int apiKeyIndex, long long accountIndex);
```

## WASM

`wasm/main.go` exposes the same three functions. The first argument is a plain JS object with the same field names as the C structs above (`MarketSlot`, `StartTimestampMs`, `EndTimestampMs`, `Symbol`, ...), followed by `skipNonce, nonce, apiKeyIndex, accountIndex`:

```js
SignBinaryOptionsOperatorCreateMarket(req, skipNonce, nonce, apiKeyIndex, accountIndex)
SignBinaryOptionsOperatorUpdateMarket(req, skipNonce, nonce, apiKeyIndex, accountIndex)
SignBinaryOptionsOperatorSettleMarket(req, skipNonce, nonce, apiKeyIndex, accountIndex)
```

Each returns `{ txType, txInfo, txHash }` or `{ error }`.

`web-wasm/` exposes the same as `_signBinaryOptionsOperatorCreateMarket/UpdateMarket/SettleMarket(accountIndex, req, nonce)`, returning a promise like the other `_sign*` functions.
