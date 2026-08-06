import { useAssets } from 'expo-asset';
import { useCallback, useMemo, useRef, useState } from 'react';
import type { StyleProp, ViewStyle } from 'react-native';
import { Platform, View } from 'react-native';
import WebView from 'react-native-webview';
import type { WebViewErrorEvent, WebViewMessageEvent } from 'react-native-webview/lib/WebViewTypes';

import { wasmMessageSchema } from './schema';
import { initWasm, resetWasm } from './wasm';

// Standalone SDK page bundled with the app (WASM embedded as base64), served
// from a local file instead of a hosted domain.
const defaultHtmlAsset = require('./wasm-wrapper.standalone.html');

const hiddenStyle: ViewStyle = {
  height: 0,
  width: 0,
  flex: 0,
  maxWidth: 0,
  maxHeight: 0,
  overflow: 'hidden',
};

const executePromises: Record<
  string,
  { resolve: (value: unknown) => void; reject: (reason?: unknown) => void } | undefined
> = {};

export type LighterSdkWebViewProps = {
  /** Result of `require()`-ing an alternate standalone SDK html file. Defaults to the bundled one. */
  htmlAsset?: number;
  style?: StyleProp<ViewStyle>;
  onInfo?: (message: string | string[]) => void;
  onLog?: (message: string | string[]) => void;
  onWarn?: (message: string | string[]) => void;
  onError?: (message: string | string[]) => void;
};

export const LighterSdkWebView = ({
  htmlAsset = defaultHtmlAsset,
  style,
  onInfo,
  onLog,
  onWarn,
  onError,
}: LighterSdkWebViewProps) => {
  const webviewRef = useRef<WebView<{}>>(null);
  const [refetchKey, setRefetchKey] = useState(0);

  const [assets] = useAssets(htmlAsset);

  const refreshSdk = useCallback(() => {
    resetWasm();
    setRefetchKey((prev) => prev + 1);
  }, []);

  const localUri = assets?.[0]?.localUri;

  const sdkUrl = useMemo(() => {
    if (!localUri) {
      return undefined;
    }
    return `${localUri}?platform=${Platform.OS}&refetchKey=${refetchKey}`;
  }, [localUri, refetchKey]);

  if (!sdkUrl) {
    return null;
  }

  return (
    <View style={[hiddenStyle, style]}>
      <WebView<{}>
        style={hiddenStyle}
        originWhitelist={['*']}
        allowFileAccess
        allowFileAccessFromFileURLs
        allowingReadAccessToURL={localUri ?? undefined}
        source={{
          uri: sdkUrl,
        }}
        onContentProcessDidTerminate={refreshSdk} // fixes the trading view white screen on iOS
        onMessage={(e: WebViewMessageEvent) => {
          const parsed = wasmMessageSchema.safeParse(e.nativeEvent.data);

          if (!parsed.success) {
            onError?.(`Go SDK WebView received invalid message: ${e.nativeEvent.data}`);
            return;
          }

          const msg = parsed.data;

          switch (msg.type) {
            case 'ready':
              initWasm((args) => {
                return new Promise((resolve, reject) => {
                  const executeId = `${args.function}_${Date.now()}__${Math.random().toString(36).slice(2)}`;
                  executePromises[executeId] = { resolve, reject };
                  webviewRef.current?.postMessage(
                    JSON.stringify({ type: 'execute', ...args, executeId }),
                  );
                });
              });

              break;
            case 'executeResult':
              executePromises[msg.executeId]?.resolve(msg.result);
              delete executePromises[msg.executeId];
              break;
            case 'executeError':
              executePromises[msg.executeId]?.reject(new Error(msg.message));
              delete executePromises[msg.executeId];
              onError?.(`Go SDK WebView execute error: ${msg.message}`);
              break;

            case 'log':
              onLog?.(msg.message);
              break;
            case 'warn':
              onWarn?.(msg.message);
              break;
            case 'error':
              onError?.(msg.message);
              break;
          }
        }}
        ref={webviewRef}
        onError={(event: WebViewErrorEvent) => {
          const { nativeEvent } = event;
          onError?.(`Go SDK WebView load error: ${nativeEvent.description} (code: ${nativeEvent.code}, url: ${nativeEvent.url})`);
          if (nativeEvent.code === -6) {
            // net::ERR_CONNECTION_CLOSED sentry error
            refreshSdk();
          }
        }}
        onHttpError={(event) => {
          const { nativeEvent } = event;
          onError?.(`Go SDK WebView HTTP error: ${nativeEvent.statusCode} ${nativeEvent.url}`);
        }}
        onLoadEnd={(event) => {
          const { nativeEvent } = event;
          onInfo?.(`Go SDK WebView load end: ${nativeEvent.url} (loading: ${nativeEvent.loading})`);
        }}
      />
    </View>
  );
};
