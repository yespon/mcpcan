<template>
  <div class="h-full flex overflow-hidden">
    <el-splitter class="h-full">
      <el-splitter-panel
        :size="isSidebarOpen ? '300px' : '40px'"
        :resizable="false"
        :class="[
          'transition-all h-full duration-300 ease-in-out',
          !isSidebarOpen ? 'bg-[var(--ep-bg-color)] border-r border-[var(--ep-border-color)]' : '',
        ]"
        :min-size="isSidebarOpen ? '300px' : '40px'"
      >
        <!-- Left Sidebar (History & Collapsed State) -->
        <div class="h-full flex flex-col overflow-hidden">
          <!-- Expanded State Content -->
          <div
            v-if="isSidebarOpen"
            class="h-full flex flex-col transition-all duration-300 ease-in-out opacity-100"
          >
            <div
              class="flex items-center justify-between p-4 border-b border-[var(--ep-border-color)]"
            >
              <div class="flex flex-col gap-3 w-full">
                <div class="flex items-center justify-between">
                  <span class="font-medium">{{ t('aiChat.history') }}</span>
                  <el-button
                    class="!rounded-xl !border-[var(--el-color-primary)] !text-[var(--el-color-primary)] hover:!bg-[var(--el-color-primary-light-9)]"
                    plain
                    size="small"
                    @click="handleNewChat"
                  >
                    <el-icon class="mr-2"><Plus /></el-icon> {{ t('aiChat.new') }}
                  </el-button>
                </div>
              </div>
            </div>
            <div class="flex-1 overflow-hidden flex flex-col">
              <SessionList
                :sessions="sessions"
                :current-session-id="currentSession?.id"
                :models="models"
                :supported-providers="supportedProviders"
                @select="loadSession"
                @delete="deleteSession"
                @rename="handleRenameSession"
              />
            </div>
          </div>

          <!-- Collapsed State Content -->
          <div
            v-else
            class="h-full flex flex-col items-center justify-center py-4 gap-4 transition-all duration-300 ease-in-out hover:bg-[var(--ep-fill-color-light)] cursor-pointer rounded"
            @click="isSidebarOpen = true"
          >
            <el-tooltip :content="t('aiChat.expand')" placement="right">
              <el-button text circle>
                <el-icon><Expand /></el-icon>
              </el-button>
            </el-tooltip>
          </div>
        </div>
      </el-splitter-panel>
      <el-splitter-panel :min-size="200">
        <!-- Main Chat Area -->
        <div class="h-full flex-1 flex flex-col min-w-0 relative bg-[var(--ep-bg-color-page)]">
          <!-- Header -->
          <div
            class="h-14 flex items-center justify-between px-6 border-b border-[var(--ep-border-color)] bg-[var(--ep-bg-color)] relative"
          >
            <div class="flex items-center gap-3 w-1/3">
              <el-button link @click="isSidebarOpen = !isSidebarOpen">
                <el-icon><Fold v-if="isSidebarOpen" /></el-icon>
              </el-button>
            </div>

            <div
              class="absolute left-1/2 -translate-x-1/2 flex items-center justify-center gap-2 cursor-pointer hover:bg-[var(--ep-fill-color-light)] px-3 py-1 rounded transition-colors group"
              @click="currentSession && handleRenameSession(currentSession)"
            >
              <div class="font-medium truncate max-w-sm select-none">
                {{ currentSession?.name || t('aiChat.newChat') }}
              </div>
              <el-icon
                v-if="currentSession"
                class="opacity-0 group-hover:opacity-100 transition-opacity text-[var(--ep-text-color-secondary)]"
              >
                <EditPen />
              </el-icon>
            </div>

            <div class="flex items-center justify-end space-x-2 w-1/3"></div>
          </div>

          <!-- Toggle Sidebar Button (Floating) - Removed as it's now in Header -->

          <!-- Messages Area -->
          <div
            class="flex-1 overflow-y-auto hide-scrollbar px-6 md:px-20 lg:px-40 py-6"
            ref="messagesContainer"
          >
            <ChatMessage v-for="msg in messages" :key="msg.id" :message="msg" />
            <div
              v-if="messages.length === 0"
              class="h-full flex flex-col items-center justify-center text-[var(--ep-text-color-placeholder)]"
            >
              <el-icon class="text-6xl mb-4"><ChatDotRound /></el-icon>
              <p>{{ t('aiChat.startConversation') }}</p>
            </div>
          </div>

          <!-- Input Area -->
          <div class="p-6 md:px-20 lg:px-40 pb-8 bg-[var(--ep-bg-color-page)]">
            <ChatInput
              v-model:currentModel="currentModel"
              v-model:currentTargetModel="currentTargetModel"
              v-model:mcpConfig="currentMcpConfig"
              v-model:systemPrompt="sessionSettings.systemPrompt"
              v-model:temperature="sessionSettings.temperature"
              :models="models"
              :supported-providers="supportedProviders"
              :disabled="isStreaming"
              :mcp-instances="mcpInstances"
              :mcp-loading="mcpLoading"
              :mcp-has-more="mcpHasMore"
              :session-id="currentSession?.id"
              @send="handleSend"
              @add-model="addCustomModel"
              @save-settings="handleSaveSettings"
              @model-change-confirmed="handleModelChangeConfirmed"
              @save-mcp-config="handleSaveMcpConfig"
              @load-mcp="handleLoadMcp"
            />
          </div>
        </div>
      </el-splitter-panel>
    </el-splitter>
    <!-- Create Session Dialog -->
    <el-dialog
      v-model="createSessionDialogVisible"
      :title="$t('aiChat.createSession')"
      width="500px"
      top="8vh"
    >
      <el-form label-position="top" size="large">
        <el-form-item :label="t('aiChat.sessionName')">
          <el-input
            v-model="newSessionForm.name"
            :placeholder="t('aiChat.sessionNamePlaceholder')"
          />
        </el-form-item>

        <el-form-item :label="t('aiChat.modelAccess')" required>
          <el-select
            v-model="newSessionForm.modelAccessID"
            :placeholder="t('aiChat.selectModelAccess')"
            class="w-full"
            filterable
            @change="handleModelAccessChange"
          >
            <el-option v-for="m in models" :key="m.id" :label="m.name" :value="m.id" />
          </el-select>
        </el-form-item>

        <el-form-item :label="t('aiChat.targetModel')" required>
          <el-select
            v-model="newSessionForm.modelName"
            :placeholder="t('aiChat.targetModelPlaceholder')"
            class="w-full"
            filterable
            allow-create
            default-first-option
          >
            <el-option v-for="m in targetModelOptions" :key="m" :label="m" :value="m" />
          </el-select>
          <div class="text-xs text-[var(--ep-text-color-secondary)] mt-1">
            {{ t('aiChat.targetModelHint') }}
          </div>
        </el-form-item>

        <el-form-item :label="t('aiChat.systemPrompt')">
          <el-input
            v-model="newSessionForm.systemPrompt"
            type="textarea"
            :rows="3"
            :placeholder="t('aiChat.systemPromptPlaceholder')"
          />
        </el-form-item>

        <el-form-item :label="t('aiChat.temperature')">
          <div class="flex items-center gap-4 w-full">
            <el-slider
              v-model="newSessionForm.temperature"
              :min="0"
              :max="2"
              :step="0.1"
              :show-input-controls="false"
              class="flex-1 !mb-0"
              size="small"
            />
            {{ newSessionForm.temperature.toFixed(1) }}
          </div>
          <div class="text-xs text-[var(--ep-text-color-secondary)] mt-1">
            {{ t('aiChat.temperatureHint') }}
          </div>
        </el-form-item>

        <el-form-item :label="t('aiChat.mcpConfig')">
          <el-input
            v-model="newSessionForm.toolsConfig"
            type="textarea"
            :rows="3"
            placeholder="{ 'mcpServers': ... }"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="createSessionDialogVisible = false">{{
            t('aiChat.cancel')
          }}</el-button>
          <el-button type="primary" @click="submitCreateSession" :loading="creatingSession">
            {{ t('aiChat.create') }}
          </el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import ChatMessage from './components/ChatMessage.vue'
