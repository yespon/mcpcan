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
        <el-icon class="cursor-pointer hover:text-[var(--el-color-danger)]" @click="clearFile">
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
        <el-popover
          placement="top-start"
          :width="500"
          trigger="click"
          v-model:visible="modelSelectorVisible"
          popper-class="!p-0 !rounded-xl overflow-hidden"
          @show="handlePopoverShow"
        >
          <template #reference>
            <span
              class="flex items-center gap-1 cursor-pointer hover:text-[var(--el-color-primary)] transition-colors select-none"
            >
              <el-icon><Cpu /></el-icon>
              {{ currentModelDisplayName }}
              <el-icon><ArrowDown /></el-icon>
            </span>
          </template>

          <div class="flex h-[320px] bg-[var(--ep-bg-color)]">
            <!-- Left: Model Access (Providers) -->
            <div
              class="w-1/3 border-r border-[var(--ep-border-color)] overflow-y-auto flex flex-col"
            >
              <div
                class="p-3 text-xs font-bold text-[var(--ep-text-color-secondary)] uppercase tracking-wider"
              >
                Provider Config
              </div>
              <div
                v-for="m in models"
                :key="m.id"
                class="px-4 py-3 cursor-pointer text-sm transition-colors flex items-center justify-between group"
                :class="
                  selectedAccessId === m.id
                    ? 'bg-[var(--el-color-primary-light-9)] text-[var(--el-color-primary)] font-medium'
                    : 'hover:bg-[var(--ep-fill-color-light)]'
                "
                @mouseenter="selectedAccessId = m.id"
              >
                <div class="truncate flex items-center gap-2">
                  <img
                    v-if="getProviderIcon(m.provider)"
                    :src="getProviderIcon(m.provider)"
                    class="w-4 h-4 object-contain"
                    :alt="m.provider"
                  />
                  {{ m.name }}
                </div>
                <el-icon v-if="selectedAccessId === m.id"><ArrowRight /></el-icon>
              </div>

              <div class="mt-auto border-t border-[var(--ep-border-color)] p-2">
                <el-button
                  link
                  size="small"
                  class="w-full justify-start text-[var(--ep-text-color-secondary)] hover:!text-[var(--el-color-primary)]"
                  @click="handleAddCustom"
                >
                  <el-icon class="mr-1"><Plus /></el-icon> Add Custom
                </el-button>
              </div>
            </div>

            <!-- Right: Specific Models -->
            <div class="flex-1 overflow-y-auto flex flex-col bg-[var(--ep-bg-color-page)]">
              <div
                class="p-3 text-xs font-bold text-[var(--ep-text-color-secondary)] uppercase tracking-wider sticky top-0 bg-[var(--ep-bg-color-page)] z-10 backdrop-blur-sm"
              >
                Models
              </div>
              <div
                v-if="previewModels.length === 0"
                class="p-4 text-center text-[var(--ep-text-color-secondary)] text-sm"
              >
                No models available
              </div>
              <div
                v-for="modelName in previewModels"
                :key="modelName"
                class="px-4 py-2 cursor-pointer text-sm hover:bg-[var(--ep-fill-color)] flex items-center justify-between"
                @click="handleSelectModel(modelName)"
              >
                <span>{{ modelName }}</span>
                <el-icon v-if="isModelSelected(modelName)" class="text-[var(--el-color-primary)]"
                  ><Check
                /></el-icon>
              </div>
            </div>
          </div>
        </el-popover>

        <!-- System Prompt & Temperature Settings Trigger -->
        <span
          class="flex items-center gap-1 cursor-pointer hover:text-[var(--el-color-primary)] transition-colors select-none"
          @click="settingsVisible = true"
        >
          <el-icon><Setting /></el-icon>
          System Prompt & Temperature
        </span>
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
  </div>

  <!-- System Prompt & Temperature Settings Dialog -->
  <el-dialog
    v-model="settingsVisible"
    title="System Prompt & Temperature"
    width="500px"
    center
    append-to-body
  >
    <div class="flex flex-col gap-4">
      <div>
        <div class="text-xs text-[var(--ep-text-color-secondary)] mb-1">System Prompt</div>
        <el-input
          v-model="settingsForm.systemPrompt"
          type="textarea"
          :rows="6"
          placeholder="You are a helpful assistant..."
          class="!bg-[var(--ep-bg-color-page)]"
        />
      </div>
      <div>
        <div class="flex items-center justify-between mb-1">
          <span class="text-xs text-[var(--ep-text-color-secondary)]">Temperature</span>
          <span class="text-xs font-mono">{{ settingsForm.temperature }}</span>
        </div>
        <el-slider
          v-model="settingsForm.temperature"
          :min="0"
          :max="2"
          :step="0.1"
          :show-tooltip="false"
          size="small"
        />
        <div class="text-[10px] text-[var(--ep-text-color-secondary)]">
          Focus (0.0 - 0.3) <span class="mx-1">|</span> Creative (0.7 - 2.0)
        </div>
      </div>
    </div>
    <template #footer>
      <span class="dialog-footer">
        <el-button @click="settingsVisible = false">Cancel</el-button>
        <el-button type="primary" @click="saveSettings">Save</el-button>
      </span>
    </template>
  </el-dialog>

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
</template>

