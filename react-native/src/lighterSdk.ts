import { goWasmExecute } from './wasm';

// Functions that carry their own apiKeyIndex argument (0-based position); every
// other identity-bound function uses the key the client was created with.
const explicitKeyPosition: Record<string, number> = {
  _createAuthToken: 1,
  _getChangePubKeyTransaction: 2,
  _getRevokePubKeyTransaction: 2,
  _signChangePubKey: 3,
  _signRevokePubKey: 3,
  _getTransferTransaction: 2,
  _signTransfer: 3,
  _getApproveIntegratorTransaction: 2,
  _signApproveIntegrator: 3,
};

const identityFree = new Set(['_createClient', '_createClientByPrv', '_lighterTest', '_resetClients']);

type SignerIdentity = { chainId: number; apiKeyIndex: number };
const identities = new Map<number, SignerIdentity>();

/**
 * Registers which chain/api key a signer account was created with, so the
 * page can select that client before every call. `createClient` and
 * `createClientByPrv` do this automatically.
 */
export const setSignerIdentity = (accountIndex: number, chainId: number, apiKeyIndex: number) => {
  identities.set(accountIndex, { chainId, apiKeyIndex });
};

export const clearSignerIdentities = () => {
  identities.clear();
};

const execute = async (fn: string, params: Array<unknown>): Promise<unknown> => {
  if (identityFree.has(fn)) {
    return goWasmExecute({ function: fn, params });
  }
  const accountIndex = params[0];
  const identity = typeof accountIndex === 'number' ? identities.get(accountIndex) : undefined;
  if (!identity) {
    throw new Error(`Signer identity not initialized for account ${String(accountIndex)}`);
  }
  const position = explicitKeyPosition[fn];
  const apiKeyIndex = position === undefined ? identity.apiKeyIndex : params[position];
  if (typeof apiKeyIndex !== 'number') {
    throw new Error(`Invalid apiKeyIndex for ${fn}`);
  }
  return goWasmExecute({ function: fn, params, chainId: identity.chainId, apiKeyIndex });
};

// Shape returned by every Go-side signing function that doesn't need a more
// specific response type. Defined locally (rather than imported from
// `@elliottech/react-store`) so this package has no dependency on it -
// consumers already using that package's `TxResponse` type can freely
// substitute it, since the shapes match.
export type TxResponse = {
  txInfo: string;
  txHash: string;
};

// The page strips `prv` from the response before it reaches native code.
export type LighterSdkCreateClientResponse = {
  body: string;
  pk: string;
};

export type LighterSdkCreateAuthTokenResponse = {
  token: string;
  deadline: number;
};

export type LighterSdkSignChangePubKeyResponse = {
  txInfo: string;
};

export type LighterSdkGetTransferTransactionResponse = {
  pk: string;
  body: string;
};

export type LighterSdkCreateOrderArgs = {
  accountIndex: number;
  marketIndex: number;
  clientOrderIndex: number;
  baseAmount: number;
  price: number;
  isAsk: number;
  orderType: number;
  timeInForce: number;
  reduceOnly: number;
  triggerPrice: number;
  orderExpiry: number;
  nonce: number;
};

export type LighterSdkTransferArgs = {
  accountIndex: number;
  signedMessage: string;
  nonce: number;
  apiKeyIndex: number;
  toAccountIndex: number;
  assetIndex: number;
  fromRouteType: number;
  toRouteType: number;
  amount: number;
  USDCFee: number;
};

export type LighterSdkGetTransferTransactionArgs = {
  accountIndex: number;
  nonce: number;
  apiKeyIndex: number;
  toAccountIndex: number;
  assetIndex: number;
  fromRouteType: number;
  toRouteType: number;
  amount: number;
  USDCFee: number;
};

export type LighterSdkGroupedOrder = {
  marketIndex: number;
  clientOrderIndex: number;
  baseAmount: string;
  price: string;
  isAsk: boolean;
  orderType: number;
  timeInForce: number;
  reduceOnly: boolean;
  triggerPrice: string;
  orderExpiry: number;
};

