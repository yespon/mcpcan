import type { ChatAttachment } from '@/api/agent'
export type { ChatAttachment }

export interface ChatMessage {
  id: string
  role: 'user' | 'assistant' | 'system'
  content: string
  timestamp: number
  isStreaming?: boolean
  attachments?: ChatAttachment[]
  usage?: {
    totalTokens: number
    promptTokens?: number
    completionTokens?: number
  }
  toolCalls?: any
  tools?: {
    name: string
    args: string
    result: string
  }[]
}

export interface AIModel {
  id: string
  name: string
  provider: string
  isCustom?: boolean
  description?: string
  allowedModels?: string[]
}

export interface ModelInfo {
  id: string
  name: string
  description?: string
  contextLength?: number
  supportTools?: boolean
  supportSystemPrompt?: boolean
  supportTemperature?: boolean
  supportThinking?: boolean
  supportsVision?: boolean
  imageMimeTypes?: string[] | null
  maxImageSize?: number
  maxImageCount?: number
  supportsDocument?: boolean
  documentMimeTypes?: string[] | null
  maxDocumentSize?: number
  maxDocumentCount?: number
}

export interface SupportedProvider {
  id: string
  name: string
  models: string[]
  baseUrl: string
  iconUrl?: string
  modelInfos?: ModelInfo[]
}

export interface AiSession {
  id: number
  name: string
  modelAccessID: number
  modelName: string
  createTime: number
  updateTime: number
  systemPrompt?: string
  temperature?: number
  toolsConfig?: string
}
