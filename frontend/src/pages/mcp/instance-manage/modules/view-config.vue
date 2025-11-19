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
        :disabled="disabledTokenStatus"
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
        <el-scrollbar ref="scrollbarRef" always>
          <div class="token-list">
            <div class="font-bold flex justify-between items-center">
              <mcp-button :icon="Plus" size="small" @click="handleAddToken">{{
                t('mcp.instance.formData.addToken')
              }}</mcp-button>
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
            </div>

            <div
              class="token-card border-rounded-2 mt-4 p-4 cursor-pointer line-height-6"
              :class="[
                {
                  active: dialogInfo.currentTokenIndex === index,
                  disabled: token.expireAt !== 0 && token.expireAt < Date.now(),
                },
              ]"
              v-for="(token, index) in tokenList"
              :key="index"
              @click="handleSelectedToken(index)"
            >
              <div class="flex">
                <div class="ellipsis-one flex-sub">
                  {{ token.token }}
                </div>
                <el-dropdown
                  v-if="index !== 0"
                  trigger="click"
                  class="ml-2"
                  style="cursor: pointer !important"
                  @click.stop
                  :show-arrow="false"
                >
                  <el-icon size="18"><Operation /></el-icon>
                  <template #dropdown>
                    <el-dropdown-menu>
                      <el-dropdown-item command="handleViewLog" @click="handleEditToken(index)">
                        {{ t('env.run.action.edit') }}
                      </el-dropdown-item>
                      <el-dropdown-item command="handleViewLog" @click="handleDeleteToken(index)">
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
                <div class="ellipsis-one grid-cols-span-1">
                  {{ t('mcp.instance.token.tag') }}：<el-tag
                    v-for="(tag, num) in token.usages"
                    :key="num"
                    effect="plain"
                    class="mr-2"
                  >
                    {{ tag }}
                  </el-tag>
                </div>
                <div class="grid-cols-span-1" v-if="true || 'API服务'">
                  是否开启透传:
                  <el-switch
                    v-model="token.enabledTransport"
                    style="--el-switch-on-color: #13ce66"
                    inline-prompt
                    :loading="dialogInfo.instanceInfo.loading"
                    :active-text="t('common.on')"
                    :inactive-text="t('common.off')"
                    @change="handleChangeTransport(token, index)"
                  ></el-switch>
                </div>
              </div>
            </div>
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
    <el-scrollbar
      :class="[
        'config-col logs-box  mt-3 p-3',
        {
          collapsed:
            !dialogInfo.instanceInfo.enabledToken ||
            dialogInfo.instanceInfo.accessType === AccessType.DIRECT,
        },
      ]"
      max-height="420px"
      always
    >
      <div ref="logEl">
        <div>令牌日志</div>
        <div class="py-5 px-5">{{ tokenConfig }}</div>
        <el-icon v-if="!isFullscreen" class="base-btn-link copy-icon" size="18" @click="toggle">
          <FullScreen />
        </el-icon>
      </div>
    </el-scrollbar>
    <template #footer>
      <div class="center">
        <!-- <mcp-button @click="handleCopy" class="w100">{{
          t('mcp.instance.action.copy')
        }}</mcp-button> -->
      </div>
    </template>
  </el-dialog>
  <el-dialog
    v-model="formData.visible"
    width="60%"
    top="20vh"
    :show-close="false"
    @close="formRef?.resetFields()"
    header-class="token-header-border"
    footer-class="token-footer-border"
  >
    <template #header>
      <div class="center mb-4">{{ t('mcp.instance.token.title') }}</div>
    </template>
    <el-row :gutter="12" v-loading="dialogInfo.instanceInfo.loading" class="mt-4">
      <el-col :span="12" class="collapsed">
        <el-scrollbar ref="scrollbarRef" height="50vh" always class="pr-4">
          <el-form
            ref="formRef"
            :model="formData"
            :rules="rules"
            label-width="auto"
            label-position="top"
            class="mx-2"
          >
            <el-form-item :label="t('mcp.instance.token.lifespan')" prop="expireAt">
              <template #label>
                <div class="center">
                  <span class="mr-2">{{ t('mcp.instance.token.lifespan') }}</span>
                  <el-button
                    class="base-btn"
                    type="primary"
                    size="small"
                    @click.stop="handleAddExpireAt(7)"
                    >7{{ t('mcp.instance.token.day') }}
                  </el-button>
                  <el-button
                    class="base-btn"
                    type="primary"
                    size="small"
                    @click.stop="handleAddExpireAt(15)"
                    >15{{ t('mcp.instance.token.day') }}
                  </el-button>
                  <el-button
                    class="base-btn"
                    type="primary"
                    size="small"
                    @click.stop="handleAddExpireAt(30)"
                    >30{{ t('mcp.instance.token.day') }}
                  </el-button>
                </div>
              </template>
              <el-date-picker
                ref="datePicker"
                v-model="formData.expireAt"
                type="datetime"
                value-format="x"
                :placeholder="t('mcp.instance.token.placeholderDate')"
                style="width: 100%"
                :disabled-date="(date: Date) => date.getTime() < Date.now()"
              ></el-date-picker>
            </el-form-item>
            <el-form-item :label="'API 认证凭证'" prop="tokenType">
              <el-select
                v-model="formData.tokenType"
                :placeholder="'请选择API 认证凭证'"
                clearable
                @change="handleTokenTypeChange"
              >
                <el-option label="Bearer" :value="1" />
                <el-option label="Api-Key" :value="2" />
                <el-option label="X-API-key" :value="3" />
                <el-option label="Basic" :value="4" />
              </el-select>

              <div v-if="Number(formData.tokenType) === 1" class="center my-2 w-full">
                Authorization：<el-input
                  v-model="formData.token"
                  :placeholder="t('mcp.instance.token.placeholderToken')"
                  class="flex-sub"
                  clearable
                />
                <el-tag
                  class="ml-2 base-btn cursor-pointer"
                  effect="dark"
                  @click="handleRandomToken"
                  >{{ t('mcp.instance.token.random') }}
                </el-tag>
              </div>
              <div v-if="Number(formData.tokenType) === 2" class="center my-2 w-full">
                Api-Key：
                <el-input
                  v-model="formData.token"
                  :placeholder="'请输入token值或自动随机生成'"
                  class="flex-sub"
                  clearable
                />
                <el-tag
                  class="ml-2 base-btn cursor-pointer"
                  effect="dark"
                  @click="handleRandomToken"
                  >{{ t('mcp.instance.token.random') }}
                </el-tag>
              </div>
              <div v-if="Number(formData.tokenType) === 3" class="center my-2 w-full">
                X-API-Key：
                <el-input
                  v-model="formData.token"
                  :placeholder="'请输入token值或自动随机生成'"
                  class="flex-sub"
                  clearable
                />
                <el-tag
                  class="ml-2 base-btn cursor-pointer"
                  effect="dark"
                  @click="handleRandomToken"
                  >{{ t('mcp.instance.token.random') }}
                </el-tag>
              </div>
              <div v-if="Number(formData.tokenType) === 4" class="center my-2 w-full">
                <el-input
                  v-model="userDataKey.username"
                  placeholder="username"
                  class="flex-sub mr-2"
                  clearable
                />
                <el-input
                  v-model="userDataKey.password"
                  placeholder="password"
                  class="flex-sub"
                  clearable
                />
              </div>
            </el-form-item>
            <el-form-item prop="enabledTransport" class="enabledTransport">
              <template #label>
                <div class="w-full flex justify-between items-center">
                  <div class="center">
                    <span class="mr-2"> HTTP 请求头 </span>
                    <el-popover placement="top" width="250">
                      <div>{{ '开启透传之后；请求头的值将透传至后续服务' }}</div>
                      <template #reference>
                        <el-icon class="cursor-pointer"><Warning /></el-icon>
                      </template>
                    </el-popover>
                  </div>
                  <div class="center">
                    <el-switch
                      v-model="formData.enabledTransport"
                      style="--el-switch-on-color: #13ce66"
                      inline-prompt
                      :active-text="'透传'"
                      :inactive-text="'不透传'"
                    ></el-switch>
                    <div
                      class="cursor-pointer border border-style-solid border-rd-md border-white ml-2 p-1 center bg-gray-600 color-white hover-scale-110"
                      @click="handleAddHeader"
                    >
                      <el-icon>
                        <Plus />
                      </el-icon>
                    </div>
                  </div>
                </div>
              </template>
              <div
                v-for="(header, index) in formData.headers"
                :key="index"
                class="flex items-center my-2 pr-3"
              >
                <el-input
                  v-model="header.key"
                  :placeholder="'请求头键名'"
                  class="flex-sub mr-2"
                ></el-input>
                ：
                <el-input
                  v-model="header.value"
                  :placeholder="'请求头键值'"
                  class="flex-sub mr-2"
                ></el-input>
                <div
                  class="cursor-pointer border border-style-solid border-rd-md border-white p-1 center bg-red-100/50 color-white hover-bg-red-400/90 hover-scale-105"
                  @click="formData.headers.splice(index, 1)"
                >
                  <el-icon><Minus /></el-icon>
                </div>
              </div>
            </el-form-item>
            <el-form-item :label="t('mcp.instance.token.tag')" prop="usages">
              <el-input-tag
                v-model="formData.usages"
                collapse-tags
                collapse-tags-tooltip
                :max-collapse-tags="3"
                clearable
                draggable
                tag-type="primary"
                tag-effect="plain"
                :placeholder="t('mcp.instance.token.placeholderTag')"
                class="tag-input"
              >
                <template #tag="{ value }">
                  <div class="flex items-center">
                    <span>{{ value }}</span>
                  </div>
                </template>
              </el-input-tag>
            </el-form-item>
          </el-form>
        </el-scrollbar>
      </el-col>
      <el-col :span="12" class="expanded">
        <el-scrollbar ref="scrollbarRef" height="50vh" always class="config-info">
          <div class="py-5 px-5">{{ config }}</div>
        </el-scrollbar>
      </el-col>
    </el-row>
    <template #footer>
      <div class="center">
        <el-button @click="formData.visible = false" class="mr-4 w-25">{{
          t('common.cancel')
        }}</el-button>
        <mcp-button @click="handleConfirmToken" class="w-25">{{ t('common.ok') }}</mcp-button>
      </div>
    </template>
  </el-dialog>
