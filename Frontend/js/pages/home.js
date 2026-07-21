import { createShortUrl } from "../api/url.service.js";
import { renderFeatureCards } from "../components/card.js";
import { setButtonLoading } from "../components/loader.js";
import { copyText } from "../utils/clipboard.js";
import { getErrorMessage } from "../utils/helpers.js";
import { validateUrl } from "../utils/validators.js";
import { store } from "../state/store.js";

const linkIcon = `
  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
    <path d="M10 13a5 5 0 0 0 7.5.5l2-2a5 5 0 0 0-7-7l-1.15 1.15"/>
    <path d="M14 11a5 5 0 0 0-7.5-.5l-2 2a5 5 0 0 0 7 7l1.15-1.15"/>
  </svg>`;

const checkIcon = `
  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
    <path d="m5 12 4 4L19 6"/>
  </svg>`;

export function renderHome(container, toast) {
  container.innerHTML = `
    <section class="hero" aria-labelledby="hero-title">
      <p class="eyebrow"><span class="eyebrow__spark">✦</span> Simple. Fast. Reliable.</p>
      <h1 class="hero__title" id="hero-title">Big links.<br><em>Small footprint.</em></h1>
      <p class="hero__subtitle">Turn unwieldy URLs into memorable, shareable links—without the clutter.</p>

      <div class="shortener">
        <form class="shortener__form" id="shorten-form" novalidate>
          <div class="field">
            <span class="field__icon">${linkIcon}</span>
            <label class="sr-only" for="long-url">Long URL</label>
            <input class="field__input" id="long-url" name="url" type="url" inputmode="url" autocomplete="url" maxlength="2048" placeholder="Paste your long URL here…" aria-describedby="url-error">
          </div>
          <button class="button" id="shorten-button" type="submit">
            <span>Shorten URL</span><span aria-hidden="true">→</span>
          </button>
        </form>
        <p class="form__error" id="url-error" role="alert"></p>
      </div>

      <section class="result" id="result" aria-label="Shortened URL" hidden>
        <span class="result__check" aria-hidden="true">${checkIcon}</span>
        <div class="result__body">
          <p class="result__label">Your short link</p>
          <a class="result__link" id="short-link" href="#" target="_blank" rel="noopener noreferrer"></a>
        </div>
        <button class="button button--copy" id="copy-button" type="button">Copy link</button>
      </section>

      <div class="trust-row" aria-label="Product benefits">
        <span>${checkIcon} No sign-up needed</span>
        <span>${checkIcon} Instant results</span>
        <span>${checkIcon} Backend powered</span>
      </div>
    </section>

    <section class="features" id="how-it-works" aria-label="How it works">
      ${renderFeatureCards()}
    </section>
  `;

  const form = container.querySelector("#shorten-form");
  const input = container.querySelector("#long-url");
  const error = container.querySelector("#url-error");
  const submitButton = container.querySelector("#shorten-button");
  const result = container.querySelector("#result");
  const shortLink = container.querySelector("#short-link");
  const copyButton = container.querySelector("#copy-button");

  function showValidation(message) {
    error.textContent = message;
    input.setAttribute("aria-invalid", String(Boolean(message)));
  }

  input.addEventListener("input", () => showValidation(""));

  form.addEventListener("submit", async (event) => {
    event.preventDefault();
    const validation = validateUrl(input.value);
    if (!validation.valid) {
      showValidation(validation.message);
      input.focus();
      return;
    }

    showValidation("");
    store.setState({ status: "loading", error: null });
    setButtonLoading(submitButton, true);

    try {
      const shortened = await createShortUrl(validation.value);
      store.setState({ status: "success", result: shortened, error: null });
      shortLink.href = shortened.shortUrl;
      shortLink.textContent = shortened.shortUrl;
      result.hidden = false;
      toast.success("Your short link is ready.");
    } catch (requestError) {
      const message = getErrorMessage(requestError);
      store.setState({ status: "error", result: null, error: message });
      result.hidden = true;
      showValidation(message);
      toast.error(message);
    } finally {
      setButtonLoading(submitButton, false);
    }
  });

  copyButton.addEventListener("click", async () => {
    const value = store.getState().result?.shortUrl;
    if (!value) return;
    try {
      await copyText(value);
      copyButton.textContent = "Copied!";
      toast.success("Copied to clipboard.");
      window.setTimeout(() => { copyButton.textContent = "Copy link"; }, 1_800);
    } catch {
      toast.error("Could not copy the link. Please copy it manually.");
    }
  });
}
