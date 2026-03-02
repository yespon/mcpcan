<template>
  <div
    class="rounded-xl p-4 border border-[var(--ep-border-color)] bg-[var(--ep-bg-color)] dark:bg-[rgba(255,255,255,0.05)] dark:border-[rgba(255,255,255,0.1)] shadow-sm transition-colors duration-300"
    :class="{ 'opacity-60 cursor-not-allowed': disabled }"
  >
    <el-input
      v-model="input"
      type="textarea"
      :rows="3"
      :placeholder="disabled ? t('aiChat.createOrSelectSession') : t('aiChat.askSomething')"
      resize="none"
      class="!border-none !shadow-none bg-transparent chat-input"
      :disabled="disabled"
      @keydown.enter.prevent="handleSend"
    />

    <!-- Selected File Preview -->
    <div v-if="attachments.length > 0" class="mt-2 flex items-center gap-2">
      <div
        v-for="(file, index) in attachments"
        :key="index"
        class="text-xs text-[var(--ep-text-color-secondary)] flex items-center gap-1 bg-[var(--ep-bg-color)] px-2 py-1 rounded border border-[var(--ep-border-color)]"
      >
        <el-icon><Paperclip /></el-icon> {{ file.name }}
        <el-icon
          class="cursor-pointer hover:text-[var(--el-color-danger)]"
          @click="clearFile(index)"
        >
          <Close />
        </el-icon>
      </div>
      <div
        v-if="isUploading"
        class="text-xs text-[var(--ep-text-color-secondary)] flex items-center gap-1"
      >
        <el-icon class="is-loading"><Loading /></el-icon> Uploading...
      </div>
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
          :disabled="disabled || !supportsFileUpload"
          :accept="acceptFileTypes"
        >
          <el-tooltip
            :content="supportsFileUpload ? t('aiChat.uploadFile') : t('aiChat.modelNoFileSupport')"
            placement="top"
          >
            <el-button
              link
              class="!text-[var(--ep-text-color-secondary)] hover:!text-[var(--el-color-primary)]"
              :disabled="disabled || !supportsFileUpload"
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
              <img
                v-if="currentModelIcon"
                :src="currentModelIcon"
                class="w-4 h-4 object-contain"
                alt="provider"
              />
              <el-icon v-else><Cpu /></el-icon>
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
                {{ t('aiChat.providerConfig') }}
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
                  <el-icon class="mr-1"><Plus /></el-icon> {{ t('aiChat.addCustom') }}
                </el-button>
              </div>
            </div>

            <!-- Right: Specific Models -->
            <div class="flex-1 overflow-y-auto flex flex-col bg-[var(--ep-bg-color-page)]">
              <div
                class="p-3 text-xs font-bold text-[var(--ep-text-color-secondary)] uppercase tracking-wider sticky top-0 bg-[var(--ep-bg-color-page)] z-10 backdrop-blur-sm"
              >
                {{ t('aiChat.models') }}
              </div>
              <div
                v-if="previewModels.length === 0"
                class="p-4 text-center text-[var(--ep-text-color-secondary)] text-sm"
              >
                {{ t('aiChat.noModelsAvailable') }}
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
          {{ t('aiChat.systemPromptAndTemperature') }}
        </span>

        <!-- MCP Instance Selector -->
        <span
          class="flex items-center gap-1 cursor-pointer hover:text-[var(--el-color-primary)] transition-colors select-none"
          @click="openMcpSelector"
        >
          <el-icon><Connection /></el-icon>
          {{ currentMcpDisplayName }}
          <el-icon
            v-if="mcpConfig && mcpConfig !== '{}'"
            @click="clearMcpConfig"
            class="text-xs hover:text-red-500"
            ><Close
          /></el-icon>
        </span>
      </div>

      <div class="flex items-center gap-4">
        <!-- <span>0.6%</span> -->
        <el-button
          type="primary"
          circle
          class="send-btn !bg-[var(--el-color-primary)] !border-none !text-white hover:!bg-[var(--el-color-primary-light-3)] transition-colors"
          :disabled="disabled || isUploading || (!input.trim() && attachments.length === 0)"
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
    :title="t('aiChat.systemPromptAndTemperature')"
    width="500px"
    center
    append-to-body
  >
    <div class="flex flex-col gap-4">
      <div>
        <div class="text-xs text-[var(--ep-text-color-secondary)] mb-1 flex items-center gap-1">
          <el-icon><Document /></el-icon>
          {{ t('aiChat.systemPrompt') }}
        </div>
        <el-input
          v-model="settingsForm.systemPrompt"
          type="textarea"
          :rows="6"
          :placeholder="t('aiChat.systemPromptPlaceholder')"
          class="!bg-[var(--ep-bg-color-page)]"
        />
      </div>
      <div>
        <div class="flex items-center justify-between mb-1">
          <span class="text-xs text-[var(--ep-text-color-secondary)] flex items-center gap-1">
            <el-icon><Odometer /></el-icon>
            {{ t('aiChat.temperature') }}
          </span>
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

  <!-- MCP Selection Dialog -->
  <McpSelector
    v-model:visible="mcpSelectorVisible"
    :title="t('aiChat.mcpConfig')"
    :mcp-instances="mcpInstances"
    :mcp-config="mcpConfig"
    :mcp-loading="mcpLoading"
    :mcp-has-more="mcpHasMore"
    @update:mcpConfig="$emit('update:mcpConfig', $event)"
    @save-mcp-config="$emit('save-mcp-config', $event)"
    @load-mcp="(page, append, query) => $emit('load-mcp', page, append, query)"
  />
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
  Document,
  Odometer,
  Setting,
  Connection,
  Loading,
} from '@element-plus/icons-vue'
import McpSelector from './McpSelector.vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, type UploadFile } from 'element-plus'
import { useRouterHooks } from '@/utils/url'
import type { AIModel, SupportedProvider } from '../types'
import type { InstanceResult } from '@/types/instance'
import { InstanceStatus } from '@/types/instance'
import { useChat } from '../composables/useChat'
import type { ChatAttachment } from '@/api/agent'