</template>
<script setup lang="ts">
import { setClipboardData, timestampToDate } from '@/utils/system'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Plus,
  Operation,
  CopyDocument,
  Link,
  Key,
  FullScreen,
  Warning,
  Minus,
} from '@element-plus/icons-vue'
import McpButton from '@/components/mcp-button/index.vue'
import { AccessType, type InstanceResult } from '@/types'
import { InstanceAPI } from '@/api/mcp/instance'
import { getToken } from '@/utils/system'
import { useUserStore } from '@/stores'
import { cloneDeep } from 'lodash-es'
import { JsonFormatter } from '@/utils/json'

const logEl = ref(null)
// 将日志容器作为全屏目标。el-scrollbar 的 ref 返回组件实例，需取其 $el
const targetEl = computed(() => {
  const el = (logEl as any).value
  return (el && (el.$el || el)) as any
})
const disabledTokenStatus = computed(() => {
  return false // !'存在一个透传的时候必须开启'
})
const { isFullscreen, toggle } = useFullscreen(targetEl)
const { t } = useI18n()
const { userInfo } = useUserStore()
const emit = defineEmits<{
  (e: 'on-refresh'): void
}>()
const formRef = ref()
const formData = ref({
  visible: false,
  token: '',
  headers: [] as { key: string; value: string }[],
  tokenType: '',
  enabledTransport: false,
  expireAt: null as number | null,
  usages: [] as string[],
})
const userDataKey = ref({
  username: '',
  password: '',
})
const rules = reactive({
  tokenType: [{ required: true, message: '请选择API凭证类型', trigger: 'change' }],
  token: [{ required: true, message: t('mcp.instance.token.mustToken'), trigger: 'blur' }],
})
const dialogInfo = ref({
  visible: false,
  title: t('mcp.instance.config'),
  instanceInfo: {} as InstanceResult,
  currentTokenIndex: null as number | null,
  currentEditIndex: null as number | null,
})
// logs with token
const tokenConfig = ref('')
const configUrl = computed(
  () => `${window.location.origin}${dialogInfo.value.instanceInfo.publicProxyPath}`,
)
const configToken = computed(
  () =>
    `${
      dialogInfo.value.currentTokenIndex !== null
        ? dialogInfo.value.instanceInfo.tokens[dialogInfo.value.currentTokenIndex].token
        : ''
    }`,
)
// config Info
const config = computed(() => {
  // "type": "${Object.keys(McpProtocol).filter((key) => isNaN(Number(key)))[dialogInfo.value.instanceInfo.proxyProtocol]}",
  if (dialogInfo.value.instanceInfo.accessType === AccessType.DIRECT) {
    return JsonFormatter.format(dialogInfo.value.instanceInfo.sourceConfig, 4)
  }
  if (dialogInfo.value.instanceInfo.enabledToken) {
    return `{
      "mcpServers": {
            "mcp-${dialogInfo.value.instanceInfo.instanceId.slice(0, 8)}": {
                  "url": "${configUrl.value}",
                  "headers": {
                        "Authorization": "${
                          dialogInfo.value.currentTokenIndex !== null
                            ? dialogInfo.value.instanceInfo.tokens[
                                dialogInfo.value.currentTokenIndex
                              ].token
                            : ''
                        }"
                  }
            }
      }
  }`
  }
  return `{
      "mcpServers": {
          "mcp-${dialogInfo.value.instanceInfo.instanceId.slice(0, 8)}": {
              "url": "${configUrl.value}"
          }
      }
  }`
})

