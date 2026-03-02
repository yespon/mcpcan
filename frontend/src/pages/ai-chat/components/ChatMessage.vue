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
        <!-- Attachments Preview -->
        <div
          v-if="message.attachments && message.attachments.length > 0"
          class="mb-2 flex flex-wrap gap-2"
        >
          <template v-for="(att, idx) in message.attachments" :key="idx">
            <el-image
              v-if="att.type === 'image'"
              :src="att.url"
              :preview-src-list="imagePreviewList"
              :initial-index="idx"
              fit="cover"
              class="w-40 h-40 rounded-lg border border-[var(--ep-border-color)] cursor-pointer object-cover"
              :alt="att.name"
            />
            <a
              v-else
              :href="att.url"
              target="_blank"
              class="flex items-center gap-1 px-3 py-2 rounded-lg border border-[var(--ep-border-color)] bg-[var(--ep-bg-color-page)] text-xs text-[var(--ep-text-color-regular)] hover:text-[var(--el-color-primary)] transition-colors"
            >
              <el-icon><Document /></el-icon>
              {{ att.name || 'File' }}
            </a>
          </template>
        </div>

        <div class="leading-relaxed break-words markdown-body">
          <div v-html="renderedContent"></div>
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
import { User, Service, UserFilled, Document } from '@element-plus/icons-vue'
import type { ChatMessage } from '../types'
import { useI18n } from 'vue-i18n'
import { useUserStore } from '@/stores'
import { storeToRefs } from 'pinia'
import { computed } from 'vue'
import MarkdownIt from 'markdown-it'

const { t } = useI18n()
const { userInfo } = storeToRefs(useUserStore())

const props = defineProps<{
  message: ChatMessage
}>()

const isUser = computed(() => props.message.role === 'user')

const imagePreviewList = computed(() => {
  if (!props.message.attachments) return []
  return props.message.attachments.filter((a) => a.type === 'image').map((a) => a.url)
})

const md = new MarkdownIt({
  html: false,
  linkify: true,
  typographer: true,
})

const renderedContent = computed(() => {
  return md.render(props.message.content || '')
})
</script>

<style scoped>
:deep(.markdown-body) {
  font-size: 14px;

  p {
    margin: 0;
    line-height: 1.6;
    &:not(:last-child) {
      margin-bottom: 0.5em;
    }
  }

  img {
    max-width: 100%;
    border-radius: 8px;
    margin-top: 0.5rem;
    display: block;
  }

  pre {
    background-color: var(--ep-fill-color-light);
    padding: 0.75rem;
    border-radius: 6px;
    overflow-x: auto;
    margin: 0.5rem 0;
    font-family:
      ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono', 'Courier New',
      monospace;
  }

  code {
    font-family:
      ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono', 'Courier New',
      monospace;
    background-color: var(--ep-fill-color-light);
    padding: 0.1rem 0.3rem;
    border-radius: 4px;
    font-size: 0.9em;
  }

  pre code {
    background-color: transparent;
    padding: 0;
    font-size: 1em;
  }

  ul,
  ol {
    padding-left: 1.5em;
    margin: 0.5em 0;
  }

  a {
    color: var(--el-color-primary);
    text-decoration: none;
    &:hover {
      text-decoration: underline;
    }
  }

  blockquote {
    border-left: 3px solid var(--ep-border-color);
    margin: 0.5em 0;
    padding-left: 1em;
    color: var(--ep-text-color-secondary);
  }
}
</style>
