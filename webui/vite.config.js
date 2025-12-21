import { fileURLToPath, URL } from "node:url";

import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";

export default defineConfig(() => {
  return {
    plugins: [vue()],
    resolve: {
      alias: {
        "@": fileURLToPath(new URL("./src", import.meta.url)),
      },
    },

    // ✅ DEV ONLY: proxy /api -> backend
    server: {
      proxy: {
        "/api": {
          target: "http://localhost:3000",
          changeOrigin: true,
          // /api/foo -> /foo
          rewrite: (path) => path.replace(/^\/api/, ""),
        },
      },
    },

    define: {
      // ✅ IMPORTANT: niente URL assoluto nel codice applicativo
      "__API_URL__": JSON.stringify("/api"),
    },
  };
});
