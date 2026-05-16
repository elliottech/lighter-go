# Lighter Signing Server - C# Integration

The HTTP signing server provides process isolation between Go's signing logic and .NET's garbage collector, solving GC compatibility issues that arise when using Go's `c-shared` build mode as a native library inside .NET.

## Why use the HTTP server instead of the shared library?

Go's `c-shared` libraries embed their own runtime and garbage collector. When loaded into a .NET process, the two GC runtimes can conflict, causing crashes and unpredictable behavior. The HTTP server runs as a separate process, keeping both runtimes fully isolated.

## Running the server

### From source
```bash
go run ./cmd/server
```

### From binary
```bash
# Build
just build-server

# Run
./build/lighter-server
```

### Environment variables / flags
| Variable | Flag | Default | Description |
|---|---|---|---|
| `LIGHTER_HOST` | `--host` | `0.0.0.0` | Host to listen on |
| `LIGHTER_PORT` | `--port` | `8080` | Port to listen on |

## Health check

```
GET /health
→ {"status": "ok"}
```

## C# example using HttpClient

```csharp
using System.Net.Http;
using System.Text;
using System.Text.Json;

var http = new HttpClient { BaseAddress = new Uri("http://localhost:8080") };

// 1. Create a client (call once on startup)
var createClientReq = new
{
    url = "https://mainnet.lighter.xyz",
    privateKey = "0x...",
    chainId = 42,
    apiKeyIndex = 255,
    accountIndex = 12345L
};
var resp = await http.PostAsync("/create-client",
    new StringContent(JsonSerializer.Serialize(createClientReq), Encoding.UTF8, "application/json"));
Console.WriteLine(await resp.Content.ReadAsStringAsync());

// 2. Sign a create-order transaction
var createOrderReq = new
{
    marketIndex = 0,
    clientOrderIndex = 1L,
    baseAmount = 1000000L,
    price = 50000u,
    isAsk = 0,
    orderType = 0,
    timeInForce = 0,
    reduceOnly = 0,
    triggerPrice = 0u,
    orderExpiry = -1L,        // defaults to 28 days
    integratorAccountIndex = 0L,
    integratorTakerFee = 0u,
    integratorMakerFee = 0u,
    skipNonce = 0,
    nonce = -1L,              // auto-fetch from server
    apiKeyIndex = 255,
    accountIndex = 12345L
};
resp = await http.PostAsync("/sign-create-order",
    new StringContent(JsonSerializer.Serialize(createOrderReq), Encoding.UTF8, "application/json"));
var result = await resp.Content.ReadAsStringAsync();
Console.WriteLine(result);

// The response JSON contains:
// {
//   "txType": 3,
//   "txInfo": "<hex-encoded transaction>",
//   "txHash": "<hex-encoded hash>",
//   "error": ""             // empty on success
// }
```

## JSON request/response formats

### POST /create-client
**Request:**
```json
{
  "url": "https://mainnet.lighter.xyz",
  "privateKey": "0x...",
  "chainId": 42,
  "apiKeyIndex": 255,
  "accountIndex": 12345
}
```
**Response (success):**
```json
{ "error": "" }
```

### POST /check-client
**Request:**
```json
{
  "apiKeyIndex": 255,
  "accountIndex": 12345
}
```
**Response (success):**
```json
{ "error": "" }
```

### POST /generate-api-key
**Request:** empty body  
**Response:**
```json
{
  "privateKey": "0x...",
  "publicKey": "0x...",
  "error": ""
}
```

### POST /create-auth-token
**Request:**
```json
{
  "deadline": 0,
  "apiKeyIndex": 255,
  "accountIndex": 12345
}
```
**Response:**
```json
{
  "result": "<auth-token-string>",
  "error": ""
}
```

### Signing endpoints

All signing endpoints return the same response format:
```json
{
  "txType": 3,
  "txInfo": "<hex-encoded transaction data>",
  "txHash": "<hex-encoded transaction hash>",
  "messageToSign": "",
  "error": ""
}
```

**Common fields** shared by most signing requests:
```json
{
  "skipNonce": 0,
  "nonce": -1,
  "apiKeyIndex": 255,
  "accountIndex": 12345
}
```

**Integrator fields** (for order-related endpoints):
```json
{
  "integratorAccountIndex": 0,
  "integratorTakerFee": 0,
  "integratorMakerFee": 0
}
```

### Available signing endpoints

| Endpoint | Additional fields |
|---|---|
| `POST /sign-change-pub-key` | `pubKey` |
| `POST /sign-create-order` | `marketIndex`, `clientOrderIndex`, `baseAmount`, `price`, `isAsk`, `orderType`, `timeInForce`, `reduceOnly`, `triggerPrice`, `orderExpiry` + integrator fields |
| `POST /sign-create-grouped-orders` | `groupingType`, `orders[]` (array of order objects) + integrator fields |
| `POST /sign-cancel-order` | `marketIndex`, `orderIndex` |
| `POST /sign-cancel-all-orders` | `timeInForce`, `time` |
| `POST /sign-modify-order` | `marketIndex`, `index`, `baseAmount`, `price`, `triggerPrice` + integrator fields |
| `POST /sign-withdraw` | `assetIndex`, `routeType`, `amount` |
| `POST /sign-transfer` | `toAccountIndex`, `assetIndex`, `fromRouteType`, `toRouteType`, `amount`, `usdcFee`, `memo` |
| `POST /sign-create-sub-account` | *(common fields only)* |
| `POST /sign-create-public-pool` | `operatorFee`, `initialTotalShares`, `minOperatorShareRate` |
| `POST /sign-update-public-pool` | `publicPoolIndex`, `status`, `operatorFee`, `minOperatorShareRate` |
| `POST /sign-mint-shares` | `publicPoolIndex`, `shareAmount` |
| `POST /sign-burn-shares` | `publicPoolIndex`, `shareAmount` |
| `POST /sign-update-leverage` | `marketIndex`, `initialMarginFraction`, `marginMode` |
| `POST /sign-update-margin` | `marketIndex`, `usdcAmount`, `direction` |
| `POST /sign-stake-assets` | `stakingPoolIndex`, `shareAmount` |
| `POST /sign-unstake-assets` | `stakingPoolIndex`, `shareAmount` |
| `POST /sign-approve-integrator` | `integratorIndex`, `maxPerpsTakerFee`, `maxPerpsMakerFee`, `maxSpotTakerFee`, `maxSpotMakerFee`, `approvalExpiry` |
| `POST /sign-update-account-config` | `accountTradingMode` |
| `POST /sign-update-account-asset-config` | `assetIndex`, `assetMarginMode` |

## Deployment

### Standalone binary
```bash
# Build for your platform
just build-server

# Or cross-compile for Linux
just build-server-linux-amd64-docker
just build-server-linux-arm64-docker

# Run alongside your C# application
./lighter-server --port 8080 &
dotnet run
```

### Docker
```dockerfile
FROM golang:1.23.2-bullseye AS builder
WORKDIR /app
COPY . .
RUN go mod vendor && CGO_ENABLED=1 go build -trimpath -o lighter-server ./cmd/server

FROM debian:bullseye-slim
COPY --from=builder /app/lighter-server /usr/local/bin/
EXPOSE 8080
CMD ["lighter-server"]
```

### systemd
```ini
[Unit]
Description=Lighter Signing Server
After=network.target

[Service]
ExecStart=/usr/local/bin/lighter-server --host 127.0.0.1 --port 8080
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```
