<template>
  <div class="task-list center">
    <el-badge
      v-if="!taskInfo.visible"
      :is-dot="taskInfo.list.some((task) => task.status === 1)"
      :offset="[-10, 5]"
      color="red"
    >
      <div class="center cursor-pointer task-tag" title="Task List" @click="handleShowTaskList">
        <el-icon size="28"><List /></el-icon>
      </div>
    </el-badge>
  </div>
  <el-dialog
    v-model="taskInfo.visible"
    :close-on-click-modal="true"
    width="800px"
    footer-class="footer-border"
  >
    <template #title> {{ t('agent.task.title') }} </template>
    <div class="task-list-scroll">
      <template v-if="taskInfo.list?.length">
        <div v-for="(task, idx) in taskInfo.list" :key="task.id" class="task-item">
          <div class="task-progress-bar" v-if="typeof task.progress === 'number'">
            <el-progress
              :percentage="task.progress"
              :stroke-width="16"
              :text-inside="true"
              status="success"
              style="margin-bottom: 8px"
            />
          </div>
          <div class="task-item-main" @click="toggleExpand(idx)">
            <span class="task-icon">
              <el-icon v-if="task.status === 1" color="#409EFF" size="22" class="icon-rotate">
                <Refresh />
              </el-icon>
              <el-icon v-else-if="task.status === 2" color="#67C23A" size="22">
                <CircleCheckFilled />
              </el-icon>
              <el-icon v-else-if="task.status === 3" color="#F56C6C" size="22">
                <WarningFilled />
              </el-icon>
              <el-icon v-else-if="task.status === 4" color="#E6A23C" size="22">
                <CircleCloseFilled />
              </el-icon>
            </span>
            <span class="task-title">
              {{ task.desc }}
              <el-button
                v-if="task.status === 1"
                type="danger"
                size="small"
                class="mx-2"
                round
                @click.stop="handleCancelTask(task)"
              >
                {{ t('agent.task.cancel') }}
              </el-button>
            </span>
            <span
              class="task-status"
              :class="['doing', 'done', 'danger', 'error'][task.status - 1]"
            >
              {{
                [
                  t('agent.task.status.doing'),
                  t('agent.task.status.done'),
                  t('agent.task.status.error'),
                  t('agent.task.status.cancel'),
                ][task.status - 1]
              }}
            </span>
            <span class="task-expand-icon">
              <el-icon size="18" color="#fff" v-if="task.expanded"><ArrowUp /></el-icon>
              <el-icon size="18" color="#fff" v-else><ArrowDown /></el-icon>
            </span>
          </div>
          <transition name="fade">
            <div v-if="task.expanded" v-loading="task.loading" class="task-detail">
              <div class="task-detail-title flex items-center justify-between">
                <span>{{ t('agent.task.details') }}</span>
                <el-icon
                  class="cursor-pointer hover:rotate-90 transition-all"
                  size="20"
                  @click.stop="handleGetTaskDetail(task)"
                >
                  <Refresh />
                </el-icon>
              </div>
              <div class="my-2">{{ t('agent.task.createTime') }}：{{ task.createdAt }}</div>
              <div class="my-2">
                {{ t('agent.task.platformName') }}：{{ task.intelligentAccessName }}
              </div>
              <div v-if="Array.isArray(task.logs)">
                <div v-for="log in task.logs" :key="log.mcpInstanceID" class="task-mcp-card mb-4">
                  <div
                    class="flex items-center justify-between cursor-pointer mx-4 my-3"
                    @click="toggleLogExpand(log)"
                  >
                    <div class="font-bold text-lg">
                      {{ log.mcpInstanceName
                      }}<el-tag
                        :type="log.status ? 'success' : 'danger'"
                        size="small"
                        effect="dark"
                        class="mx-2"
                      >
                        {{ log.status ? t('status.success') : t('status.fail') }}
                      </el-tag>
                      <span class="font-500 font-size-4">
                        {{ t('agent.sync.executed') }}
                        <span class="color-green">{{
                          log.insertIntelligentLogs.filter((item: any) => item.status).length
                        }}</span>
                        <span class="color-red ml-4">
                          {{ log.insertIntelligentLogs.filter((item: any) => !item.status).length }}
                        </span>
                      </span>
                    </div>
                    <div class="flex items-center">
                      <el-icon size="16" color="#fff">
                        <component :is="(log._expanded ?? false) ? ArrowUp : ArrowDown" />
                      </el-icon>
                    </div>
                  </div>

                  <div v-show="log._expanded ?? false">
                    <div
                      v-if="
                        Array.isArray(log.insertIntelligentLogs) && log.insertIntelligentLogs.length
                      "
                    >
                      <div class="text-sm text-gray-400 mb-1 ml-2">
                        {{ t('agent.task.syncSpaceName') }}：
                      </div>
                      <div class="sync-space-list">
                        <DynamicScroller
                          class="logs-list-scroll mt-4"
                          :items="log.insertIntelligentLogs"
                          :min-item-size="80"
                          key-field="id"
                        >
                          <template v-slot="{ item, index: idx, active }">
                            <DynamicScrollerItem
                              :item="item"
                              :data-index="idx"
                              :active="active"
                              :size-dependencies="[
                                item.errorLog,
                                item.insertIntelligentInfo?.difySpaceName,
                              ]"
                              class="py-1"
                            >
                              <div class="sync-space-item">
                                <span class="font-bold">{{
                                  item.insertIntelligentInfo?.difySpaceName || '-'
                                }}</span>
                                <el-tag
                                  :type="item.status ? 'success' : 'danger'"
                                  size="small"
                                  effect="plain"
                                  class="mx-2"
                                >
                                  {{ item.status ? t('status.success') : t('status.fail') }}
                                </el-tag>
                                <div
                                  v-if="!item.status && item.errorLog"
                                  class="sync-error-subcard my-1 mx-2 flex-sub"
                                >
                                  <el-icon
                                    color="#F56C6C"
                                    class="mr-1"
                                    style="vertical-align: middle"
                                    size="16"
                                  >
                                    <WarningFilled />
                                  </el-icon>
                                  <span class="sync-error-text">{{ item.errorLog }}</span>
                                </div>
                              </div>
                            </DynamicScrollerItem>
                          </template>
                        </DynamicScroller>
                      </div>
                    </div>
                    <div v-if="!log.status && log.errorLog" class="sync-error-maincard mt-3">
                      <el-icon
                        color="#F56C6C"
                        class="mr-1"
                        style="vertical-align: middle"
                        size="18"
                      >
                        <WarningFilled />
                      </el-icon>
                      <span class="sync-error-text">{{ log.errorLog }}</span>
                    </div>
                  </div>
                </div>
              </div>
              <div v-else>{{ task.logs }}</div>
            </div>
          </transition>
        </div>
      </template>
      <el-empty v-else class="mt-20" :image-size="200" :description="t('status.noData')"></el-empty>
    </div>
  </el-dialog>