import ChatInput from './components/ChatInput.vue'
import SessionList from './components/SessionList.vue'
import { useChat } from './composables/useChat'
import {
  Fold,
  Expand,
  ChatDotRound,
  Plus,
  ArrowLeft,
  ArrowRight,
  EditPen,
} from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'
import type { AiSession, ChatAttachment } from './types'

const { t } = useI18n()

const {
  messages,
  models,
  sessions,
  // currentModel is less relevant for globally creating sessions now,
  // but we might use it for default values
  currentModel,
  currentTargetModel,
  currentSession,
  addMessage,
  addCustomModel,
  loadSession,
  createNewSession,
  updateSessionName,
  deleteSession,
  supportedProviders,
  fetchSupportedProviders,
  uploadFile,
  updateSessionSettings,
  mcpInstances,
  fetchMcpInstances,
  mcpHasMore,
  mcpLoading,
  isStreaming,
  fetchModels,
  fetchSessions,
} = useChat()

const isSidebarOpen = ref(true)
const messagesContainer = ref<HTMLElement | null>(null)

const sessionSettings = reactive({
  systemPrompt: '',
  temperature: 0.7,
  toolsConfig: '',
})

// Expose a currentMcpConfig ref for v-model:mcpConfig in the template and keep it in sync
const currentMcpConfig = ref(sessionSettings.toolsConfig)

