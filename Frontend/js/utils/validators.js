const MAX_URL_LENGTH = 2_048;

export function validateUrl(value) {
  const candidate = value.trim();
  if (!candidate) return { valid: false, message: "Enter a URL to shorten." };
  if (candidate.length > MAX_URL_LENGTH) return { valid: false, message: "The URL is too long." };

  try {
    const url = new URL(candidate);
    if (!["http:", "https:"].includes(url.protocol)) {
      return { valid: false, message: "Use a URL beginning with http:// or https://." };
    }
    return { valid: true, value: url.toString(), message: "" };
  } catch {
    return { valid: false, message: "Enter a complete URL, including https://." };
  }
}
