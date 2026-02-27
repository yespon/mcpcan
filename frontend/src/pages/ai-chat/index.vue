<template>
  <div class="h-full flex overflow-hidden">
    <el-splitter class="h-full">
      <el-splitter-panel
        :size="isSidebarOpen ? '20%' : '40px'"
        :resizable="false"
        :class="[
          'transition-all h-full duration-300 ease-in-out',
          !isSidebarOpen ? 'bg-[var(--ep-bg-color)] border-r border-[var(--ep-border-color)]' : '',
        ]"
        :min-size="isSidebarOpen ? '200px' : '40px'"
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
              <div class="flex items-center gap-2">
                <span class="font-medium">Chat History</span>
                <el-button type="primary" link size="small" @click="openCreateSessionDialog">
                  <el-icon><Plus /></el-icon> New
                </el-button>
              </div>
              <el-button link @click="isSidebarOpen = false">
                <el-icon><Fold /></el-icon>
              </el-button>
            </div>
            <div class="flex-1 overflow-hidden flex flex-col">
              <SessionList
                :sessions="sessions"
                :current-session-id="currentSession?.id"
                @select="loadSession"
                @delete="deleteSession"
                @rename="handleRenameSession"
              />
            </div>
          </div>

          <!-- Collapsed State Content -->
          <div
            v-else
            class="h-full flex flex-col items-center justify-center py-4 gap-4 transition-all duration-300 ease-in-out"
          >
            <el-tooltip content="Expand" placement="right">
              <el-button text circle @click="isSidebarOpen = true">
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
            <div class="flex items-center gap-3 w-1/3"></div>

            <div
              class="absolute left-1/2 -translate-x-1/2 flex items-center justify-center gap-2 cursor-pointer hover:bg-[var(--ep-fill-color-light)] px-3 py-1 rounded transition-colors group"
              @click="currentSession && handleRenameSession(currentSession)"
            >
              <div class="font-medium truncate max-w-sm select-none">
                {{ currentSession?.name || 'New Chat' }}
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

          <!-- Toggle Sidebar Button (Floating) - Only visible when sidebar is open, as per requirement -->
          <div
            v-if="isSidebarOpen"
            class="absolute top-1/2 -translate-y-1/2 z-10 transition-all duration-300"
            :class="[isSidebarOpen ? 'left-0' : 'left-0']"
          >
            <el-tooltip content="Close sidebar" placement="right" :show-after="500">
              <div
                class="cursor-pointer bg-[var(--ep-bg-color)] border border-[var(--ep-border-color)] border-l-0 rounded-r-md py-4 pr-1 pl-0.5 shadow-sm hover:bg-[var(--ep-bg-color-page)] text-[var(--ep-text-color-secondary)] hover:text-[var(--ep-text-color-primary)] transition-colors flex items-center justify-center opacity-0 hover:opacity-100"
                @click="isSidebarOpen = !isSidebarOpen"
                style="width: 24px; height: 60px"
              >
                <el-icon class="text-xs">
                  <ArrowLeft />
                </el-icon>
              </div>
            </el-tooltip>
          </div>

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
              <p>Start a conversation</p>
            </div>
          </div>

          <!-- Input Area -->
          <div class="p-6 md:px-20 lg:px-40 pb-8 bg-[var(--ep-bg-color-page)]">
            <ChatInput
              v-model:currentModel="currentModel"
              :models="models"
              :supported-providers="supportedProviders"
              :disabled="!currentSession"
              @send="handleSend"
              @add-model="addCustomModel"
            />
          </div>
        </div>
      </el-splitter-panel>
    </el-splitter>
    <!-- Create Session Dialog -->
    <el-dialog
      v-model="createSessionDialogVisible"
      title="Create New Session"
      width="500px"
      top="8vh"
    >
      <el-form label-position="top" size="large">
        <el-form-item label="Session Name">
          <el-input v-model="newSessionForm.name" placeholder="e.g. Code Helper (Optional)" />
        </el-form-item>

        <el-form-item label="Model Access" required>
          <el-select
            v-model="newSessionForm.modelAccessID"
            placeholder="Select Model Access"
            class="w-full"
            filterable
            @change="handleModelAccessChange"
          >
            <el-option v-for="m in models" :key="m.id" :label="m.name" :value="m.id" />
          </el-select>
        </el-form-item>

        <el-form-item label="Target Model" required>
          <el-select
            v-model="newSessionForm.modelName"
            placeholder="Select or enter model ID"
            class="w-full"
            filterable
            allow-create
            default-first-option
          >
            <el-option v-for="m in targetModelOptions" :key="m" :label="m" :value="m" />
          </el-select>
          <div class="text-xs text-[var(--ep-text-color-secondary)] mt-1">
            Specific model ID to use (e.g. gpt-4)
          </div>
        </el-form-item>

        <el-form-item label="System Prompt">
          <el-input
            v-model="newSessionForm.systemPrompt"
            type="textarea"
            :rows="3"
            placeholder="You are a helpful assistant..."
          />
        </el-form-item>

        <el-form-item label="Temperature">
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
            Higher values make output more random, lower values more deterministic.
          </div>
        </el-form-item>

        <el-form-item label="MCP Config (JSON)">
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
          <el-button @click="createSessionDialogVisible = false">Cancel</el-button>
          <el-button type="primary" @click="submitCreateSession" :loading="creatingSession">
            Create
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

const {
  messages,
  models,
  sessions,
  // currentModel is less relevant for globally creating sessions now,
  // but we might use it for default values
  currentModel,
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
} = useChat()

const isSidebarOpen = ref(true)
const messagesContainer = ref<HTMLElement | null>(null)

onMounted(() => {
  fetchSupportedProviders()
})

const handleSend = async (content: string, file?: File) => {
  addMessage(content, 'user', [], undefined, undefined, file)
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
    ElMessage.warning('Please select a Model Access')
    return
  }
  if (!newSessionForm.modelName) {
    ElMessage.warning('Please select a Target Model')
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
    ElMessage.error('Invalid MCP Config JSON')
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
    ElMessage.error(e.message || 'Failed to create session')
  } finally {
    creatingSession.value = false
  }
}

const handleRenameSession = async (session: { id: number; name: string }) => {
  try {
    const { value } = await ElMessageBox.prompt('Enter new session name', 'Rename Session', {
      confirmButtonText: 'OK',
      cancelButtonText: 'Cancel',
      inputValue: session.name,
      inputPattern: /\S/,
      inputErrorMessage: 'Name cannot be empty',
    })

    if (value && value.trim() !== session.name) {
      await updateSessionName(session.id, value.trim())
    }
  } catch {
    // cancelled
  }
}
</script>
