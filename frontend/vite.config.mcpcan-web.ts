import { fileURLToPath, URL } from 'node:url'
import path from 'node:path'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import Unocss from 'unocss/vite'
import AutoImport from 'unplugin-auto-import/vite'

// 专门用于 /mcpcan-web 路径部署的配置文件
// 使用绝对路径 base: '/mcpcan-web/'，确保所有静态资源路径都包含 /mcpcan-web 前缀
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

