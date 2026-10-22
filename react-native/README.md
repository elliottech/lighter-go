# @elliottech/lighter-sdk-react-native

React Native WebView wrapper around the Lighter Go WASM SDK. The Go SDK is
compiled to WASM, embedded as base64 into a self-contained HTML page (see
`../web-wasm`), and bundled with this package. `LighterSdkWebView` loads that page
in a hidden WebView and exposes the Go functions it registers via a small
promise-based bridge.

## Install

```sh
npm install @elliottech/lighter-sdk-react-native expo-asset react-native-webview
```

`react`, `react-native`, `react-native-webview`, and `expo-asset` are peer
dependencies - install whichever versions match your app.

### Metro config

The bundled SDK page is a `.html` file loaded via `require()`, so Metro needs
to know how to treat it as an asset. Add `html` to `resolver.assetExts` in
your `metro.config.js`:

```js
const { getDefaultConfig } = require('expo/metro-config'); // or `@react-native/metro-config`

const config = getDefaultConfig(__dirname);
config.resolver.assetExts.push('html');

module.exports = config;
```

## Usage

Mount `LighterSdkWebView` once near the root of your app (it renders nothing
visible):

```tsx
import { LighterSdkWebView, goWasmExecute } from '@elliottech/lighter-sdk-react-native';

export function App() {
  return (
    <>
      <LighterSdkWebView />
      {/* rest of your app */}
    </>
  );
}
```

Then call Go functions exposed by the SDK from anywhere:

```ts
import { goWasmExecute } from '@elliottech/lighter-sdk-react-native';

const result = await goWasmExecute({
  function: 'SomeGoFunction',
  params: [arg1, arg2],
});
```

`goWasmExecute` resolves once the WebView has finished loading the SDK, so
it's safe to call it before `LighterSdkWebView` has mounted - the call just waits.

### `LighterSDK`

`LighterSDK` is a typed convenience wrapper around `goWasmExecute` - one
method per Go SDK function, with typed arguments/return values instead of
raw `function`/`params` calls:

```ts
import { LighterSDK } from '@elliottech/lighter-sdk-react-native';

const { pk, body } = await LighterSDK.createClient(seed, chainId, accountIndex, nonce, apiKeyIndex);

const { txInfo } = await LighterSDK.signCreateOrder({
  accountIndex,
  marketIndex,
  clientOrderIndex,
  baseAmount,
  price,
  isAsk,
  orderType,
  timeInForce,
  reduceOnly,
  triggerPrice,
  orderExpiry,
  nonce,
});
```

`createClient` / `createClientByPrv` also remember the `chainId` and
`apiKeyIndex` the account was created with; every other method sends that
identity along so the page selects the matching signer before signing. Use
`setSignerIdentity(accountIndex, chainId, apiKeyIndex)` if you create clients
through raw `goWasmExecute` calls. The page never returns the private key
(`prv` is stripped before it reaches native code).

Available methods: `createClient`, `createClientByPrv`, `resetClients`, `createAuthToken`, `signChangePubKey`,
`signUpdateLeverage`, `signStakeAssets`, `signUnstakeAssets`,
`signMintShares`, `signBurnShares`, `signWithdraw`, `signTransfer`,
`getTransferTransaction`, `signUpdateMargin`,
`signUpdateAccountAssetConfig`, `signModifyOrder`, `signCancelOrder`,
`signCancelAllOrders`, `signCreateOrder`, `signCreateGroupedOrders`,
`signUpdateAccountConfig`, `signCreateSubAccount`. Each method's argument
and response types are also exported (e.g. `TxResponse`,
`LighterSdkCreateOrderArgs`, `LighterSdkTransferArgs`).

### `LighterSdkWebView` props

| Prop        | Type                                            | Description                                                              |
| ----------- | ------------------------------------------------ | ------------------------------------------------------------------------ |
| `htmlAsset` | `number`                                          | Result of `require()`-ing an alternate standalone SDK html file.         |
| `style`     | `StyleProp<ViewStyle>`                            | Merged into the (hidden, 0x0) wrapper view's style.                      |
| `onInfo`    | `(message: string) => void`                       | Called when the WebView finishes loading.                                |
| `onError`   | `(message: string) => void`                       | Called on load failures, runtime failures and rejected requests. The page never forwards error details or console output. |

## Building

`dist/wasm-wrapper.standalone.html` (the file this package requires at
runtime) is a copy of `../web-wasm/sdk.html`, which `../web-wasm/build.py`
generates from `../web-wasm/wasm-template.html` (Go runtime, WASM as base64,
export allow-list and CSP script hashes embedded). It is never hand-written:

```sh
npm run build:js     # bundles src/ into dist/ (tsdown)
npm run build:wasm   # runs ../web-wasm/build.py and copies sdk.html + manifest into dist/
npm run build        # runs both, in that order
```

The order matters: `build:js` cleans and (re)creates `dist/` before
`build:wasm` writes the generated html into it. `npm run build` (used by
`prepublishOnly`) always runs them in the right order. `build.py` needs
`go` (it pins `GOTOOLCHAIN=go1.26.8`) and `brotli` on `PATH`.
