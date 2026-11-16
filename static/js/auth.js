import {
  b64,
  deriveKeys,
  encryptRecord,
  exportDataKey,
  fromB64,
  generateDataKey,
  makeNonExtractable,
  newRecoveryCode,
  newSalt,
  recoveryVerifier,
  unwrapDataKey,
  unwrapFromRecovery,
  wrapDataKey,
  wrapForRecovery,
} from "./crypto.js";
import {
  clearDataKey,
  currentUserID,
  getDataKey,
  setCurrentUserID,
  storeDataKey,
} from "./keystore.js";

async function requestJSON(url, options = {}) {
  const response = await fetch(url, {
    credentials: "same-origin",
    headers: { "Content-Type": "application/json", ...(options.headers || {}) },
    ...options,
  });
  let data = {};
  try {
    data = await response.json();
  } catch (_) {
    data = {};
  }
  return { response, data };
}

function showError(targetID, message) {
  const target = document.getElementById(targetID);
  if (!target) return;
  target.innerHTML = "";
  if (!message) return;
  const banner = document.createElement("div");
  banner.className = "error-banner";
  banner.setAttribute("role", "alert");
  banner.textContent = message;
  target.appendChild(banner);
}

function setBusy(form, busy) {
  form.querySelectorAll("button, input").forEach((el) => {
    el.disabled = busy;
  });
}

function requirePassword(password) {
  if (password.length < 12) {
    throw new Error("Password must be at least 12 characters.");
  }
}

async function authParams(email) {
  const response = await fetch(`/auth/params?email=${encodeURIComponent(email)}`, {
    credentials: "same-origin",
    cache: "no-store",
  });
  if (!response.ok) throw new Error("Unable to start login.");
  return response.json();
}

async function createKeyMaterial(password) {
  requirePassword(password);
  const salt = newSalt();
  const keys = await deriveKeys(password, salt);
  const dataKey = await generateDataKey();
  const wrapped = await wrapDataKey(dataKey, keys.wrappingKey);
  const recoveryCode = newRecoveryCode();
  const recoveryWrapped = await wrapForRecovery(dataKey, recoveryCode);
  const verifier = await recoveryVerifier(dataKey);
  const rawFallback = await exportDataKey(dataKey);
  const storedKey = await makeNonExtractable(dataKey);

  return {
    keys,
    wrapped,
    recoveryCode,
    recoveryWrapped,
    verifier,
    rawFallback,
    storedKey,
  };
}

function recoveryPanel(code, redirect) {
  let panel = document.getElementById("recovery-panel");
  if (!panel) {
    panel = document.createElement("div");
    panel.id = "recovery-panel";
    panel.className = "recovery-panel";
    const form = document.querySelector(".auth-container form");
    if (form) form.insertAdjacentElement("afterend", panel);
  }
  panel.hidden = false;
  panel.innerHTML = "";

  const heading = document.createElement("h2");
  heading.textContent = "Save your recovery code";
  const copy = document.createElement("p");
  copy.textContent = "This code can unlock your data if you forget your password. It is shown only once.";
  const codeEl = document.createElement("code");
  codeEl.id = "recovery-code";
  codeEl.textContent = code;
  const button = document.createElement("button");
  button.type = "button";
  button.textContent = "I saved it — continue";
  button.addEventListener("click", () => {
    window.location.href = redirect || "/";
  });

  panel.append(heading, copy, codeEl, button);

  const form = document.querySelector(".auth-container form");
  if (form) form.hidden = true;
}

async function importLegacyRows(userID, dataKey) {
  const rowsResponse = await fetch("/auth/legacy/rows", {
    credentials: "same-origin",
    cache: "no-store",
  });
  if (!rowsResponse.ok) throw new Error("Unable to load your existing data.");
  const rows = await rowsResponse.json();

  const questions = [];
  for (const question of rows.questions || []) {
    const clientID = crypto.randomUUID();
    const aad = `khayal:v1:question:${userID}:${clientID}`;
    const ciphertext = await encryptRecord(dataKey, question.question, aad);
    questions.push({
      id: question.id,
      client_id: clientID,
      ciphertext: b64(ciphertext),
      encryption_version: 1,
    });
  }

  const khayals = [];
  for (const khayal of rows.khayals || []) {
    const clientID = crypto.randomUUID();
    const aad = `khayal:v1:khayal:${userID}:${khayal.question_id}:${clientID}`;
    const ciphertext = await encryptRecord(dataKey, khayal.khayal, aad);
    khayals.push({
      id: khayal.id,
      question_id: khayal.question_id,
      client_id: clientID,
      ciphertext: b64(ciphertext),
      encryption_version: 1,
    });
  }

  const imported = await requestJSON("/auth/legacy/import", {
    method: "POST",
    body: JSON.stringify({ questions, khayals }),
  });
  if (!imported.response.ok) throw new Error(imported.data.error || "Unable to migrate data.");
}

