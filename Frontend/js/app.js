import { authService } from "./api/auth.service.js";
import { renderAuthStatus, renderNavbar } from "./components/navbar.js";
import { createToastManager } from "./components/toast.js";
import { renderHome } from "./pages/home.js";
import { getErrorMessage } from "./utils/helpers.js";

async function loadAuthStatus(container, toast) {
  const render = (user) => renderAuthStatus(container, {
    user,
    loginUrl: authService.loginUrl(),
    onLogout: async () => {
      try {
        await authService.logout();
        render(null);
        toast.success("Signed out.");
      } catch (error) {
        toast.error(getErrorMessage(error));
      }
    },
  });

  try {
    render(await authService.getCurrentUser());
  } catch {
    render(null);
  }
}

function bootstrap() {
  const app = document.querySelector("#app");
  const navbarRoot = document.querySelector("#navbar-root");
  const toastRoot = document.querySelector("#toast-root");

  if (!app || !navbarRoot || !toastRoot) throw new Error("Required application root is missing.");

  const toast = createToastManager(toastRoot);
  renderNavbar(navbarRoot);
  renderHome(app, toast);
  loadAuthStatus(navbarRoot.querySelector("#nav-auth"), toast);
}

bootstrap();
