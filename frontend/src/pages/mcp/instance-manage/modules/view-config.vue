<template>
  <el-dialog v-model="dialogInfo.visible" width="80%" top="5vh">
    <template #header>
      <div class="center">{{ dialogInfo.title }}</div>
    </template>
    <div
      class="mb-2 flex items-center"
      v-if="dialogInfo.instanceInfo.accessType !== AccessType.DIRECT"
    >
      <el-switch
        v-model="dialogInfo.instanceInfo.enabledToken"
        style="--el-switch-on-color: #13ce66"
        inline-prompt
        :loading="dialogInfo.instanceInfo.loading"
        :active-text="t('common.on')"
        :inactive-text="t('common.off')"
        @change="handleEabledToken()"
      ></el-switch>
      <span class="ml-2">{{ t('mcp.instance.tableHeadDesc.token') }}</span>
    </div>
    <el-row :gutter="12" v-loading="dialogInfo.instanceInfo.loading">
      <el-col
        :span="12"
        :class="[
          'config-col left',
          {
            collapsed:
              !dialogInfo.instanceInfo.enabledToken ||
              dialogInfo.instanceInfo.accessType === AccessType.DIRECT,
          },
        ]"
      >
        <el-scrollbar ref="scrollbarRef" max-height="75vh" always>
          <div class="token-list w-full">
            <div class="font-bold flex justify-between items-center">
              <!-- <mcp-button :icon="Plus" size="small" @click="handleAddToken">{{
                t('mcp.instance.formData.addToken')
              }}</mcp-button> -->
              <el-checkbox
                v-model="allTokenChecked"
                :indeterminate="isTokenIndeterminate"
                @change="handleCheckAllToken"
              >
                <span style="color: var(--el-color-primary)">{{ t('agent.sync.selectAll') }}</span>
              </el-checkbox>
              <div>
                <span class="mr-4 color-green">
                  {{ t('mcp.instance.token.active') }}:
                  {{ tokenList.filter((token) => !token.expire).length }}
                </span>
                <span class="color-red">
                  {{ t('mcp.instance.token.expired') }}:{{
                    tokenList.filter((token) => token.expire).length
                  }}
                </span>
              </div>
              <div class="flex">
                <SearchForm
                  :formConfig="tokenFormConfig"
                  :formData="tokenFormData"
                  @handle-query="handleQueryToken"
                  @reset-fields="handleQueryToken"
                  @on-refresh="handleTokenList"
                >
                  <template #operation>
                    <el-dropdown trigger="click" class="ml-3" :show-arrow="false">
                      <el-button style="width: 32px">
                        <el-icon><Operation /></el-icon>
                      </el-button>
                      <template #dropdown>
                        <el-dropdown-menu>
                          <el-dropdown-item command="handleAddInstance" @click="handleAddToken">
                            <el-icon><i class="icon iconfont MCP-a-1"></i></el-icon>
                            {{ t('mcp.instance.formData.addToken') }}
                          </el-dropdown-item>
                          <el-dropdown-item
                            command="handleAddInstance"
                            @click="handleBatchChangeHeader"
                          >
                            <el-icon><Edit /></el-icon>
                            {{ t('mcp.instance.formData.batchChangeHeader') }}
                          </el-dropdown-item>
                        </el-dropdown-menu>
                      </template>
                    </el-dropdown>
                  </template>
                </SearchForm>
              </div>
            </div>
            <RecycleScroller
              class="scroller hide-scrollbar mt-4"
              :items="showTokenList"
              :item-size="140"
              key-field="id"
              v-slot="{ item: token, index }"
            >
              <div class="w-full">
                <div class="flex items-center w-full">
                  <div class="mr-2 flex-shrink-0">
                    <el-checkbox
                      :model-value="selectedTokens.includes(token.id)"
                      @change="toggleTokenSelection(token.id)"
                      @click.stop
                    ></el-checkbox>
                  </div>
                  <div
                    class="token-card border-rounded-2 p-4 cursor-pointer line-height-6 flex-1 min-w-0"
                    :class="[
                      {
                        active: dialogInfo.currentTokenIndex === index,
                        disabled: token.expireAt !== 0 && token.expireAt < Date.now(),
                      },
                    ]"
                    @click="handleSelectedToken(index)"
                  >
                    <div class="flex">
                      <div class="ellipsis-one flex-sub">
                        {{ token.token }}
                      </div>
                      <el-dropdown
                        trigger="click"
                        class="ml-2"
                        style="cursor: pointer !important"
                        @click.stop
                        :show-arrow="false"
                      >
                        <el-icon size="18"><Operation /></el-icon>
                        <template #dropdown>
                          <el-dropdown-menu>
                            <el-dropdown-item @click="handleEditToken(index)">
                              <el-button link>
                                {{ t('env.run.action.edit') }}
                              </el-button>
                            </el-dropdown-item>
                            <el-dropdown-item @click="handleViewLog(index)">
                              <el-button link>
                                {{ t('mcp.instance.action.accessLogs') }}
                              </el-button>
                            </el-dropdown-item>
                            <el-dropdown-item
                              v-if="showDeleteBtn(token)"
                              @click="handleDeleteToken(index)"
                            >
                              <el-button type="danger" link>
                                {{ t('mcp.instance.action.delete') }}
                              </el-button>
                            </el-dropdown-item>
                          </el-dropdown-menu>
                        </template>
                      </el-dropdown>
                    </div>
                    <div class="grid grid-cols-2 mt-2">
                      <div class="ellipsis-one grid-cols-span-1">
                        {{ t('mcp.instance.token.expireAt') }}：<span
                          :class="
                            !token.expireAt
                              ? 'color-green'
                              : token.expireAt < Date.now()
                                ? 'color-red'
                                : ''
                          "
                        >
                          {{
                            !token.expireAt
                              ? t('mcp.instance.token.placeholderAlways')
                              : timestampToDate(token.expireAt) +
                                (token.expireAt < Date.now() ? t('mcp.instance.token.expired') : '')
                          }}
                        </span>
                      </div>
                      <div class="ellipsis-one grid-cols-span-1">
                        {{ t('mcp.instance.createTime') }}：{{ timestampToDate(token.publishAt) }}
                      </div>
                    </div>
                    <div class="grid grid-cols-2 mt-2">
                      <div class="ellipsis-one grid-cols-span-1 w-full">
                        {{ t('mcp.instance.token.tag') }}：<el-tag
                          v-for="(tag, num) in token.usages"
                          :key="num"
                          effect="plain"
                          class="mr-2"
                        >
                          <div class="ellipsis-one max-w-25">
                            {{ tag }}
                          </div>
                        </el-tag>
                      </div>
                      <div class="grid-cols-span-1">
                        {{ t('mcp.token.isEnable') }}:
                        <el-switch
                          v-model="token.enabled"
                          style="--el-switch-on-color: #13ce66"
                          inline-prompt
                          :loading="dialogInfo.instanceInfo.loading"
                          :active-text="t('status.active')"
                          :inactive-text="t('status.inactive')"
                          @change="handleChangeTransport(token, index)"
                        ></el-switch>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </RecycleScroller>
          </div>
        </el-scrollbar>
      </el-col>
      <el-col
        :span="12"
        :class="[
          'config-col right',
          {
            expanded:
              !dialogInfo.instanceInfo.enabledToken ||
              dialogInfo.instanceInfo.accessType === AccessType.DIRECT,
          },
        ]"
      >
        <el-scrollbar ref="scrollbarRef" always class="config-info">
          <div class="py-5 px-5">{{ config }}</div>
          <el-tooltip
            class="box-item"
            effect="dark"
            :content="t('mcp.instance.token.copyUrl')"
            placement="top"
          >
            <el-icon class="base-btn-link copy-icon-url" size="18" @click="handleCopy('url')">
              <Link />
            </el-icon>
          </el-tooltip>
          <el-tooltip
            v-if="dialogInfo.currentTokenIndex !== null"
            class="box-item"
            effect="dark"
            :content="t('mcp.instance.token.copyToken')"
            placement="top"
          >
            <el-icon class="base-btn-link copy-icon-token" size="18" @click="handleCopy('token')">
              <Key />
            </el-icon>
          </el-tooltip>
          <el-tooltip
            class="box-item"
            effect="dark"
            :content="t('mcp.instance.token.copyAll')"
            placement="top"
          >
            <el-icon class="base-btn-link copy-icon" size="18" @click="handleCopy('config')">
              <CopyDocument />
            </el-icon>
          </el-tooltip>
        </el-scrollbar>
      </el-col>
    </el-row>
    <template #footer>
      <div class="center">
        <!-- <mcp-button @click="handleCopy" class="w100">{{
          t('mcp.instance.action.copy')
        }}</mcp-button> -->
      </div>
    </template>
  </el-dialog>

  <TokenForm
    v-model="formData"
    :instance-info="dialogInfo.instanceInfo"
    :token-list="tokenList"
    :current-edit-index="dialogInfo.currentEditIndex"
    @on-confirm="handleConfirmToken"
    @on-cancel="handleTokenList"
  ></TokenForm>

  <BatchChangeHeader
    ref="batchChangeHeaderRef"
    @on-confirm="handleCommitHeader"
  ></BatchChangeHeader>
