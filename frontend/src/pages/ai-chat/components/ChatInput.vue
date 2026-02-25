<template>
  <div
    class="rounded-xl p-4 border border-[var(--ep-border-color)] bg-[var(--ep-bg-color)] dark:bg-[rgba(255,255,255,0.05)] dark:border-[rgba(255,255,255,0.1)] shadow-sm transition-colors duration-300"
  >
    <el-input
      v-model="input"
      type="textarea"
      :rows="3"
      placeholder="Ask something... Use Slash '/' commands for Skills & MCP prompts"
      resize="none"
      class="!border-none !shadow-none bg-transparent chat-input"
      @keydown.enter.prevent="handleSend"
    />

    <div
      class="flex items-center justify-between mt-4 text-[var(--ep-text-color-secondary)] text-sm"
    >
      <div class="flex items-center gap-4">
        <el-tooltip content="Upload File" placement="top">
          <el-button
            link
            class="!text-[var(--ep-text-color-secondary)] hover:!text-[var(--el-color-primary)]"
          >
            <el-icon class="text-lg"><Paperclip /></el-icon>
          </el-button>
        </el-tooltip>

        <!-- Model Selector -->
        <el-dropdown trigger="click" @command="handleModelSelect">
          <span
            class="flex items-center gap-1 cursor-pointer hover:text-[var(--el-color-primary)] transition-colors"
          >
            <el-icon><Cpu /></el-icon>
            {{ currentModelName }}
            <el-icon><ArrowDown /></el-icon>
          </span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item v-for="model in models" :key="model.id" :command="model.id">
                {{ model.name }}
              </el-dropdown-item>
              <el-dropdown-item divided command="add_custom">
                <el-icon><Plus /></el-icon> Add Custom Model
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>

        <!-- <el-tooltip content="Setting" placement="top">
          <el-button
            link
            class="!text-[var(--ep-text-color-secondary)] hover:!text-[var(--el-color-primary)]"
          >
            <el-icon><Setting /></el-icon> System Prompt & Temper...
          </el-button>
        </el-tooltip>

        <el-tooltip content="Tool Approval" placement="top">
          <el-button
            link
            class="!text-[var(--ep-text-color-secondary)] hover:!text-[var(--el-color-primary)]"
          >
            <el-icon><CircleCheck /></el-icon> Tool Approval
          </el-button>
        </el-tooltip>

        <el-tooltip content="X-Ray" placement="top">
          <el-button
            link
            class="!text-[var(--ep-text-color-secondary)] hover:!text-[var(--el-color-primary)]"
          >
            <el-icon><View /></el-icon> X-Ray
          </el-button>
        </el-tooltip> -->
      </div>

      <div class="flex items-center gap-4">
        <span>0.6%</span>
        <el-button
          type="primary"
          circle
          class="send-btn !bg-[var(--el-color-primary)] !border-none !text-white hover:!bg-[var(--el-color-primary-light-3)] transition-colors"
          @click="handleSend"
        >
          <el-icon><Top /></el-icon>
        </el-button>
      </div>
    </div>

    <!-- Add Custom Model Dialog -->
    <el-dialog v-model="dialogVisible" title="Add Custom Model" width="400px">
      <el-form :model="customModelForm" label-width="80px">
        <el-form-item label="Name">
          <el-input v-model="customModelForm.name" placeholder="e.g. My Custom Model" />
        </el-form-item>
        <el-form-item label="Provider">
          <el-input v-model="customModelForm.provider" placeholder="e.g. Local API" />
        </el-form-item>
        <el-form-item label="Model ID">
          <el-input v-model="customModelForm.id" placeholder="e.g. custom-model-v1" />
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="dialogVisible = false">Cancel</el-button>
          <el-button type="primary" @click="submitCustomModel">Add</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import {
  Paperclip,
  Cpu,
  ArrowDown,
  Setting,
  CircleCheck,
  View,
  Top,
  Plus,
} from '@element-plus/icons-vue'
import type { AIModel } from '../types'

const props = defineProps<{
  models: AIModel[]
  currentModel: string
}>()

const emit = defineEmits<{
  (e: 'send', content: string): void
  (e: 'update:currentModel', id: string): void
  (e: 'add-model', model: AIModel): void
}>()

const input = ref('')
const dialogVisible = ref(false)
const customModelForm = ref({
  name: '',
  provider: '',
  id: '',
})

const currentModelName = computed(() => {
  const model = props.models.find((m) => m.id === props.currentModel)
  return model ? model.name : 'Select Model'
})

const handleSend = () => {
  if (!input.value.trim()) return
  emit('send', input.value.trim())
  input.value = ''
}

const handleModelSelect = (command: string) => {
  if (command === 'add_custom') {
    dialogVisible.value = true
  } else {
    emit('update:currentModel', command)
  }
}

const submitCustomModel = () => {
  if (customModelForm.value.name && customModelForm.value.id) {
    emit('add-model', { ...customModelForm.value })
    dialogVisible.value = false
    customModelForm.value = { name: '', provider: '', id: '' }
  }
}
</script>

<style scoped>
:deep(.el-textarea__inner) {
  box-shadow: none !important;
  background-color: transparent !important;
  color: var(--ep-text-color-primary) !important;
  padding: 0;
}
:deep(.chat-input .el-textarea__inner::placeholder) {
  /* color: var(--ep-text-color-placeholder); */
}
:deep(.el-textarea__inner:focus) {
  box-shadow: none !important;
}
</style>