async function legacySetupAndMigrate(email, password, userID) {
  const material = await createKeyMaterial(password);
  const setup = await requestJSON("/auth/legacy/setup", {
    method: "POST",
    body: JSON.stringify({
      auth_key: material.keys.authKey,
      ...material.keys.parameters,
      wrapped_data_key: b64(material.wrapped.wrapped),
      wrapped_data_key_iv: b64(material.wrapped.iv),
      recovery_wrapped_data_key: b64(material.recoveryWrapped.wrapped),
      recovery_wrapped_data_key_iv: b64(material.recoveryWrapped.iv),
      recovery_verifier: material.verifier,
    }),
  });
  if (!setup.response.ok) throw new Error(setup.data.error || "Unable to set up encryption.");

  await importLegacyRows(userID, material.storedKey);
  await storeDataKey(userID, material.storedKey, material.rawFallback);
  recoveryPanel(material.recoveryCode, "/");
}

async function resumeLegacyMigration(email, password, userID, params) {
  const keys = await deriveKeys(password, fromB64(params.salt), params.kdf_iterations);
  const keyResponse = await fetch("/auth/keys", { credentials: "same-origin", cache: "no-store" });
  if (!keyResponse.ok) throw new Error("Unable to unlock your data key.");
  const keyMaterial = await keyResponse.json();
  const dataKey = await unwrapDataKey(
    fromB64(keyMaterial.wrapped_data_key),
    fromB64(keyMaterial.wrapped_data_key_iv),
    keys.wrappingKey,
    false,
  );
  await importLegacyRows(userID, dataKey);
  await storeDataKey(userID, dataKey);
  window.location.href = "/";
}

async function legacyFlow(email, password) {
  const login = await requestJSON("/login/legacy", {
    method: "POST",
    body: JSON.stringify({ email, password }),
  });
  if (!login.response.ok) throw new Error(login.data.error || "Invalid email or password.");

  const params = await authParams(email);
  if (!params.legacy && params.salt) {
    await resumeLegacyMigration(email, password, login.data.user_id, params);
    return;
  }
  await legacySetupAndMigrate(email, password, login.data.user_id);
}

async function loginFlow(form) {
  const email = form.querySelector('[name="email"]').value.trim();
  const password = form.querySelector('[name="password"]').value;
  showError("login-error", "");
  setBusy(form, true);

  try {
    const params = await authParams(email);
    if (params.legacy) {
      await legacyFlow(email, password);
      return;
    }
    if (!params.salt) throw new Error("Unable to start login.");

    const keys = await deriveKeys(password, fromB64(params.salt), params.kdf_iterations);
    const result = await requestJSON("/login", {
      method: "POST",
      body: JSON.stringify({ email, auth_key: keys.authKey }),
    });

    if (result.response.status === 409 && result.data.legacy_login_required) {
      await legacyFlow(email, password);
      return;
    }
    if (!result.response.ok) {
      throw new Error(result.data.error || "Invalid email or password.");
    }

    const dataKey = await unwrapDataKey(
      fromB64(result.data.wrapped_data_key),
      fromB64(result.data.wrapped_data_key_iv),
      keys.wrappingKey,
      false,
    );
    setCurrentUserID(result.data.user_id);
    await storeDataKey(result.data.user_id, dataKey);
    window.location.href = result.data.redirect || "/";
  } catch (err) {
    showError("login-error", err.message);
  } finally {
    setBusy(form, false);
  }
}

async function registerFlow(form) {
  const username = form.querySelector('[name="username"]').value.trim();
  const email = form.querySelector('[name="email"]').value.trim();
  const password = form.querySelector('[name="password"]').value;
  showError("register-error", "");
  setBusy(form, true);

  try {
    requirePassword(password);
    const material = await createKeyMaterial(password);
    const result = await requestJSON("/register", {
      method: "POST",
      body: JSON.stringify({
        username,
        email,
        auth_key: material.keys.authKey,
        ...material.keys.parameters,
        wrapped_data_key: b64(material.wrapped.wrapped),
        wrapped_data_key_iv: b64(material.wrapped.iv),
        recovery_wrapped_data_key: b64(material.recoveryWrapped.wrapped),
        recovery_wrapped_data_key_iv: b64(material.recoveryWrapped.iv),
        recovery_verifier: material.verifier,
      }),
    });
    if (!result.response.ok) {
      throw new Error(result.data.error || "Unable to create account.");
    }

    await storeDataKey(result.data.user_id, material.storedKey, material.rawFallback);
    recoveryPanel(material.recoveryCode, result.data.redirect || "/");
  } catch (err) {
    showError("register-error", err.message);
  } finally {
    setBusy(form, false);
  }
}

