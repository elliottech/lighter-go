import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const dirname = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(dirname, "..", "..");
const buildDir = path.join(repoRoot, "build");

await import(path.join(buildDir, "wasm_exec.js"));

const go = new globalThis.Go();
const wasmBytes = fs.readFileSync(path.join(buildDir, "lighter-signer.wasm"));
const { instance } = await WebAssembly.instantiate(wasmBytes, go.importObject);

void go.run(instance);

async function callWasm(name, ...params) {
  const exportedFunction = globalThis[name];
  assert.equal(typeof exportedFunction, "function", `${name} should be exported`);

  const invoke = exportedFunction(...params);
  assert.equal(typeof invoke, "function", `${name} should return an invoker`);
  return invoke();
}

function assertNoError(name, result) {
  assert.ok(result && typeof result === "object", `${name} should return an object`);
  assert.equal(result.error, undefined, `${name} failed: ${result.error}`);
}

function assertSignedTx(name, result) {
  assertNoError(name, result);
  assert.match(result.txHash, /^[0-9a-f]+$/i, `${name} txHash should be hex`);
  assert.equal(typeof result.txInfo, "string", `${name} should return txInfo`);
  return JSON.parse(result.txInfo);
}

const seed = "11".repeat(32);
const chainId = 304;
const accountIndex = 1;

const createResult = await callWasm(
  "_createClient",
  seed,
  chainId,
  accountIndex,
  42,
  0,
  true,
);
console.log("_createClient (skipNonce=true):", createResult);
assertNoError("_createClient", createResult);
assert.equal(createResult.success, true);
assert.equal(createResult.pubKeySuccess, true);
assert.match(createResult.pk, /^[0-9a-f]+$/i);
assert.match(createResult.prv, /^[0-9a-f]+$/i);

const cancelWithSkip = await callWasm(
  "_signCancelOrder",
  accountIndex,
  0,
  "12345",
  43,
);
console.log("_signCancelOrder (skipNonce=true):", cancelWithSkip);
const cancelWithSkipInfo = assertSignedTx("_signCancelOrder", cancelWithSkip);
assert.equal(cancelWithSkipInfo.AccountIndex, accountIndex);
assert.equal(cancelWithSkipInfo.MarketIndex, 0);
assert.equal(cancelWithSkipInfo.Index, 12345);
assert.equal(cancelWithSkipInfo.Nonce, 43);
assert.equal(cancelWithSkipInfo.L2TxAttributes?.["4"], 1);

await callWasm("_createClient", seed, chainId, accountIndex, 42, 0, false);
const cancelWithoutSkip = await callWasm(
  "_signCancelOrder",
  accountIndex,
  0,
  "12345",
  43,
);
console.log("_signCancelOrder (skipNonce=false):", cancelWithoutSkip);
const cancelWithoutSkipInfo = assertSignedTx("_signCancelOrder", cancelWithoutSkip);
assert.equal(cancelWithoutSkipInfo.L2TxAttributes, null);
assert.notEqual(cancelWithSkip.txHash, cancelWithoutSkip.txHash);

const orderResult = await callWasm(
  "_signCreateOrder",
  accountIndex,
  0,
  1,
  "1000",
  "50000",
  0,
  0,
  0,
  0,
  "0",
  0,
  44,
);
console.log("_signCreateOrder:", orderResult);
const orderInfo = assertSignedTx("_signCreateOrder", orderResult);
assert.equal(orderInfo.AccountIndex, accountIndex);
assert.equal(orderInfo.ClientOrderIndex, 1);
assert.equal(orderInfo.BaseAmount, 1000);
assert.equal(orderInfo.Price, 50000);
assert.equal(orderInfo.Nonce, 44);

const allocationResult = await callWasm(
  "_getAirdropAllocationMessage",
  accountIndex,
  "1:100,2:200",
);
console.log("_getAirdropAllocationMessage:", allocationResult);
assertNoError("_getAirdropAllocationMessage", allocationResult);
assert.match(allocationResult.message, /allocations: 1:100,2:200/);
assert.match(allocationResult.message, /chainId: 0x0000000000000130/);

console.log("\n--- All WASM assertions passed ---");
