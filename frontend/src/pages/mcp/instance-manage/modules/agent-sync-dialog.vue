<template>
  <el-dialog
    :title="dialogInfo.title"
    v-model="dialogInfo.visible"
    :close-on-click-modal="false"
    width="1200px"
    footer-class="footer-border"
  >
    <el-steps :active="dialogInfo.currentStep" align-center class="mb-6">
      <el-step
        title="同步平台"
        :description="
          dialogInfo.selectedPlatform
            ? '平台：' + { community: '社区版', business: '商业版' }[dialogInfo.selectedPlatform]
            : '请选择一个平台'
        "
      />
      <el-step
        title="同步的智能体"
        :description="
          dialogInfo.selectedAgentId
            ? '智能体：' + dialogInfo.selectedAgentId
            : '请选择一个你需要同步的智能体'
        "
      >
        <template #description>
          <div v-if="dialogInfo.selectedAgentId" class="u-line-1">
            智能体：
            {{
              dialogInfo.instanceList.find(
                (agent) => agent.instanceId === dialogInfo.selectedAgentId,
              )?.name || dialogInfo.selectedAgentId
            }}
          </div>
          <div v-else>请选择一个你需要同步的智能体</div>
        </template>
      </el-step>
      <el-step
        title="同步的空间"
        :description="
          dialogInfo.selectedNamespaces.length
            ? '已选择空间：' + (dialogInfo.selectedNamespaces.length || 0) + '个'
            : '请选择一个你需要同步的空间'
        "
      />
    </el-steps>
    <div v-if="dialogInfo.currentStep === 1" class="step-content">
      <!-- Step 1: 选择平台 -->
      <div class="platform-container">
        <div
          class="platform-card"
          :class="{ active: dialogInfo.selectedPlatform === 'community' }"
          @click="dialogInfo.selectedPlatform = 'community'"
        >
          <div class="platform-logo">
            <img src="@/assets/images/dify-logo.svg" alt="Dify" class="logo-img" />
          </div>
          <div class="platform-badge community">社区版</div>
        </div>

        <div
          class="platform-card"
          :class="{ active: dialogInfo.selectedPlatform === 'business' }"
          @click="dialogInfo.selectedPlatform = 'business'"
        >
          <div class="platform-logo">
            <img src="@/assets/images/dify-logo.svg" alt="Dify" class="logo-img" />
          </div>
          <div class="platform-badge business">商业版</div>
        </div>
      </div>
    </div>
    <div v-else-if="dialogInfo.currentStep === 2" class="step-content">
      <!-- Step 2: 选择智能体 -->

      <div class="agent-grid grid grid-cols-12 gap-4 px-3">
        <div
          v-for="(agent, index) in dialogInfo.instanceList"
          :key="index"
          class="agent-card col-span-3 center py-3 px-4"
          :class="{ active: dialogInfo.selectedAgentId === agent.instanceId }"
          @click="toggleAgentSelection(agent.instanceId)"
        >
          <div class="w-full flex">
            <div
              v-if="dialogInfo.selectedAgentId === agent.instanceId"
              class="selected-badge"
            ></div>
            <div class="agent-icon">
              <el-icon class="cursor-pointer" size="48" color="var(--ep-purple-color)"
                ><i class="icon iconfont MCP-zhinengti"></i
              ></el-icon>
            </div>
            <div class="agent-info flex-sub u-line-1">
              <div class="agent-name">{{ agent.name || agent.instanceName || '未命名智能体' }}</div>
              <div class="agent-desc">{{ agent.description || agent.remark || '暂无描述' }}</div>
            </div>
          </div>
        </div>
      </div>
      <div v-if="!dialogInfo.instanceList.length" class="empty-state">
        <el-empty description="暂无可同步的智能体" />
      </div>
    </div>
    <div v-else-if="dialogInfo.currentStep === 3" class="step-content">
      <!-- Step 3: 选择空间 -->
      <div class="agent-grid grid grid-cols-12 gap-4 px-3">
        <div
          v-for="(namespace, index) in dialogInfo.namespaceList"
          :key="index"
          class="agent-card col-span-3 center py-3 px-4"
          :class="{ active: dialogInfo.selectedNamespaces.includes(namespace.id) }"
          @click="toggleNamespaceSelection(namespace.id)"
        >
          <div class="w-full flex">
            <div
              v-if="dialogInfo.selectedNamespaces.includes(namespace.id)"
              class="selected-badge"
            ></div>
            <div class="agent-icon">
              <el-icon class="cursor-pointer" size="48" color="var(--ep-purple-color)"
                ><i class="icon iconfont MCP-mingmingkongjian"></i
              ></el-icon>
            </div>
            <div class="agent-info flex-sub u-line-1">
              <div class="agent-name">{{ namespace.name || '未命名空间' }}</div>
              <div class="agent-desc">{{ namespace.description || '暂无描述' }}</div>
            </div>
          </div>
        </div>
      </div>
      <div v-if="!dialogInfo.namespaceList.length" class="empty-state">
        <el-empty description="暂无可同步的空间" />
      </div>
    </div>
    <template #footer>
      <div class="center">
        <el-button
          type="primary"
          class="base-btn"
          v-if="dialogInfo.currentStep > 1"
          @click="dialogInfo.currentStep -= 1"
          >上一步</el-button
        >
        <el-button
          type="primary"
          class="base-btn"
          v-if="dialogInfo.currentStep < 3"
          :disabled="
            (dialogInfo.currentStep === 1 && !dialogInfo.selectedPlatform) ||
            (dialogInfo.currentStep === 2 && !dialogInfo.selectedAgentId)
          "
          @click="dialogInfo.currentStep += 1"
          >下一步</el-button
        >
        <el-button
          type="primary"
          class="base-btn"
          v-else
          :disabled="dialogInfo.selectedNamespaces.length === 0"
          @click="dialogInfo.visible = false"
        >
          完 成
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
const dialogInfo = reactive({
  title: '同步至 Dify',
  visible: false,
  instanceList: [] as any[],
  namespaceList: [] as any[],
  currentStep: 1,
  selectedPlatform: '' as 'community' | 'business' | '',
  selectedAgentId: '',
  selectedNamespaces: [] as string[],
})