const { t } = useI18n()

const props = defineProps<{
  models: AIModel[]
  currentModel: string
  currentTargetModel?: string
  supportedProviders?: SupportedProvider[]
  disabled?: boolean
  systemPrompt?: string
  temperature?: number
  mcpInstances?: InstanceResult[]
  mcpConfig?: string
  mcpLoading?: boolean
  mcpHasMore?: boolean
  sessionId?: number
}>()

const emit = defineEmits<{
  (e: 'send', content: string, attachments: ChatAttachment[]): void
  (e: 'update:currentModel', id: string): void
  (e: 'update:currentTargetModel', name: string): void
  (e: 'update:systemPrompt', prompt: string): void
  (e: 'update:temperature', temp: number): void
  (e: 'add-model', model: any): void
  (e: 'save-settings'): void
  (e: 'model-change-confirmed', modelName: string, accessId: string): void
  (e: 'update:mcpConfig', config: string): void
  (e: 'save-mcp-config', config: string): void
  (e: 'load-mcp', page: number, append: boolean, query: string): void
}>()

const { jumpToPage } = useRouterHooks()
const { uploadFile: useChatUploadFile } = useChat(true)

const input = ref('')
const attachments = ref<ChatAttachment[]>([])
const isUploading = ref(false)
const dialogVisible = ref(false)
const loading = ref(false)
const modelSelectorVisible = ref(false)
const mcpSelectorVisible = ref(false)
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

// service status
const activeOptions = {
  active: {
    label: t('status.' + InstanceStatus.ACTIVE),
    type: 'success',
    value: InstanceStatus.ACTIVE,
  },
  inactive: {
    label: t('status.' + InstanceStatus.INACTIVE),
    type: 'danger',
    value: InstanceStatus.INACTIVE,
  },
}

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

const currentModelIcon = computed(() => {
  if (!props.currentModel) return ''
  const model = props.models.find((m) => m.id === props.currentModel)
  if (!model) return ''
  return getProviderIcon(model.provider)
})

