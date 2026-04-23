import assert from "node:assert/strict";
import fs from "fs";
import path from "path";
import { fileURLToPath } from "url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(__dirname, "..", "..");
const buildDir = path.join(repoRoot, "build");

await import(path.join(buildDir, "wasm_exec.js"));

const go = new globalThis.Go();
const wasmBytes = fs.readFileSync(path.join(buildDir, "lighter-signer.wasm"));
const { instance } = await WebAssembly.instantiate(wasmBytes, go.importObject);

go.run(instance);

function assertHex(name, value) {
    assert.equal(typeof value, "string", `${name} should be a string`);
    assert.match(value, /^0x[0-9a-f]+$/i, `${name} should be a hex string`);
}

function assertNoError(name, result) {
    assert.ok(result && typeof result === "object", `${name} should return an object`);
    assert.equal(result.error, undefined, `${name} failed: ${result.error}`);
}

function assertSignedTx(name, result, expectedTxType) {
    assertNoError(name, result);
    assert.equal(result.txType, expectedTxType, `${name} returned an unexpected txType`);
    assert.equal(typeof result.txInfo, "string", `${name} should return txInfo`);
    assert.equal(typeof result.txHash, "string", `${name} should return txHash`);
    assert.match(result.txHash, /^[0-9a-f]+$/i, `${name} txHash should be hex`);

    const txInfo = JSON.parse(result.txInfo);
    assert.equal(typeof txInfo.Nonce, "number", `${name} txInfo should include Nonce`);
    assert.equal(typeof txInfo.Sig, "string", `${name} txInfo should include Sig`);
    return txInfo;
}

function assertSkipNonceAttr(name, txInfo, expected) {
    if (expected) {
        assert.equal(txInfo.L2TxAttributes?.["4"], 1, `${name} should include skipNonce attribute`);
        return;
    }
    assert.equal(txInfo.L2TxAttributes, null, `${name} should not include tx attributes`);
}

const keyResult = globalThis.GenerateAPIKey();
console.log("GenerateAPIKey:", keyResult);
assertNoError("GenerateAPIKey", keyResult);
assertHex("GenerateAPIKey.privateKey", keyResult.privateKey);
assertHex("GenerateAPIKey.publicKey", keyResult.publicKey);
const privateKey = keyResult.privateKey;

const createResult = globalThis.CreateClient("http://localhost:1234", privateKey, 304, 0, 1);
console.log("CreateClient:", createResult);
assertNoError("CreateClient", createResult);

const cancelResult = globalThis.SignCancelOrder(0, 12345, 1, 42, 0, 1);
console.log("SignCancelOrder (skipNonce=1):", cancelResult);
const cancelTxInfo = assertSignedTx("SignCancelOrder (skipNonce=1)", cancelResult, 15);
assert.equal(cancelTxInfo.AccountIndex, 1);
assert.equal(cancelTxInfo.ApiKeyIndex, 0);
assert.equal(cancelTxInfo.MarketIndex, 0);
assert.equal(cancelTxInfo.Index, 12345);
assert.equal(cancelTxInfo.Nonce, 42);
assertSkipNonceAttr("SignCancelOrder (skipNonce=1)", cancelTxInfo, true);

const cancelResult2 = globalThis.SignCancelOrder(0, 12345, 0, 42, 0, 1);
console.log("SignCancelOrder (skipNonce=0):", cancelResult2);
const cancelTxInfo2 = assertSignedTx("SignCancelOrder (skipNonce=0)", cancelResult2, 15);
assert.equal(cancelTxInfo2.Nonce, 42);
assertSkipNonceAttr("SignCancelOrder (skipNonce=0)", cancelTxInfo2, false);
assert.notEqual(cancelResult.txHash, cancelResult2.txHash, "skipNonce should affect the signed tx hash");