</template>
<script setup lang="ts">
import { setClipboardData, timestampToDate } from '@/utils/system'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Operation, CopyDocument, Link, Key, Edit } from '@element-plus/icons-vue'
// import McpButton from '@/components/mcp-button/index.vue'
import { AccessType, TokenType, type InstanceResult } from '@/types'
import { InstanceAPI, TokenAPI } from '@/api/mcp/instance'
import { getToken } from '@/utils/system'
import { useUserStore } from '@/stores'
import { cloneDeep } from 'lodash-es'
import { JsonFormatter } from '@/utils/json'
import { useRouterHooks } from '@/utils/url'
import SearchForm from '@/components/SearchForm/index.vue'
import TokenForm from './components/token-form-list.vue'
import BatchChangeHeader from './components/batch-change-header.vue'
// @ts-expect-error - vue-virtual-scroller 缺少类型定义
import { RecycleScroller } from 'vue-virtual-scroller'
import 'vue-virtual-scroller/dist/vue-virtual-scroller.css'

const { jumpToPage } = useRouterHooks()
const { t } = useI18n()
const { userInfo } = useUserStore()
const emit = defineEmits<{
  (e: 'on-refresh'): void
}>()
const batchChangeHeaderRef = ref()
const formRef = ref()
const formData = ref<any>({
  visible: false,
  token: '',
  headers: [{ key: 'Authorization', value: '' }],
  tokenType: 1 as TokenType | null,
  enabled: true,
  expireAt: null as number | null,
  usages: [] as string[],
})
const tokenFormConfig = ref([
  {
    span: 5,
    key: 'name',
    component: 'el-input',
    label: t('mcp.instance.token.tagName'),
    labelWidth: '60px',
    props: { placeholder: t('mcp.instance.token.tagNamePlaceholder') },
  },
  {
    span: 5,
    key: 'type',
    label: t('mcp.instance.token.tagType'),
    component: 'el-select',
    props: {
      placeholder: t('mcp.instance.token.tagTypePlaceholder'),
      options: [
        { label: 'dify_user_id', value: 'dify_user_id' },
        { label: 'dify_user_name', value: 'dify_user_name' },
        { label: 'dify_space_id', value: 'dify_space_id' },
        { label: 'dify_space_name', value: 'dify_space_name' },
        { label: 'intelligent_access_id', value: 'intelligent_access_id' },
        { label: 'intelligent_access_name', value: 'intelligent_access_name' },
        { label: 'intelligent_access_type', value: 'intelligent_access_type' },
        { label: 'default', value: 'default' },
      ],
    },
  },
  {
    span: 5,
    key: 'value',
    label: t('mcp.instance.token.tagValue'),
    component: 'el-input',
    props: {
      placeholder: t('mcp.instance.token.tagValuePlaceholder'),
    },
  },
  {
    span: 5,
    key: 'handler',
    component: 'slot',
    slotName: 'handler',
  },
])
const tokenFormData = ref({
  name: '',
  type: '',
  value: '',
})

