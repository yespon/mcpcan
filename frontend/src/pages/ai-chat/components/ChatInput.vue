<template>
  <div
    class="rounded-xl p-4 border border-[var(--ep-border-color)] bg-[var(--ep-bg-color)] dark:bg-[rgba(255,255,255,0.05)] dark:border-[rgba(255,255,255,0.1)] shadow-sm transition-colors duration-300"
    :class="{ 'opacity-60 cursor-not-allowed': disabled }"
  >
    <el-input
      v-model="input"
      type="textarea"
      :rows="3"
      :placeholder="
        disabled
          ? 'Create or select a session to start chatting...'
          : 'Ask something... Use Slash \'/\' commands for Skills & MCP prompts'
      "
      resize="none"
      class="!border-none !shadow-none bg-transparent chat-input"
      :disabled="disabled"
      @keydown.enter.prevent="handleSend"
    />

    <!-- Selected File Preview -->
    <div v-if="selectedFile" class="mt-2 flex items-center gap-2">
      <span
        class="text-xs text-[var(--ep-text-color-secondary)] flex items-center gap-1 bg-[var(--ep-bg-color)] px-2 py-1 rounded border border-[var(--ep-border-color)]"
      >
        <el-icon><Paperclip /></el-icon> {{ selectedFile.name }}
        <el-icon class="cursor-pointer hover:text-[var(--ep-color-danger)]" @click="clearFile">
          <Close />
        </el-icon>
      </span>
    </div>

    <div
      class="flex items-center justify-between mt-4 text-[var(--ep-text-color-secondary)] text-sm"
    >
      <div class="flex items-center gap-4">
        <el-upload
          action="#"
          :auto-upload="false"
          :show-file-list="false"
          :on-change="handleFileChange"
          :disabled="disabled"
        >
          <el-tooltip content="Upload File" placement="top">
            <el-button
              link
              class="!text-[var(--ep-text-color-secondary)] hover:!text-[var(--el-color-primary)]"
              :disabled="disabled"
            >
              <el-icon class="text-lg"><Paperclip /></el-icon>
            </el-button>
          </el-tooltip>
        </el-upload>

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
        <!-- <span>0.6%</span> -->
        <el-button
          type="primary"
          circle
          class="send-btn !bg-[var(--el-color-primary)] !border-none !text-white hover:!bg-[var(--el-color-primary-light-3)] transition-colors"
          :disabled="disabled || (!input.trim() && !selectedFile)"
          @click="handleSend"
        >
          <el-icon><Top /></el-icon>
        </el-button>
      </div>
    </div>

    <!-- Add Custom Model Dialog -->
    <el-dialog v-model="dialogVisible" title="Add Custom Model" width="500px">
      <el-form :model="customModelForm" label-width="120px">
        <el-form-item label="Display Name" required>
          <el-input v-model="customModelForm.name" placeholder="e.g. My Custom Model" />
        </el-form-item>
        <el-form-item label="Provider" required>
          <el-select
            v-model="customModelForm.provider"
            placeholder="Select provider"
            filterable
            class="w-full"
            @change="handleProviderChange"
          >
            <el-option v-for="p in supportedProviders" :key="p.id" :label="p.name" :value="p.id">
              <span class="float-left">{{ p.name }}</span>
              <span class="float-right text-gray-400 text-xs ml-2">{{ p.id }}</span>
            </el-option>
          </el-select>
        </el-form-item>

        <el-form-item label="API Key" required>
          <el-input
            v-model="customModelForm.apiKey"
            type="password"
            show-password
            placeholder="sk-..."
          />
        </el-form-item>
        <el-form-item label="Base URL">
          <el-input v-model="customModelForm.baseUrl" placeholder="https://api.openai.com/v1" />
          <div
            class="text-xs text-[var(--ep-text-color-secondary)] mt-1"
            v-if="customModelForm.provider && getProviderBaseUrl(customModelForm.provider)"
          >
            Default: {{ getProviderBaseUrl(customModelForm.provider) }}
            <el-button
              type="primary"
              link
              size="small"
              @click="customModelForm.baseUrl = getProviderBaseUrl(customModelForm.provider)"
            >
              Use
            </el-button>
          </div>
        </el-form-item>

        <el-form-item label="Allowed Models">
          <el-select
            v-model="customModelForm.allowedModels"
            multiple
            filterable
            allow-create
            default-first-option
            placeholder="Leave empty = allow all"
            class="w-full"
          >
            <el-option v-for="m in selectedProviderModels" :key="m" :label="m" :value="m" />
          </el-select>
          <div class="text-xs text-[var(--ep-text-color-secondary)] mt-1">
            Restrict available models. Leave empty to allow all.
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="dialogVisible = false">Cancel</el-button>
          <el-button type="primary" @click="submitCustomModel" :loading="loading">Save</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { Paperclip, Cpu, ArrowDown, Top, Plus, Close } from '@element-plus/icons-vue'
