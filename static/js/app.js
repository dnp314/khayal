import { initAuth } from "./auth.js";
import { initContent } from "./content.js";

function start() {
  initAuth();
  initContent();
}

if (document.readyState === "loading") {
  document.addEventListener("DOMContentLoaded", start, { once: true });
} else {
  start();
}
