import { apiClient } from "./client.js";
import { CONFIG } from "../config.js";

export const authService = Object.freeze({
  // Sign-in is a full-page redirect to the backend, which sends the browser on to Google.
  loginUrl: () => `${CONFIG.api.baseUrl}${CONFIG.api.endpoints.login}`,

  // Returns { id, email, name } for a signed-in user, or null for a guest.
  async getCurrentUser() {
    try {
      return await apiClient.get(CONFIG.api.endpoints.me, { retries: 0 });
    } catch (error) {
      if (error.status === 401) return null;
      throw error;
    }
  },

  logout: () => apiClient.post(CONFIG.api.endpoints.logout, undefined, { retries: 0 }),
});
