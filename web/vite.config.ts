import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import VueI18nPlugin from '@intlify/unplugin-vue-i18n/vite'

export default defineConfig({
  plugins: [
    vue(),
    tailwindcss(),
    VueI18nPlugin({}),
  ],
  server: {
    proxy: {
      '/api': 'http://localhost:8000',
    },
  },
})
