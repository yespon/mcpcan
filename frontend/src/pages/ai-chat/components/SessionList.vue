<template>
  <div class="flex flex-col h-full bg-[var(--ep-bg-color-page)]">
    <!-- Session List -->
    <div class="flex-1 overflow-y-auto p-2 space-y-2 custom-scrollbar">
      <div
        v-for="session in sessions"
        :key="session.id"
        class="group flex flex-col p-3 rounded-lg cursor-pointer text-sm transition-all duration-200 border border-[var(--ep-border-color)] bg-[var(--ep-bg-color)] shadow-sm hover:shadow-md"
        :class="[
          currentSessionId === session.id
            ? 'border-[var(--el-color-primary)] ring-1 ring-[var(--el-color-primary)]'
            : 'hover:border-[var(--el-color-primary-light-5)]',
        ]"
        @click="$emit('select', session.id)"
      >
        <div class="flex items-center justify-between mb-2">
          <div class="flex items-center gap-2 truncate font-medium flex-1">
            <el-icon class="text-base shrink-0 text-[var(--el-color-primary)]"
              ><ChatDotRound
            /></el-icon>
            <span class="truncate">{{ session.name || t('aiChat.newChat') }}</span>
          </div>

          <div class="opacity-0 group-hover:opacity-100 transition-opacity" @click.stop>
            <el-dropdown trigger="click" @command="(cmd: string) => handleCommand(cmd, session)">
              <el-icon
                class="text-[var(--ep-text-color-secondary)] hover:text-[var(--el-color-primary)] cursor-pointer"
              >
                <MoreFilled />
              </el-icon>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="rename">
                    <el-icon><Edit /></el-icon> {{ t('aiChat.rename') }}
                  </el-dropdown-item>
                  <el-dropdown-item command="delete" divided class="!text-red-500">
                    <el-icon><Delete /></el-icon> {{ t('aiChat.delete') }}
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </div>

        <div class="text-xs text-[var(--ep-text-color-secondary)] grid grid-cols-4 gap-y-1 gap-x-2">
          <div class="truncate flex items-center gap-1 col-span-3">
            <template v-if="getModelIcon(session)">
              <img
                :src="getModelIcon(session)"
                class="w-4 h-4 object-contain"
                :alt="session.modelName"
              />
            </template>
            <el-icon v-else><Cpu /></el-icon>
            <span class="truncate">{{ session.modelName || '-' }}</span>
          </div>
          <div class="truncate flex items-center gap-1 col-span-1">
            <el-icon><Odometer /></el-icon> {{ session.temperature ?? '-' }}
          </div>

          <div class="col-span-4 flex items-center gap-1 truncate mt-2" @click.stop>
            <el-icon>
              <i class="icon iconfont MCP-MCPshili"></i>
            </el-icon>
            <el-popover
              placement="right"
              :width="320"
              trigger="hover"
              popper-class="!rounded-lg"
              v-if="session.toolsConfig && session.toolsConfig !== '{}'"
            >
              <template #reference>
                <span class="truncate border-b cursor-pointer hover:text-[var(--el-color-primary)]">
                  MCP ({{ getMcpCount(session.toolsConfig) }})
                </span>
              </template>
              <div class="flex flex-col gap-2">
                <div
                  class="text-xs font-bold text-[var(--ep-text-color-secondary)] uppercase tracking-wider"
                >
                  {{ t('aiChat.mcpInstances') }}
                </div>
                <div class="flex flex-wrap gap-1.5">
                  <el-tag
                    v-for="(name, idx) in getMcpInstanceNames(session.toolsConfig)"
                    :key="idx"
                    size="small"
                    effect="plain"
                    type="primary"
                    class="!rounded-md"
                  >
                    {{ name }}
                  </el-tag>
                </div>
              </div>
            </el-popover>
            <span v-else class="text-[var(--ep-text-color-placeholder)]">MCP (0)</span>
          </div>
        </div>
      </div>

      <div
        v-if="sessions.length === 0"
        class="text-center py-8 text-[var(--ep-text-color-placeholder)] text-xs"
      >
        {{ t('aiChat.noChatHistory') }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import {
  ChatDotRound,
  Delete,
  MoreFilled,
  Edit,
  Cpu,
  Odometer,
  Tools,
} from '@element-plus/icons-vue'
import type { AiSession, AIModel, SupportedProvider } from '../types'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const props = defineProps<{
  sessions: AiSession[]
  currentSessionId?: number
  models?: AIModel[]
  supportedProviders?: SupportedProvider[]
}>()

const getModelIcon = (session: AiSession) => {
  if (!session.modelAccessID || !props.models || !props.supportedProviders) return ''

  // Find the model access configuration
  const modelAccess = props.models.find((m) => m.id === session.modelAccessID.toString())
  if (!modelAccess) return ''

  // Find the provider
  const provider = props.supportedProviders.find((p) => p.id === modelAccess.provider)
  return provider ? provider.iconUrl : ''
}

const emit = defineEmits<{
  (e: 'create'): void
  (e: 'select', id: number): void
  (e: 'delete', id: number): void
  (e: 'rename', session: AiSession): void
}>()

const handleCommand = (cmd: string, session: AiSession) => {
  if (cmd === 'rename') {
    emit('rename', session)
  } else if (cmd === 'delete') {
    emit('delete', session.id)
  }
}

const getMcpCount = (configStr?: string) => {
  if (!configStr || configStr === '{}') return 0
  try {
    const config = JSON.parse(configStr)
    if (config.mcpServers) {
      return Object.keys(config.mcpServers).length
    }
  } catch (e) {
    return 0
  }
  return 0
}

const getMcpInstanceNames = (configStr?: string): string[] => {
  if (!configStr || configStr === '{}') return []
  try {
    const config = JSON.parse(configStr)
    if (config.mcpServers) {
      return Object.values(config.mcpServers).map((server: any) => {
        return server.instanceName || t('aiChat.unknown')
      })
    }
  } catch (e) {
    return []
  }
  return []
}
</script>

<style scoped>
.custom-scrollbar::-webkit-scrollbar {
  width: 4px;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background-color: var(--ep-border-color);
  border-radius: 4px;
}
</style>
