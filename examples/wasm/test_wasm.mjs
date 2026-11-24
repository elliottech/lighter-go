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

// GenerateAPIKey returns a secp256k1 ECDSA key: 32-byte private key, 33-byte compressed public key.
assert.equal(privateKey.length, 2 + 64, "GenerateAPIKey.privateKey should be 32 bytes");
assert.equal(keyResult.publicKey.length, 2 + 66, "GenerateAPIKey.publicKey should be 33 bytes");

// It is registered with SignChangeECDSAPubKey (tx type 54); the L1 wallet signs messageToSign.
const changeECDSAResult = globalThis.SignChangeECDSAPubKey(keyResult.publicKey, 0, 42, 0, 1);
console.log("SignChangeECDSAPubKey:", changeECDSAResult);
const changeECDSATxInfo = assertSignedTx("SignChangeECDSAPubKey", changeECDSAResult, 54);
assert.equal(changeECDSATxInfo.AccountIndex, 1);
assert.equal(changeECDSATxInfo.ApiKeyIndex, 0);
assert.equal(changeECDSATxInfo.PubKey.length, 33);
assert.equal(Buffer.from(changeECDSATxInfo.PubKey).toString("hex"), keyResult.publicKey.slice(2));
assert.equal(Buffer.from(changeECDSATxInfo.Sig, "base64").length, 64, "ECDSA signatures are 64 bytes");
assert.match(changeECDSAResult.messageToSign, /^Register Lighter Account\n\npubkey: 0x[0-9a-f]{66}\n/);
assert.match(changeECDSAResult.messageToSign, /\nexpired at: 0x[0-9a-f]+\nchainId: 0x[0-9a-f]+$/);

// SignChangePubKey is the Schnorr registration and rejects an ECDSA public key.
const wrongSchemeResult = globalThis.SignChangePubKey(keyResult.publicKey, 0, 42, 0, 1);
assert.match(wrongSchemeResult.error, /invalid pub key length/);

// Legacy Schnorr key: 40-byte private and public keys, registered with SignChangePubKey (tx type 8).
const schnorrKeyResult = globalThis.GenerateSchnorrAPIKey();
console.log("GenerateSchnorrAPIKey:", schnorrKeyResult);
assertNoError("GenerateSchnorrAPIKey", schnorrKeyResult);
assert.equal(schnorrKeyResult.privateKey.length, 2 + 80, "GenerateSchnorrAPIKey.privateKey should be 40 bytes");
assert.equal(schnorrKeyResult.publicKey.length, 2 + 80, "GenerateSchnorrAPIKey.publicKey should be 40 bytes");

const schnorrCreateResult = globalThis.CreateClient("http://localhost:1234", schnorrKeyResult.privateKey, 304, 1, 1);
assertNoError("CreateClient (Schnorr)", schnorrCreateResult);

const changeSchnorrResult = globalThis.SignChangePubKey(schnorrKeyResult.publicKey, 0, 42, 1, 1);
console.log("SignChangePubKey:", changeSchnorrResult);
const changeSchnorrTxInfo = assertSignedTx("SignChangePubKey", changeSchnorrResult, 8);
assert.equal(changeSchnorrTxInfo.ApiKeyIndex, 1);
assert.equal(changeSchnorrTxInfo.PubKey.length, 40);
assert.equal(Buffer.from(changeSchnorrTxInfo.Sig, "base64").length, 80, "Schnorr signatures are 80 bytes");

const schnorrOrderResult = globalThis.SignCreateOrder(
    0, 1, 1000, 50000, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 42, 1, 1
);
const schnorrOrderTxInfo = assertSignedTx("SignCreateOrder (Schnorr)", schnorrOrderResult, 14);
assert.equal(Buffer.from(schnorrOrderTxInfo.Sig, "base64").length, 80);

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

const cancelAllResult = globalThis.SignCancelAllOrders(0, 0, 255, 1, 42, 0, 1);
console.log("SignCancelAllOrders (skipNonce=1):", cancelAllResult);
const cancelAllTxInfo = assertSignedTx("SignCancelAllOrders (skipNonce=1)", cancelAllResult, 16);
assert.equal(cancelAllTxInfo.TimeInForce, 0);
assert.equal(cancelAllTxInfo.Time, 0);
assertSkipNonceAttr("SignCancelAllOrders (skipNonce=1)", cancelAllTxInfo, true);

const orderResult = globalThis.SignCreateOrder(
    0, 1, 1000, 50000, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 42, 0, 1
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
    0, 0, 0, 0, 0, 1, 42, 0, 1
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
