import { fileURLToPath, URL } from 'node:url'
import path from 'node:path'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import Unocss from 'unocss/vite'
import AutoImport from 'unplugin-auto-import/vite'

// 专门用于 /mcpcan-web 路径部署的配置文件; 部署位置有自定义路径的情况，base：'./' || '/自定义路径/'
// https://vite.dev/config/
export default defineConfig({
  base: '/mcpcan-web/',
  server: {
    open: true,
    host: '0.0.0.0',
    proxy: {
      '/api/authz': {
        target: 'https://mcp-dev.itqm.com',
        changeOrigin: true,
      },
      '/api/': {
        target: 'https://mcp-dev.itqm.com',
        changeOrigin: true,
      },
    },
  },
  css: {
    preprocessorOptions: {
      scss: {
        additionalData: `@use "~/styles/element/index.scss" as *;`,
      },
    },
  },
  plugins: [
    vue(),
    Unocss(),
    AutoImport({
      imports: ['vue', '@vueuse/core', 'pinia', 'vue-router', 'vue-i18n'],
      defaultExportByFilename: false,
      dts: 'src/auto-imports.d.ts',
    }),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
      '~/': `${path.resolve(__dirname, 'src')}/`,
    },
  },
})
