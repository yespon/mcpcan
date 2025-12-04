<template>
  <el-dialog
    v-model="dialogInfo.visible"
    :close-on-click-modal="false"
    width="1200px"
    footer-class="footer-border"
  >
    <template #title>
      <div class="mb-4">
        {{ dialogInfo.title }} : 已选择
        <span class="count-highlight">{{ dialogInfo.instanceList.length }}</span> 个MCP服务
      </div>
    </template>
    <div v-loading="dialogInfo.loading" :element-loading-text="dialogInfo.loadingText">
      <el-steps :active="dialogInfo.currentStep" align-center class="mb-6">
        <el-step
          title="智能体平台选择"
          :description="
            dialogInfo.selectedAgentPlatformId
              ? '平台：' +
                  dialogInfo.platformList.find(
                    (agent) => agent.accessID === dialogInfo.selectedAgentPlatformId,
                  )?.accessName || dialogInfo.selectedAgentPlatformId
              : '请选择一个你需要同步的智能体平台'
          "
        />
        <el-step
          title="同步的命名空间"
          :description="
            dialogInfo.selectedNamespaces.length
              ? '已选择空间：' + (dialogInfo.selectedNamespaces.length || 0) + '个'
              : '请选择一个你需要同步的空间'
          "
        />
        <el-step title="鉴权设置" :description="'请给命名空间进行鉴权设置'"> </el-step>
      </el-steps>

      <div v-if="dialogInfo.currentStep === 1" class="step-content">
        <!-- Step 1: 智能体平台选择 -->
        <div class="mb-6 flex items-center gap-2 mx-4">
          <div class="flex items-center w-full">
            <span class="color-red mr-1">*</span>
            同步任务：<el-input
              v-model="dialogInfo.desc"
              style="width: 300px"
              placeholder="请输入同步任务名称"
            ></el-input>
          </div>
        </div>
        <div class="agent-grid grid grid-cols-12 gap-4 px-3">
          <div
            v-for="(agent, index) in dialogInfo.platformList"
            :key="index"
            class="agent-card col-span-3 center py-3 px-4"
            :class="{ active: dialogInfo.selectedAgentPlatformId === agent.accessID }"
            @click="toggleAgentSelection(agent.accessID)"
          >
            <div class="w-full flex">
              <div
                v-if="dialogInfo.selectedAgentPlatformId === agent.accessID"
                class="selected-badge"
              ></div>
              <div class="agent-icon">
                <el-icon class="cursor-pointer" size="48" color="var(--ep-purple-color)"
                  ><i class="icon iconfont MCP-zhinengti"></i
                ></el-icon>
              </div>
              <div class="agent-info flex-sub u-line-1">
                <div class="agent-name">{{ agent.accessName || '未命名智能体' }}</div>
                <div class="agent-desc my-1">
                  类型：{{ agent.accessType === 'Dify' ? '社区版' : '商业版' }}
                </div>
              </div>
            </div>
          </div>
        </div>
        <div v-if="!dialogInfo.instanceList.length" class="empty-state">
          <el-empty description="暂无可同步的智能体" />
        </div>
      </div>
      <div v-else-if="dialogInfo.currentStep === 2" class="step-content">
        <!-- Step 2: 选择空间 -->
        <div class="agent-grid grid grid-cols-12 gap-4 px-3">
          <div
            v-for="(namespace, index) in dialogInfo.namespaceList"
            :key="index"
            class="agent-card col-span-3 center py-3 px-4"
            :class="{ active: dialogInfo.selectedNamespaces.includes(namespace.tenantID) }"
            @click="toggleNamespaceSelection(namespace.tenantID)"
          >
            <div class="w-full flex">
              <div
                v-if="dialogInfo.selectedNamespaces.includes(namespace.tenantID)"
                class="selected-badge"
              ></div>
              <div class="agent-icon">
                <el-icon class="cursor-pointer" size="48" color="var(--ep-purple-color)"
                  ><i class="icon iconfont MCP-mingmingkongjian"></i
                ></el-icon>
              </div>
              <div class="agent-info flex-sub u-line-1">
                <div class="agent-name">{{ namespace.tenantName || '未命名空间' }}</div>
                <div class="agent-desc my-1">用户：{{ namespace.userName || '暂无描述' }}</div>
              </div>
            </div>
          </div>
        </div>
        <div v-if="!dialogInfo.namespaceList.length" class="empty-state">
          <el-empty description="暂无可同步的空间" />
        </div>
      </div>
      <div v-else-if="dialogInfo.currentStep === 3" class="step-content">
        <!-- Step 3: 鉴权设置 -->
        <el-splitter class="h-full" direction="horizontal" :gutter="8">
          <el-splitter-panel size="30%" class="px-2">
            <div
              v-for="(namespace, index) in selectedNamespaceList"
              :key="index"
              class="agent-card center py-3 px-4 mb-2"
              :class="{ active: dialogInfo.selectedNamespaceId === namespace.tenantID }"
              @click="dialogInfo.selectedNamespaceId = namespace.tenantID"
            >
              <div class="w-full flex">
                <div
                  v-if="dialogInfo.selectedNamespaceId === namespace.tenantID"
                  class="selected-badge"
                ></div>
                <div class="agent-icon">
                  <el-icon class="cursor-pointer" size="48" color="var(--ep-purple-color)"
                    ><i class="icon iconfont MCP-zhinengti"></i
                  ></el-icon>
                </div>
                <div class="agent-info flex-sub u-line-1">
                  <div class="agent-name">
                    {{ namespace.tenantName }}
                  </div>
                  <div class="agent-desc">
                    {{ namespace.userName }}
                  </div>
                </div>
              </div>
            </div>
          </el-splitter-panel>
          <el-splitter-panel size="70%" class="px-2">
            <div
              v-for="(item, index) in selectedNamespaceList.find(
                (ns) => ns.tenantID === dialogInfo.selectedNamespaceId,
              )?.headers || []"
              :key="index"
              class="py-3 px-4 mb-2"
            >
              <div class="pl-2 mb-2">MCP名称: {{ item?.instanceName }}</div>
              <TokenFormSync :formData="item.value"></TokenFormSync>

              <!-- <div class="pl-2">
                <div class="w-full u-line-1 pl-2 py-2" style="white-space: nowrap">
                  Authorization：{{ (instance as { Token: string }).Token }}
                </div>
                <div class="flex justify-between items-center">
                  Header 请求头
                  <div
                    class="cursor-pointer border border-style-solid border-rd-md border-white ml-2 p-1 center bg-gray-600 color-white hover-scale-110"
                  >
                    <el-icon>
                      <Plus />
                    </el-icon>
                  </div>
                </div>
                <div v-for="head in (instance as { headers?: any[] }).headers" :key="head.id">
                  <el-row>
                    <el-col :span="8">
                      <el-input
                        v-model="head.key"
                        :placeholder="t('mcp.instance.token.headersKey')"
                        class="flex-sub"
                      />
                    </el-col>
                    <el-col :span="14">
                      <el-input
                        v-model="head.value"
                        :placeholder="t('mcp.instance.token.headersValue')"
                        class="flex-sub"
                      ></el-input>
                    </el-col>
                    <el-col :span="2">
                      <el-icon><Minus /></el-icon>
                    </el-col>
                  </el-row>
                </div>
              </div> -->
            </div>
          </el-splitter-panel>
        </el-splitter>
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
            (dialogInfo.currentStep === 1 && !dialogInfo.selectedAgentPlatformId) ||
            (dialogInfo.currentStep === 2 && dialogInfo.selectedNamespaces.length === 0)
          "
          @click="handleNextStep"
          >下一步</el-button
        >
        <el-button
          type="primary"
          class="base-btn"
          v-else
          :disabled="dialogInfo.selectedNamespaces.length === 0"
          @click="handleConfirmSync"
        >
          完 成
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { AgentAPI } from '@/api/agent'
import TokenFormSync from './components/token-form-sync.vue'
import { useBusinessStoreHook } from '@/stores/modules/business-store'

