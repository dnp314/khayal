const enc = new TextEncoder();
const dec = new TextDecoder();

const PBKDF2_ITERATIONS = 600000;
const KDF_ALGORITHM = "PBKDF2-SHA256";
const KDF_VERSION = 1;
const AES_GCM_IV_BYTES = 12;
const HKDF_SALT = new Uint8Array(0);

export function randomBytes(length) {
  const out = new Uint8Array(length);
  crypto.getRandomValues(out);
  return out;
}

export function b64(bytes) {
  const view = bytes instanceof Uint8Array ? bytes : new Uint8Array(bytes);
  let binary = "";
  for (const byte of view) binary += String.fromCharCode(byte);
  return btoa(binary).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/, "");
}

export function fromB64(value) {
  if (!value) return new Uint8Array(0);
  let s = value.replace(/-/g, "+").replace(/_/g, "/");
  while (s.length % 4) s += "=";
  const binary = atob(s);
  const out = new Uint8Array(binary.length);
  for (let i = 0; i < binary.length; i++) out[i] = binary.charCodeAt(i);
  return out;
}

export function utf8(value) {
  return enc.encode(value);
}

export function text(bytes) {
  return dec.decode(bytes);
}

async function hkdfBits(masterBits, info) {
  const key = await crypto.subtle.importKey("raw", masterBits, "HKDF", false, ["deriveBits"]);
  return new Uint8Array(
    await crypto.subtle.deriveBits(
      { name: "HKDF", hash: "SHA-256", salt: HKDF_SALT, info: utf8(info) },
      key,
      256,
    ),
  );
}

export async function deriveKeys(password, salt, iterations = PBKDF2_ITERATIONS) {
  const passwordKey = await crypto.subtle.importKey(
    "raw",
    utf8(password),
    "PBKDF2",
    false,
    ["deriveBits"],
  );
  const masterBits = await crypto.subtle.deriveBits(
    { name: "PBKDF2", hash: "SHA-256", salt, iterations },
    passwordKey,
    256,
  );

  const authBits = await hkdfBits(masterBits, "khayal-auth-v1");
  const wrapBits = await hkdfBits(masterBits, "khayal-wrap-v1");
  const wrappingKey = await crypto.subtle.importKey(
    "raw",
    wrapBits,
    { name: "AES-GCM", length: 256 },
    false,
    ["wrapKey", "unwrapKey"],
  );

  return {
    authKey: b64(authBits),
    wrappingKey,
    parameters: {
      kdf_algorithm: KDF_ALGORITHM,
      kdf_iterations: iterations,
      kdf_version: KDF_VERSION,
      salt: b64(salt),
    },
  };
}

export function newSalt() {
  return randomBytes(16);
}

export async function generateDataKey() {
  return crypto.subtle.generateKey(
    { name: "AES-GCM", length: 256 },
    true,
    ["encrypt", "decrypt"],
  );
}

export async function exportDataKey(dataKey) {
  return b64(new Uint8Array(await crypto.subtle.exportKey("raw", dataKey)));
}

export async function importDataKey(rawB64, extractable = false) {
  return crypto.subtle.importKey(
    "raw",
    fromB64(rawB64),
    { name: "AES-GCM", length: 256 },
    extractable,
    ["encrypt", "decrypt"],
  );
}

export async function makeNonExtractable(dataKey) {
  const raw = await crypto.subtle.exportKey("raw", dataKey);
  return crypto.subtle.importKey("raw", raw, { name: "AES-GCM", length: 256 }, false, [
    "encrypt",
    "decrypt",
  ]);
}

export async function wrapDataKey(dataKey, wrappingKey) {
  const iv = randomBytes(AES_GCM_IV_BYTES);
  const wrapped = await crypto.subtle.wrapKey("raw", dataKey, wrappingKey, {
    name: "AES-GCM",
    iv,
    tagLength: 128,
  });
  return { wrapped: new Uint8Array(wrapped), iv };
}

export async function unwrapDataKey(wrapped, iv, wrappingKey, extractable = false) {
  return crypto.subtle.unwrapKey(
    "raw",
    wrapped,
    wrappingKey,
    { name: "AES-GCM", iv, tagLength: 128 },
    { name: "AES-GCM", length: 256 },
    extractable,
    ["encrypt", "decrypt"],
  );
}

export async function encryptRecord(dataKey, plaintext, aad) {
  const iv = randomBytes(AES_GCM_IV_BYTES);
  const ciphertext = new Uint8Array(
    await crypto.subtle.encrypt(
      { name: "AES-GCM", iv, additionalData: utf8(aad), tagLength: 128 },
      dataKey,
      utf8(plaintext),
    ),
  );
  const blob = new Uint8Array(iv.length + ciphertext.length);
  blob.set(iv, 0);
  blob.set(ciphertext, iv.length);
  return blob;
}

export async function decryptRecord(dataKey, blob, aad) {
  if (blob.length < AES_GCM_IV_BYTES + 16) throw new Error("ciphertext is too short");
  const iv = blob.slice(0, AES_GCM_IV_BYTES);
  const ciphertext = blob.slice(AES_GCM_IV_BYTES);
  const plaintext = await crypto.subtle.decrypt(
    { name: "AES-GCM", iv, additionalData: utf8(aad), tagLength: 128 },
    dataKey,
    ciphertext,
  );
  return text(plaintext);
}

export function newRecoveryCode() {
  return b64(randomBytes(32));
}

async function recoveryKey(code) {
  const raw = fromB64(code);
  if (raw.length < 16) throw new Error("recovery code is too short");
  const bits = await hkdfBits(raw, "khayal-recovery-wrap-v1");
  return crypto.subtle.importKey("raw", bits, { name: "AES-GCM", length: 256 }, false, [
    "wrapKey",
    "unwrapKey",
  ]);
}

export async function wrapForRecovery(dataKey, code) {
  return wrapDataKey(dataKey, await recoveryKey(code));
}

export async function unwrapFromRecovery(wrapped, iv, code) {
  return unwrapDataKey(wrapped, iv, await recoveryKey(code), true);
}

export async function recoveryVerifier(dataKey) {
  const raw = await crypto.subtle.exportKey("raw", dataKey);
  const bits = await hkdfBits(raw, "khayal-recovery-verifier-v1");
  return b64(bits);
}

export const kdfDefaults = {
  algorithm: KDF_ALGORITHM,
  iterations: PBKDF2_ITERATIONS,
  version: KDF_VERSION,
};
