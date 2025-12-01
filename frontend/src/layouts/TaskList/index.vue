<template>
  <div class="task-list center">
    <el-badge is-dot :offset="[-10, 5]" color="red">
      <div class="center cursor-pointer task-tag" title="Task List" @click="handleShowTaskList">
        <el-icon size="28"><List /></el-icon>
      </div>
    </el-badge>
  </div>
  <el-dialog
    v-model="taskInfo.visible"
    :close-on-click-modal="false"
    width="800px"
    footer-class="footer-border"
  >
    <template #title> 任务列表 </template>

    <div class="task-list-scroll">
      <div v-for="(task, idx) in taskInfo.list" :key="task.id" class="task-item">
        <div class="task-item-main" @click="toggleExpand(idx)">
          <span class="task-icon">
            <el-icon v-if="task.status === 'doing'" color="#fff" size="22"><Refresh /></el-icon>
            <el-icon v-else-if="task.status === 'done'" color="#fff" size="22"
              ><CircleCheckFilled
            /></el-icon>
            <el-icon v-else-if="task.status === 'error'" color="#fff" size="22"
              ><WarningFilled
            /></el-icon>
          </span>
          <span class="task-title">{{ task.title }}</span>
          <span class="task-status" :class="task.status">
            <template v-if="task.status === 'doing'">进行中</template>
            <template v-else-if="task.status === 'done'">已完成</template>
            <template v-else>异常</template>
          </span>
          <span class="task-expand-icon">
            <el-icon size="18" color="#fff" v-if="task.expanded"><ArrowUp /></el-icon>
            <el-icon size="18" color="#fff" v-else><ArrowDown /></el-icon>
          </span>
        </div>
        <transition name="fade">
          <div v-if="task.expanded" class="task-detail">
            <div class="task-detail-title">详细信息</div>
            <div class="task-detail-content">{{ task.detail }}</div>
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
} from '@element-plus/icons-vue'
import { useBusinessStoreHook } from '@/stores/modules/business-store'

const { taskInfo } = toRefs(useBusinessStoreHook())
const toggleExpand = (idx: number) => {
  taskInfo.value.list[idx].expanded = !taskInfo.value.list[idx].expanded
}

// 显示任务列表
const handleShowTaskList = () => {
  taskInfo.value.visible = true
}
/**
 * 初始化任务列表
 */
const init = () => {}
onMounted(() => {
  init()
})
</script>

<style scoped lang="scss">
.task-list {
  position: fixed;
  bottom: 20px;
  right: 20px;
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
    color: #3ec46d;
  }
  &.done {
    color: #4ea1ff;
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
  color: #ffd700;
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
