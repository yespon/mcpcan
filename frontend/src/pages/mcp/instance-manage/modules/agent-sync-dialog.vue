<template>
  <el-dialog
    v-model="dialogInfo.visible"
    :close-on-click-modal="false"
    width="1200px"
    footer-class="footer-border"
  >
    <template #title>
      <div class="mb-4">
        {{ dialogInfo.title }} : {{ t('agent.sync.selected') }}
        <span class="count-highlight">{{ dialogInfo.instanceList.length }}</span>
        {{ t('agent.sync.unitMCP') }}
      </div>
    </template>
    <div v-loading="dialogInfo.loading" :element-loading-text="dialogInfo.loadingText">
      <el-steps :active="dialogInfo.currentStep" align-center class="mb-6">
        <el-step
          :title="t('agent.sync.stepOne')"
          :description="
            dialogInfo.selectedAgentPlatformId
              ? t('agent.sync.platform') +
                  dialogInfo.platformList.find(
                    (agent) => agent.accessID === dialogInfo.selectedAgentPlatformId,
                  )?.accessName || dialogInfo.selectedAgentPlatformId
              : t('agent.sync.descStepOne')
          "
        />
        <el-step
          :title="t('agent.sync.stepTwo')"
          :description="
            dialogInfo.selectedNamespaces.length
              ? t('agent.sync.selectSpace') +
                (dialogInfo.selectedNamespaces.length || 0) +
                t('agent.sync.unit')
              : t('agent.sync.descStepTwo')
          "
        />
        <el-step :title="t('agent.sync.stepThree')" :description="t('agent.sync.descStepThree')">
        </el-step>
      </el-steps>

      <div v-if="dialogInfo.currentStep === 1" class="step-content">
        <!-- Step 1: platform selection -->
        <div class="mb-6 flex items-center gap-2 mx-4">
          <div class="flex items-center w-full">
            <span class="color-red mr-1">*</span>
            {{ t('agent.sync.taskName') }}：<el-input
              v-model="dialogInfo.desc"
              style="width: 300px"
              :placeholder="t('agent.sync.taskNamePlaceholder')"
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
                <el-icon class="cursor-pointer" size="48" color="var(--el-color-primary)"
                  ><i class="icon iconfont MCP-zhinengti"></i
                ></el-icon>
              </div>
              <div class="agent-info flex-sub u-line-1">
                <div class="agent-name">{{ agent.accessName || t('agent.sync.noAccessName') }}</div>
                <div class="agent-desc my-1">
                  {{ t('agent.sync.type')
                  }}{{
                    agent.accessType === 'Dify'
                      ? t('agent.action.community')
                      : t('agent.action.enterprise')
                  }}
                </div>
              </div>
            </div>
          </div>
        </div>
        <div v-if="!dialogInfo.platformList.length" class="empty-state">
          <el-empty :description="t('agent.sync.emptyDesc')">
            <el-button link class="base-btn-link" @click="handleJumptoAgent"
              >去添加智能体平台</el-button
            >
          </el-empty>
        </div>
      </div>
      <div v-else-if="dialogInfo.currentStep === 2" class="step-content">
        <!-- Step 2: selection nameSpace -->
        <div class="flex items-center mb-2 px-3 justify-between">
          <div class="center">
            <el-checkbox
              v-model="allNamespaceChecked"
              :indeterminate="isNamespaceIndeterminate"
              @change="handleCheckAllNamespace"
              >{{ t('agent.sync.selectAll') }}</el-checkbox
            >
            <span class="ml-2"
              >{{ t('agent.sync.selected') }}{{ dialogInfo.selectedNamespaces.length
              }}{{ t('agent.sync.unitSpace') }}</span
            >
          </div>

          <el-input
            v-model="namespaceStep2Search"
            :placeholder="t('agent.sync.searchPlaceholder')"
            clearable
            class="ml-4"
            size="small"
            style="width: 260px"
          />
        </div>
        <div class="agent-grid grid grid-cols-12 gap-4 px-3">
          <div
            v-for="(namespace, index) in filteredNamespaceStep2List"
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
                <el-icon class="cursor-pointer" size="48" color="var(--el-color-primary)"
                  ><i class="icon iconfont MCP-mingmingkongjian"></i
                ></el-icon>
              </div>
              <div class="agent-info flex-sub u-line-1">
                <div class="agent-name">
                  {{ namespace.tenantName || t('agent.sync.noNamespace') }}
                </div>
                <div class="agent-desc my-1">
                  {{ t('agent.sync.user') }}：{{ namespace.userName || t('agent.sync.noUserName') }}
                </div>
              </div>
            </div>
          </div>
        </div>
        <div v-if="!filteredNamespaceStep2List.length" class="empty-state">
          <el-empty :description="t('agent.sync.emptyNamespaceDesc')" />
        </div>
      </div>
      <div v-else-if="dialogInfo.currentStep === 3" class="step-content">
        <!-- Step 3 -->
        <el-splitter class="h-full" direction="horizontal" :gutter="8">
          <el-splitter-panel size="30%" class="px-2">
            <div class="center search-bar">
              <el-input
                v-model="namespaceSearch"
                :placeholder="t('agent.sync.searchPlaceholder')"
                clearable
                class="mb-2 mt-1 flex-sub"
                size="small"
              />
              <span class="ml-2"
                >{{ t('agent.sync.selected') }}{{ dialogInfo.selectedNamespaces.length
                }}{{ t('agent.sync.unitSpace') }}</span
              >
            </div>
            <div
              v-for="(namespace, index) in filteredNamespaceList"
              :key="index"
              class="flex items-center"
            >
              <el-checkbox :model-value="true" @click.stop disabled></el-checkbox>
              <div
                class="agent-card center py-3 px-4 mb-2 ml-2 flex-sub min-w-0"
                :class="{ active: dialogInfo.selectedNamespaceId === namespace.tenantID }"
                @click="dialogInfo.selectedNamespaceId = namespace.tenantID"
              >
                <div class="w-full flex">
                  <div
                    v-if="dialogInfo.selectedNamespaceId === namespace.tenantID"
                    class="selected-badge"
                  ></div>
                  <div class="agent-icon">
                    <el-icon class="cursor-pointer" size="48" color="var(--el-color-primary)"
                      ><i class="icon iconfont MCP-zhinengti"></i
                    ></el-icon>
                  </div>
                  <div class="agent-info flex-sub u-line-1 w-full">
                    <div class="agent-name ellipsis-one w-full">
                      {{ namespace.tenantName }}
                    </div>
                    <div class="agent-desc">
                      {{ namespace.userName }}
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </el-splitter-panel>
          <el-splitter-panel size="70%" class="px-2">
            <div class="search-bar">
              <el-input
                v-model="instanceSearch"
                :placeholder="t('agent.sync.instanceSearchPlaceholder')"
                clearable
                class="mb-2 mt-1"
                size="small"
                style="width: 50%"
              />
            </div>
            <div v-for="(item, index) in filteredInstanceList" :key="index" class="py-3 px-4 mb-2">
              <div class="pl-2 mb-2">{{ t('agent.sync.MCPName') }}: {{ item?.instanceName }}</div>
              <TokenFormSync :formData="item.value"></TokenFormSync>
              <el-divider />
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
          >{{ t('agent.sync.stepUp') }}</el-button
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
          >{{ t('agent.sync.stepNext') }}</el-button
        >
        <el-button
          type="primary"
          class="base-btn"
          v-else
          :disabled="dialogInfo.selectedNamespaces.length === 0"
          @click="handleConfirmSync"
        >
          {{ t('agent.sync.confirmSync') }}
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { AgentAPI } from '@/api/agent'
import TokenFormSync from './components/token-form-sync.vue'
import { useBusinessStoreHook } from '@/stores/modules/business-store'
import { useRouterHooks } from '@/utils/url'

