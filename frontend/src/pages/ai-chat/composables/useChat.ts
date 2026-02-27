import {
  ChatAPI,
  type CreateSessionRequest,
  type CreateModelAccessRequest,
  type ChatAttachment,
  type McpProfile,
} from '@/api/agent'
import type { ChatMessage, AIModel, AiSession, SupportedProvider } from '../types/index'
import { ElMessage } from 'element-plus'
import baseConfig from '@/config/base_config.ts'
import { Storage } from '@/utils/storage'

export function useChat() {
  const messages = ref<ChatMessage[]>([])
  const models = ref<AIModel[]>([])
  const supportedProviders = ref<SupportedProvider[]>([])
  const sessions = ref<AiSession[]>([])
  const currentModel = ref<string>('')
  const currentSession = ref<AiSession | null>(null)
  const isStreaming = ref(false)

  // Fetch supported providers (needed for model options)
  const fetchSupportedProviders = async () => {
    try {
      const res = await ChatAPI.getSupportedModels()
      if (res && res.providers) {
        supportedProviders.value = res.providers
      }
    } catch (error) {
      console.error('Failed to fetch supported providers:', error)
    }
  }

  // Fetch available models (Model Access Configurations)
  const fetchModels = async () => {
    try {
      const response = await ChatAPI.listModelAccess({ page: 1, pageSize: 100 })
      if (response && response.list) {
        models.value = response.list.map((m: any) => ({
          id: m.id.toString(), // Use string ID for frontend consistency
          name: m.name,
          provider: m.provider,
          description: m.modelName,
          // Support both array and JSON string for allowedModels
          allowedModels: Array.isArray(m.allowedModels)
            ? m.allowedModels
            : m.allowedModels
              ? JSON.parse(m.allowedModels)
              : [],
        }))
        if (models.value.length > 0 && !currentModel.value) {
          currentModel.value = models.value[0].id
        }
      }
    } catch (error) {
      console.error('Failed to fetch models:', error)
      ElMessage.error('Failed to load AI models')
    }
  }

  // Fetch session list
  const fetchSessions = async () => {
    try {
      const res = await ChatAPI.listSessions({ page: 1, pageSize: 100 })
      if (res && res.list) {
        sessions.value = res.list
      }
    } catch (error) {
      console.error('Failed to fetch sessions:', error)
    }
  }

  // Load session history
  const loadSession = async (sessionId: number) => {
    try {
      const session = sessions.value.find((s) => s.id === sessionId)
      if (session) {
        currentSession.value = session
        // Load messages
        const res = await ChatAPI.getSessionMessages(sessionId, { page: 1, pageSize: 100 })
        if (res && res.list) {
          // Map backend messages to frontend format
          messages.value = res.list
            .map((m: any) => ({
              id: m.id.toString(),
              role: m.role || 'user',
              content: m.content || '',
              timestamp: m.createTime ? m.createTime * 1000 : Date.now(),
              // Add Usage info if available
              usage: m.totalTokens
                ? {
                    promptTokens: m.promptTokens,
                    completionTokens: m.completionTokens,
                    totalTokens: m.totalTokens,
                  }
                : undefined,
              // Parse tools if needed
              toolCalls: m.toolCalls ? JSON.parse(m.toolCalls) : undefined,
              tools: m.toolCalls && m.toolCalls !== '[]' ? JSON.parse(m.toolCalls) : [],
            }))
            .reverse() // Backend might return latest first? Check sorting. Assuming list is chrono or reverse chrono.
          // Usually chat messages are stored chrono. If backend returns reverse chrono (latest first), we need to reverse.
          // Let's assume chrono for now or adjust based on observation.
        } else {
          messages.value = []
        }
      }
    } catch (error) {
      console.error('Failed to load session:', error)
      ElMessage.error('Failed to load chat history')
    }
  }

  // Create new session with full config
  const createNewSession = async (config?: Partial<CreateSessionRequest>) => {
    try {
      let req: CreateSessionRequest

      if (config && config.modelAccessID && config.modelName) {
        req = {
          name: config.name || 'New Chat',
          modelAccessID: config.modelAccessID,
          modelName: config.modelName,
          maxContext: config.maxContext || 10,
          toolsConfig: config.toolsConfig || '{}',
          temperature: config.temperature,
          systemPrompt: config.systemPrompt,
        }
      } else {
        // Default fallback logic
        const modelId = parseInt(currentModel.value) || 0
        let selectedModel = models.value.find((m) => m.id === currentModel.value)
        if (!selectedModel && models.value.length > 0) {
          selectedModel = models.value[0]
          currentModel.value = selectedModel.id
        }

        if (!selectedModel) {
          ElMessage.warning('Please select a model first')
          return
        }
        req = {
          name: 'New Chat',
          modelAccessID: parseInt(selectedModel.id),
          modelName: selectedModel.description || selectedModel.name, // Use description/modelName as fallback target model
        }
      }

      const res = await ChatAPI.createSession(req)
      if (res && res.session) {
        sessions.value.unshift(res.session)
        currentSession.value = res.session
        messages.value = []

        // Add optional welcome message
        messages.value.push({
          id: Date.now().toString(),
          role: 'assistant',
          content: `Session "${req.name}" created.`,
          timestamp: Date.now(),
        })
      }
    } catch (error) {
      console.error('Failed to create session:', error)
      ElMessage.error('Failed to create new chat')
    }
  }

  // Delete session
  const deleteSession = async (id: number) => {
    try {
      await ChatAPI.deleteSession(id)
      sessions.value = sessions.value.filter((s) => s.id !== id)
      if (currentSession.value?.id === id) {
        currentSession.value = null
        messages.value = []
        // Optionally load next available session
        if (sessions.value.length > 0) {
          loadSession(sessions.value[0].id)
        }
      }
      ElMessage.success('Chat deleted')
    } catch (error) {
      console.error('Failed to delete session:', error)
      ElMessage.error('Failed to delete chat')
    }
  }

  // Create or get session (Auto-create logic used by sendMessage)
  const initSession = async () => {
    if (!currentSession.value) {
      await createNewSession()
    }
  }

  const addMessage = async (
    content: string,
    role: 'user' | 'assistant' = 'user',
    attachments: ChatAttachment[] = [],
    tools?: string,
    mcpProfile?: McpProfile,
    file?: File,
  ) => {
    if (!content.trim() && (!attachments || attachments.length === 0) && !file) return

    // Add user message immediately
    const userMsg: ChatMessage = {
      id: Date.now().toString(),
      role: 'user',
      content:
        content +
        (file ? `\n[File: ${file.name}]` : '') +
        (attachments.length > 0 && !file ? `\n[File: ${attachments[0].name}]` : ''),
      timestamp: Date.now(),
      attachments,
    }
    messages.value.push(userMsg)

    if (role === 'user') {
      await sendMessageToBackend(content, attachments, tools, mcpProfile, file)
    }
  }

  const sendMessageToBackend = async (
    content: string,
    attachments: ChatAttachment[] = [],
    tools?: string,
    mcpProfile?: McpProfile,
    file?: File,
  ) => {
    if (!currentSession.value) {
      await initSession()
    }

    if (!currentSession.value) {
      ElMessage.error('No active session')
      return
    }

    isStreaming.value = true
    const assistantMsgId = Date.now().toString()
    const assistantMsg: ChatMessage = {
      id: assistantMsgId,
      role: 'assistant',
      content: '',
      timestamp: Date.now(),
      isStreaming: true,
      tools: [], // Initialize tools array
    }
    messages.value.push(assistantMsg)

    // Get the reactive version of the message from the array
    const reactiveMsg = messages.value.find((m) => m.id === assistantMsgId) || assistantMsg

    try {
      // Upload file if exists
      if (file) {
        try {
          const res = await uploadFile(file)
          if (res && res.url) {
            attachments.push({
              type: 'image',
              name: file.name,
              url: res.url,
            })
          }
        } catch (e) {
          console.error(e)
          throw new Error('File upload failed')
        }
      }

      // Use fetch for streaming response validation
      // Construct URL manually as axios wrapper might handle response differently
      // Need to check backend streaming format (NDJSON or similar)

      await fetchStreamResponse(
        currentSession.value.id,
        content,
        reactiveMsg,
        attachments,
        tools,
        mcpProfile,
      )
    } catch (err: any) {
      console.error('Chat error:', err)
      reactiveMsg.content += `\n[Error: ${err.message || 'Failed to get response'}]`
      ElMessage.error('Failed to send message')
    } finally {
      isStreaming.value = false
      reactiveMsg.isStreaming = false
    }
  }

  const fetchStreamResponse = async (
    sessionId: number,
    content: string,
    msgRef: ChatMessage,
    attachments: ChatAttachment[] = [],
    tools?: string,
    mcpProfile?: McpProfile,
  ) => {
    const token = Storage.get('token')
    // Get API base path from config. Default is '/api'
    const apiBase = (window as any).__APP_CONFIG__?.API_BASE || '/api'

    // Use relative path to allow proxy to work correctly in dev mode
    // and relative path in production
    const normalizedApiBase = apiBase.startsWith('/') ? apiBase : `/${apiBase}`
    const url = `${normalizedApiBase}/market/ai/sessions/${sessionId}/chat`

    try {
      console.log('Starting chat stream to:', url)
      const response = await fetch(url, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: token ? `Bearer ${token}` : '',
        },
        body: JSON.stringify({
          sessionID: sessionId,
          content: content,
          tools,
          mcpProfile,
          attachments,
        }),
      })

      if (!response.ok) {
        const errorText = await response.text()
        console.error('Stream Error Body:', errorText)
        throw new Error(`HTTP error! status: ${response.status}: ${response.statusText}`)
      }

      const reader = response.body?.getReader()
      if (!reader) return

      const decoder = new TextDecoder()
      let buffer = ''

      try {
        while (true) {
          const { done, value } = await reader.read()
          if (done) {
            break
          }

          const chunk = decoder.decode(value, { stream: true })
          buffer += chunk

          // Split by double newline which is standard for SSE messages
          const parts = buffer.split('\n\n')

          // Keep the last part in buffer as it might be incomplete
          buffer = parts.pop() || ''

          for (const part of parts) {
            const lines = part.split('\n')
            for (const line of lines) {
              const trimmedLine = line.trim()
              if (!trimmedLine) continue

              if (trimmedLine.startsWith('data: ')) {
                const jsonStr = trimmedLine.slice(6)
                if (jsonStr.trim() === '[DONE]') {
                  continue
                }
                try {
                  console.log('Stream chunk:', jsonStr) // Add log
                  const json = JSON.parse(jsonStr)
                  // Handle different message types match chat_test.html logic
                  if (json.type === 'text') {
                    msgRef.content += json.content
                  } else if (json.type === 'tool_start') {
                    // console.log(`Starting ${json.content}`)
                  } else if (json.type === 'tool_result') {
                    if (!msgRef.tools) msgRef.tools = []
                    msgRef.tools.push({
                      name: 'Tool',
                      args: '?',
                      result: json.content,
                    })
                  } else if (json.type === 'error') {
                    ElMessage.error(json.content)
                    msgRef.content += `\n[System Error: ${json.content}]`
                  } else if (json.type === 'usage') {
                    try {
                      const usage =
                        typeof json.content === 'string' ? JSON.parse(json.content) : json.content
                      msgRef.usage = usage
                    } catch (e) {}
                  }
                } catch (e) {
                  console.error('JSON parse error:', e, jsonStr)
                  // Ignore parse errors for partial chunks
                }
              }
            }
          }
        }
        // Check for buffer residue that might not end with \n\n but is valid JSON
        if (buffer.trim().startsWith('data: ')) {
          const jsonStr = buffer.trim().slice(6)
          try {
            const json = JSON.parse(jsonStr)
            if (json.type === 'text' && json.content) {
              msgRef.content += json.content
            }
          } catch (e) {
            // ignore
          }
        }
      } catch (err: any) {
        console.error('Stream reading error:', err)
        msgRef.content += `\n[Error: ${err.message}]`
        throw err
      }
    } catch (error) {
      console.error('Stream fetch error:', error)
      msgRef.content += '\n[Network Error: Failed to receive complete response]'
      // Re-throw to handle error in component
      throw error
    }
  }

  const addCustomModel = async (model: AIModel) => {
    try {
      const createReq: CreateModelAccessRequest = {
        name: model.name,
        provider: model.provider,
        apiKey: (model as any).apiKey,
        baseUrl: (model as any).baseUrl,
        modelName: (model as any).modelName || model.description || model.name,
        allowedModels: JSON.stringify((model as any).allowedModels || []),
      }
      await ChatAPI.createModelAccess(createReq)
      ElMessage.success('Model added successfully')
      // Refresh models
      await fetchModels()
    } catch (error) {
      console.error('Failed to add custom model:', error)
      ElMessage.error('Failed to add custom model')
    }
  }

  const uploadFile = async (file: File) => {
    try {
      const formData = new FormData()
      formData.append('file', file)
      const res = await ChatAPI.uploadFile(formData)
      return res // expected { url: string, ... }
    } catch (error) {
      console.error('Failed to upload file:', error)
      throw error
    }
  }

  onMounted(() => {
    fetchModels()
    fetchSessions()
  })

  return {
    messages,
    models,
    sessions,
    currentModel,
    currentSession,
    isStreaming,
    addMessage,
    addCustomModel,
    fetchSessions,
    loadSession,
    createNewSession,
    deleteSession,
    supportedProviders, // Expose supportedProviders
    fetchSupportedProviders, // Expose fetchSupportedProviders
    uploadFile, // Expose uploadFile method
  }
}
