<template>
  <div class="token-log-page p-4">
    <!-- 筛选栏 -->
    <div class="filter-bar mb-4 p-3 border-rd-2">
      <!-- 第一行：日志级别 -->
      <div class="flex items-center gap-3 mb-3">
        <span class="font-bold filter-label">日志级别：</span>
        <div class="level-tags flex items-center gap-2">
          <el-tag
            v-for="item in levelOptions"
            :key="item.value"
            :type="item.type"
            :effect="level === item.value ? 'dark' : 'light'"
            size="mini"
            class="level-tag"
            @click="handleLevelChange(item.value)"
          >
            {{ item.label }}
          </el-tag>
        </div>
      </div>

      <!-- 第二行：时间范围 -->
      <div class="flex items-center gap-3 mb-3">
        <span class="font-bold filter-label">时间范围：</span>
        <div class="w-25">
          <el-date-picker
            v-model="dateRange"
            type="datetimerange"
            range-separator="至"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
            size="small"
            @change="handleGetLogs"
          />
        </div>
      </div>

      <!-- 第三行：Trace ID + 刷新按钮 -->
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-3">
          <span class="font-bold filter-label">Trace ID：</span>
          <el-input
            v-model="traceId"
            placeholder="请输入 Trace ID"
            size="small"
            clearable
            @clear="handleGetLogs"
            @keyup.enter="handleGetLogs"
            style="width: 400px"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
        </div>
        <el-button size="small" @click="handleGetLogs" :loading="loading">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>
    </div>

    <!-- 日志内容 -->
    <div class="logs-container" v-loading="loading">
      <el-scrollbar class="logs-box" v-if="logList.length > 0">
        <div class="log-line" v-for="item in logList" :key="item.id">
          <span class="log-time">{{ formatTime(item.createdAt) }}</span>
          <el-tag :type="getLevelType(item.level)" size="small" class="log-level-tag">
            {{ getLevelLabel(item.level) }}
          </el-tag>
          <span class="log-event">{{ item.event }}</span>
          <span class="log-separator">|</span>
          <span class="log-detail">{{ formatLogOneLine(item.log) }}</span>
        </div>
      </el-scrollbar>
      <el-empty v-else description="暂无日志数据" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { InstanceAPI } from '@/api/mcp/instance'
import { Refresh, Search } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

interface LogItem {
  id: number
  instanceId: string
  tokenHeader: string
  token: string
  usages: string[]
  level: number
  event: string
  createdAt: string
  log: string
  traceId: string
}

const route = useRoute()
const logList = ref<LogItem[]>([])
const loading = ref(false)
const dateRange = ref<[Date, Date] | null>(null)
const level = ref('')
const traceId = ref('')

// 日志级别选项
const levelOptions = [
  { label: '全部', value: '', type: '' as const },
  { label: 'Trace', value: 1, type: 'primary' as const },
  { label: 'Debug', value: 2, type: 'info' as const },
  { label: 'Info', value: 3, type: 'primary' as const },
  { label: 'Warn', value: 4, type: 'warning' as const },
  { label: 'Error', value: 5, type: 'danger' as const },
]

// 从路由参数获取 instanceId 和 token
const instanceId = computed(() => (route.query.instanceId as string) || '')
const token = computed(() => (route.query.token as string) || '')

// 切换日志级别
const handleLevelChange = (val: string) => {
  level.value = val
  handleGetLogs()
}

// 格式化时间
const formatTime = (time: string) => {
  if (!time) return ''
  return new Date(time).toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  })
}

// 获取日志级别类型
const getLevelType = (
  level: number,
): '' | 'primary' | 'success' | 'warning' | 'danger' | 'info' => {
  const map: Record<number, '' | 'primary' | 'success' | 'warning' | 'danger' | 'info'> = {
    1: 'info',
    2: 'primary',
    3: 'success',
    4: 'warning',
    5: 'danger',
  }
  return map[level] || ''
}

// 获取日志级别标签
const getLevelLabel = (level: number): string => {
  const map: Record<number, string> = {
    1: 'Trace',
    2: 'Debug',
    3: 'Info',
    4: 'Warn',
    5: 'Error',
  }
  return map[level] || 'Unknown'
}

// 格式化日志为一行
const formatLogOneLine = (log: string) => {
  if (!log) return ''
  try {
    const logObj = JSON.parse(log)
    // 将 JSON 对象压缩为一行字符串
    return JSON.stringify(logObj)
  } catch {
    return log
  }
}

const handleGetLogs = async () => {
  if (!instanceId.value || !token.value) {
    ElMessage.warning('缺少必要参数：instanceId 或 token')
    return
  }

  loading.value = true
  try {
    const params: any = {
      instanceId: instanceId.value,
      token: token.value,
      startTime: dateRange.value ? new Date(dateRange.value[0]).getTime().toString() : '',
      endTime: dateRange.value ? new Date(dateRange.value[1]).getTime().toString() : '',
      level: level.value,
      traceId: traceId.value,
      pageNum: 1,
      pageSize: 999,
    }

    const { logs: logData } = await InstanceAPI.logsByToken(params)
    logList.value = logData || []
  } catch (error) {
    console.error('获取日志失败', error)
    ElMessage.error('获取日志失败')
    logList.value = []
  } finally {
    loading.value = false
  }
}

const init = () => {
  handleGetLogs()
}

onMounted(init)
</script>

<style lang="scss" scoped>
.token-log-page {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.filter-bar {
  // background: var(--ep-bg-color);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.08);

  .filter-label {
    display: inline-block;
    min-width: 80px;
    text-align: right;
    color: var(--ep-text-color-regular);
  }
}

.level-tags {
  .level-tag {
    cursor: pointer;
    transition: all 0.3s;
    user-select: none;
    font-weight: 500;

    &:hover {
      transform: translateY(-2px);
      box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
    }
  }
}

.logs-container {
  flex: 1;
  height: 0;
  background: var(--ep-bg-color-page);
  border-radius: 8px;
  overflow: hidden;
}

.logs-box {
  height: 100%;
  padding: 12px;
}

.log-line {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 12px;
  border-bottom: 1px solid var(--ep-border-color-lighter);
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 12px;
  line-height: 1.8;
  transition: background-color 0.2s;

  &:hover {
    background-color: var(--ep-bg-color-overlay);
  }

  &:last-child {
    border-bottom: none;
  }

  .log-time {
    flex-shrink: 0;
    color: var(--ep-text-color-secondary);
    font-size: 12px;
    min-width: 160px;
  }

  .log-level-tag {
    flex-shrink: 0;
    font-weight: 600;
  }

  .log-event {
    flex-shrink: 0;
    color: var(--ep-color-primary);
    font-weight: 500;
    min-width: 200px;
  }

  .log-separator {
    flex-shrink: 0;
    color: var(--ep-border-color);
  }

  .log-detail {
    flex: 1;
    color: var(--ep-text-color-primary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}
</style>