export const LighterSDK = {
  createClient: async (
    ...args: [
      seed: string,
      chainId: number,
      accountIndex: number,
      nonce: number,
      apiKeyIndex: number,
      skipNonce?: boolean,
    ]
  ): Promise<LighterSdkCreateClientResponse> => {
    const response = await execute('_createClient', args);
    setSignerIdentity(args[2], args[1], args[4]);
    return response as LighterSdkCreateClientResponse;
  },

  createClientByPrv: async (
    ...args: [
      privateKey: string,
      chainId: number,
      accountIndex: number,
      nonce: number,
      apiKeyIndex: number,
      skipNonce?: boolean,
    ]
  ): Promise<LighterSdkCreateClientResponse> => {
    const response = await execute('_createClientByPrv', args);
    setSignerIdentity(args[2], args[1], args[4]);
    return response as LighterSdkCreateClientResponse;
  },

  resetClients: async (): Promise<{ success: boolean }> => {
    const response = await execute('_resetClients', []);
    clearSignerIdentities();
    return response as { success: boolean };
  },

  createAuthToken: async (
    ...args: [accountIndex: number, apiKeyIndex: number]
  ): Promise<LighterSdkCreateAuthTokenResponse> => {
    const response = await execute('_createAuthToken', args);
    return response as LighterSdkCreateAuthTokenResponse;
  },

  signChangePubKey: async (
    ...args: [accountIndex: number, signedMessage: string, nonce: number, apiKeyIndex: number]
  ): Promise<LighterSdkSignChangePubKeyResponse> => {
    const response = await execute('_signChangePubKey', args);
    return response as LighterSdkSignChangePubKeyResponse;
  },

  signUpdateLeverage: async (
    ...args: [
      accountIndex: number,
      marketIndex: number,
      initialMarginFraction: number,
      marginMode: number,
      nonce: number,
    ]
  ): Promise<TxResponse> => {
    const response = await execute('_signUpdateLeverage', args);
    return response as TxResponse;
  },

  signStakeAssets: async (
    ...args: [accountIndex: number, stakingPoolIndex: number, shareAmount: string, nonce: number]
  ): Promise<TxResponse> => {
    const response = await execute('_signStakeAssets', args);
    return response as TxResponse;
  },

  signUnstakeAssets: async (
    ...args: [accountIndex: number, stakingPoolIndex: number, shareAmount: string, nonce: number]
  ): Promise<TxResponse> => {
    const response = await execute('_signUnstakeAssets', args);
    return response as TxResponse;
  },

  signMintShares: async (
    ...args: [accountIndex: number, publicPoolIndex: number, shareAmount: string, nonce: number]
  ): Promise<TxResponse> => {
    const response = await execute('_signMintShares', args);
    return response as TxResponse;
  },

  signBurnShares: async (
    ...args: [accountIndex: number, publicPoolIndex: number, shareAmount: string, nonce: number]
  ): Promise<TxResponse> => {
    const response = await execute('_signBurnShares', args);
    return response as TxResponse;
  },

  signWithdraw: async (
    ...args: [
      accountIndex: number,
      assetIndex: number,
      routeType: number,
      amountStr: string,
      nonce: number,
    ]
  ): Promise<TxResponse> => {
    const response = await execute('_signWithdraw', args);
    return response as TxResponse;
  },

  signTransfer: async (
    ...args: [transfer: LighterSdkTransferArgs, memo: Uint8Array]
  ): Promise<TxResponse> => {
    const response = await execute('_signTransfer', [
        args[0].accountIndex,
        args[0].signedMessage,
        args[0].nonce,
        args[0].apiKeyIndex,
        args[0].toAccountIndex,
        args[0].assetIndex,
        args[0].fromRouteType,
        args[0].toRouteType,
        args[0].amount,
        args[0].USDCFee,
        args[1],
      ]);
    return response as TxResponse;
  },

  getTransferTransaction: async (
    ...args: [transfer: LighterSdkGetTransferTransactionArgs, memo: Uint8Array]
  ): Promise<LighterSdkGetTransferTransactionResponse> => {
    const response = await execute('_getTransferTransaction', [
        args[0].accountIndex,
        args[0].nonce,
        args[0].apiKeyIndex,
        args[0].toAccountIndex,
        args[0].assetIndex,
        args[0].fromRouteType,
        args[0].toRouteType,
        args[0].amount,
        args[0].USDCFee,
        args[1],
      ]);
    return response as LighterSdkGetTransferTransactionResponse;
  },

  signUpdateMargin: async (
    ...args: [
      accountIndex: number,
      marketIndex: number,
      usdcAmount: number,
      direction: number,
      nonce: number,
    ]
  ): Promise<TxResponse> => {
    const response = await execute('_signUpdateMargin', args);
    return response as TxResponse;
  },

  signUpdateAccountAssetConfig: async (
    ...args: [accountIndex: number, assetIndex: number, assetMarginMode: number, nonce: number]
  ): Promise<TxResponse> => {
    const response = await execute('_signUpdateAccountAssetConfig', args);
    return response as TxResponse;
  },

  signModifyOrder: async (
    ...args: [
      accountIndex: number,
      marketIndex: number,
      orderIndex: string,
      baseAmount: number,
      price: number,
      triggerPrice: number,
      nonce: number,
    ]
  ): Promise<TxResponse> => {
    const response = await execute('_signModifyOrder', args);
    return response as TxResponse;
  },

  signCancelOrder: async (
    ...args: [accountIndex: number, marketIndex: number, orderIndex: string, nonce: number]
  ): Promise<TxResponse> => {
    const response = await execute('_signCancelOrder', args);
    return response as TxResponse;
  },

  signCancelAllOrders: async (
    ...args: [
      accountIndex: number,
      timeInForce: number,
      time: number,
      nonce: number,
      marketIndex?: number,
    ]
  ): Promise<TxResponse> => {
    const response = await execute('_signCancelAllOrders', args);
    return response as TxResponse;
  },

  signCreateOrder: async (
    ...args: [order: LighterSdkCreateOrderArgs]
  ): Promise<TxResponse> => {
    const response = await execute('_signCreateOrder', [
        args[0].accountIndex,
        args[0].marketIndex,
        args[0].clientOrderIndex,
        args[0].baseAmount.toString(),
        args[0].price.toString(),
        args[0].isAsk,
        args[0].orderType,
        args[0].timeInForce,
        args[0].reduceOnly,
        args[0].triggerPrice.toString(),
        args[0].orderExpiry,
        args[0].nonce,
      ]);
    return response as TxResponse;
  },

  signCreateGroupedOrders: async (
    ...args: [accountIndex: number, groupingType: number, ordersJsonStr: string, nonce: number]
  ): Promise<TxResponse> => {
    const orders = JSON.parse(args[2]) as Array<LighterSdkGroupedOrder>;
    const params = orders.flatMap((order) => [
      order.marketIndex,
      order.clientOrderIndex,
      order.baseAmount.toString(),
      order.price.toString(),
      order.isAsk,
      order.orderType,
      order.timeInForce,
      order.reduceOnly,
      order.triggerPrice.toString(),
      order.orderExpiry,
    ]);
    const response = await execute('_signCreateGroupedOrders', [
        args[0], // accountIndex
        args[1], // groupingType
        orders.length, // number of orders
        ...params,
        args[3], // nonce
      ]);
    return response as TxResponse;
  },

  signUpdateAccountConfig: async (
    ...args: [accountIndex: number, accountTradingMode: number, nonce: number]
  ): Promise<TxResponse> => {
    const response = await execute('_signUpdateAccountConfig', [args[0], args[1], args[2]]);
    return response as TxResponse;
  },

  signCreateSubAccount: async (
    ...args: [accountIndex: number, nonce: number]
  ): Promise<TxResponse> => {
    const response = await execute('_signCreateSubAccount', [args[0], args[1]]);
    return response as TxResponse;
  },
};