</template>

<script setup lang="ts">
import {
  List,
  Refresh,
  CircleCheckFilled,
  WarningFilled,
  ArrowDown,
  ArrowUp,
  CircleCloseFilled,
} from '@element-plus/icons-vue'
import { useBusinessStoreHook } from '@/stores/modules/business-store'
import { AgentAPI } from '@/api/agent'
// @ts-expect-error - vue-virtual-scroller 缺少类型定义
import { DynamicScroller, DynamicScrollerItem } from 'vue-virtual-scroller'

// import { RecycleScroller } from 'vue-virtual-scroller'
import 'vue-virtual-scroller/dist/vue-virtual-scroller.css'

const { t } = useI18n()
const { taskInfo, handleGetTaskList } = toRefs(useBusinessStoreHook())
const toggleExpand = (idx: number) => {
  taskInfo.value.list[idx].expanded = !taskInfo.value.list[idx].expanded
  handleGetTaskDetail(taskInfo.value.list[idx])
}

// cancel task
const handleCancelTask = async (task: any) => {
  await AgentAPI.cancelTask(task.id)
  task.status = 4
}

// handle get task detail
const handleGetTaskDetail = async (baseTask: any, updateProgress?: boolean) => {
  try {
    baseTask.loading = true
    const { task: detail } = await AgentAPI.taskDetail(baseTask.id)

    taskInfo.value.list.forEach((item) => {
      if (item.id === detail.id) {
        // 保持每个 log 的 _expanded 状态：按 mcpInstanceID 进行匹配合并
        const prevLogs = Array.isArray(item.logs) ? item.logs : []
        const prevMap = new Map(prevLogs.map((l: any) => [l.mcpInstanceID, l]))
        const mergedLogs = Array.isArray(detail.installLogs)
          ? detail.installLogs.map((newLog: any) => {
              const oldLog: any = prevMap.get(newLog.mcpInstanceID) as any
              if (oldLog && Object.prototype.hasOwnProperty.call(oldLog, '_expanded')) {
                // 使用 Reflect.set 以确保响应性
                Reflect.set(newLog as any, '_expanded', (oldLog as any)._expanded)
              } else {
                Reflect.set(newLog as any, '_expanded', false)
              }
              return newLog
            })
          : detail.installLogs
        // 唯一ID处理；用于虚拟列表渲染
        mergedLogs.forEach((log: any) => {
          log.insertIntelligentLogs.forEach((info: any) => {
            Reflect.set(
              info as any,
              'id',
              info.insertIntelligentInfo?.difySpaceID ||
                Date.now().toString() + Math.random().toString(),
            )
          })
        })
        item.logs = mergedLogs
        // calculate progress
        const total = detail.insertIntelligentInfos.length * detail.mcpInstanceIDs.length
        let done = 0
        if (Array.isArray(detail.installLogs)) {
          for (const log of detail.installLogs) {
            if (Array.isArray(log.insertIntelligentLogs)) {
              done += log.insertIntelligentLogs.filter((l: any) => l.status === true).length
            }
          }
        }
        if (updateProgress) {
          item.progress = total ? Math.round((done / total) * 100) : 0
        }
      }
    })
  } finally {
    baseTask.loading = false
  }
}

