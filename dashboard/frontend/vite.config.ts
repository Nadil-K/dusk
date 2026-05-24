import { defineConfig } from 'vite'
import { svelte, vitePreprocess } from '@sveltejs/vite-plugin-svelte'

export default defineConfig({
  plugins: [svelte({ preprocess: vitePreprocess() })],
  server: {
    proxy: {
      '/api': 'http://localhost:9001',
    },
  },
  build: {
    outDir: 'dist',
    emptyOutDir: true,
  },
})
