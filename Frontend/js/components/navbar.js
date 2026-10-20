export function renderNavbar(container) {
  container.innerHTML = `
    <nav class="navbar" aria-label="Main navigation">
      <a class="brand" href="./" aria-label="Snip home">
        <span class="brand__mark" aria-hidden="true">↗</span>
        <span>snip.</span>
      </a>
      <div class="nav__links">
        <a class="nav__link" href="#how-it-works">How it works</a>
        <div class="nav__auth" id="nav-auth"></div>
      </div>
    </nav>
  `;
}

// Shows "Sign in with Google" for guests, or the user's name and a sign-out button.
// User data is set with textContent, never innerHTML, so a name can't inject HTML.
export function renderAuthStatus(container, { user, loginUrl, onLogout }) {
  container.replaceChildren();

  if (!user) {
    const signIn = document.createElement("a");
    signIn.className = "button button--nav";
    signIn.href = loginUrl;
    signIn.textContent = "Sign in with Google";
    container.append(signIn);
    return;
  }

  const status = document.createElement("span");
  status.className = "nav__status";
  status.textContent = user.name || user.email;
  status.title = user.email;

  const signOut = document.createElement("button");
  signOut.className = "button button--nav button--ghost";
  signOut.type = "button";
  signOut.textContent = "Sign out";
  signOut.addEventListener("click", onLogout);

  container.append(status, signOut);
}
