const DB_NAME = "khayal-keys";
const STORE = "keys";
const USER_KEY = "khayal_user_id";
const FALLBACK_KEY = "khayal_data_key_fallback";

function openDB() {
  return new Promise((resolve, reject) => {
    const req = indexedDB.open(DB_NAME, 1);
    req.onupgradeneeded = () => {
      if (!req.result.objectStoreNames.contains(STORE)) req.result.createObjectStore(STORE);
    };
    req.onsuccess = () => resolve(req.result);
    req.onerror = () => reject(req.error);
  });
}

export function currentUserID() {
  return localStorage.getItem(USER_KEY);
}

export function setCurrentUserID(userID) {
  localStorage.setItem(USER_KEY, String(userID));
}

export async function storeDataKey(userID, dataKey, rawFallback = "") {
  setCurrentUserID(userID);
  try {
    const db = await openDB();
    await new Promise((resolve, reject) => {
      const tx = db.transaction(STORE, "readwrite");
      tx.objectStore(STORE).put(dataKey, String(userID));
      tx.oncomplete = resolve;
      tx.onerror = () => reject(tx.error);
      tx.onabort = () => reject(tx.error || new Error("key store aborted"));
    });
    db.close();
    sessionStorage.removeItem(FALLBACK_KEY);
  } catch (err) {
    if (!rawFallback) throw err;
    sessionStorage.setItem(FALLBACK_KEY, rawFallback);
  }
}

export async function getDataKey() {
  const userID = currentUserID();
  if (!userID) return null;
  try {
    const db = await openDB();
    const key = await new Promise((resolve, reject) => {
      const tx = db.transaction(STORE, "readonly");
      const req = tx.objectStore(STORE).get(String(userID));
      req.onsuccess = () => resolve(req.result || null);
      req.onerror = () => reject(req.error);
    });
    db.close();
    if (key) return key;
  } catch (_) {
    // Fall through to the session fallback.
  }

  const raw = sessionStorage.getItem(FALLBACK_KEY);
  if (!raw) return null;
  const { importDataKey } = await import("./crypto.js");
  const key = await importDataKey(raw, false);
  await storeDataKey(userID, key);
  return key;
}

export async function clearDataKey() {
  const userID = currentUserID();
  if (userID) {
    try {
      const db = await openDB();
      await new Promise((resolve, reject) => {
        const tx = db.transaction(STORE, "readwrite");
        tx.objectStore(STORE).delete(String(userID));
        tx.oncomplete = resolve;
        tx.onerror = () => reject(tx.error);
      });
      db.close();
    } catch (_) {
      // Best effort: the session cookie is still cleared server-side.
    }
  }
  sessionStorage.removeItem(FALLBACK_KEY);
  localStorage.removeItem(USER_KEY);
}
