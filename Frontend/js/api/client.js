import { CONFIG } from "../config.js";

export class ApiError extends Error {
  constructor(message, { status = 0, code = "UNKNOWN_ERROR", details = null } = {}) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.code = code;
    this.details = details;
  }
}

let authTokenProvider = () => null;

export function setAuthTokenProvider(provider) {
  authTokenProvider = typeof provider === "function" ? provider : () => null;
}

function buildUrl(path, query) {
  const url = new URL(`${CONFIG.api.baseUrl}${path}`);
  Object.entries(query || {}).forEach(([key, value]) => {
    if (value !== undefined && value !== null && value !== "") url.searchParams.set(key, value);
  });
  return url;
}

function createHeaders(customHeaders, hasBody) {
  const token = authTokenProvider();
  return {
    Accept: "application/json",
    ...(hasBody ? { "Content-Type": "application/json" } : {}),
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
    ...customHeaders,
  };
}

async function parseResponse(response) {
  if (response.status === 204) return null;
  const contentType = response.headers.get("content-type") || "";
  if (contentType.includes("application/json")) return response.json();
  const text = await response.text();
  return text ? { message: text } : null;
}

function errorFrom(response, body) {
  const fallback = response.status >= 500
    ? "The server is temporarily unavailable. Please try again."
    : "The request could not be completed.";
  return new ApiError(body?.message || body?.error || fallback, {
    status: response.status,
    code: body?.code || `HTTP_${response.status}`,
    details: body,
  });
}

function shouldRetry(error, attempt, retries, method) {
  if (attempt >= retries) return false;
  if (!["GET", "HEAD", "OPTIONS"].includes(method)) return false;
  return error.status === 0 || error.status === 408 || error.status === 429 || error.status >= 500;
}

function delay(ms) {
  return new Promise((resolve) => window.setTimeout(resolve, ms));
}

export async function request(path, options = {}) {
  const method = (options.method || "GET").toUpperCase();
  const retries = options.retries ?? CONFIG.api.retries;
  let attempt = 0;

  while (true) {
    const controller = new AbortController();
    const timeout = window.setTimeout(() => controller.abort(), options.timeoutMs ?? CONFIG.api.timeoutMs);

    try {
      const hasBody = options.body !== undefined;
      const response = await fetch(buildUrl(path, options.query), {
        method,
        mode: "cors",
        credentials: options.credentials || "omit",
        headers: createHeaders(options.headers, hasBody),
        body: hasBody ? JSON.stringify(options.body) : undefined,
        signal: controller.signal,
      });
      const body = await parseResponse(response);
      if (!response.ok) throw errorFrom(response, body);
      return body;
    } catch (error) {
      const apiError = error instanceof ApiError
        ? error
        : new ApiError(
          error.name === "AbortError" ? "The request timed out. Please try again." : "Could not reach the server. Check your connection.",
          { code: error.name === "AbortError" ? "TIMEOUT" : "NETWORK_ERROR" },
        );
      if (!shouldRetry(apiError, attempt, retries, method)) throw apiError;
      await delay(300 * (2 ** attempt));
      attempt += 1;
    } finally {
      window.clearTimeout(timeout);
    }
  }
}

export const apiClient = Object.freeze({
  get: (path, options) => request(path, { ...options, method: "GET" }),
  post: (path, body, options) => request(path, { ...options, method: "POST", body }),
  put: (path, body, options) => request(path, { ...options, method: "PUT", body }),
  patch: (path, body, options) => request(path, { ...options, method: "PATCH", body }),
  delete: (path, options) => request(path, { ...options, method: "DELETE" }),
});
