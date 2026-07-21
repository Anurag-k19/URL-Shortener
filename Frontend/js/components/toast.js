const DEFAULT_DURATION = 3_500;

export function createToastManager(container) {
  function show(message, type = "info", duration = DEFAULT_DURATION) {
    const toast = document.createElement("div");
    toast.className = `toast${type === "error" ? " toast--error" : ""}`;
    toast.setAttribute("role", type === "error" ? "alert" : "status");
    toast.textContent = message;
    container.append(toast);

    window.setTimeout(() => toast.remove(), duration);
  }

  return Object.freeze({
    success: (message) => show(message),
    error: (message) => show(message, "error", 5_000),
  });
}
