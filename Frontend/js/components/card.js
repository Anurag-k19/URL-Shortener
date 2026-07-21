export function renderFeatureCards() {
  const features = [
    ["01", "Paste your link", "Drop in any valid web address. Validation happens before it reaches the API."],
    ["02", "We make it tiny", "Your Go backend creates and owns the short link as the single source of truth."],
    ["03", "Share anywhere", "Copy the clean result and use it in messages, campaigns, or social posts."],
  ];

  return features.map(([number, title, copy]) => `
    <article class="feature">
      <span class="feature__number">${number}</span>
      <h2 class="feature__title">${title}</h2>
      <p class="feature__copy">${copy}</p>
    </article>
  `).join("");
}
