<template>
  <div class="h-full flex overflow-hidden">
    <!-- Main Chat Area -->
    <div class="flex-1 flex flex-col min-w-0 relative">
      <!-- Header -->
      <div
        class="h-14 flex items-center justify-end px-6 absolute top-0 right-0 z-10"
        v-if="!isSidebarOpen"
      >
        <el-tooltip content="Show Logs" placement="bottom">
          <el-button
            type="info"
            text
            circle
            @click="isSidebarOpen = true"
            class="!text-[var(--ep-text-color-secondary)] hover:!text-[var(--el-color-primary)]"
          >
            <el-icon class="text-lg"><Fold /></el-icon>
          </el-button>
        </el-tooltip>
      </div>

      <!-- Messages Area -->
      <div class="flex-1 overflow-y-auto hide-scrollbar px-6 md:px-20 lg:px-40 py-6">
        <ChatMessage v-for="msg in messages" :key="msg.id" :message="msg" />
      </div>

      <!-- Input Area -->
      <div class="p-6 md:px-20 lg:px-40 pb-8">
        <ChatInput
          :models="models"
          :current-model="currentModel"
          @update:current-model="currentModel = $event"
          @send="handleSend"
          @add-model="addCustomModel"
        />
      </div>
    </div>

    <!-- Right Sidebar (Logs) -->
    <div
      class="shrink-0 transition-all duration-300 ease-in-out border-l border-[var(--ep-border-color)] bg-[var(--ep-bg-color)]"
      :class="[
        isSidebarOpen ? 'w-80 translate-x-0' : 'w-0 translate-x-full overflow-hidden border-l-0',
      ]"
    >
      <LogSidebar @close="isSidebarOpen = false" />
    </div>
  </div>
</template>

<script setup lang="ts">
import ChatMessage from './components/ChatMessage.vue'
import ChatInput from './components/ChatInput.vue'
import LogSidebar from './components/LogSidebar.vue'
import { useChat } from './composables/useChat'
import { Fold } from '@element-plus/icons-vue'

const { messages, models, currentModel, addMessage, addCustomModel } = useChat()
const isSidebarOpen = ref(true)

const handleSend = (content: string) => {
  addMessage(content, 'user')
  // Simulate AI response
  setTimeout(() => {
    addMessage(`This is a simulated response from ${currentModel.value}.`, 'assistant')
  }, 1000)
}
</script>