async function recoverFlow(form) {
  const email = form.querySelector('[name="email"]').value.trim();
  const code = form.querySelector('[name="recovery_code"]').value.trim();
  const password = form.querySelector('[name="new_password"]').value;
  showError("recover-error", "");
  setBusy(form, true);

  try {
    requirePassword(password);
    const response = await fetch(`/auth/recovery/params?email=${encodeURIComponent(email)}`, {
      credentials: "same-origin",
      cache: "no-store",
    });
    if (!response.ok) throw new Error("Unable to start recovery.");
    const params = await response.json();

    const dataKey = await unwrapFromRecovery(
      fromB64(params.wrapped_data_key),
      fromB64(params.wrapped_data_key_iv),
      code,
    );
    const verifier = await recoveryVerifier(dataKey);
    const salt = newSalt();
    const keys = await deriveKeys(password, salt);
    const wrapped = await wrapDataKey(dataKey, keys.wrappingKey);
    const storedKey = await makeNonExtractable(dataKey);
    const rawFallback = await exportDataKey(dataKey);

    const result = await requestJSON("/auth/recover", {
      method: "POST",
      body: JSON.stringify({
        email,
        recovery_verifier: verifier,
        auth_key: keys.authKey,
        ...keys.parameters,
        wrapped_data_key: b64(wrapped.wrapped),
        wrapped_data_key_iv: b64(wrapped.iv),
      }),
    });
    if (!result.response.ok) {
      throw new Error(result.data.error || "Invalid recovery code.");
    }

    await storeDataKey(result.data.user_id, storedKey, rawFallback);
    window.location.href = result.data.redirect || "/";
  } catch (err) {
    showError("recover-error", err.message || "Invalid recovery code.");
  } finally {
    setBusy(form, false);
  }
}

async function changePasswordFlow(form) {
  const email = form.dataset.email;
  const currentPassword = form.querySelector('[name="current_password"]').value;
  const newPassword = form.querySelector('[name="new_password"]').value;
  const status = document.getElementById("account-status");
  if (status) status.textContent = "";
  setBusy(form, true);

  try {
    requirePassword(newPassword);
    const params = await authParams(email);
    if (!params.salt) throw new Error("Unable to read account parameters.");

    const oldKeys = await deriveKeys(currentPassword, fromB64(params.salt), params.kdf_iterations);
    const keyResponse = await fetch("/auth/keys", { credentials: "same-origin", cache: "no-store" });
    if (!keyResponse.ok) throw new Error("Unable to unlock your data key.");
    const keyMaterial = await keyResponse.json();
    let dataKey;
    try {
      dataKey = await unwrapDataKey(
        fromB64(keyMaterial.wrapped_data_key),
        fromB64(keyMaterial.wrapped_data_key_iv),
        oldKeys.wrappingKey,
        true,
      );
    } catch (_) {
      throw new Error("Current password is incorrect.");
    }

    const salt = newSalt();
    const newKeys = await deriveKeys(newPassword, salt);
    const wrapped = await wrapDataKey(dataKey, newKeys.wrappingKey);
    const storedKey = await makeNonExtractable(dataKey);
    const rawFallback = await exportDataKey(dataKey);

    const result = await requestJSON("/auth/password", {
      method: "POST",
      body: JSON.stringify({
        current_auth_key: oldKeys.authKey,
        auth_key: newKeys.authKey,
        ...newKeys.parameters,
        wrapped_data_key: b64(wrapped.wrapped),
        wrapped_data_key_iv: b64(wrapped.iv),
      }),
    });
    if (!result.response.ok) {
      throw new Error(result.data.error || "Unable to change password.");
    }

    const activeUserID = document.body.dataset.userId || currentUserID();
    await storeDataKey(activeUserID, storedKey, rawFallback);
    if (status) status.textContent = "Password changed.";
    form.reset();
  } catch (err) {
    if (status) status.textContent = err.message || "Unable to change password.";
  } finally {
    setBusy(form, false);
  }
}

async function logoutFlow() {
  try {
    await fetch("/logout", { method: "POST", credentials: "same-origin" });
  } finally {
    await clearDataKey();
    window.location.href = "/login";
  }
}

export function initAuth() {
  const loginForm = document.querySelector('[data-auth-form="login"]');
  if (loginForm) {
    loginForm.addEventListener("submit", (event) => {
      event.preventDefault();
      loginFlow(loginForm);
    });
  }

  const registerForm = document.querySelector('[data-auth-form="register"]');
  if (registerForm) {
    registerForm.addEventListener("submit", (event) => {
      event.preventDefault();
      registerFlow(registerForm);
    });
  }

  const recoverForm = document.querySelector('[data-auth-form="recover"]');
  if (recoverForm) {
    recoverForm.addEventListener("submit", (event) => {
      event.preventDefault();
      recoverFlow(recoverForm);
    });
  }

  const changePasswordForm = document.querySelector('[data-auth-form="change-password"]');
  if (changePasswordForm) {
    changePasswordForm.addEventListener("submit", (event) => {
      event.preventDefault();
      changePasswordFlow(changePasswordForm);
    });
  }

  const logoutButton = document.getElementById("logout-button");
  if (logoutButton) logoutButton.addEventListener("click", logoutFlow);
}