<script setup lang="ts">
import {
  Paperclip,
  Cpu,
  ArrowDown,
  Top,
  Plus,
  Close,
  ArrowRight,
  Check,
} from '@element-plus/icons-vue'
import { ElMessage, type UploadFile } from 'element-plus'
import { useRouterHooks } from '@/utils/url'
import type { AIModel, SupportedProvider } from '../types'

const props = defineProps<{
  models: AIModel[]
  currentModel: string
  currentTargetModel?: string
  supportedProviders?: SupportedProvider[]
  disabled?: boolean
  systemPrompt?: string
  temperature?: number
}>()

const emit = defineEmits<{
  (e: 'send', content: string, file?: File): void
  (e: 'update:currentModel', id: string): void
  (e: 'update:currentTargetModel', name: string): void
  (e: 'update:systemPrompt', prompt: string): void
  (e: 'update:temperature', temp: number): void
  (e: 'add-model', model: any): void
  (e: 'save-settings'): void
  (e: 'model-change-confirmed', modelName: string, accessId: string): void
}>()

const { jumpToPage } = useRouterHooks()
const input = ref('')
const selectedFile = ref<File | null>(null)
const dialogVisible = ref(false)
const loading = ref(false)
const modelSelectorVisible = ref(false)
const settingsVisible = ref(false)
const selectedAccessId = ref('') // For temporary selection in popover
const customModelForm = ref({
  name: '',
  provider: '',
  apiKey: '',
  baseUrl: '',
  modelName: '',
  allowedModels: [] as string[],
})

const settingsForm = reactive({
  systemPrompt: '',
  temperature: 0.7,
})

watch(
  () => props.systemPrompt,
  (val) => {
    settingsForm.systemPrompt = val || ''
  },
  { immediate: true },
)

watch(
  () => props.temperature,
  (val) => {
    settingsForm.temperature = val !== undefined ? val : 0.7
  },
  { immediate: true },
)

const currentModelDisplayName = computed(() => {
  // If we have a specific target model, show it. Otherwise show Access Name
  if (props.currentTargetModel) return props.currentTargetModel

  const model = props.models.find((m) => m.id === props.currentModel)
  return model ? model.description || model.name : 'Select Model'
})

const previewModels = computed(() => {
  if (!selectedAccessId.value) return []
  const access = props.models.find((m) => m.id === selectedAccessId.value)
  if (!access) return []

  // Case 1: Access key has restricted allowedModels
  if (access.allowedModels && access.allowedModels.length > 0) {
    return access.allowedModels
  }

  // Case 2: Fetch from supported Providers using access.provider
  if (access.provider && props.supportedProviders) {
    const provider = props.supportedProviders.find((p) => p.id === access.provider)
    if (provider) return provider.models
  }

  return []
})

const handlePopoverShow = () => {
  selectedAccessId.value = props.currentModel
  // If no models exist, select first
  if (!selectedAccessId.value && props.models.length > 0) {
    selectedAccessId.value = props.models[0].id
  }
}

const isModelSelected = (modelName: string) => {
  return props.currentModel === selectedAccessId.value && props.currentTargetModel === modelName
}

const handleSelectModel = (modelName: string) => {
  if (!selectedAccessId.value) return

  // If no session is active (disabled=false means we are ready to chat or chatting)
  // Check if session exists via prop context if possible, but here we only have disabled.
  // Actually, disabled updates if streaming.
  // We need to know if there is an active session ID to update.
  // Let's assume if currentTargetModel is set and different, we ask.

  if (props.currentTargetModel && props.currentTargetModel !== modelName) {
    ElMessageBox.confirm(
      '现在更改模型将会导致聊天会话重新更新。这个动作不能被返回。',
      '重启聊天会话？',
      {
        confirmButtonText: '重启聊天会话',
        cancelButtonText: '取消',
        type: 'warning',
        center: true,
      },
    )
      .then(() => {
        // Confirmed
        emit('update:currentModel', selectedAccessId.value)
        emit('update:currentTargetModel', modelName)
        emit('model-change-confirmed', modelName, selectedAccessId.value)
      })
      .catch(() => {
        // Cancelled
      })
  } else {
    // No active session or just initial selection
    emit('update:currentModel', selectedAccessId.value)
    emit('update:currentTargetModel', modelName)
  }

  modelSelectorVisible.value = false
}

const handleAddCustom = () => {
  jumpToPage({
    url: '/model-manage',
  })
  modelSelectorVisible.value = false
}

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

const getProviderIcon = (providerId: string) => {
  if (!props.supportedProviders) return ''
  const p = props.supportedProviders.find((x) => x.id === providerId)
  return p ? p.iconUrl : ''
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
    await emit('add-model', {
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

const saveSettings = () => {
  emit('update:systemPrompt', settingsForm.systemPrompt)
  emit('update:temperature', settingsForm.temperature)
  emit('save-settings')
  settingsVisible.value = false
}
</script>

<style scoped>
:deep(.el-textarea__inner) {
  box-shadow: none !important;
  background-color: transparent !important;
  color: var(--ep-text-color-primary) !important;
}
</style>
