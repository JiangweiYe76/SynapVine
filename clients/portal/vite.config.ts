import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import VueI18nPlugin from '@intlify/unplugin-vue-i18n/vite'
import path from 'node:path'

export default defineConfig({
  plugins: [
    vue(),
    tailwindcss(),
    VueI18nPlugin({}),
  ],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    proxy: {
      '/api': 'http://localhost:8000',
    },
  },
  build: {
    rolldownOptions: {
      output: {
        codeSplitting: {
          groups: [
            {
              // Keep the heavy 3D rendering stack in its own chunk with a
              // stable hash so it stays cached across app-code releases.
              name: 'three-vendor',
              test: /[\\/]node_modules[\\/](three|three-spritetext|three-render-objects|3d-force-graph|force-graph|d3-force-3d|d3-quadtree|d3-dispatch|d3-timer|internmap)[\\/]/,
              priority: 10,
            },
          ],
        },
      },
    },
  },
})
