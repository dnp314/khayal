import { b64, decryptRecord, encryptRecord, fromB64 } from "./crypto.js";
import { getDataKey } from "./keystore.js";

let keyPromise = null;

async function dataKey() {
  if (!keyPromise) keyPromise = getDataKey();
  return keyPromise;
}

function formError(form, message) {
  let banner = form.querySelector(".error-banner");
  if (!banner) {
    banner = document.createElement("div");
    banner.className = "error-banner";
    banner.setAttribute("role", "alert");
    form.appendChild(banner);
  }
  banner.textContent = message;
}

function clearFormError(form) {
  const banner = form.querySelector(".error-banner");
  if (banner) banner.remove();
}

async function decryptElement(el) {
  if (el.dataset.decrypted === "1") return;
  const ciphertext = el.dataset.ciphertext;
  const aad = el.dataset.aad;
  if (!ciphertext || !aad) return;

  try {
    const key = await dataKey();
    if (!key) {
      el.textContent = "[log in again to unlock]";
      return;
    }
    el.textContent = await decryptRecord(key, fromB64(ciphertext), aad);
    el.dataset.decrypted = "1";
  } catch (_) {
    el.textContent = "[unable to decrypt]";
    el.dataset.decrypted = "1";
  }
}

export function decryptAll(root = document) {
  root.querySelectorAll("[data-ciphertext]").forEach((el) => decryptElement(el));
}

function growTextarea(el) {
  if (!el || el.tagName !== "TEXTAREA") return;
  const cs = getComputedStyle(el);
  const min = parseFloat(cs.minHeight) || 0;
  const max = parseFloat(cs.maxHeight) || Infinity;
  const start = el.offsetHeight;
  el.style.height = "auto";
  const end = Math.max(min, Math.min(el.scrollHeight + 2, max));
  el.style.height = end + "px";
  if (start !== end) {
    el.animate([{ height: start + "px" }, { height: end + "px" }], {
      duration: 120,
      easing: "ease-out",
    });
  }
}

function addHidden(form, name, value) {
  const input = document.createElement("input");
  input.type = "hidden";
  input.name = name;
  input.value = value;
  input.dataset.encryptedField = "1";
  form.appendChild(input);
}

async function encryptForm(form) {
  const field = form.dataset.encrypt;
  const textarea = form.querySelector("textarea");
  const plaintext = textarea ? textarea.value : "";
  if (!plaintext.trim()) throw new Error("Please write something first.");

  const key = await dataKey();
  if (!key) throw new Error("Please log in again to unlock your data.");

  const userID = form.dataset.userId;
  const clientID = crypto.randomUUID();
  let aad;
  let ciphertextField;

  if (field === "question") {
    aad = `khayal:v1:question:${userID}:${clientID}`;
    ciphertextField = "question_ciphertext";
  } else if (field === "khayal") {
    aad = `khayal:v1:khayal:${userID}:${form.dataset.questionId}:${clientID}`;
    ciphertextField = "khayal_ciphertext";
  } else {
    throw new Error("Unknown encrypted form type.");
  }

  const blob = await encryptRecord(key, plaintext, aad);
  form.querySelectorAll("[data-encrypted-field]").forEach((el) => el.remove());
  addHidden(form, "client_id", clientID);
  addHidden(form, ciphertextField, b64(blob));
  addHidden(form, "encryption_version", "1");
}

function initHTMXHooks() {
  document.body.addEventListener("htmx:confirm", async (event) => {
    const form = event.detail.elt.closest && event.detail.elt.closest("form[data-encrypt]");
    if (!form) return;

    event.preventDefault();
    clearFormError(form);
    try {
      await encryptForm(form);
      event.detail.issueRequest(true);
    } catch (err) {
      formError(form, err.message || "Unable to encrypt.");
    }
  });

  document.body.addEventListener("htmx:afterRequest", (event) => {
    const form = event.detail.elt.closest && event.detail.elt.closest("form[data-encrypt]");
    if (!form || !event.detail.successful) return;
    const textarea = form.querySelector("textarea");
    if (textarea) {
      textarea.value = "";
      growTextarea(textarea);
    }
    form.querySelectorAll("[data-encrypted-field]").forEach((el) => el.remove());
  });

  document.body.addEventListener("htmx:load", (event) => {
    decryptAll(event.target);
  });

  document.body.addEventListener("input", (event) => {
    if (event.target.matches("form textarea")) growTextarea(event.target);
  });

  document.body.addEventListener("keydown", (event) => {
    if (
      event.target.matches("textarea[data-enter-submit]") &&
      event.key === "Enter" &&
      !event.shiftKey &&
      !event.isComposing
    ) {
      event.preventDefault();
      if (event.target.form.requestSubmit) event.target.form.requestSubmit();
    }
  });
}

export function initContent() {
  initHTMXHooks();
  decryptAll(document);
  window.addEventListener("pageshow", () => decryptAll(document));
}