const userDataKey = ref({
  visible: false,
  username: '',
  password: '',
  index: 0,
})
const dialogInfo = ref({
  visible: false,
  title: t('mcp.instance.config'),
  instanceInfo: {} as InstanceResult,
  currentTokenIndex: null as number | null,
  currentEditIndex: null as number | null,
})

// token selection
const selectedTokens = ref<number[]>([])

// select all tokens
const allTokenChecked = computed({
  get() {
    return (
      showTokenList.value.length > 0 && selectedTokens.value.length === showTokenList.value.length
    )
  },
  set(val: boolean) {
    if (val) {
      selectedTokens.value = showTokenList.value.map((token) => token.id)
    } else {
      selectedTokens.value = []
    }
  },
})

// select all status
const isTokenIndeterminate = computed(() => {
  return selectedTokens.value.length > 0 && selectedTokens.value.length < showTokenList.value.length
})

// select all or cancel all
const handleCheckAllToken = (val: boolean) => {
  allTokenChecked.value = val
}

// toggle token selection
const toggleTokenSelection = (id: number) => {
  const idx = selectedTokens.value.findIndex((tokenId) => tokenId === id)
  if (idx > -1) {
    selectedTokens.value.splice(idx, 1)
  } else {
    selectedTokens.value.push(id)
  }
}

