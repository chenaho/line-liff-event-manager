import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vite.dev/config/
export default defineConfig(({ mode }) => ({
  plugins: [vue()],
  // Use '/' for development (ngrok), '/line-liff-event-manager/' for production (GitHub Pages)
  base: mode === 'production' ? '/line-liff-event-manager/' : '/',
  server: {
    allowedHosts: ['localhost', '127.0.0.1', '0.0.0.0', 'line-liff-event-manager', '2b50cb88cfb6.ngrok-free.app']
  }
}))
