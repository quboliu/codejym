import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'

const API_TARGET = process.env.VITE_DEV_API ?? 'http://localhost:8080'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    proxy: {
      '/api': API_TARGET,
      '/healthz': API_TARGET,
    },
  },
})