const configUrl = computed(() => {
  if (dialogInfo.value.instanceInfo.accessType === AccessType.DIRECT) {
    const mcpServers = JSON.parse(dialogInfo.value.instanceInfo.sourceConfig).mcpServers
    return mcpServers[Object.keys(mcpServers)[0]].url
  }
  return `${window.location.origin}${(window as any).__APP_CONFIG__?.PUBLIC_PATH}${dialogInfo.value.instanceInfo.publicProxyPath}`
})
const configToken = computed(() => {
  if (dialogInfo.value.instanceInfo.accessType === AccessType.DIRECT) {
    const mcpServers = JSON.parse(dialogInfo.value.instanceInfo.sourceConfig).mcpServers
    return mcpServers[Object.keys(mcpServers)[0]].token || 'No Data'
  }
  console.log(111, dialogInfo.value.currentTokenIndex)

  return `${
    dialogInfo.value.currentTokenIndex !== null
      ? tokenList.value[dialogInfo.value.currentTokenIndex].token
      : ''
  }`
})
// config Info
const config = computed(() => {
  // "type": "${Object.keys(McpProtocol).filter((key) => isNaN(Number(key)))[dialogInfo.value.instanceInfo.proxyProtocol]}",
  if (dialogInfo.value.instanceInfo.accessType === AccessType.DIRECT) {
    return JsonFormatter.format(dialogInfo.value.instanceInfo.sourceConfig, 4)
  }
  if (dialogInfo.value.instanceInfo.enabledToken) {
    if (!tokenList.value) return JsonFormatter.format(`{}`, 4)
    if (dialogInfo.value.currentTokenIndex !== null && tokenList.value.length) {
      return JsonFormatter.format(
        `{
          "mcpServers": {
                "mcp-${dialogInfo.value.instanceInfo.instanceId.slice(0, 8)}": {
                      "url": "${configUrl.value}",
                      "headers": {
                            "Authorization": "${tokenList.value[dialogInfo.value.currentTokenIndex].token}"
                      }
                }
          }
      }`,
        4,
      )
    }
  }
  return JsonFormatter.format(
    `{
      "mcpServers": {
          "mcp-${dialogInfo.value.instanceInfo.instanceId.slice(0, 8)}": {
              "url": "${configUrl.value}"
          }
      }
  }`,
    4,
  )
})

