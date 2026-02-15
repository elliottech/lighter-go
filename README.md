# lighter-go

The official Go SDK for the Lighter trading platform. This SDK provides full feature parity with the [Python SDK](https://github.com/elliottech/lighter-python), including:

- **Transaction signing** - Sign all Lighter L2 transactions
- **Full HTTP API client** - Access all REST endpoints
- **WebSocket client** - Real-time order book and account streaming
- **Convenience methods** - High-level trading functions
- **Nonce management** - Optimistic and API-based nonce strategies

## Installation

```bash
go get github.com/elliottech/lighter-go
```

## Quick Start

### HTTP API

```go
package main

import (
    "fmt"
    "github.com/elliottech/lighter-go/client/http"
    "github.com/elliottech/lighter-go/types/api"
)

func main() {
    // Create HTTP client
    client := http.NewFullClient("https://mainnet.zklighter.elliot.ai")

    // Get order book
    marketIndex := int16(0) // ETH-USD
    orderBooks, _ := client.Order().GetOrderBooks(&marketIndex, api.MarketFilterAll)

    for _, ob := range orderBooks.OrderBooks {
        fmt.Printf("Best Bid: %s @ %s\n", ob.Bids[0].Size, ob.Bids[0].Price)
        fmt.Printf("Best Ask: %s @ %s\n", ob.Asks[0].Size, ob.Asks[0].Price)
    }
}
```

### Creating Orders

```go
package main

import (
    "github.com/elliottech/lighter-go/client"
    "github.com/elliottech/lighter-go/client/http"
    "github.com/elliottech/lighter-go/types"
)

func main() {
    httpClient := http.NewFullClient("https://mainnet.zklighter.elliot.ai")

    // Create signer client with your private key
    signerClient, _ := client.NewSignerClient(
        httpClient,
        "your-private-key",
        1,    // chainId
        0,    // apiKeyIndex
        0,    // accountIndex
        nil,  // nonceManager (nil = use optimistic)
    )

    // Create a market order
    txInfo, _ := signerClient.CreateMarketOrder(
        0,           // marketIndex (ETH-USD)
        1000000,     // size (0.01 ETH scaled)
        true,        // isBuy
        &types.TransactOpts{Nonce: types.NewInt64(-1)},
    )

    // Submit to API
    resp, _ := signerClient.SendAndSubmit(txInfo)
    fmt.Printf("Order submitted: %s\n", resp.TxHash)
}
```

### WebSocket Streaming

```go
package main

import (
    "context"
    "fmt"
    "github.com/elliottech/lighter-go/client/ws"
)

func main() {
    client := ws.NewClient("wss://mainnet.zklighter.elliot.ai/ws", ws.DefaultOptions())

    ctx := context.Background()
    client.Connect(ctx)
    defer client.Close()

    // Subscribe to order book
    client.SubscribeOrderBook(0) // ETH-USD

    // Process updates
    for update := range client.OrderBookUpdates() {
        if update.State != nil {
            bestBid := update.State.GetBestBid()
            bestAsk := update.State.GetBestAsk()
            fmt.Printf("Bid: %s @ %s | Ask: %s @ %s\n",
                bestBid.Size, bestBid.Price,
                bestAsk.Size, bestAsk.Price)
        }
    }
}
```

## API Reference

### HTTP Client

The `FullHTTPClient` provides access to all API endpoints organized by domain:

| API Group | Methods |
|-----------|---------|
| `Account()` | GetAccount, GetAccountLimits, GetLiquidations, GetPnL, etc. |
| `Order()` | GetActiveOrders, GetOrderBooks, GetRecentTrades, GetExchangeStats, etc. |
| `Transaction()` | SendTx, SendTxBatch, GetTx, GetDepositHistory, GetWithdrawHistory, etc. |
| `Candlestick()` | GetCandlesticks, GetFundings, GetFundingRates |
| `Block()` | GetBlock, GetBlocks, GetCurrentHeight |
| `Bridge()` | GetBridges, GetIsNextBridgeFast, GetFastBridgeInfo |
| `Info()` | GetStatus, GetInfo, GetAnnouncements |

### SignerClient Convenience Methods

| Method | Description |
|--------|-------------|
| `CreateMarketOrder()` | Create a market order |
| `CreateMarketOrderWithSlippage()` | Market order with slippage protection |
| `CreateLimitOrder()` | Create a limit order |
| `CreateTakeProfitOrder()` | Create a take-profit order |
| `CreateStopLossOrder()` | Create a stop-loss order |
| `CancelAllOrders()` | Cancel all open orders |
| `SendAndSubmit()` | Sign and submit a transaction |
| `SendTxBatch()` | Submit multiple transactions |

### WebSocket Client

| Method | Description |
|--------|-------------|
| `Connect()` | Establish WebSocket connection |
| `SubscribeOrderBook()` | Subscribe to order book updates |
| `SubscribeAccount()` | Subscribe to account updates (requires auth) |
| `OrderBookUpdates()` | Channel for order book updates |
| `AccountUpdates()` | Channel for account updates |
| `GetOrderBookState()` | Get current order book state |

### Nonce Management

Two nonce management strategies are available:

- **OptimisticManager** (default): Assumes transactions succeed, increments locally. Fast but requires failure acknowledgment.
- **APIManager**: Queries API for every nonce. Slower but always accurate.

```go
import "github.com/elliottech/lighter-go/nonce"

// Use optimistic (default)
manager := nonce.NewOptimisticManager(httpClient)

// Use API-based
manager := nonce.NewAPIManager(httpClient)

// Create signer with custom nonce manager
signerClient, _ := client.NewSignerClient(httpClient, privateKey, chainId, 0, 0, manager)
```

## Transactions

All L2 transaction types are supported:

```
=== Client ===
CreateClient
CheckClient

=== API Key ===
CreateAuthToken
SignChangePubKey
GenerateAPIKey

=== Order ===
SignCreateOrder
SignCreateGroupedOrders
SignCancelOrder
SignCancelAllOrders
SignModifyOrder

=== Leverage & Margin ===
SignUpdateLeverage
SignUpdateMargin

=== Transfers ===
SignWithdraw
SignTransfer

=== Sub account & pools ===
SignCreateSubAccount
SignCreatePublicPool
SignUpdatePublicPool
SignMintShares
SignBurnShares

=== Staking ===
SignStakeAssets
SignUnstakeAssets
```

## Shared Libraries

Pre-compiled shared libraries are available for FFI usage:
- macOS (darwin) dynamic library (.dylib) for arm architecture
- Linux shared object (.so) for amd64 and arm architectures
- Windows DLL for amd64 architecture

All libraries follow the naming convention `lighter-{os}-{arch}`.

The build & accompanying `.h` files can be found in the [releases](https://github.com/elliottech/lighter-go/releases).

To compile your own binaries, see the commands in the `justfile`.

## Examples

See the [examples](./examples) directory for complete working examples:

- `examples/orders/` - Order creation and management
- `examples/account/` - Account information
- `examples/market_data/` - Market data fetching
- `examples/websocket/` - Real-time streaming

## Auth Tokens

Auth tokens are used to call HTTP & WS endpoints that require authentication (e.g., open orders).

```go
// Auth tokens are valid for up to 8 hours
deadline := time.Now().Add(8 * time.Hour)
authToken, _ := txClient.GetAuthToken(deadline)
```

**Note:** Auth tokens are bound to an API key. Changing the API key will invalidate all generated auth tokens.

## License

See [LICENSE](./LICENSE) for details.