const { t } = useI18n()
const { taskInfo } = toRefs(useBusinessStoreHook())
const dialogInfo = reactive({
  title: '智能体平台同步',
  visible: false,
  loading: false,
  loadingText: 'Loading...',
  desc: '',
  instanceList: [] as any[],
  platformList: [] as any[],
  namespaceList: [] as any[],
  currentStep: 1,
  selectedPlatform: '',
  selectedAgentPlatformId: '',
  selectedNamespaces: [] as string[],
  selectedNamespaceId: '',
})

// 选中的命名空间列表
const selectedNamespaceList = computed(() => {
  return dialogInfo.namespaceList.filter((namespace: any) =>
    dialogInfo.selectedNamespaces.some((id) => namespace.tenantID === id),
  )
})

const handleNextStep = () => {
  dialogInfo.currentStep += 1
  if (dialogInfo.currentStep === 2) {
    handleGetNamespaceList(dialogInfo.selectedAgentPlatformId)
  }
}

// 选择智能体平台
const toggleAgentSelection = (accessID: string) => {
  dialogInfo.selectedAgentPlatformId =
    dialogInfo.selectedAgentPlatformId === accessID ? '' : accessID
  dialogInfo.desc =
    '同步任务-' +
      dialogInfo.platformList.find((agent) => agent.accessID === dialogInfo.selectedAgentPlatformId)
        ?.accessName || accessID
}