const cancelAllResult = globalThis.SignCancelAllOrders(0, 0, 1, 42, 0, 1);
console.log("SignCancelAllOrders (skipNonce=1):", cancelAllResult);
const cancelAllTxInfo = assertSignedTx("SignCancelAllOrders (skipNonce=1)", cancelAllResult, 16);
assert.equal(cancelAllTxInfo.TimeInForce, 0);
assert.equal(cancelAllTxInfo.Time, 0);
assertSkipNonceAttr("SignCancelAllOrders (skipNonce=1)", cancelAllTxInfo, true);

const orderResult = globalThis.SignCreateOrder(
    0, 1, 1000, 50000, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 42, 0, 1
);
console.log("SignCreateOrder (skipNonce=1):", orderResult);
const orderTxInfo = assertSignedTx("SignCreateOrder (skipNonce=1)", orderResult, 14);
assert.equal(orderTxInfo.MarketIndex, 0);
assert.equal(orderTxInfo.ClientOrderIndex, 1);
assert.equal(orderTxInfo.BaseAmount, 1000);
assert.equal(orderTxInfo.Price, 50000);
assert.equal(orderTxInfo.TimeInForce, 0);
assert.equal(orderTxInfo.OrderExpiry, 0);
assertSkipNonceAttr("SignCreateOrder (skipNonce=1)", orderTxInfo, true);

const subAccResult = globalThis.SignCreateSubAccount(1, 42, 0, 1);
console.log("SignCreateSubAccount (skipNonce=1):", subAccResult);
const subAccTxInfo = assertSignedTx("SignCreateSubAccount (skipNonce=1)", subAccResult, 9);
assert.equal(subAccTxInfo.AccountIndex, 1);
assert.equal(subAccTxInfo.Nonce, 42);
assertSkipNonceAttr("SignCreateSubAccount (skipNonce=1)", subAccTxInfo, true);

const levResult = globalThis.SignUpdateLeverage(0, 100, 0, 1, 42, 0, 1);
console.log("SignUpdateLeverage (skipNonce=1):", levResult);
const levTxInfo = assertSignedTx("SignUpdateLeverage (skipNonce=1)", levResult, 20);
assert.equal(levTxInfo.MarketIndex, 0);
assert.equal(levTxInfo.InitialMarginFraction, 100);
assert.equal(levTxInfo.MarginMode, 0);
assertSkipNonceAttr("SignUpdateLeverage (skipNonce=1)", levTxInfo, true);

const expiry = Date.now() + 7 * 24 * 60 * 60 * 1000;

const groupedResult = globalThis.SignCreateGroupedOrders(
    1,
    [
        {
            MarketIndex: 0, ClientOrderIndex: 0, BaseAmount: 1000, Price: 50000,
            IsAsk: 0, Type: 0, TimeInForce: 1, ReduceOnly: 0, TriggerPrice: 0, OrderExpiry: expiry,
        },
        {
            MarketIndex: 0, ClientOrderIndex: 0, BaseAmount: 0, Price: 51000,
            IsAsk: 1, Type: 4, TimeInForce: 0, ReduceOnly: 1, TriggerPrice: 49000, OrderExpiry: expiry,
        },
    ],
    0, 0, 0, 1, 42, 0, 1
);
console.log("SignCreateGroupedOrders (skipNonce=1):", groupedResult);
const groupedTxInfo = assertSignedTx("SignCreateGroupedOrders (skipNonce=1)", groupedResult, 28);
assert.equal(groupedTxInfo.GroupingType, 1);
assert.equal(groupedTxInfo.Orders.length, 2);
assert.equal(groupedTxInfo.Orders[0].TimeInForce, 1);
assert.equal(groupedTxInfo.Orders[1].Type, 4);
assert.equal(groupedTxInfo.Orders[1].TriggerPrice, 49000);
assertSkipNonceAttr("SignCreateGroupedOrders (skipNonce=1)", groupedTxInfo, true);

console.log("\n--- All assertions passed ---");
