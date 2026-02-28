<template>
  <div
    class="h-full flex flex-col bg-[var(--ep-bg-color)] dark:bg-[var(--ep-bg-color-hover)] border-l border-[var(--ep-border-color)] dark:border-l-[rgba(255,255,255,0.15)] transition-all duration-300"
  >
    <!-- Header -->
    <div
      class="px-4 py-3 border-b border-[var(--ep-border-color)] dark:border-b-[rgba(255,255,255,0.08)] flex items-center justify-between shrink-0"
    >
      <span class="text-sm font-medium text-[var(--ep-text-color-primary)]">{{
        t('aiChat.logs')
      }}</span>
      <el-button
        link
        class="!text-[var(--ep-text-color-secondary)] hover:!text-[var(--ep-color-primary)]"
        @click="$emit('close')"
      >
        <el-icon><Expand /></el-icon>
      </el-button>
    </div>

    <!-- Toolbar -->
    <div
      class="px-4 py-2 border-b border-[var(--ep-border-color)] dark:border-b-[rgba(255,255,255,0.08)] flex items-center gap-2 shrink-0 bg-[var(--ep-bg-color-overlay)]/50"
    >
      <el-input
        v-model="searchQuery"
        :placeholder="t('aiChat.filterLogs')"
        prefix-icon="Search"
        class="w-full !mr-0"
        size="small"
        clearable
      />
      <div class="flex items-center shrink-0">
        <el-tooltip :content="t('aiChat.clearLogs')" placement="top">
          <el-button
            link
            class="!text-[var(--ep-text-color-secondary)] hover:!text-[var(--ep-color-danger)]"
          >
            <el-icon><Delete /></el-icon>
          </el-button>
        </el-tooltip>
      </div>
    </div>

    <!-- Content Area -->
    <div class="flex-1 overflow-y-auto p-4 custom-scrollbar">
      <div
        class="flex flex-col items-center justify-center h-full text-[var(--ep-text-color-secondary)] opacity-60"
      >
        <el-icon :size="48" class="mb-2"><Document /></el-icon>
        <p class="text-sm font-medium">{{ t('aiChat.noLogsAvailable') }}</p>
        <p class="text-xs mt-1 text-[var(--ep-text-color-placeholder)]">
          {{ t('aiChat.systemEventsHere') }}
        </p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Search, Delete, Expand, Document } from '@element-plus/icons-vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const searchQuery = ref('')
const emit = defineEmits(['close'])
</script>

<style scoped>
.custom-sidebar-input :deep(.el-input__wrapper) {
  background-color: transparent;
  box-shadow: none;
  padding-left: 0;
}
.custom-sidebar-input :deep(.el-input__wrapper.is-focus) {
  box-shadow: none;
}
.custom-sidebar-input :deep(.el-input__inner) {
  font-size: 12px;
}

/* Custom Scrollbar */
.custom-scrollbar::-webkit-scrollbar {
  width: 4px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background-color: var(--ep-border-color);
  border-radius: 4px;
}
.custom-scrollbar:hover::-webkit-scrollbar-thumb {
  background-color: var(--ep-text-color-placeholder);
}
</style>
