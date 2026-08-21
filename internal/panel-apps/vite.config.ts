import { defineConfig } from "vite"
import viteReact from "@vitejs/plugin-react"
import viteTsConfigPaths from "vite-tsconfig-paths"
import tailwindcss from "@tailwindcss/vite"

const apiPort = process.env.CHAUFFEUR_WEBUI_API_PORT || "3083"

const config = defineConfig({
  plugins: [
    viteTsConfigPaths({
      projects: ["./tsconfig.json"],
    }),
    tailwindcss(),
    viteReact(),
  ],
  ssr: {
    noExternal: ["lucide-react"],
  },
  server: {
    port: 5173,
    strictPort: true,
    proxy: {
      "/api": {
        target: `http://localhost:${apiPort}`,
        changeOrigin: true,
      },
    },
  },
})

export default config
