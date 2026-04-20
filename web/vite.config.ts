import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  build: {
    outDir: '../services/gateway/static',
    emptyOutDir: true,
  },
  server: {
    proxy: {
      '/products': 'http://localhost:8080',
      '/orders': 'http://localhost:8080',
      '/inventory': 'http://localhost:8080',
    },
  },
})