// token list
const tokenList = computed(
  () =>
    dialogInfo.value.instanceInfo.tokens.map((item) => ({
      ...item,
      expire: item.expireAt !== 0 && item.expireAt < Date.now(),
    })) || [],
)

const handleChangeTransport = async (token: any, index: number) => {
  try {
    dialogInfo.value.instanceInfo.loading = true
    dialogInfo.value.instanceInfo.tokens[index].enabledTransport = token.enabledTransport
    await handleSaveTokens()
    emit('on-refresh')
  } catch {
    dialogInfo.value.instanceInfo.tokens[index].enabledTransport = token.enabledTransport
  } finally {
    dialogInfo.value.instanceInfo.loading = false
  }
}

// handle add token
const handleAddToken = () => {
  formData.value.token = ''
  formData.value.expireAt = null
  formData.value.usages = []
  dialogInfo.value.currentEditIndex = null
  formData.value.visible = true
  formRef.value?.resetFields()
}

// handle add header
const handleAddHeader = () => {
  formData.value.headers.push({ key: '', value: '' })
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

// handle token type change and clear token value
const handleTokenTypeChange = () => {
  formData.value.token = ''
}

// handle random token
const handleRandomToken = () => {
  if (dialogInfo.value.currentEditIndex !== null) {
    ElMessageBox.confirm(t('mcp.instance.action.random'), t('common.warn'), {
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
    }).then(() => {
      handleGetTokenValue()
    })
    return
  }

  handleGetTokenValue()
}

// handle get token value
const handleGetTokenValue = () => {
  if (Number(formData.value.tokenType) === 1) {
    formData.value.token =
      'Bearer ' +
      getToken(
        JSON.stringify({
          expireAt: formData.value.expireAt,
          userId: userInfo.userId,
          username: userInfo.username,
        }),
      )
  } else if (Number(formData.value.tokenType) === 2) {
    formData.value.token = getToken(
      JSON.stringify({
        expireAt: formData.value.expireAt,
        userId: userInfo.userId,
        username: userInfo.username,
      }),
    )
  } else if (Number(formData.value.tokenType) === 3) {
    formData.value.token = getToken(
      JSON.stringify({
        expireAt: formData.value.expireAt,
        userId: userInfo.userId,
        username: userInfo.username,
      }),
    )
  }
}
// handle add expire at
const handleAddExpireAt = (days: number) => {
  const expireDate = new Date()
  expireDate.setDate(expireDate.getDate() + days)
  formData.value.expireAt = expireDate.getTime()
}
// handle edit token
const handleEditToken = (index: number) => {
  formRef.value?.resetFields()
  formData.value.visible = true
  dialogInfo.value.currentEditIndex = index
  const token = dialogInfo.value.instanceInfo.tokens[index]
  formData.value.token = token.token
  formData.value.expireAt = token.expireAt
  formData.value.usages = token.usages || []
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
  }).then(() => {
    dialogInfo.value.instanceInfo.tokens.splice(index, 1)
    if (dialogInfo.value.currentTokenIndex === index) {
      dialogInfo.value.currentTokenIndex = null
    } else if (dialogInfo.value.currentTokenIndex && dialogInfo.value.currentTokenIndex > index) {
      dialogInfo.value.currentTokenIndex!--
    }
    handleSaveTokens()
  })
}

