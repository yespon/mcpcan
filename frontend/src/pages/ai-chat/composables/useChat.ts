import type { ChatMessage, AIModel } from '../types'

export function useChat() {
  const messages = ref<ChatMessage[]>([
    {
      id: '1',
      role: 'user',
      content: '1 + 1 = 2',
      timestamp: Date.now() - 10000,
    },
    {
      id: '2',
      role: 'assistant',
      content:
        "That's a straightforward arithmetic problem! Is there anything else I can help you with?",
      timestamp: Date.now(),
    },
  ])

  const models = ref<AIModel[]>([
    { id: 'claude-3-haiku', name: 'Claude Haiku 4.5 (Free)', provider: 'Anthropic' },
    { id: 'gpt-4o', name: 'GPT-4o', provider: 'OpenAI' },
  ])

  const currentModel = ref<string>('claude-3-haiku')

  const addMessage = (content: string, role: 'user' | 'assistant' = 'user') => {
    messages.value.push({
      id: Date.now().toString(),
      role,
      content,
      timestamp: Date.now(),
    })
  }

  const addCustomModel = (model: AIModel) => {
    models.value.push({ ...model, isCustom: true })
    currentModel.value = model.id
  }

  return {
    messages,
    models,
    currentModel,
    addMessage,
    addCustomModel,
  }
}
