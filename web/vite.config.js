import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// The Go server owns port 3000 in every release and serves the built bundle
// itself. The dev server keeps the hot-reload workflow on its own port and
// proxies the API to the backend.
// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    port: 3001,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true
      }
    }
  }
})
