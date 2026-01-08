// webui/src/services/axios.js
import axios from "axios";

/**
 * Axios instance used by the whole frontend.
 *
 * Responsibilities:
 * - Use the correct baseURL (dev proxy or nginx /api in production)
 * - Add a global response interceptor to handle "invalid session" (401)
 *   in ONE place instead of duplicating logic in every view.
 */
const instance = axios.create({
  baseURL: `${__API_URL__}/v1`,
  timeout: 1000 * 5,
});

/**
 * Clear local session info stored in localStorage.
 * NOTE: Keep keys aligned with services/api.js logout() implementation.
 */
function clearLocalSession() {
  localStorage.removeItem("token");
  localStorage.removeItem("username");
  localStorage.removeItem("photoUrl");
  localStorage.removeItem("theme");
  localStorage.removeItem("lastChat");
}

/**
 * Redirect to login preserving "next" so after login you can return.
 * Uses window.location to avoid circular imports with the Vue router.
 */
function redirectToLogin() {
  const currentPath = window.location.pathname + window.location.search;

  // Avoid infinite redirect loop if we are already on /login
  if (window.location.pathname.startsWith("/login")) return;

  window.location.href = `/login?next=${encodeURIComponent(currentPath)}`;
}

/**
 * Global response interceptor:
 * - If backend returns 401 => session is invalid/expired => logout client-side and redirect.
 * - Otherwise just pass the error through.
 */
instance.interceptors.response.use(
  (response) => response,
  (error) => {
    const status = error?.response?.status;
    const message = error?.response?.data?.message;

    // Handle invalid session (stale token / expired) centrally.
    if (status === 401) {
      clearLocalSession();

      // Optional: notify UI (navbar/profile) to refresh immediately
      window.dispatchEvent(new Event("profile-updated"));

      // Redirect to login
      redirectToLogin();
    }

    // Some backends may return 200 with message "invalid session" (rare),
    // but in your case backend now correctly returns 401, so status check is enough.
    // If you want extra safety, uncomment:
    // if (message === "invalid or expired session") { ... }

    return Promise.reject(error);
  }
);

export default instance;
