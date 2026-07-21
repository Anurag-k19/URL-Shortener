export function renderNavbar(container) {
  container.innerHTML = `
    <nav class="navbar" aria-label="Main navigation">
      <a class="brand" href="./" aria-label="Snip home">
        <span class="brand__mark" aria-hidden="true">↗</span>
        <span>snip.</span>
      </a>
      <div class="nav__links">
        <a class="nav__link" href="#how-it-works">How it works</a>
        <span class="nav__status">API ready</span>
      </div>
    </nav>
  `;
}
