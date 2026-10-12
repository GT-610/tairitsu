import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// The Go server serves both the console and the API on a fixed port. For hot
// reload, this dev server takes over the UI and proxies the API to that port.
// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    port: 3001,
    proxy: {
      '/api': {
        target: 'http://localhost:3000',
        changeOrigin: true
      }
    }
  }
})