const showDeleteBtn = computed(() => {
  return (token: any) => {
    if (
      [
        'dify_user_id',
        'dify_user_name',
        'dify_space_id',
        'dify_space_name',
        'intelligent_access_id',
        'intelligent_access_name',
        'intelligent_access_type',
        'default',
      ].some((keyword) => token.usages.some((usage: string) => usage.includes(keyword)))
    ) {
      return false
    }
    return true
  }
})

const showTokenList = computed(() => {
  // 筛选逻辑：根据tokenFormData的name/type/value
  const list = tokenList.value || []
  const { name, type, value } = tokenFormData.value || {}
  // 如果都为空，直接返回全部
  if (!name && !type && !value) return list

  return list.filter((token) => {
    // usages为标签数组
    let match = true
    // name精准匹配usages
    if (name) {
      match = token.usages?.includes(name)
    }
    // type模糊匹配usages
    if (match && type && !value) {
      match = token.usages?.some((u: any) => u.includes(type))
    }
    // type和value同时存在，精准匹配 type=value
    if (match && type && value) {
      match = token.usages?.includes(`${type}=${value}`)
    }
    return match
  })
})
// token list
const tokenList = ref<Array<any>>([])

const handleQueryToken = (formData: any) => {
  tokenFormData.value = formData
  // 清空选择，避免筛选后索引不匹配
  selectedTokens.value = []
}

const handleChangeTransport = async (token: any, index: number) => {
  try {
    dialogInfo.value.instanceInfo.loading = true
    tokenList.value[index].enabled = token.enabled
    await handleSaveTokens()
    emit('on-refresh')
  } catch {
    tokenList.value[index].enabled = token.enabled
  } finally {
    dialogInfo.value.instanceInfo.loading = false
  }
}

const handleBatchChangeHeader = () => {
  if (selectedTokens.value.length === 0) {
    ElMessage.warning(t('mcp.instance.token.mustSelectToken'))
    return
  }
  batchChangeHeaderRef.value.init()
}

// confirm change headers
const handleCommitHeader = async (formData: any) => {
  await TokenAPI.edit({
    tokens: tokenList.value
      .filter((token) => selectedTokens.value.includes(token.id))
      .map((token) => ({
        ...token,
        headers: {
          ...token.headers,
          ...Object.fromEntries(formData.headers.map((header: any) => [header.key, header.value])),
        },
        usages: [...token.usages, ...formData.usages],
      })),
  })
  handleTokenList()
  batchChangeHeaderRef.value.finish()
}

// handle add token
const handleAddToken = () => {
  const baseToken = `Bearer ${getToken(
    JSON.stringify({
      expireAt: formData.value.expireAt,
      userId: userInfo.userId,
      username: userInfo.username,
    }),
  )}`
  formData.value = {
    visible: true,
    token: baseToken,
    headers: [{ key: 'Authorization', value: baseToken }],
    tokenType: 1 as TokenType | null,
    enabled: true,
    expireAt: null as number | null,
    usages: [] as string[],
  }
  userDataKey.value = {
    visible: false,
    username: '',
    password: '',
    index: 0,
  }
  dialogInfo.value.currentEditIndex = null
}

// handle edit token
const handleEditToken = (index: number) => {
  formRef.value?.resetFields()
  formData.value.visible = true
  dialogInfo.value.currentEditIndex = index
  const token = tokenList.value[index]
  formData.value.token = token.token
  formData.value.expireAt = token.expireAt
  formData.value.usages = token.usages || []
  formData.value.headers = Object.entries(token.headers || {}).map(([key, value]) => ({
    key: key,
    value: value,
  }))
  formData.value.tokenType = token.tokenType
  formData.value.enabled = token.enabled
  if (formData.value.headers[0]?.value?.startsWith('Basic')) {
    formData.value.tokenType = 4
    formData.value.headers[0].tokenType = 4
  }
}