// show task list
const handleShowTaskList = () => {
  taskInfo.value.visible = true
}
onMounted(() => {
  handleGetTaskList.value()
})

// 切换单条日志的展开状态：以当前展示状态为准取反，修复首次点击不生效
const toggleLogExpand = (log: any) => {
  const current = log._expanded ?? false
  // 使用 Reflect.set 以确保属性新增也可响应
  Reflect.set(log, '_expanded', !current)
}

watch(
  () => taskInfo.value.list.map((t) => t.status),
  async (newStatuses, oldStatuses = [], onCleanup) => {
    let stop = false
    onCleanup(() => {
      stop = true
    })
    for (let i = 0; i < taskInfo.value.list.length; i++) {
      const task = taskInfo.value.list[i]
      const prev = oldStatuses[i]
      const curr = newStatuses[i]
      if (curr === 1) {
        handleGetTaskDetail(task, true)
        if (stop) break
      } else if (prev === 1 && curr !== 1) {
        // 从进行中变为其他状态：再请求一次详情
        try {
          handleGetTaskDetail(task, true)
          setTimeout(() => {
            task.progress = undefined
          }, 500)
        } catch {}
      }
    }
  },
  { immediate: true, deep: true },
)
</script>

<style scoped lang="scss">
.task-list {
  position: fixed;
  bottom: 20px;
  right: 20px;
  z-index: 9999;
  &:hover {
    transform: scale(1.1);
  }
  transition: all 0.3s ease-in-out;
  .task-tag {
    width: 60px;
    height: 60px;
    background: var(--ep-purple-color);
    border-radius: 50%;
    box-shadow: 0 4px 12px var(--ep-bg-purple-color-deep);
  }
}
:deep(.el-badge__content.is-dot) {
  width: 12px;
  height: 12px;
}

