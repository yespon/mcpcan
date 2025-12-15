<template>
  <el-watermark
    :font="{ color: '12px' }"
    :content="true ? '' : 'qm-mcp-box'"
    :z-index="9999"
    class="wh-full"
  >
    <router-view />
  </el-watermark>
</template>

<script setup lang="ts">
import { Storage } from '@/utils/storage'
import { initThemeInfo, initUseI18n } from '@/utils/system'
// import { useSystemStoreHook } from '@/stores/modules/system-store'
// const systemStore = useSystemStoreHook()

/**
 * Handle init dark theme
 */
const init = async () => {
  initUseI18n()
  await initThemeInfo()
  const root = document.documentElement
  root.className = Storage.get('theme') || 'dark'
}
onMounted(init)
</script>

<style scoped>
.wh-full {
  width: 100%;
  height: 100vh;
}
</style>
