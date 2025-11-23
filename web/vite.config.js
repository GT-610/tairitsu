import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true
      }
    }
  },
  build: {
    // 确保更好的浏览器兼容性
    target: ['es2020', 'edge88', 'firefox78', 'chrome87', 'safari14'],
    // 启用预构建依赖项
    commonjsOptions: {
      transformMixedEsModules: true
    },
    // 优化静态资源
    assetsInlineLimit: 4096,
    // 启用源代码映射便于调试
    sourcemap: false
  },
  // 优化依赖解析
  resolve: {
    alias: {
      '@': '/src'
    }
  }
})