// Update settings when session changes
watch(
  currentSession,
  (sess) => {
    if (sess) {
      sessionSettings.systemPrompt = sess.systemPrompt || ''
      sessionSettings.temperature = sess.temperature !== undefined ? sess.temperature : 0.7
      sessionSettings.toolsConfig = sess.toolsConfig || ''
    } else {
      // defaults for new session
      sessionSettings.systemPrompt = ''
      sessionSettings.temperature = 0.7
      sessionSettings.toolsConfig = ''
    }

    // ensure the mcp config ref is updated whenever sessionSettings.toolsConfig changes on session switch
    currentMcpConfig.value = sessionSettings.toolsConfig
  },
  { immediate: true },
)

// Keep sessionSettings.toolsConfig in sync when the child component updates currentMcpConfig
watch(currentMcpConfig, (val) => {
  sessionSettings.toolsConfig = val
})

const handleNewChat = () => {
  currentSession.value = null
  messages.value = []
}

const handleModelChangeConfirmed = async (modelName: string, accessId: string) => {
  if (currentSession.value) {
    // Current session exists, update it
    try {
      await updateSessionSettings(currentSession.value.id, {
        modelName: modelName,
        modelAccessID: parseInt(accessId),
      })
    } catch (e) {
      ElMessage.error(t('aiChat.failedToUpdateModel'))
    }
  }
}

const handleSaveMcpConfig = async () => {
  if (currentSession.value) {
    await updateSessionSettings(currentSession.value.id, {
      toolsConfig: sessionSettings.toolsConfig,
    })
  }
}

const handleSaveSettings = async () => {
  if (currentSession.value) {
    await updateSessionSettings(currentSession.value.id, {
      systemPrompt: sessionSettings.systemPrompt,
      temperature: sessionSettings.temperature,
      toolsConfig: sessionSettings.toolsConfig,
    })
  }
}

const handleSend = async (content: string, attachments: ChatAttachment[] = []) => {
  // If no session, these settings will be used to create one
  addMessage(content, 'user', attachments, undefined, undefined, undefined, {
    systemPrompt: sessionSettings.systemPrompt,
    temperature: sessionSettings.temperature,
    toolsConfig: sessionSettings.toolsConfig,
  })
}

// Auto-scroll to bottom when messages change
watch(
  messages,
  () => {
    nextTick(() => {
      if (messagesContainer.value) {
        messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
      }
    })
  },
  { deep: true },
)

const createSessionDialogVisible = ref(false)
const creatingSession = ref(false)

const newSessionForm = reactive({
  name: '',
  modelAccessID: '',
  modelName: '',
  systemPrompt: '',
  temperature: 0.7,
  toolsConfig: '{}',
})

