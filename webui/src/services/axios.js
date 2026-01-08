// webui/src/services/axios.js
import axios from "axios";

const instance = axios.create({
  baseURL: `${__API_URL__}/v1`,
  timeout: 1000 * 5,
});

function clearLocalSession() {
  localStorage.removeItem("token");
  localStorage.removeItem("username");
  localStorage.removeItem("photoUrl");
  localStorage.removeItem("theme");
  localStorage.removeItem("lastChat");

  window.dispatchEvent(new Event("profile-updated"));
}

function redirectToLogin() {
  // Avoid infinite redirect loop if we are already on /login
  if (window.location.pathname.startsWith("/login")) return;

  const currentPath = window.location.pathname + window.location.search;
  window.location.href = `/login?next=${encodeURIComponent(currentPath)}`;
}

instance.interceptors.response.use(
  (response) => response,
  (error) => {
    const status = error?.response?.status;

    if (status === 401) {
      clearLocalSession();
      redirectToLogin();
    }

    return Promise.reject(error);
  }
);

export default instance;