import { ElMessage, type UploadFile } from 'element-plus'
import type { AIModel, SupportedProvider } from '../types'

const props = defineProps<{
  models: AIModel[]
  currentModel: string
  supportedProviders?: SupportedProvider[]
  disabled?: boolean
}>()

const emit = defineEmits<{
  (e: 'send', content: string, file?: File): void
  (e: 'update:currentModel', id: string): void
  (e: 'add-model', model: any): void
}>()

const input = ref('')
const selectedFile = ref<File | null>(null)
const dialogVisible = ref(false)
const loading = ref(false)
const customModelForm = ref({
  name: '',
  provider: '',
  apiKey: '',
  baseUrl: '',
  modelName: '',
  allowedModels: [] as string[],
})

const currentModelName = computed(() => {
  const model = props.models.find((m) => m.id === props.currentModel)
  return model ? model.description || model.name : 'Select Model'
})

const selectedProviderModels = computed(() => {
  if (!customModelForm.value.provider || !props.supportedProviders) return []
  const p = props.supportedProviders.find((x) => x.id === customModelForm.value.provider)
  return p ? p.models : []
})

const getProviderBaseUrl = (pid: string) => {
  if (!props.supportedProviders) return ''
  const p = props.supportedProviders.find((x) => x.id === pid)
  return p ? p.baseUrl : ''
}

const handleProviderChange = () => {
  // Optional: Auto-fill base URL if empty?
  // Only if logic dictates. For now, keep it simple as reference implementation does not auto-fill on change, just shows helper.
  // Actually, chat_test.html does not auto-fill, just shows a helper button.
}

const handleFileChange = (uploadFile: UploadFile) => {
  if (uploadFile.raw) {
    selectedFile.value = uploadFile.raw
  }
}

const clearFile = () => {
  selectedFile.value = null
}

const handleSend = () => {
  if (!input.value.trim() && !selectedFile.value) return
  if (props.disabled) return

  emit('send', input.value.trim(), selectedFile.value || undefined)

  input.value = ''
  selectedFile.value = null
}

const handleModelSelect = (command: string) => {
  if (command === 'add_custom') {
    dialogVisible.value = true
  } else {
    emit('update:currentModel', command)
  }
}

const submitCustomModel = async () => {
  // Basic validation
  if (
    !customModelForm.value.name ||
    !customModelForm.value.provider ||
    !customModelForm.value.apiKey
  ) {
    ElMessage.warning('Please fill in required fields (Name, Provider, API Key)')
    return
  }

  // Infer modelName (Target Model) from allowedModels if available
  let finalModelName = ''
  if (customModelForm.value.allowedModels && customModelForm.value.allowedModels.length > 0) {
    finalModelName = customModelForm.value.allowedModels[0]
  }

  loading.value = true
  try {
    emit('add-model', {
      id: '',
      name: customModelForm.value.name,
      provider: customModelForm.value.provider,
      apiKey: customModelForm.value.apiKey,
      baseUrl: customModelForm.value.baseUrl,
      description: finalModelName,
      modelName: finalModelName,
      allowedModels: customModelForm.value.allowedModels,
    })

    dialogVisible.value = false
    // Reset form
    customModelForm.value = {
      name: '',
      provider: '',
      apiKey: '',
      baseUrl: '',
      modelName: '',
      allowedModels: [],
    }
  } finally {
    loading.value = false
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
</style>
