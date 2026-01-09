<template>
  <div class="tool-page h-full flex flex-col" v-loading="loading">
    <!-- Instance Detail Info Bar -->
    <el-card class="mb-4 text-sm" shadow="hover">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-4">
          <div class="flex items-center gap-2">
            <span class="font-bold text-lg">{{ instanceInfo.name || 'Unknown Instance' }}</span>
            <el-tag
              :type="activeOptions[instanceInfo.status as keyof typeof activeOptions]?.type"
              effect="dark"
              size="small"
              round
            >
              {{
                activeOptions[instanceInfo.status as keyof typeof activeOptions]?.label ||
                'Unknown status'
              }}
            </el-tag>
          </div>
          <el-divider direction="vertical" />
          <div class="text-gray-500">ID: {{ instanceInfo.instanceId }}</div>
          <el-divider direction="vertical" />
          <div class="text-gray-500">Base URL: {{ '----' }}</div>
        </div>
        <div>
          <el-button type="primary" link @click="refreshInstance">
            <el-icon class="mr-1"><Refresh /></el-icon> Refresh
          </el-button>
        </div>
      </div>
    </el-card>

    <!-- Main Splitter Area -->
    <div class="flex-1 overflow-hidden">
      <el-splitter class="h-full border rounded-lg bg-page">
        <!-- Panel 1: Tool List -->
        <el-splitter-panel size="350px" class="flex flex-col" :resizable="false">
          <div class="p-3 border-b bg-header font-medium flex justify-between">
            <span>Tools</span>
            <el-tag size="small" type="info">{{ toolList.length }}</el-tag>
          </div>
          <div class="p-2 border-b">
            <el-input
              v-model="keyword"
              placeholder="Search tools..."
              :prefix-icon="Search"
              clearable
            />
          </div>
          <el-scrollbar class="flex-1">
            <div
              v-for="tool in filteredTools"
              :key="tool.name"
              class="p-3 cursor-pointer tool-item transition-colors border-l-4 border-transparent"
              :class="{ 'active-tool': currentTool?.name === tool.name }"
              @click="handleSelectTool(tool)"
            >
              <div class="font-bold">{{ tool.name }}</div>
              <div class="text-xs mt-1 line-clamp-2">{{ tool.description }}</div>
            </div>
            <el-empty
              v-if="filteredTools.length === 0"
              description="No tools found"
              image-size="60"
            />
          </el-scrollbar>
        </el-splitter-panel>

        <!-- Panel 2: Input Parameters -->
        <el-splitter-panel class="flex flex-col relative">
          <div class="p-3 border-b bg-header font-medium flex justify-between items-center">
            <span>Input Parameters</span>
            <el-switch
              v-if="currentTool"
              v-model="jsonMode"
              active-text="JSON"
              inactive-text="Form"
              size="small"
            />
          </div>
          <el-scrollbar class="flex-1 p-4 px-2">
            <template v-if="currentTool">
              <div class="mb-4 mx-2">
                <div class="text-sm text-secondary mb-2">Description</div>
                <div class="bg-block p-2 rounded text-sm text-regular">
                  {{ currentTool.description || 'No description provided.' }}
                </div>
              </div>

              <!-- Form Mode -->
              <div v-if="!jsonMode && currentTool.inputSchema?.properties" class="mx-2">
                <div
                  v-for="(schema, key) in currentTool.inputSchema.properties"
                  :key="key"
                  class="mb-4"
                >
                  <div class="flex items-center justify-between mb-1">
                    <span class="text-sm font-bold text-regular">
                      {{ key }}
                      <span
                        v-if="isPropertyRequired(String(key), currentTool.inputSchema)"
                        class="text-red"
                        >*
                      </span>
                    </span>
                    <span class="text-xs text-secondary">{{ schema.type }}</span>
                  </div>
                  <div class="text-xs text-placeholder mb-2" v-if="schema.description">
                    {{ schema.description }}
                  </div>

                  <!-- String / Enum -->
                  <template v-if="schema.type === 'string'">
                    <el-select
                      v-if="schema.enum"
                      v-model="paramsForm[key]"
                      class="w-full"
                      @change="handleFormChange"
                      clearable
                    >
                      <el-option v-for="opt in schema.enum" :key="opt" :label="opt" :value="opt" />
                    </el-select>
                    <el-input v-else v-model="paramsForm[key]" @input="handleFormChange" />
                  </template>

                  <!-- Boolean -->
                  <template v-else-if="schema.type === 'boolean'">
                    <el-switch v-model="paramsForm[key]" @change="handleFormChange" />
                  </template>

                  <!-- Number / Integer -->
                  <template v-else-if="schema.type === 'number' || schema.type === 'integer'">
                    <el-input-number
                      v-model="paramsForm[key]"
                      class="w-full"
                      @change="handleFormChange"
                      :controls-position="'right'"
                    />
                  </template>

                  <!-- Object / Array (Fallback to JSON input for field) -->
                  <template v-else>
                    <el-input
                      type="textarea"
                      :model-value="JSON.stringify(paramsForm[key])"
                      disabled
                      placeholder="Complex type, please use JSON mode"
                    />
                  </template>
                </div>
              </div>

              <!-- JSON Mode or No Schema -->
              <div v-else class="h-full flex flex-col mx-2">
                <div class="mb-2 text-sm font-bold text-regular">Arguments (JSON)</div>
                <div class="border rounded flex-1 relative min-h-[300px]">
                  <el-input
                    v-model="inputJson"
                    type="textarea"
                    :rows="15"
                    placeholder="{}"
                    class="font-mono text-sm h-full"
                    resize="none"
                    @input="handleJsonChange"
                  />
                  <div
                    v-if="jsonError"
                    class="text-error text-xs mt-1 absolute bottom-2 left-2 right-2 bg-card p-1 opacity-90 border border-red-200"
                  >
                    {{ jsonError }}
                  </div>
                </div>
              </div>
            </template>
            <el-empty v-else description="Select a tool to configure inputs" />
          </el-scrollbar>
        </el-splitter-panel>

        <!-- Panel 3: Operation Bar -->
        <el-splitter-panel
          size="200px"
          class="flex flex-col bg-block border-l border-r"
          :resizable="false"
        >
          <div class="p-3 border-b bg-header font-medium text-center">Action</div>
          <div class="flex-1 flex flex-col items-center p-4 gap-4">
            <el-button
              type="primary"
              size="large"
              class="w-full base-btn"
              :icon="VideoPlay"
              :loading="running"
              :disabled="!currentTool || !!jsonError"
              @click="handleRunTool"
            >
              Run
            </el-button>

            <el-divider content-position="center" class="!my-2">History</el-divider>

            <el-scrollbar class="w-full flex-1">
              <div
                v-for="(item, idx) in history"
                :key="idx"
                class="history-item text-xs p-2 mb-2 rounded bg-card shadow-sm border cursor-pointer hover:bg-hover"
                @click="restoreHistory(item)"
              >
                <div class="font-bold truncate text-regular">{{ item.tool }}</div>
                <div class="text-secondary">{{ formatTime(item.timestamp) }}</div>
                <div class="mt-1">
                  <el-tag size="small" :type="item.status === 'success' ? 'success' : 'danger'">{{
                    item.status
                  }}</el-tag>
                </div>
              </div>
            </el-scrollbar>
          </div>
        </el-splitter-panel>

        <!-- Panel 4: Output Parameters -->
        <el-splitter-panel class="flex flex-col">
          <div class="p-3 border-b bg-header font-medium flex justify-between items-center">
            <span>Entry Output</span>
            <el-button v-if="outputResult" size="small" link @click="clearOutput">Clear</el-button>
          </div>
          <el-scrollbar class="flex-1 p-0">
            <div v-if="outputResult" class="h-full relative">
              <pre class="m-0 p-4 text-sm font-mono whitespace-pre-wrap break-all text-regular">{{
                outputResult
              }}</pre>
            </div>
            <el-empty v-else description="No output generated" image-size="60" />
          </el-scrollbar>
        </el-splitter-panel>
      </el-splitter>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Search, Refresh, VideoPlay } from '@element-plus/icons-vue'
