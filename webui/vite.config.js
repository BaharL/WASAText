import { fileURLToPath, URL } from "node:url";
import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";

export default defineConfig(() => {
  return {
    plugins: [vue()],

    // Alias for cleaner imports
    resolve: {
      alias: {
        "@": fileURLToPath(new URL("./src", import.meta.url)),
      },
    },

    /**
     * DEV ONLY
     * - /api  -> backend (rewrite removes /api prefix)
     * - /v1   -> backend (used for media like /v1/uploads/...)
     */
    server: {
      proxy: {
        "/api": {
          target: "http://localhost:3000",
          changeOrigin: true,
          // /api/foo -> /foo
          rewrite: (path) => path.replace(/^\/api/, ""),
        },

        // Needed to keep media URLs working exactly like before
        "/v1": {
          target: "http://localhost:3000",
          changeOrigin: true,
        },
      },
    },

    /**
     * Global API base used in services/api.js
     * Must stay relative (no absolute URLs)
     */
    define: {
      "__API_URL__": JSON.stringify("/api"),
    },
  };
});
