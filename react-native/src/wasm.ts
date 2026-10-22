export type WasmExecutorArgs = {
  function: string;
  params: Array<unknown>;
  /**
   * Signer identity the page selects before running any function that takes
   * an account index. Required for everything except `_createClient`,
   * `_createClientByPrv`, `_lighterTest` and `_resetClients`.
   */
  chainId?: number;
  apiKeyIndex?: number;
  type?: string;
};

export type WasmExecutor = (args: WasmExecutorArgs) => Promise<unknown>;

const noopExecutor: WasmExecutor = () => Promise.resolve();

let executor: WasmExecutor = noopExecutor;

let wasmReadyResolve: () => void;

let wasmReady = new Promise<void>((resolve) => {
  wasmReadyResolve = resolve;
});

/**
 * Wires up the function that actually talks to the WASM WebView. Called by
 * `LighterSdkWebView` once the page reports it's ready - not meant to be called
 * directly by consumers.
 */
export const initWasm = (newExecutor: WasmExecutor) => {
  executor = newExecutor;
  wasmReadyResolve();
};

/**
 * Reverts to the no-op executor and re-arms `wasmReady`, so calls made while
 * the WebView is reloading queue up instead of hitting a stale executor.
 */
export const resetWasm = () => {
  executor = noopExecutor;
  wasmReady = new Promise<void>((resolve) => {
    wasmReadyResolve = resolve;
  });
};

/**
 * Calls a Go function exposed by the WASM bundle. Resolves once the WebView
 * has loaded, so it's safe to call before `LighterSdkWebView` has mounted/finished
 * initializing - the call just waits.
 */
export const goWasmExecute: WasmExecutor = async (args) => {
  await wasmReady;

  return executor(args);
};
