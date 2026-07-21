import { apiClient } from "./client.js";
import { CONFIG } from "../config.js";

function normalizeShortenResponse(data) {
  const shortUrl = data?.short_url || data?.shortUrl || data?.url;
  const code = data?.code || data?.short_code || data?.shortCode;

  if (!shortUrl && !code) throw new Error("The server response did not include a short URL or code.");

  return {
    id: data?.id ?? code ?? shortUrl,
    code: code ?? null,
    shortUrl: shortUrl || getRedirectUrl(code),
    originalUrl: data?.original_url || data?.originalUrl || null,
    createdAt: data?.created_at || data?.createdAt || null,
  };
}

export async function createShortUrl(longUrl, options = {}) {
  const data = await apiClient.post(
    CONFIG.api.endpoints.shorten,
    {
      url: longUrl,
      ...(options.customAlias ? { custom_alias: options.customAlias } : {}),
    },
    { retries: 0 },
  );
  return normalizeShortenResponse(data);
}

export function getRedirectUrl(code) {
  const url = new URL(`${CONFIG.api.baseUrl}${CONFIG.api.endpoints.redirect}`);
  url.searchParams.set("code", code);
  return url.toString();
}
