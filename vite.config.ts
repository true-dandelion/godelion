import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'

// https://vite.dev/config/
export default defineConfig({
  build: {
    outDir: 'godelion_public',
    sourcemap: 'hidden',
  },
  plugins: [
    vue(),
  ],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'), // ✅ 定义 @ = src
    },
  },
  server: {
    watch: {
      ignored: ['**/node_modules/**', '**/.git/**', '**/.pnpm-store/**']
    },
    proxy: {
      '/sys': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  }
})
