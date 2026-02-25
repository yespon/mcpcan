<template>
  <div class="flex gap-4 mb-6" :class="{ 'flex-row-reverse': isUser }">
    <div
      v-if="!isUser"
      class="w-8 h-8 rounded-full flex items-center justify-center shrink-0 bg-[var(--el-color-primary-light-9)] text-[var(--el-color-primary)]"
    >
      <el-icon><Service /></el-icon>
    </div>
    <div class="flex-1 max-w-[80%] flex" :class="{ 'justify-end': isUser }">
      <div
        class="px-4 py-3 rounded-2xl"
        :class="
          isUser
            ? 'bg-[var(--ep-bg-color)] border border-[var(--ep-border-color)] text-[var(--ep-text-color-primary)] shadow-sm'
            : 'bg-transparent text-[var(--ep-text-color-primary)]'
        "
      >
        <div v-if="isUser" class="font-medium">{{ message.content }}</div>
        <div v-else class="leading-relaxed">{{ message.content }}</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { User, Service } from '@element-plus/icons-vue'
import type { ChatMessage } from '../types'

const props = defineProps<{
  message: ChatMessage
}>()

const isUser = computed(() => props.message.role === 'user')
</script>
