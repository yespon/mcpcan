<template>
  <div class="flex gap-4 mb-6" :class="{ 'flex-row-reverse': isUser }">
    <div
      v-if="!isUser"
      class="w-10 h-10 rounded-full flex items-center justify-center shrink-0 bg-[var(--el-color-primary-light-9)] text-[var(--el-color-primary)] shadow-sm"
    >
      <el-icon :size="20"><Service /></el-icon>
    </div>
    <div v-else class="w-10 h-10 rounded-full flex items-center justify-center shrink-0 shadow-sm">
      <el-avatar v-if="userInfo.avatar" :src="userInfo.avatar" fit="cover" :size="28" />
      <el-avatar v-else :icon="UserFilled" :size="28" />
    </div>

    <div class="flex-1 max-w-[85%] flex flex-col" :class="{ 'items-end': isUser }">
      <!-- Header Info -->
      <div class="text-xs text-[var(--ep-text-color-secondary)] mb-1 px-1">
        {{ new Date(message.timestamp).toLocaleTimeString() }}
      </div>

      <div
        class="px-5 py-3 rounded-2xl relative group"
        :class="
          isUser
            ? 'bg-[var(--ep-bg-color)] border border-[var(--ep-border-color)] text-[var(--ep-text-color-primary)] shadow-sm rounded-tr-sm'
            : 'bg-transparent text-[var(--ep-text-color-primary)] rounded-tl-sm'
        "
      >
        <div class="whitespace-pre-wrap leading-relaxed break-words">
          {{ message.content }}
          <span v-if="message.isStreaming && !message.content" class="italic opacity-50">{{
            t('aiChat.thinking')
          }}</span>
        </div>

        <!-- Tool Calls Display -->
        <div
          v-if="message.tools && message.tools.length > 0"
          class="mt-3 pt-3 border-t border-[var(--ep-border-color)] text-xs"
        >
          <div v-for="(tool, idx) in message.tools" :key="idx" class="mb-2 last:mb-0">
            <div class="font-mono font-bold text-[var(--el-color-primary)] mb-1">
              🔨 {{ tool.name }}
            </div>
            <div
              v-if="tool.args && tool.args !== '?'"
              class="font-mono bg-[var(--ep-bg-color)] p-2 rounded mb-1 text-[var(--ep-text-color-secondary)] break-all border border-[var(--ep-border-color)]"
            >
              {{ tool.args }}
            </div>
            <div
              v-if="tool.result"
              class="font-mono bg-[var(--ep-bg-color)] p-2 rounded text-[var(--ep-text-color-regular)] break-all border-l-2 border-[var(--el-color-success)] border border-[var(--ep-border-color)] border-l-0 border-t-0 border-r-0 border-b-0"
            >
              -> {{ tool.result }}
            </div>
          </div>
        </div>

        <!-- Token Usage -->
        <div
          v-if="message.usage"
          class="text-[10px] text-[var(--ep-text-color-secondary)] mt-2 text-right opacity-0 group-hover:opacity-100 transition-opacity"
        >
          {{ t('aiChat.tokens') }}: {{ message.usage.totalTokens }}
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { User, Service, UserFilled } from '@element-plus/icons-vue'
import type { ChatMessage } from '../types'
import { useI18n } from 'vue-i18n'
import { useUserStore } from '@/stores'
import { storeToRefs } from 'pinia'
import { computed } from 'vue'

const { t } = useI18n()
const { userInfo } = storeToRefs(useUserStore())

const props = defineProps<{
  message: ChatMessage
}>()

const isUser = computed(() => props.message.role === 'user')
</script>