// handle enabled token switch
const handleEabledToken = async () => {
  try {
    dialogInfo.value.instanceInfo.loading = true
    await InstanceAPI.updateTokenStatus({
      instanceId: dialogInfo.value.instanceInfo.instanceId,
      enabledToken: dialogInfo.value.instanceInfo.enabledToken,
    })
    // dialogInfo.value.currentTokenIndex = 0
    handleTokenList()
    emit('on-refresh')
  } catch {
    dialogInfo.value.instanceInfo.enabledToken = !dialogInfo.value.instanceInfo.enabledToken
  } finally {
    dialogInfo.value.instanceInfo.loading = false
  }
}

// handle selected a token
const handleSelectedToken = (index: number) => {
  if (
    tokenList.value[index].expireAt !== 0 &&
    tokenList.value[index].expireAt < new Date().getTime()
  ) {
    return
  }
  dialogInfo.value.currentTokenIndex = index
}

// handle delete token
const handleDeleteToken = (index: number) => {
  ElMessageBox.confirm(t('mcp.instance.action.deleteToken'), t('common.warn'), {
    confirmButtonText: t('common.ok'),
    cancelButtonText: t('common.cancel'),
    type: 'warning',
    customClass: 'tips-box',
    center: true,
    showClose: false,
    confirmButtonClass: 'is-plain el-button--danger danger-btn',
    customStyle: {
      width: '517px',
      height: '247px',
    },
  }).then(async () => {
    await TokenAPI.delete({
      id: tokenList.value[index].id,
    })
    ElMessage.success(t('action.delete'))
    if (dialogInfo.value.currentTokenIndex === index) {
      dialogInfo.value.currentTokenIndex = null
    } else if (dialogInfo.value.currentTokenIndex && dialogInfo.value.currentTokenIndex > index) {
      dialogInfo.value.currentTokenIndex!--
    }
    tokenList.value.splice(index, 1)
  })
}

const handleViewLog = (index: number) => {
  jumpToPage({
    url: '/token-log',
    data: {
      instanceId: dialogInfo.value.instanceInfo.instanceId,
      token: tokenList.value[index].token || '',
    },
  })
}

// handle confirm token
const handleConfirmToken = async (data: {
  token: string
  expireAt: number
  usages: string[]
  headers: Record<string, string>
  currentEditIndex: number | null
}) => {
  if (data.currentEditIndex !== null) {
    tokenList.value[data.currentEditIndex] = {
      token: data.token,
      expireAt: data.expireAt || 0,
      publishAt: tokenList.value[data.currentEditIndex].publishAt,
      usages: data.usages,
      enabled: tokenList.value[data.currentEditIndex].enabled,
      headers: data.headers,
    }
    dialogInfo.value.currentEditIndex = null
    await handleSaveTokens()
    formData.value = {
      visible: false,
      token: '',
      headers: [],
      enabled: true,
      expireAt: null,
      usages: [],
    }
    return
  }
  try {
    tokenList.value.push({
      enabled: true,
      token: data.token,
      expireAt: data.expireAt || 0,
      publishAt: Date.now(),
      usages: data.usages,
      headers: data.headers,
    })
    await handleSaveTokens()
    formData.value = {
      visible: false,
      token: '',
      headers: [],
      enabled: true,
      expireAt: null,
      usages: [],
    }
  } catch {
    tokenList.value.pop()
  }
}

/**
 * handle Save the tokens
 */
const handleSaveTokens = async () => {
  try {
    dialogInfo.value.instanceInfo.loading = true
    console.log('tokens', tokenList.value)
    await TokenAPI.edit({
      tokens: tokenList.value.map((token) => ({
        ...token,
        instanceId: dialogInfo.value.instanceInfo.instanceId,
      })),
    })
    handleTokenList()
    emit('on-refresh')
  } finally {
    dialogInfo.value.instanceInfo.loading = false
  }
}

const handleTokenList = async () => {
  dialogInfo.value.instanceInfo.loading = true
  try {
    const { tokens } = await TokenAPI.list({
      instanceId: dialogInfo.value.instanceInfo.instanceId,
    })
    // reverse the token list to show the latest created token on top
    tokenList.value = (tokens || [])
      .map((token: any) => ({
        ...token,
        expire: token.expireAt !== 0 && token.expireAt < Date.now(),
      }))
      .reverse()
  } finally {
    dialogInfo.value.instanceInfo.loading = false
  }
}