// 选择命名空间
const toggleNamespaceSelection = (namespaceId: string) => {
  const index = dialogInfo.selectedNamespaces.indexOf(namespaceId)
  if (index > -1) {
    dialogInfo.selectedNamespaces.splice(index, 1)
  } else {
    dialogInfo.selectedNamespaces.push(namespaceId)
  }
}

// handle get agent platform list
const handleGetAgentPlatform = async () => {
  const { list } = await AgentAPI.list({ page: 1, pageSize: 1000 })
  dialogInfo.platformList = list || []
}

// handle get namespace list
const handleGetNamespaceList = async (accessID: string) => {
  try {
    dialogInfo.loading = true
    dialogInfo.loadingText = '获取命名空间中...'
    const { userSpaces } = await AgentAPI.getNamespaces({
      accessID,
      instancesIDs: dialogInfo.instanceList.map((item) => item.instanceId),
    })

    // 处理数据结构以渲染
    userSpaces.forEach((space: any) => {
      space.headers = Object.entries(space.headers).map(([key, value]: [string, any]) => ({
        instanceId: key,
        instanceName: dialogInfo.instanceList.find((inst) => inst.instanceId === key)?.instanceName,
        value: {
          token: value.token,
          headers:
            Object.entries(value.headers || {}).map(([hKey, hValue]) => ({
              key: hKey,
              value: hValue,
            })) || [],
        },
      }))
    })
    dialogInfo.namespaceList = userSpaces || []
  } finally {
    dialogInfo.loading = false
  }
}

const handleConfirmSync = async () => {
  try {
    dialogInfo.loading = true
    dialogInfo.loadingText = '正在创建同步任务中...'
    const params = {
      desc: dialogInfo.desc,
      intelligentAccessID: dialogInfo.selectedAgentPlatformId,
      insertIntelligentInfos: dialogInfo.namespaceList.map((namespace: any) => ({
        difySpaceID: namespace.tenantID,
        difyUserID: namespace.userID,
        difySpaceName: namespace.tenantName,
        difyUserName: namespace.userName,
        headers: Object.fromEntries(
          namespace.headers.map((item: any) => [
            item.instanceId,
            Object.assign(
              { ...item.value },
              {
                headers: item.value.headers.reduce((acc: any, curr: any) => {
                  acc[curr.key] = curr.value
                  return acc
                }, {}),
              },
            ),
          ]),
        ),
      })),
      mcpInstanceIDs: dialogInfo.instanceList.map((item) => item.instanceId),
      domain: 'https://mcp-dev.itqm.com/' || window.location.origin,
    }

    await AgentAPI.createSyncTask(params)
    dialogInfo.visible = false
    taskInfo.value.visible = true
  } catch (error) {
    console.error('同步失败:', error)
  } finally {
    dialogInfo.loading = false
  }
}

// 数据初始化
const init = (list: any[]) => {
  dialogInfo.visible = true
  dialogInfo.instanceList = list
  dialogInfo.currentStep = 1
  handleGetAgentPlatform()
  dialogInfo.selectedPlatform = ''
  dialogInfo.selectedAgentPlatformId = ''
  dialogInfo.selectedNamespaces = []
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
  height: 450px;
  padding: 0 20px 20px;
}

// Selection Info
.selection-info {
  text-align: center;
  padding: 0 0 20px;
  font-size: 14px;
  color: var(--ep-text-color-regular);

  .count-total {
    font-weight: 600;
    margin: 0 4px;
  }
}
.count-highlight {
  color: var(--ep-purple-color);
  font-weight: 600;
  margin: 0 4px;
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