const toggleAgentSelection = (agentId: string) => {
  dialogInfo.selectedAgentId = dialogInfo.selectedAgentId === agentId ? '' : agentId
}

const toggleNamespaceSelection = (namespaceId: string) => {
  const index = dialogInfo.selectedNamespaces.indexOf(namespaceId)
  if (index > -1) {
    dialogInfo.selectedNamespaces.splice(index, 1)
  } else {
    dialogInfo.selectedNamespaces.push(namespaceId)
  }
}

const init = (list: any[]) => {
  dialogInfo.visible = true
  dialogInfo.instanceList = list
  dialogInfo.currentStep = 1
  dialogInfo.selectedPlatform = ''
  dialogInfo.selectedAgentId = ''
  dialogInfo.selectedNamespaces = []
  // 模拟命名空间数据，实际应该从 API 获取
  dialogInfo.namespaceList = [
    { id: 'ns-1', name: '默认空间', description: '系统默认命名空间' },
    { id: 'ns-2', name: '开发环境', description: '用于开发测试的命名空间' },
    { id: 'ns-3', name: '生产环境', description: '生产环境专用命名空间' },
  ]
}

defineExpose({
  init,
})
</script>

<style scoped lang="scss">
:deep(.is-finish .el-step__icon) {
  border-color: var(--ep-purple-color);
  color: var(--ep-purple-color);
}
:deep(.is-finish .el-step__line) {
  border-color: var(--ep-purple-color);
  color: var(--ep-purple-color);
}
:deep(.el-step__title.is-finish) {
  color: var(--ep-purple-color);
}
:deep(.el-step__description.is-finish) {
  color: var(--ep-purple-color);
}

.step-content {
  min-height: 450px;
  padding: 0 20px 20px;
}

.platform-container {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 60px;
  padding: 20px;
}

.platform-card {
  position: relative;
  width: 300px;
  height: 260px;
  background: var(--ep-bg-color);
  border: 2px solid var(--ep-border-color);
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.3s ease;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  overflow: hidden;

  &:hover {
    border-color: var(--ep-purple-color);
    transform: translateY(-4px);
    box-shadow: 0 8px 16px rgba(0, 0, 0, 0.1);
  }

  &.active {
    border-color: var(--ep-purple-color);
    border-width: 3px;
    background: var(--ep-bg-purple-color);
    box-shadow: 0 0 20px rgba(124, 77, 255, 0.2);
  }

  .platform-logo {
    margin-bottom: 30px;

    .logo-img {
      width: 120px;
      height: auto;
    }
  }

  .platform-badge {
    padding: 8px 24px;
    border-radius: 6px;
    font-size: 16px;
    font-weight: 600;
    color: white;

    &.community {
      background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    }

    &.business {
      background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
    }
  }
}

// Selection Info
.selection-info {
  text-align: center;
  padding: 0 0 20px;
  font-size: 14px;
  color: var(--ep-text-color-regular);

  .count-highlight {
    color: var(--ep-purple-color);
    font-weight: 600;
    margin: 0 4px;
  }

  .count-total {
    font-weight: 600;
    margin: 0 4px;
  }
}

// Agent Cards Grid Layout
.agent-grid {
  max-height: 380px;
  overflow-y: auto;
  padding-bottom: 20px;

  &::-webkit-scrollbar {
    width: 6px;
  }

  &::-webkit-scrollbar-track {
    background: transparent;
  }

  &::-webkit-scrollbar-thumb {
    background: var(--ep-border-color);
    border-radius: 3px;

    &:hover {
      background: var(--ep-border-color-dark);
    }
  }
}

.agent-card {
  position: relative;
  background: var(--ep-bg-color);
  border: 1px solid var(--ep-border-color);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.3s ease;
  height: 100px;

  &:hover {
    border-color: var(--ep-purple-color);
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  }

  &.active {
    border-color: var(--ep-purple-color);
    background: var(--ep-bg-purple-color);
    box-shadow: 0 0 12px rgba(124, 77, 255, 0.15);
  }

  .selected-badge {
    position: absolute;
    top: 0;
    right: 0;
    width: 0px;
    height: 0px;
    border-style: solid;
    border-width: 0 40px 40px 0;
    border-color: transparent var(--ep-purple-color) transparent transparent;
    border-top-right-radius: 6px;
  }

  .agent-icon {
    flex-shrink: 0;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .agent-info {
    display: flex;
    flex-direction: column;
    justify-content: center;
    .agent-name {
      font-size: 14px;
      font-weight: 600;
      color: var(--ep-text-color-primary);
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
      padding: 0 4px;
    }

    .agent-desc {
      font-size: 12px;
      color: var(--ep-text-color-secondary);
      line-height: 1.4;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
      padding: 0 4px;
    }
  }
}

.empty-state {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 400px;
}
</style>

<style lang="scss">
.el-dialog__footer.footer-border {
  background-color: transparent !important;
  border-top: 1px solid var(--el-border-color-light) !important;
}
</style>
