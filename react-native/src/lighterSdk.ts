import { goWasmExecute } from './wasm';

// Shape returned by every Go-side signing function that doesn't need a more
// specific response type. Defined locally (rather than imported from
// `@elliottech/react-store`) so this package has no dependency on it -
// consumers already using that package's `TxResponse` type can freely
// substitute it, since the shapes match.
export type TxResponse = {
  txInfo: string;
  txHash: string;
};

export type LighterSdkCreateClientResponse = {
  body: string;
  pk: string;
  prv: string;
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
    const response = await goWasmExecute({
      function: '_createClient',
      params: args,
    });
    return response as LighterSdkCreateClientResponse;
  },

  createAuthToken: async (
    ...args: [accountIndex: number, apiKeyIndex: number]
  ): Promise<LighterSdkCreateAuthTokenResponse> => {
    const response = await goWasmExecute({
      function: '_createAuthToken',
      params: args,
    });
    return response as LighterSdkCreateAuthTokenResponse;
  },

  signChangePubKey: async (
    ...args: [accountIndex: number, signedMessage: string, nonce: number, apiKeyIndex: number]
  ): Promise<LighterSdkSignChangePubKeyResponse> => {
    const response = await goWasmExecute({
      function: '_signChangePubKey',
      params: args,
    });
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
    const response = await goWasmExecute({
      function: '_signUpdateLeverage',
      params: args,
    });
    return response as TxResponse;
  },

  signStakeAssets: async (
    ...args: [accountIndex: number, stakingPoolIndex: number, shareAmount: string, nonce: number]
  ): Promise<TxResponse> => {
    const response = await goWasmExecute({
      function: '_signStakeAssets',
      params: args,
    });
    return response as TxResponse;
  },

  signUnstakeAssets: async (
    ...args: [accountIndex: number, stakingPoolIndex: number, shareAmount: string, nonce: number]
  ): Promise<TxResponse> => {
    const response = await goWasmExecute({
      function: '_signUnstakeAssets',
      params: args,
    });
    return response as TxResponse;
  },

  signMintShares: async (
    ...args: [accountIndex: number, publicPoolIndex: number, shareAmount: string, nonce: number]
  ): Promise<TxResponse> => {
    const response = await goWasmExecute({
      function: '_signMintShares',
      params: args,
    });
    return response as TxResponse;
  },

  signBurnShares: async (
    ...args: [accountIndex: number, publicPoolIndex: number, shareAmount: string, nonce: number]
  ): Promise<TxResponse> => {
    const response = await goWasmExecute({
      function: '_signBurnShares',
      params: args,
    });
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
    const response = await goWasmExecute({
      function: '_signWithdraw',
      params: args,
    });
    return response as TxResponse;
  },

  signTransfer: async (
    ...args: [transfer: LighterSdkTransferArgs, memo: Uint8Array]
  ): Promise<TxResponse> => {
    const response = await goWasmExecute({
      function: '_signTransfer',
      params: [
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
      ],
    });
    return response as TxResponse;
  },

  getTransferTransaction: async (
    ...args: [transfer: LighterSdkGetTransferTransactionArgs, memo: Uint8Array]
  ): Promise<LighterSdkGetTransferTransactionResponse> => {
    const response = await goWasmExecute({
      function: '_getTransferTransaction',
      params: [
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
      ],
    });
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
    const response = await goWasmExecute({
      function: '_signUpdateMargin',
      params: args,
    });
    return response as TxResponse;
  },

  signUpdateAccountAssetConfig: async (
    ...args: [accountIndex: number, assetIndex: number, assetMarginMode: number, nonce: number]
  ): Promise<TxResponse> => {
    const response = await goWasmExecute({
      function: '_signUpdateAccountAssetConfig',
      params: args,
    });
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
    const response = await goWasmExecute({
      function: '_signModifyOrder',
      params: args,
    });
    return response as TxResponse;
  },

  signCancelOrder: async (
    ...args: [accountIndex: number, marketIndex: number, orderIndex: string, nonce: number]
  ): Promise<TxResponse> => {
    const response = await goWasmExecute({
      function: '_signCancelOrder',
      params: args,
    });
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
    const response = await goWasmExecute({
      function: '_signCancelAllOrders',
      params: args,
    });
    return response as TxResponse;
  },

  signCreateOrder: async (
    ...args: [order: LighterSdkCreateOrderArgs]
  ): Promise<TxResponse> => {
    const response = await goWasmExecute({
      function: '_signCreateOrder',
      params: [
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
      ],
    });
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
    const response = await goWasmExecute({
      function: '_signCreateGroupedOrders',
      params: [
        args[0], // accountIndex
        args[1], // groupingType
        orders.length, // number of orders
        ...params,
        args[3], // nonce
      ],
    });
    return response as TxResponse;
  },

  signUpdateAccountConfig: async (
    ...args: [accountIndex: number, accountTradingMode: number, nonce: number]
  ): Promise<TxResponse> => {
    const response = await goWasmExecute({
      function: '_signUpdateAccountConfig',
      params: [args[0], args[1], args[2]],
    });
    return response as TxResponse;
  },

  signCreateSubAccount: async (
    ...args: [accountIndex: number, nonce: number]
  ): Promise<TxResponse> => {
    const response = await goWasmExecute({
      function: '_signCreateSubAccount',
      params: [args[0], args[1]],
    });
    return response as TxResponse;
  },
};
