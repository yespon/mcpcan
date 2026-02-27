<template>
  <div class="flex flex-col h-full bg-[var(--ep-bg-color-page)]">
    <!-- Session List -->
    <div class="flex-1 overflow-y-auto p-2 space-y-1 custom-scrollbar">
      <div
        v-for="session in sessions"
        :key="session.id"
        class="group flex items-center px-3 py-2 rounded-md cursor-pointer text-sm transition-colors duration-200"
        :class="[
          currentSessionId === session.id
            ? 'bg-[var(--el-color-primary-light-9)] text-[var(--el-color-primary)]'
            : 'hover:bg-[var(--ep-fill-color-light)] text-[var(--ep-text-color-primary)]',
        ]"
        @click="$emit('select', session.id)"
      >
        <el-icon class="mr-2 text-base shrink-0"><ChatDotRound /></el-icon>
        <span class="truncate flex-1">{{ session.name || 'New Chat' }}</span>

        <div class="opacity-0 group-hover:opacity-100 transition-opacity flex" @click.stop>
          <el-dropdown trigger="click" @command="(cmd: string) => handleCommand(cmd, session)">
            <el-button link size="small">
              <el-icon
                class="text-[var(--ep-text-color-secondary)] hover:text-[var(--ep-text-color-primary)]"
                ><MoreFilled
              /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="rename">
                  <el-icon><Edit /></el-icon> Rename
                </el-dropdown-item>
                <el-dropdown-item command="delete" divided class="!text-red-500">
                  <el-icon><Delete /></el-icon> Delete
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </div>

      <div
        v-if="sessions.length === 0"
        class="text-center py-8 text-[var(--ep-text-color-placeholder)] text-xs"
      >
        No chat history
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ChatDotRound, Delete, MoreFilled, Edit } from '@element-plus/icons-vue'
import type { AiSession } from '../types'

defineProps<{
  sessions: AiSession[]
  currentSessionId?: number
}>()

const emit = defineEmits<{
  (e: 'create'): void
  (e: 'select', id: number): void
  (e: 'delete', id: number): void
  (e: 'rename', session: AiSession): void
}>()

const handleCommand = (cmd: string, session: AiSession) => {
  if (cmd === 'rename') {
    emit('rename', session)
  } else if (cmd === 'delete') {
    emit('delete', session.id)
  }
}
</script>

<style scoped>
.custom-scrollbar::-webkit-scrollbar {
  width: 4px;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background-color: var(--ep-border-color);
  border-radius: 4px;
}
</style>
