export interface ChatMessage {
  id: string
  role: 'user' | 'assistant' | 'system'
  content: string
  timestamp: number
}

export interface AIModel {
  id: string
  name: string
  provider: string
  isCustom?: boolean
}
