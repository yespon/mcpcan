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
    :show-close="false"
    width="800px"
    footer-class="footer-border"
  >
    <template #title> 任务列表 </template>
    <div class="task-list-scroll hide-scrollbar">
      <div v-for="(task, idx) in taskInfo.list" :key="task.id" class="task-item">
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
              取消任务
            </el-button>
          </span>
          <span class="task-status" :class="['doing', 'done', 'danger', 'error'][task.status - 1]">
            {{ ['进行中', '已完成', '异常', '已取消'][task.status - 1] }}
          </span>
          <span class="task-expand-icon">
            <el-icon size="18" color="#fff" v-if="task.expanded"><ArrowUp /></el-icon>
            <el-icon size="18" color="#fff" v-else><ArrowDown /></el-icon>
          </span>
        </div>
        <transition name="fade">
          <div v-if="task.expanded" v-loading="task.loading" class="task-detail">
            <div class="task-detail-title flex items-center justify-between">
              <span>任务详情</span>
              <el-icon
                class="cursor-pointer hover:rotate-90 transition-all"
                size="20"
                title="刷新"
                @click.stop="handleGetTaskDetail(task)"
              >
                <Refresh />
              </el-icon>
            </div>
            <div class="my-2">任务创建时间：{{ task.createdAt }}</div>
            <div class="my-2">智能平台名称：{{ task.intelligentAccessName }}</div>
            <div v-if="Array.isArray(task.logs)">
              <div v-for="log in task.logs" :key="log.mcpInstanceID" class="task-mcp-card mb-4">
                <div class="flex items-center justify-between mb-1">
                  <div class="font-bold text-lg">{{ log.mcpInstanceName }}</div>
                  <el-tag :type="log.status ? 'success' : 'danger'" size="small" effect="dark">
                    {{ log.status ? '成功' : '失败' }}
                  </el-tag>
                </div>
                <div
                  v-if="
                    Array.isArray(log.insertIntelligentLogs) && log.insertIntelligentLogs.length
                  "
                >
                  <div class="text-sm text-gray-400 mb-1">已同步到的命名空间：</div>
                  <div class="sync-space-list">
                    <div
                      v-for="(item, idx) in log.insertIntelligentLogs"
                      :key="idx"
                      class="sync-space-item"
                    >
                      <span class="font-bold">{{
                        item.insertIntelligentInfo?.difySpaceName || '-'
                      }}</span>
                      <el-tag
                        :type="item.status ? 'success' : 'danger'"
                        size="mini"
                        effect="plain"
                        class="mx-2"
                        >{{ item.status ? '成功' : '失败' }}</el-tag
                      >
                      <div
                        v-if="!item.status && item.errorLog"
                        class="sync-error-subcard my-1 mx-2"
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
                  </div>
                </div>
                <div v-if="!log.status && log.errorLog" class="sync-error-maincard mt-3">
                  <el-icon color="#F56C6C" class="mr-1" style="vertical-align: middle" size="18"
                    ><WarningFilled
                  /></el-icon>
                  <span class="sync-error-text">{{ log.errorLog }}</span>
                </div>
              </div>
            </div>
            <div v-else>{{ task.logs }}</div>
          </div>
        </transition>
      </div>
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

const { taskInfo, handleGetTaskList } = toRefs(useBusinessStoreHook())
const toggleExpand = (idx: number) => {
  taskInfo.value.list[idx].expanded = !taskInfo.value.list[idx].expanded
  handleGetTaskDetail(taskInfo.value.list[idx])
}

// 取消任务（需补充具体API逻辑）
const handleCancelTask = async (task: any) => {
  await AgentAPI.cancelTask(task.id)
  task.status = 4
}

// 获取任务详情信息
const handleGetTaskDetail = async (baseTask: any) => {
  try {
    baseTask.loading = true
    const { task } = await AgentAPI.taskDetail(baseTask.id)
    taskInfo.value.list.forEach((item) => {
      if (item.id === task.id) {
        item.logs = task.installLogs
      }
    })
  } finally {
    baseTask.loading = false
  }
}

// 显示任务列表
const handleShowTaskList = () => {
  taskInfo.value.visible = true
}
onMounted(() => {
  handleGetTaskList.value()
})
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
  padding: 16px 18px 12px 18px;
  border: 1.5px solid #3a3a5a;
  margin-bottom: 18px;
  transition: box-shadow 0.2s;
  &:hover {
    box-shadow: 0 4px 16px #0004;
    border-color: #7c4dff;
  }
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