.task-list-scroll {
  height: 60vh;
  overflow-y: auto;
  padding: 0 12px 8px 12px;
}
.task-item {
  background: #444;
  border-radius: 12px;
  margin-bottom: 18px;
  box-shadow: 0 2px 8px var(--ep-bg-purple-color);
  transition: box-shadow 0.2s;
  border: 2px solid #222;
  &:hover {
    box-shadow: 0 4px 16px #0006;
    border-color: var(--ep-bg-purple-color-deep);
  }
}
.task-item-main {
  display: flex;
  align-items: center;
  padding: 12px 12px 12px 12px;
  cursor: pointer;
  font-size: 14px;
  position: relative;
}
.task-icon {
  margin-right: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
}
.task-title {
  flex: 1;
  font-weight: 600;
  font-size: 14px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.task-status {
  min-width: 60px;
  text-align: right;
  font-size: 14px;
  margin-left: 18px;
  &.doing {
    color: #4ea1ff;
  }
  &.done {
    color: #3ec46d;
  }
  &.danger {
    color: #f56c6c;
  }
  &.error {
    color: #ff9900;
  }
}
.task-expand-icon {
  margin-left: 12px;
  display: flex;
  align-items: center;
}
.task-detail {
  background: #333;
  border-radius: 0 0 12px 12px;
  padding: 12px 12px 12px 24px;
  color: #fff;
  font-size: 15px;
  border-top: 1px solid #222;
  animation: fadeIn 0.2s;
}
.task-detail-title {
  font-weight: bold;
  margin-bottom: 8px;
}
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>

<style scoped lang="scss">
.icon-rotate {
  animation: icon-rotate 1s linear infinite;
}
@keyframes icon-rotate {
  0% {
    transform: rotate(0deg);
  }
  100% {
    transform: rotate(360deg);
  }
}
</style>

<style scoped lang="scss">
.sync-error-maincard {
  background: #3a1a1a;
  border: 1.5px solid #f56c6c;
  color: #f56c6c;
  border-radius: 8px;
  padding: 10px 16px 10px 16px;
  margin-top: 10px;
  font-size: 15px;
  display: flex;
  align-items: center;
  font-weight: 500;
  box-shadow: 0 2px 8px #f56c6c22;
}
.sync-error-subcard {
  background: #f56c6c18;
  border-radius: 6px;
  padding: 7px 12px 7px 12px;
  margin-top: 4px;
  margin-left: 0;
  font-size: 14px;
  color: #f56c6c;
  display: flex;
  align-items: center;
}
.sync-error-text {
  color: #f56c6c;
  word-break: break-all;
}
</style>

<style scoped lang="scss">
.task-mcp-card {
  background: #23233a;
  border-radius: 10px;
  box-shadow: 0 2px 8px #0002;
  // padding: 16px 18px 12px 18px;
  border: 1.5px solid #3a3a5a;
  margin-bottom: 18px;
  transition: box-shadow 0.2s;
  &:hover {
    box-shadow: 0 4px 16px #0004;
    border-color: #7c4dff;
  }
}
.logs-list-scroll {
  height: 300px;
  overflow-y: auto;
  padding-right: 8px;
}
.sync-space-list {
  margin-top: 6px;
  padding-left: 8px;
}
.sync-space-item {
  display: flex;
  align-items: center;
  margin-bottom: 4px;
  padding: 4px 0 4px 8px;
  border-left: 3px solid #7c4dff33;
  background: #2d2d44;
  border-radius: 4px;
  font-size: 14px;
}
</style>