import { useDebugToolsHooks } from './hooks/index.ts'
import {
  generateDefaultValue,
  isPropertyRequired,
  resolveRef,
  normalizeUnionType,
} from '@/utils/schemaUtils'

const {
  activeOptions,
  instanceInfo,
  toolList,
  currentTool,
  keyword,
  inputJson,
  outputResult,
  history,
  route,
  t,
} = useDebugToolsHooks()

const loading = ref(false)
const running = ref(false)
const instanceId = computed(() => route.query.instanceId as string)
// Computed
const filteredTools = computed(() => {
  if (!keyword.value) return toolList.value
  const lower = keyword.value.toLowerCase()
  return toolList.value.filter(
    (t) =>
      t.name.toLowerCase().includes(lower) ||
      (t.description && t.description.toLowerCase().includes(lower)),
  )
})

const jsonError = computed(() => {
  try {
    JSON.parse(inputJson.value)
    return ''
  } catch (e: any) {
    return e.message
  }
})

// Methods
const formatTime = (ts: number) => {
  return new Date(ts).toLocaleTimeString()
}

const getInstanceDetail = async () => {
  if (!instanceId.value) return
  try {
    loading.value = true
    // TODO: Replace with actual API call
    // const res = await InstanceAPI.detail({ instanceId: instanceId.value })
    // instanceInfo.value = res.data || {}

    // Mock Data for UI dev
    await new Promise((r) => setTimeout(r, 500))
    instanceInfo.value = {
      name: 'Example MCP Instance',
      status: 'Running',
      instanceId: instanceId.value,
      baseUrl: 'http://localhost:3000',
    }
  } catch (err) {
    console.error(err)
  } finally {
    loading.value = false
  }
}