const { t } = useI18n()
const { taskInfo } = toRefs(useBusinessStoreHook())
const { jumpToPage } = useRouterHooks()
const dialogInfo = reactive({
  title: t('agent.sync.dialogTitle'),
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

// selected namespace list based on selectedNamespaces ids
const selectedNamespaceList = computed(() => {
  return dialogInfo.namespaceList.filter((namespace: any) =>
    dialogInfo.selectedNamespaces.some((id) => namespace.tenantID === id),
  )
})

// search nameSpace list by name
const namespaceSearch = ref('')
const filteredNamespaceList = computed(() => {
  if (!namespaceSearch.value) return selectedNamespaceList.value
  return selectedNamespaceList.value.filter((ns: any) =>
    ns.tenantName?.toLowerCase().includes(namespaceSearch.value.trim().toLowerCase()),
  )
})

// step 2 namespace search
const namespaceStep2Search = ref('')
const filteredNamespaceStep2List = computed(() => {
  if (!namespaceStep2Search.value) return dialogInfo.namespaceList
  return dialogInfo.namespaceList.filter((ns: any) =>
    ns.tenantName?.toLowerCase().includes(namespaceStep2Search.value.trim().toLowerCase()),
  )
})

// search MCP实例 by name（右侧）
const instanceSearch = ref('')
const filteredInstanceList = computed(() => {
  const ns = selectedNamespaceList.value.find(
    (ns) => ns.tenantID === dialogInfo.selectedNamespaceId,
  )
  if (!ns) return []
  if (!instanceSearch.value) return ns.headers || []
  return (ns.headers || []).filter((item: any) =>
    item.instanceName?.toLowerCase().includes(instanceSearch.value.trim().toLowerCase()),
  )
})

// select all namespaces
const allNamespaceChecked = computed({
  get() {
    return (
      dialogInfo.namespaceList.length > 0 &&
      dialogInfo.selectedNamespaces.length === dialogInfo.namespaceList.length
    )
  },
  set(val: boolean) {
    if (val) {
      dialogInfo.selectedNamespaces = dialogInfo.namespaceList.map((ns: any) => ns.tenantID)
    } else {
      dialogInfo.selectedNamespaces = []
    }
  },
})

// select all status
const isNamespaceIndeterminate = computed(() => {
  return (
    dialogInfo.selectedNamespaces.length > 0 &&
    dialogInfo.selectedNamespaces.length < dialogInfo.namespaceList.length
  )
})

const handleJumptoAgent = () => {
  jumpToPage({
    url: '/agent-manage',
  })
}

// select all or cancel all
const handleCheckAllNamespace = (val: boolean) => {
  allNamespaceChecked.value = val
}

// next step
const handleNextStep = () => {
  dialogInfo.currentStep += 1
  //  step two request space list
  if (dialogInfo.currentStep === 2) {
    handleGetNamespaceList(dialogInfo.selectedAgentPlatformId)
  }
  // step three default select first namespace
  if (dialogInfo.currentStep === 3) {
    dialogInfo.selectedNamespaceId = selectedNamespaceList.value[0]?.tenantID || ''
  }
}

// select agent platformshuaxin
const toggleAgentSelection = (accessID: string) => {
  dialogInfo.selectedAgentPlatformId =
    dialogInfo.selectedAgentPlatformId === accessID ? '' : accessID
  dialogInfo.desc =
    t('agent.sync.taskName') +
      '-' +
      dialogInfo.platformList.find((agent) => agent.accessID === dialogInfo.selectedAgentPlatformId)
        ?.accessName || accessID
}

// select namespace
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
    dialogInfo.loadingText = t('agent.sync.loadingTextNamespace')
    const { userSpaces } = await AgentAPI.getNamespaces({
      accessID,
      instancesIDs: dialogInfo.instanceList.map((item) => item.instanceId),
    })

    // handle data structure to render
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
    dialogInfo.loadingText = t('agent.sync.loadingTextTask')
    const params = {
      desc: dialogInfo.desc,
      intelligentAccessID: dialogInfo.selectedAgentPlatformId,
      insertIntelligentInfos: filteredNamespaceList.value.map((namespace: any) => ({
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
      domain: window.location.origin + (window as any).__APP_CONFIG__?.PUBLIC_PATH,
    }

    await AgentAPI.createSyncTask(params)
    dialogInfo.visible = false
    taskInfo.value.visible = true
  } catch (error) {
    console.error('sync error:', error)
  } finally {
    dialogInfo.loading = false
  }
}

// init dialog data
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
  border-color: var(--el-color-primary);
  color: var(--el-color-primary);
}
:deep(.is-finish .el-step__line) {
  border-color: var(--el-color-primary);
  color: var(--el-color-primary);
}
:deep(.el-step__title.is-finish) {
  color: var(--el-color-primary);
}
:deep(.el-step__description.is-finish) {
  color: var(--el-color-primary);
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
  color: var(--el-color-primary);
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
    border-color: var(--el-color-primary);
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  }

  &.active {
    border-color: var(--el-color-primary);
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
    border-color: transparent var(--el-color-primary) transparent transparent;
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
.search-bar {
  position: sticky;
  top: 0;
  z-index: 999;
  background-color: var(--el-bg-color);
}
.empty-state {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 400px;
}
:deep(.el-checkbox__input.is-checked + .el-checkbox__label) {
  color: var(--el-color-primary);
}
</style>

<style lang="scss">
.el-dialog__footer.footer-border {
  background-color: transparent !important;
  border-top: 1px solid var(--el-border-color-light) !important;
}
</style>