// handle confirm token
const handleConfirmToken = async () => {
  const result = await formRef.value.validate()
  if (!result) return
  if (Number(formData.value.tokenType) === 4) {
    const base64Credentials = btoa(`${userDataKey.value.username}:${userDataKey.value.password}`)
    formData.value.token = `Basic ${base64Credentials}`
  }
  if (dialogInfo.value.currentEditIndex) {
    dialogInfo.value.instanceInfo.tokens[dialogInfo.value.currentEditIndex] = {
      token: formData.value.token,
      expireAt: formData.value.expireAt || 0,
      publishAt: dialogInfo.value.instanceInfo.tokens[dialogInfo.value.currentEditIndex].publishAt,
      usages: formData.value.usages,
      enabledTransport: formData.value.enabledTransport,
      headers: formData.value.headers,
    }
    dialogInfo.value.currentEditIndex = null
    await handleSaveTokens()
    formData.value = {
      visible: false,
      token: '',
      headers: [],
      tokenType: '',
      enabledTransport: false,
      expireAt: null,
      usages: [],
    }
    return
  }
  try {
    dialogInfo.value.instanceInfo.tokens.push({
      enabledTransport: formData.value.enabledTransport,
      token: formData.value.token,
      expireAt: formData.value.expireAt || 0,
      publishAt: Date.now(),
      usages: formData.value.usages,
      headers: formData.value.headers,
    })
    await handleSaveTokens()
    formData.value = {
      visible: false,
      token: '',
      headers: [],
      tokenType: '',
      enabledTransport: false,
      expireAt: null,
      usages: [],
    }
  } catch {
    dialogInfo.value.instanceInfo.tokens.pop()
  }
}

