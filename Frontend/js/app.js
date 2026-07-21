import { renderNavbar } from "./components/navbar.js";
import { createToastManager } from "./components/toast.js";
import { renderHome } from "./pages/home.js";

function bootstrap() {
  const app = document.querySelector("#app");
  const navbarRoot = document.querySelector("#navbar-root");
  const toastRoot = document.querySelector("#toast-root");

  if (!app || !navbarRoot || !toastRoot) throw new Error("Required application root is missing.");

  renderNavbar(navbarRoot);
  renderHome(app, createToastManager(toastRoot));
}

bootstrap();