/**
 * Handle copy config info
 */
const handleCopy = async (type: string) => {
  if (type === 'url') {
    await setClipboardData(configUrl.value)
  } else if (type === 'token') {
    await setClipboardData(configToken.value)
  } else {
    await setClipboardData(config.value)
  }
  ElMessage.success(t('action.copy'))
}

/**
 * Handle init model data
 * @param config - public proxy config
 */
const init = (instanceInfo: InstanceResult) => {
  dialogInfo.value.visible = true
  dialogInfo.value.instanceInfo = cloneDeep(instanceInfo)
  selectedTokens.value = []
  if (instanceInfo.enabledToken) {
    dialogInfo.value.currentTokenIndex = 0
    handleTokenList()
  } else {
    dialogInfo.value.currentTokenIndex = null
  }
}
defineExpose({
  init,
})
</script>

<style lang="scss" scoped>
.w100 {
  width: 100px;
}
.token-list {
  min-height: 75vh;
  border-radius: 8px;
  background: var(--ep-bg-color-deep);
  padding: 24px;
  .scroller {
    height: calc(75vh - 100px);
    overflow: auto;
  }
  .token-card {
    flex: 1;
    min-width: 0;
    border: 1px solid var(--ep-border-color);
    &.active {
      border-color: var(--el-color-primary);
      background-color: var(--ep-bg-purple-color-deep);
    }
    &:hover {
      border-color: var(--el-color-primary);
    }
    &.disabled {
      cursor: not-allowed;
      background-color: #e6e8eb30;
      &:hover {
        border-color: var(--ep-border-color);
      }
    }
    :deep(.el-tag) {
      color: var(--ep-color);
      border-color: var(--ep-pager-border);
      background-color: var(--ep-bg-purple-color);
    }
  }
}
.config-info {
  font-family: 'Monaco, Menlo, "Ubuntu Mono", monospace';
  font-size: 12px;
  line-height: 1.8;
  white-space: pre;
  word-break: normal;
  border-radius: 8px;
  background: var(--ep-bg-color-deep);
  border-radius: 8px;
  box-sizing: border-box;
}

.copy-icon-url {
  position: absolute;
  top: 12px;
  right: 72px;
  cursor: pointer;
}
.copy-icon-token {
  position: absolute;
  top: 12px;
  right: 42px;
  cursor: pointer;
}
.copy-icon {
  position: absolute;
  top: 12px;
  right: 12px;
  cursor: pointer;
}

/* config columns: left and right */
.config-col {
  transition: all 350ms ease;
  overflow: hidden;
  display: block;
}

.config-col.left {
  /* left column expanded width */
  flex-basis: 50%;
  max-width: 50%;
}

.config-col.right {
  /* right column default 50% */
  flex-basis: 50%;
  max-width: 50%;
}

.config-col.left.collapsed {
  opacity: 0;
  padding: 0 !important;
  max-width: 0 !important;
  flex-basis: 0 !important;
}

.config-col.right.expanded {
  /* when left collapsed, right expands to full width */
  flex-basis: 100% !important;
  max-width: 100% !important;
}

.tag-input {
  :deep(.el-tag) {
    color: var(--el-color-primary);
    border-color: var(--ep-border-color);
    .el-tag__close {
      color: var(--el-color-primary);
      &:hover {
        background-color: var(--ep-bg-purple-color-deep);
      }
    }
  }
}
.enabledTransport {
  :deep(.el-form-item__label) {
    width: 100%;
  }
  :deep(.el-form-item__content) {
    display: block;
    width: 100%;
  }
}
.delete-header {
  width: 24px;
  height: 24px;
  border-radius: 4px;
}
</style>
<style lang="scss">
.el-dialog__header.token-header-border {
  background-color: transparent !important;
  border-bottom: 1px solid var(--el-border-color-light) !important;
}
.el-dialog__footer.token-footer-border {
  background-color: transparent !important;
  border-top: 1px solid var(--el-border-color-light) !important;
}
</style>