/**
 * handle Save the tokens
 */
const handleSaveTokens = async () => {
  try {
    dialogInfo.value.instanceInfo.loading = true
    await InstanceAPI.updateInstanceTokens({
      instanceId: dialogInfo.value.instanceInfo.instanceId,
      tokens: dialogInfo.value.instanceInfo.tokens,
    })
    emit('on-refresh')
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

const handleGetLogs = async () => {
  if (dialogInfo.value.currentTokenIndex === null) {
    ElMessage.warning('请先选择一个Token')
    return
  }
  try {
    dialogInfo.value.instanceInfo.loading = true
    const { logs } = await InstanceAPI.logsByToken({
      instanceId: dialogInfo.value.instanceInfo.instanceId,
      token: dialogInfo.value.instanceInfo.tokens[dialogInfo.value.currentTokenIndex].token,
      pageNum: 1,
      pageSize: 100,
    })
    tokenConfig.value = logs
  } finally {
    dialogInfo.value.instanceInfo.loading = false
  }
}

/**
 * Handle init model data
 * @param config - public proxy config
 */
const init = (instanceInfo: InstanceResult) => {
  dialogInfo.value.visible = true
  dialogInfo.value.instanceInfo = cloneDeep(instanceInfo)
  if (instanceInfo.enabledToken) {
    dialogInfo.value.currentTokenIndex = 0
    handleGetLogs()
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
  height: 40vh;
  border-radius: 8px;
  background: var(--ep-bg-color-deep);
  padding: 24px;
  .token-card {
    border: 1px solid var(--ep-border-color);
    &.active {
      border-color: var(--ep-purple-color);
      background-color: var(--ep-bg-purple-color-deep);
    }
    &:hover {
      border-color: var(--ep-purple-color);
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
.logs-info {
  min-height: 36vh;
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
.logs-box {
  height: 36vh;
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
.logs-box.collapsed {
  opacity: 0;
  padding: 0 !important;
  max-width: 0 !important;
  min-height: 0;
  height: 0 !important;
}
.tag-input {
  :deep(.el-tag) {
    color: var(--ep-purple-color);
    border-color: var(--ep-border-color);
    .el-tag__close {
      color: var(--ep-purple-color);
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