const targetModelOptions = computed(() => {
  if (!newSessionForm.modelAccessID) return []

  const modelAccess = models.value.find((m) => m.id === newSessionForm.modelAccessID)
  if (!modelAccess) return []

  // 1. If Access has configured allowedModels, strictly use them
  if (modelAccess.allowedModels && modelAccess.allowedModels.length > 0) {
    return modelAccess.allowedModels
  }

  // 2. Otherwise, return all models from the provider
  if (!modelAccess.provider) return []

  const provider = supportedProviders.value.find((p) => p.id === modelAccess.provider)
  return provider ? provider.models : []
})

const handleModelAccessChange = () => {
  // Reset target model when access changes to prevent invalid selection
  newSessionForm.modelName = ''

  const options = targetModelOptions.value
  // Auto-select first available if any
  if (options.length > 0) {
    newSessionForm.modelName = options[0]
  }
}

const openCreateSessionDialog = () => {
  // Reset form
  newSessionForm.name = ''
  newSessionForm.systemPrompt = ''
  newSessionForm.temperature = 0.7
  newSessionForm.toolsConfig = '{}'
  newSessionForm.modelName = ''

  // Default model access selection logic (try current, or first available)
  if (currentModel.value) {
    newSessionForm.modelAccessID = currentModel.value
  } else if (models.value.length > 0) {
    newSessionForm.modelAccessID = models.value[0].id
  }

  handleModelAccessChange() // Update target model options & default selection

  createSessionDialogVisible.value = true
}

const submitCreateSession = async () => {
  if (!newSessionForm.modelAccessID) {
    ElMessage.warning(t('aiChat.pleaseSelectModelAccess'))
    return
  }
  if (!newSessionForm.modelName) {
    ElMessage.warning(t('aiChat.pleaseSelectTargetModel'))
    return
  }

  // Validate JSON config
  let parsedTools = '{}'
  try {
    if (newSessionForm.toolsConfig && newSessionForm.toolsConfig.trim()) {
      JSON.parse(newSessionForm.toolsConfig) // Check validity
      parsedTools = newSessionForm.toolsConfig
    }
  } catch (e) {
    ElMessage.error(t('aiChat.invalidMcpConfig'))
    return
  }

  creatingSession.value = true
  try {
    const sessionName = newSessionForm.name.trim() || `Session ${new Date().toLocaleTimeString()}`

    await createNewSession({
      name: sessionName,
      modelAccessID: parseInt(newSessionForm.modelAccessID),
      modelName: newSessionForm.modelName,
      systemPrompt: newSessionForm.systemPrompt,
      temperature: newSessionForm.temperature,
      toolsConfig: parsedTools,
    })

    createSessionDialogVisible.value = false
  } catch (e: any) {
    ElMessage.error(e.message || t('aiChat.failedToCreateSession'))
  } finally {
    creatingSession.value = false
  }
}

// Initial data fetch
onMounted(async () => {
  await fetchSupportedProviders()
  await fetchModels()
  await fetchSessions()
  // Ensure we select the first session if available and no current session
  if (sessions.value.length > 0 && !currentSession.value) {
    loadSession(sessions.value[0].id)
  } else if (!currentSession.value) {
    // Or open create dialog if needed? No, just wait user action.
    // Maybe collapse sidebar on mobile by default?
  }
})

const handleLoadMcp = async (page: number, append: boolean, query: string = '') => {
  await fetchMcpInstances(page, 20, query, append)
}

const handleRenameSession = async (session: AiSession) => {
  try {
    const { value } = await ElMessageBox.prompt(
      t('aiChat.enterNewName'),
      t('aiChat.renameSession'),
      {
        confirmButtonText: 'OK',
        cancelButtonText: t('aiChat.cancel'),
        inputValue: session.name,
        inputPattern: /\S/,
        inputErrorMessage: t('aiChat.nameEmpty'),
      },
    )

    if (value && value.trim() !== session.name) {
      await updateSessionName(session.id, value.trim())
    }
  } catch {
    // cancelled
  }
}
</script>
