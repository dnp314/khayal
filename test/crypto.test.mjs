import test from "node:test";
import assert from "node:assert/strict";

import {
  b64,
  decryptRecord,
  deriveKeys,
  encryptRecord,
  fromB64,
  generateDataKey,
  newSalt,
  unwrapDataKey,
  wrapDataKey,
} from "../static/js/crypto.js";

test("same password and salt derive the same auth key", async () => {
  const salt = newSalt();
  const a = await deriveKeys("correct horse battery staple", salt);
  const b = await deriveKeys("correct horse battery staple", salt);
  assert.equal(a.authKey, b.authKey);
});

test("different passwords derive different auth keys", async () => {
  const salt = newSalt();
  const a = await deriveKeys("correct horse battery staple", salt);
  const b = await deriveKeys("incorrect horse battery staple", salt);
  assert.notEqual(a.authKey, b.authKey);
});

test("record encryption round-trips", async () => {
  const key = await generateDataKey();
  const aad = "khayal:v1:question:1:11111111-1111-1111-1111-111111111111";
  const blob = await encryptRecord(key, "hello world", aad);
  assert.equal(await decryptRecord(key, blob, aad), "hello world");
});

test("wrong AAD fails", async () => {
  const key = await generateDataKey();
  const blob = await encryptRecord(key, "hello world", "khayal:v1:question:1:a");
  await assert.rejects(() => decryptRecord(key, blob, "khayal:v1:question:1:b"));
});

test("tampered ciphertext fails", async () => {
  const key = await generateDataKey();
  const blob = await encryptRecord(key, "hello world", "khayal:v1:question:1:a");
  blob[15] ^= 1;
  await assert.rejects(() => decryptRecord(key, blob, "khayal:v1:question:1:a"));
});

test("IVs are unique across encryptions", async () => {
  const key = await generateDataKey();
  const aad = "khayal:v1:question:1:a";
  const a = await encryptRecord(key, "same text", aad);
  const b = await encryptRecord(key, "same text", aad);
  assert.notDeepEqual(a.slice(0, 12), b.slice(0, 12));
});

test("wrong password cannot unwrap the data key", async () => {
  const salt = newSalt();
  const right = await deriveKeys("correct horse battery staple", salt);
  const wrong = await deriveKeys("incorrect horse battery staple", salt);
  const dataKey = await generateDataKey();
  const wrapped = await wrapDataKey(dataKey, right.wrappingKey);

  await assert.rejects(() =>
    unwrapDataKey(wrapped.wrapped, wrapped.iv, wrong.wrappingKey, false),
  );
  const unwrapped = await unwrapDataKey(wrapped.wrapped, wrapped.iv, right.wrappingKey, false);
  assert.equal(unwrapped.algorithm.name, "AES-GCM");
});

test("base64url helpers round-trip binary values", async () => {
  const value = new Uint8Array([0, 1, 2, 250, 251, 252]);
  assert.deepEqual(fromB64(b64(value)), value);
});
