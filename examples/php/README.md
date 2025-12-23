# PHP Wrapper for Lighter Signer

This PHP wrapper provides an interface to the Lighter signing library, enabling transaction signing for the Lighter decentralized exchange.

## Requirements

- PHP 7.4 or higher
- PHP FFI extension enabled
- The appropriate shared library for your platform

## Installation

1. Download the shared library for your platform from the [lighter-go releases](https://github.com/elliottech/lighter-go/releases). The library should be placed in the `build/` directory relative to the repository root, or you can specify a custom path.

Supported platforms:
- `lighter-signer-darwin-arm64.dylib` - macOS ARM64 (Apple Silicon)
- `lighter-signer-linux-amd64.so` - Linux x86_64
- `lighter-signer-linux-arm64.so` - Linux ARM64
- `lighter-signer-windows-amd64.dll` - Windows x86_64

2. Enable the FFI extension in your `php.ini`:
```ini
extension=ffi
ffi.enable=true
```

3. Include the `LighterSigner.php` file in your project:
```php
require_once 'path/to/LighterSigner.php';
```

## Usage

### Basic Setup

```php
<?php
require_once 'LighterSigner.php';

// Get the singleton instance (auto-detects platform)
$signer = LighterSigner::getInstance();

// Or specify a custom library path
$signer = LighterSigner::getInstance('/path/to/lighter-signer.so');
```

### Generate API Key

```php
$keyPair = $signer->generateAPIKey();
echo "Private Key: " . $keyPair['privateKey'] . "\n";
echo "Public Key: " . $keyPair['publicKey'] . "\n";
```

### Create Client

Before signing transactions, you must create a client:

```php
$signer->createClient(
    'https://mainnet.zklighter.elliot.ai',  // API URL
    $privateKey,                             // Your private key
    270,                                     // Chain ID
    0,                                       // API key index (0-254)
    12345                                    // Account index
);
```

### Sign Create Order

```php
$result = $signer->signCreateOrder(
    0,          // marketIndex (0 = ETH)
    1001,       // clientOrderIndex
    1000000,    // baseAmount
    350000,     // price
    1,          // isAsk (1 = sell, 0 = buy)
    0,          // orderType (0 = limit)
    0,          // timeInForce
    0,          // reduceOnly
    0,          // triggerPrice
    -1,         // orderExpiry (-1 = default 28 days)
    42,         // nonce
    0,          // apiKeyIndex
    12345       // accountIndex
);

echo "Transaction Type: " . $result['txType'] . "\n";
echo "Transaction Info: " . $result['txInfo'] . "\n";
echo "Transaction Hash: " . $result['txHash'] . "\n";
```

### Sign Cancel Order

```php
$result = $signer->signCancelOrder(
    0,          // marketIndex
    5001,       // orderIndex
    43,         // nonce
    0,          // apiKeyIndex
    12345       // accountIndex
);
```

### Sign Withdraw

```php
$result = $signer->signWithdraw(
    3,          // assetIndex (3 = USDC)
    0,          // routeType (0 = Perps, 1 = Spot)
    1000000,    // amount (1 USDC = 1000000)
    44,         // nonce
    0,          // apiKeyIndex
    12345       // accountIndex
);
```

### Sign Update Leverage

```php
$result = $signer->signUpdateLeverage(
    0,          // marketIndex
    500,        // initialMarginFraction (for 20x leverage)
    0,          // marginMode (0 = Cross, 1 = Isolated)
    45,         // nonce
    0,          // apiKeyIndex
    12345       // accountIndex
);
```

### Create Auth Token

```php
$token = $signer->createAuthToken(
    0,          // deadline (0 = default 7 hours)
    0,          // apiKeyIndex
    12345       // accountIndex
);
echo "Auth Token: " . $token . "\n";
```

### Sign Grouped Orders (OCO/OTO)

```php
$orders = [
    [
        'marketIndex' => 0,
        'clientOrderIndex' => 1001,
        'baseAmount' => 1000000,
        'price' => 350000,
        'isAsk' => 1,
        'type' => 0,
        'timeInForce' => 0,
        'reduceOnly' => 0,
        'triggerPrice' => 0,
        'orderExpiry' => -1,
    ],
    [
        'marketIndex' => 0,
        'clientOrderIndex' => 1002,
        'baseAmount' => 1000000,
        'price' => 340000,
        'isAsk' => 1,
        'type' => 0,
        'timeInForce' => 0,
        'reduceOnly' => 0,
        'triggerPrice' => 0,
        'orderExpiry' => -1,
    ],
];

$result = $signer->signCreateGroupedOrders(
    1,          // groupingType (1 = OCO)
    $orders,
    46,         // nonce
    0,          // apiKeyIndex
    12345       // accountIndex
);
```

## Available Methods

| Method | Description |
|--------|-------------|
| `generateAPIKey()` | Generate a new API key pair |
| `createClient()` | Create a signing client |
| `checkClient()` | Verify client exists and is valid |
| `signChangePubKey()` | Sign a change public key transaction |
| `signCreateOrder()` | Sign a create order transaction |
| `signCreateGroupedOrders()` | Sign grouped orders (OCO/OTO/OTOCO) |
| `signCancelOrder()` | Sign a cancel order transaction |
| `signWithdraw()` | Sign a withdraw transaction |
| `signCreateSubAccount()` | Sign a create sub-account transaction |
| `signCancelAllOrders()` | Sign a cancel all orders transaction |
| `signModifyOrder()` | Sign a modify order transaction |
| `signTransfer()` | Sign a transfer transaction |
| `signCreatePublicPool()` | Sign a create public pool transaction |
| `signUpdatePublicPool()` | Sign an update public pool transaction |
| `signMintShares()` | Sign a mint shares transaction |
| `signBurnShares()` | Sign a burn shares transaction |
| `signUpdateLeverage()` | Sign an update leverage transaction |
| `signUpdateMargin()` | Sign an update margin transaction |
| `createAuthToken()` | Create an authentication token |

## Error Handling

All methods throw `RuntimeException` on failure:

```php
try {
    $result = $signer->signCreateOrder(...);
} catch (RuntimeException $e) {
    echo "Error: " . $e->getMessage() . "\n";
}
```

## Constants Reference

### Order Types
- `0` - Limit
- `1` - Market
- `2` - Stop Loss
- `3` - Take Profit

### Time In Force
- `0` - Good Till Time
- `1` - Immediate Or Cancel
- `2` - Post Only

### Route Types
- `0` - Perps
- `1` - Spot

### Margin Modes
- `0` - Cross Margin
- `1` - Isolated Margin

### Common Asset Indices
- `1` - Native (ETH)
- `3` - USDC
