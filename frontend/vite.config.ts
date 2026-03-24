import { fileURLToPath, URL } from 'node:url'
import path from 'node:path'
import fs from 'node:fs'
// import { defineConfig } from 'vite'
import { defineConfig } from 'vitest/config'
import vue from '@vitejs/plugin-vue'
import Unocss from 'unocss/vite'
import AutoImport from 'unplugin-auto-import/vite' // 自动根据需求导入vue的相关API如；ref、reactive等

const packageJson = JSON.parse(fs.readFileSync(path.resolve(__dirname, './package.json'), 'utf-8')) as {
  version?: string
}

// https://vite.dev/config/
export default defineConfig({
  base: './', // 适配 自定义部署路径的情况
  define: {
    __APP_VERSION__: JSON.stringify(packageJson.version || 'dev'),
  },
  server: {
    open: false,
    host: '0.0.0.0',
    port: 3000,
    watch: {
      // macOS + Docker Desktop volume 挂载时 inotify 失效，需要开启轮询
      usePolling: true,
      interval: 500,
    },
    proxy: {
      '/api/authz': {
        target: 'http://mcp-entry-svc',
        changeOrigin: true,
      },
      '/api': {
        target: 'http://mcp-entry-svc',
        changeOrigin: true,
      },
      '/mcp-gateway': {
        target: 'http://mcp-entry-svc',
        changeOrigin: true,
      },
    },
  },
  css: {
    preprocessorOptions: {
      scss: {
        additionalData: `@use "~/styles/element/index.scss" as *;`,
        // api: 'modern-compiler',
      },
    },
  },
  plugins: [
    vue(),
    Unocss(),
    AutoImport({
      // 1. 自动导入 Vue 的 Composition API
      imports: [
        'vue',
        '@vueuse/core',
        'pinia',
        'vue-router',
        'vue-i18n',
        {
          'element-plus': ['ElMessage', 'ElMessageBox', 'ElNotification'],
        },
      ],
      // 禁用默认导入（若有冲突）
      defaultExportByFilename: false,
      // 会自动导入 vue 中的 ref、reactive、onMounted 等所有 API
      // 2. 自动导入 composables 目录下的文件
      dirs: ['src/composables'],
      // 3. 生成声明文件（可选，解决 TypeScript 类型提示问题）
      dts: 'src/auto-imports.d.ts', // 生成的声明文件路径，需手动创建空文件
      // 4. 全局注册的组件（可选，若需自动导入组件）
      // components: ['vue'],
    }),
  ],

  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
      '~/': `${path.resolve(__dirname, 'src')}/`,
    },
  },
  test: {
    environment: 'happy-dom',
    globals: true,
    coverage: {
      provider: 'v8',
      reporter: ['text', 'html'],
    },
  },
})
