# Examples

## Go

Run any Go example directly:

```bash
# Account information
go run ./examples/account/get_account.go

# Market data
go run ./examples/market_data/get_orderbook.go

# Order creation (requires private key)
LIGHTER_PRIVATE_KEY=your-key go run ./examples/orders/create_market_order.go
LIGHTER_PRIVATE_KEY=your-key go run ./examples/orders/create_limit_order.go

# WebSocket streaming
go run ./examples/websocket/orderbook_stream.go
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `LIGHTER_API_URL` | `https://mainnet.zklighter.elliot.ai` | HTTP API URL |
| `LIGHTER_WS_URL` | `wss://mainnet.zklighter.elliot.ai/ws` | WebSocket URL |
| `LIGHTER_PRIVATE_KEY` | - | Private key for signing transactions |

## C++

The C++ example demonstrates FFI usage with the shared library.

Compile (select the correct shared library for your platform):
```bash
clang++ -std=c++20 -O3 ./examples/cpp/example.cpp ./build/lighter-signer-darwin-arm64.dylib -o ./build/example-cpp
```

Run from the `./build` folder:
```bash
./example-cpp
```