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

### `LighterSdkWebView` props

| Prop        | Type                                            | Description                                                              |
| ----------- | ------------------------------------------------ | ------------------------------------------------------------------------ |
| `htmlAsset` | `number`                                          | Result of `require()`-ing an alternate standalone SDK html file.         |
| `style`     | `StyleProp<ViewStyle>`                            | Merged into the (hidden, 0x0) wrapper view's style.                      |
| `onLog`     | `(message: string \| string[]) => void`           | Called on `console.log` messages forwarded from the page.                |
| `onWarn`    | `(message: string \| string[]) => void`           | Called on `console.warn` messages forwarded from the page.               |
| `onError`   | `(message: string \| string[]) => void`           | Called on uncaught errors reported by the page.                          |

## Building

`dist/wasm-wrapper.standalone.html` (the file this package requires at
runtime) is generated from `src/wasm-template.html`, not hand-written:

```sh
npm run build:js     # bundles src/ into dist/ (tsdown)
npm run build:wasm   # builds ../web-wasm/main.wasm and regenerates dist/wasm-wrapper.standalone.html
npm run build        # runs both, in that order
```

The order matters: `build:js` cleans and (re)creates `dist/` before
`build:wasm` writes the generated html into it. `npm run build` (used by
`prepublishOnly`) always runs them in the right order.
