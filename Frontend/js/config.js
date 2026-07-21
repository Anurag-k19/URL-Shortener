const metaBaseUrl = document.querySelector('meta[name="api-base-url"]')?.content?.trim();

export const CONFIG = Object.freeze({
  appName: "Snip",
  api: Object.freeze({
    baseUrl: (metaBaseUrl || "http://localhost:8080").replace(/\/$/, ""),
    timeoutMs: 10_000,
    retries: 2,
    endpoints: Object.freeze({
      shorten: "/shorten",
      redirect: "/redirect",
    }),
  }),
});