const getTools = async () => {
  // TODO: Call API to list tools for this instance
  toolList.value = [
    {
      name: 'read_file',
      description: 'Read a file from the filesystem',
      inputSchema: {
        type: 'object',
        properties: {
          path: { type: 'string', description: 'The absolute path to the file to read' },
          encoding: { type: 'string', description: 'File encoding', default: 'utf-8' },
        },
        required: ['path'],
      },
    },
    {
      name: 'write_file',
      description: 'Write content to a file',
      inputSchema: {
        type: 'object',
        properties: {
          path: { type: 'string', description: 'The absolute path to the file to write' },
          content: { type: 'string', description: 'The content to write' },
          create_dirs: {
            type: 'boolean',
            description: 'Create missing directories',
            default: false,
          },
        },
        required: ['path', 'content'],
      },
    },
    {
      name: 'list_directory',
      description: 'List contents of a directory',
      inputSchema: {
        type: 'object',
        properties: {
          path: { type: 'string', description: 'The directory path' },
        },
        required: ['path'],
      },
    },
    {
      name: 'search_project',
      description: 'Semantic search across project files',
      inputSchema: {
        type: 'object',
        properties: {
          query: { type: 'string', description: 'Search query' },
          max_results: { type: 'integer', description: 'Maximum results', default: 10 },
          include_pattern: { type: 'string', description: 'Glob pattern to include' },
        },
        required: ['query'],
      },
    },
  ]
}

const paramsForm = ref<Record<string, any>>({})
const jsonMode = ref(false)

const handleSelectTool = (tool: any) => {
  currentTool.value = tool
  jsonMode.value = false

  // Generate default values from schema
  const defaults: Record<string, any> = {}
  if (tool.inputSchema?.properties) {
    Object.entries(tool.inputSchema.properties).forEach(([key, schema]: [string, any]) => {
      const resolved = resolveRef(schema, tool.inputSchema)
      const val = generateDefaultValue(resolved, key, tool.inputSchema)
      if (val !== undefined) {
        defaults[key] = val
      }
    })
  }

  paramsForm.value = defaults
  inputJson.value = JSON.stringify(defaults, null, 2)
  outputResult.value = ''
}

const handleFormChange = () => {
  inputJson.value = JSON.stringify(paramsForm.value, null, 2)
}

const handleJsonChange = (val: string) => {
  try {
    paramsForm.value = JSON.parse(val)
  } catch (e) {
    // Ignore parse errors while typing
  }
}

const handleRunTool = async () => {
  if (jsonError.value) return
  try {
    running.value = true
    const params = JSON.parse(inputJson.value)

    // TODO: Call API executeTool(instanceId, tool.name, params)
    await new Promise((r) => setTimeout(r, 1000)) // Mock delay

    const mockOutput = {
      tool: currentTool.value.name,
      params,
      result: {
        success: true,
        data: 'This is a mock response from the tool execution.\nFile content would appear here.',
      },
    }

    outputResult.value = JSON.stringify(mockOutput, null, 2)

    // Add to history
    history.value.unshift({
      tool: currentTool.value.name,
      params: inputJson.value,
      timestamp: Date.now(),
      status: 'success',
    })
  } catch (err) {
    outputResult.value = JSON.stringify({ error: 'Execution Failed', details: err }, null, 2)
  } finally {
    running.value = false
  }
}

const restoreHistory = (item: any) => {
  const tool = toolList.value.find((t) => t.name === item.tool)
  if (tool) {
    currentTool.value = tool
    inputJson.value = item.params
  }
}

const refreshInstance = () => {
  getInstanceDetail()
  getTools()
}

const clearOutput = () => {
  outputResult.value = ''
}

onMounted(() => {
  if (instanceId.value) {
    getInstanceDetail()
    getTools()
  } else {
    // Demo mode if no ID
    getInstanceDetail()
    getTools()
  }
})
</script>

<style lang="scss" scoped>
.tool-page {
  /* Fix height to fill remaining viewport if needed, assuming header is ~60px */
  height: calc(100vh - 60px);
  margin: -20px;
  padding: 16px;
  background-color: var(--ep-bg-color-page);
}

.bg-page {
  border: 1px solid var(--ep-border-color);
}
.bg-header {
  background-color: var(--ep-fill-color);
  border-bottom: 1px solid var(--ep-border-color-light);
}
.bg-block {
  background-color: var(--ep-fill-color);
  border: 1px solid var(--ep-border-color-lighter);
}
.bg-card {
  background-color: var(--ep-bg-color-overlay);
}
.text-primary {
  color: var(--ep-text-color-primary);
}
.text-regular {
  color: var(--ep-text-color-regular);
}
.text-secondary {
  color: var(--ep-text-color-secondary);
}
.text-placeholder {
  color: var(--ep-text-color-placeholder);
}

.hover\:bg-hover:hover {
  background-color: var(--ep-fill-color);
}
.tool-item {
  border-bottom: 1px solid var(--ep-border-color-lighter);
  transition: all 0.3s ease;
}
.tool-item:hover {
  background-color: var(--el-color-primary-hover);
  transform: scale(1.02);
}

.active-tool {
  background-color: var(--el-color-primary);
}
:deep(.el-splitter) {
  --el-splitter-border-color: var(--ep-border-color);
}
</style>
