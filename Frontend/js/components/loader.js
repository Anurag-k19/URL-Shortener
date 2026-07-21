export function setButtonLoading(button, isLoading, idleLabel = "Shorten URL") {
  button.disabled = isLoading;
  button.setAttribute("aria-busy", String(isLoading));
  button.innerHTML = isLoading
    ? '<span class="spinner" aria-hidden="true"></span><span>Shortening…</span>'
    : `<span>${idleLabel}</span><span aria-hidden="true">→</span>`;
}
