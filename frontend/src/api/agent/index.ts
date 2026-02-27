import request from '@/utils/request'
import baseConfig from '@/config/base_config.ts'
import { update } from 'lodash-es'

export const AgentAPI = {
  // agent list
  list(params: TableData | null) {
    return request<any, any>({
      url: `/market/intelligent_access/list`,
      method: 'GET',
      params,
    })
  },
  // create agent
  create(data: CreateAgentRequest) {
    return request<any, any>({
      url: `/market/intelligent_access`,
      method: 'POST',
      data,
    })
  },
  // connection test
  connectionTest(data: CreateAgentRequest) {
    return request<any, any>({
      url: `/market/intelligent_access/test-connection`,
      method: 'POST',
      data,
    })
  },
  // delete agent platform
  delete(accessID: string) {
    return request<any, any>({
      url: `/market/intelligent_access/delete`,
      method: 'DELETE',
      data: { accessID },
    })
  },
  // update agent platform
  update(data: CreateAgentRequest) {
    return request<any, any>({
      url: `/market/intelligent_access/edit`,
      method: 'PUT',
      data,
    })
  },
  // get namespaces by platform
  getNamespaces(data: any) {
    return request<any, any>({
      url: `/market/intelligent_access/list-user-space`,
      method: 'POST',
      data,
    })
  },
  // create a sync task
  createSyncTask(data: any) {
    return request<any, any>({
      url: `/market/mcp_to_intelligent_task`,
      method: 'POST',
      data,
    })
  },
  // get task list
  taskList(params: any) {
    return request<any, any>({
      url: `/market/mcp_to_intelligent_task/list`,
      method: 'GET',
      params,
    })
  },
  // get task detail
  taskDetail(id: string) {
    return request<any, any>({
      url: `/market/mcp_to_intelligent_task/${id}`,
      method: 'GET',
    })
  },
  // cancel task
  cancelTask(id: string) {
    return request<any, any>({
      url: `/market/mcp_to_intelligent_task/${id}/cancel`,
      method: 'POST',
    })
  },
  // check N8N
  checkN8n(data: CreateAgentRequest) {
    return request<any, any>({
      url: `/market/intelligent_access/check-n8n`,
      method: 'POST',
      data,
    })
  },
  // install N8N plugin
  installPlugin(data: CreateAgentRequest) {
    return request<any, any>({
      url: `/market/intelligent_access/install-n8n-plugin`,
      method: 'POST',
      data,
      timeout: 300000,
    })
  },
}

export const ChatAPI = {
  // === Session Management ===
  createSession(data: CreateSessionRequest) {
    return request<any, any>({
      url: `/market/ai/sessions`,
      method: 'POST',
      data,
    })
  },
  updateSession(data: any | UpdateSessionRequest) {
    return request<any, any>({
      url: `/market/ai/sessions`,
      method: 'PUT',
      data,
    })
  },
  deleteSession(id: number) {
    return request<any, any>({
      url: `/market/ai/sessions/${id}`,
      method: 'DELETE',
    })
  },
  getSession(id: number) {
    return request<any, any>({
      url: `/market/ai/sessions/${id}`,
      method: 'GET',
    })
  },
  listSessions(params: { page: number; pageSize: number }) {
    return request<any, any>({
      url: `/market/ai/sessions`,
      method: 'GET',
      params,
    })
  },
  getSessionMessages(sessionID: number, params: { page: number; pageSize: number }) {
    return request<any, any>({
      url: `/market/ai/sessions/${sessionID}/messages`,
      method: 'GET',
      params,
    })
  },
  getSessionUsage(sessionID: number) {
    return request<any, any>({
      url: `/market/ai/sessions/${sessionID}/usage`,
      method: 'GET',
    })
  },
  // Chat stream is usually handled differently, but here's a basic post request if needed for non-stream or setup
  // For SSE/streaming, you might use fetch or specific logic in the component
  chat(sessionID: number, data: ChatRequest) {
    return request<any, any>({
      url: `/market/ai/sessions/${sessionID}/chat`,
      method: 'POST',
      data,
      responseType: 'stream', // Indicate stream response if axios supports it, mostly for handling
    })
  },
  uploadFile(data: FormData) {
    return request<any, any>({
      url: `/market/ai/files/upload`,
      method: 'POST',
      data,
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    })
  },

  // === Model Access Management ===
  createModelAccess(data: CreateModelAccessRequest) {
    return request<any, any>({
      url: `/market/ai/models`,
      method: 'POST',
      data,
    })
  },
  updateModelAccess(data: UpdateModelAccessRequest) {
    return request<any, any>({
      url: `/market/ai/models`,
      method: 'PUT',
      data,
    })
  },
  deleteModelAccess(id: number) {
    return request<any, any>({
      url: `/market/ai/models/${id}`,
      method: 'DELETE',
    })
  },
  getModelAccess(id: number) {
    return request<any, any>({
      url: `/market/ai/models/${id}`,
      method: 'GET',
    })
  },
  listModelAccess(params: { page: number; pageSize: number }) {
    return request<any, any>({
      url: `/market/ai/models`,
      method: 'GET',
      params,
    })
  },
  getSupportedModels() {
    return request<any, any>({
      url: `/market/ai/models/supported`,
      method: 'GET',
    })
  },
  getAvailableModels() {
    return request<any, any>({
      url: `/market/ai/models/available`,
      method: 'GET',
    })
  },
  testConnectionById(id: number) {
    return request<any, any>({
      url: `/market/ai/models/${id}/test`,
      method: 'POST',
    })
  },
  testConnectionNew(data: TestConnectionRequest) {
    return request<any, any>({
      url: `/market/ai/models/test`,
      method: 'POST',
      data,
    })
  },
}

// === Type Definitions ===

export interface CreateAgentRequest {
  accessID?: string
  accessName?: string
  accessType?: string
  dbHost?: string
  dbPort?: number
  dbUser?: string
  dbPassword?: string
  dbName?: string
  enterpriseId?: string
  subType?: string
  baseUrl?: string
  username?: string
  password?: string
}
export interface TableData {
  page: number
  pageSize: number
  [key: string]: any
}

export interface CreateSessionRequest {
  name: string
  modelAccessID: number
  toolsConfig?: string
  maxContext?: number
  modelName: string
  temperature?: number
  systemPrompt?: string
}

export interface UpdateSessionRequest {
  id: number
  name?: string
  toolsConfig?: string
  maxContext?: number
  modelAccessID?: number
  modelName?: string
  temperature?: number
  systemPrompt?: string
}

export interface ChatAttachment {
  type: string
  url: string
  id?: string
  name?: string
}

export interface McpProfile {
  instanceId: string
  includeTools?: string[]
  enableAll?: boolean
}

export interface ChatRequest {
  sessionID: number
  content: string
  tools?: string
  mcpProfile?: McpProfile
  attachments?: ChatAttachment[]
}

export interface SessionUsage {
  sessionId: number
  totalMessages: number
  promptTokens: number
  completionTokens: number
  totalTokens: number
}

export interface CreateModelAccessRequest {
  name: string
  provider: string
  apiKey: string
  baseUrl?: string
  modelName: string
  allowedModels?: any[]
}

export interface UpdateModelAccessRequest {
  id: number
  name?: string
  provider?: string
  apiKey?: string
  baseUrl?: string
  modelName?: string
  allowedModels?: any[]
}

export interface TestConnectionRequest {
  id: number
  modelName: string
}