// Find the current model's ModelInfo from supportedProviders
const currentModelInfo = computed(() => {
  if (!props.currentTargetModel || !props.currentModel || !props.supportedProviders) return null
  // Find the access config to get the provider id
  const access = props.models.find((m) => m.id === props.currentModel)
  if (!access) return null
  // Find the provider
  const provider = props.supportedProviders.find((p) => p.id === access.provider)
  if (!provider || !provider.modelInfos) return null
  // Find the specific model info
  return provider.modelInfos.find((mi) => mi.id === props.currentTargetModel) || null
})

// Whether the current model supports file upload (image or document)
const supportsFileUpload = computed(() => {
  if (!currentModelInfo.value) return false
  return !!(currentModelInfo.value.supportsVision || currentModelInfo.value.supportsDocument)
})

// Build accept string from imageMimeTypes + documentMimeTypes
const acceptFileTypes = computed(() => {
  if (!currentModelInfo.value) return ''
  const mimeTypes: string[] = []
  if (currentModelInfo.value.supportsVision && currentModelInfo.value.imageMimeTypes) {
    mimeTypes.push(...currentModelInfo.value.imageMimeTypes)
  }
  if (currentModelInfo.value.supportsDocument && currentModelInfo.value.documentMimeTypes) {
    mimeTypes.push(...currentModelInfo.value.documentMimeTypes)
  }
  return mimeTypes.join(',')
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

const handleFileChange = async (file: UploadFile) => {
  if (file.raw) {
    const modelInfo = currentModelInfo.value

    // Validate file type against model capabilities
    if (modelInfo) {
      const fileMime = file.raw.type
      const isImage = modelInfo.supportsVision && modelInfo.imageMimeTypes?.includes(fileMime)
      const isDocument =
        modelInfo.supportsDocument && modelInfo.documentMimeTypes?.includes(fileMime)

      if (!isImage && !isDocument) {
        ElMessage.warning(t('aiChat.unsupportedFileType'))
        return
      }

      // Validate file size
      const maxSize = isImage ? modelInfo.maxImageSize || 0 : modelInfo.maxDocumentSize || 0
      if (maxSize > 0 && file.raw.size > maxSize) {
        const maxSizeMB = (maxSize / 1024 / 1024).toFixed(1)
        ElMessage.warning(`${t('aiChat.fileTooLarge')} (max ${maxSizeMB}MB)`)
        return
      }
    }

    isUploading.value = true
    try {
      const res = await useChatUploadFile(file.raw, props.sessionId!)
      if (res && res.url) {
        // Determine attachment type based on MIME
        const fileType = file.raw.type.startsWith('image/') ? 'image' : 'file'
        attachments.value.push({
          name: file.name,
          url: res.url,
          type: fileType,
        })
      }
    } catch (error) {
      ElMessage.error(t('aiChat.uploadFailed'))
    } finally {
      isUploading.value = false
    }
  }
}

const clearFile = (index: number) => {
  attachments.value.splice(index, 1)
}

const handleSend = () => {
  if (!input.value.trim() && attachments.value.length === 0) return
  if (props.disabled) return

  emit('send', input.value.trim(), [...attachments.value])

  input.value = ''
  attachments.value = []
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

const currentMcpDisplayName = computed(() => {
  // Check if current mcpConfig matches any instance
  if (!props.mcpInstances || !props.mcpConfig) return 'Select MCP'
  // If not config string is empty
  if (!props.mcpConfig || props.mcpConfig === '{}') return 'Select MCP'
  // If multiple instances selected
  try {
    const config = JSON.parse(props.mcpConfig)
    // Basic check if it looks like mcpServers config
    if (config.mcpServers) {
      const serverCount = Object.keys(config.mcpServers).length
      if (serverCount > 0) return `${serverCount} MCP(s)`
    }
  } catch (e) {
    // failed parse
  }
  return 'MCP'
})

const openMcpSelector = () => {
  mcpSelectorVisible.value = true
}

const clearMcpConfig = (e: Event) => {
  e.stopPropagation()
  emit('update:mcpConfig', '{}')
  emit('save-mcp-config', '{}')
}
</script>

<style scoped>
:deep(.el-textarea__inner) {
  box-shadow: none !important;
  background-color: transparent !important;
  color: var(--ep-text-color-primary) !important;
}

.editor-container {
  height: calc(100% - 32px); /* Adjust based on your layout */
  /* Optional: Add more styles as needed */
}
</style>
