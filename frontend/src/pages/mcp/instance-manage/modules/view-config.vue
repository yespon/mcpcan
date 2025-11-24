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
        <el-scrollbar ref="scrollbarRef" max-height="75vh" always>
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
                  trigger="click"
                  class="ml-2"
                  style="cursor: pointer !important"
                  @click.stop
                  :show-arrow="false"
                >
                  <el-icon size="18"><Operation /></el-icon>
                  <template #dropdown>
                    <el-dropdown-menu>
                      <el-dropdown-item v-if="index !== 0" @click="handleEditToken(index)">
                        <el-button link>
                          {{ t('env.run.action.edit') }}
                        </el-button>
                      </el-dropdown-item>
                      <el-dropdown-item @click="handleViewLog(index)">
                        <el-button link>
                          {{ t('mcp.instance.action.accessLogs') }}
                        </el-button>
                      </el-dropdown-item>
                      <el-dropdown-item v-if="index !== 0" @click="handleDeleteToken(index)">
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
                <div class="grid-cols-span-1" v-if="dialogInfo.instanceInfo.sourceType === 4">
                  {{ t('mcp.instance.token.passthrough') }}:
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
            <el-form-item prop="tokenType">
              <template #label>
                API {{ t('mcp.instance.token.auth') }}
                <el-popover placement="top" width="250">
                  <div>{{ t('mcp.instance.token.Bearer') }}</div>
                  <div>{{ t('mcp.instance.token.Api-Key') }}</div>
                  <div>{{ t('mcp.instance.token.X-API-key') }}</div>
                  <div>{{ t('mcp.instance.token.Basic') }}</div>
                  <template #reference>
                    <el-icon class="cursor-pointer"><Warning /></el-icon>
                  </template>
                </el-popover>
              </template>
              <el-select
                v-model="formData.tokenType"
                :placeholder="t('mcp.instance.token.placeholderTokenType')"
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
              <div v-if="Number(formData.tokenType) === 3" class="center my-2 w-full">
                X-API-Key：
                <el-input
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
              <div v-if="Number(formData.tokenType) === 4" class="center my-2 w-full">
                <template v-if="userDataKey.visible">
                  <el-input
                    v-model="userDataKey.username"
                    placeholder="username"
                    class="flex-sub mr-2"
                    clearable
                  />
                  <el-input
                    v-model="userDataKey.password"
                    placeholder="password"
                    type="password"
                    show-password
                    class="flex-sub"
                    clearable
                  />
                </template>
                <template v-else>
                  Authorization：
                  <el-input
                    v-model="formData.token"
                    :placeholder="t('mcp.instance.token.placeholderToken')"
                    clearable
                  />
                </template>
                <el-button link class="ml-2 link-hover base-btn-link" @click="handleChangeBasic">
                  <el-icon><Refresh /></el-icon
                  >{{
                    userDataKey.visible
                      ? t('mcp.instance.token.custom')
                      : t('mcp.instance.token.accountPassword')
                  }}
                </el-button>
              </div>
            </el-form-item>
            <el-form-item prop="enabledTransport" class="enabledTransport">
              <template #label>
                <div class="w-full flex justify-between items-center">
                  <div class="center">
                    <span class="mr-2"> HTTP {{ t('mcp.instance.token.headers') }} </span>
                    <el-popover placement="top" width="250">
                      <div>{{ t('mcp.instance.token.headersPlaceholder') }}</div>
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
                      :active-text="t('mcp.instance.token.passthrough')"
                      :inactive-text="t('mcp.instance.token.passthroughOff')"
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
                  :placeholder="t('mcp.instance.token.headersKey')"
                  class="flex-sub mr-2"
                ></el-input>
                ：
                <el-input
                  v-model="header.value"
                  :placeholder="t('mcp.instance.token.headersValue')"
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
          <div class="py-5 px-5">{{ editConfig }}</div>
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
  Warning,
  Minus,
  Refresh,
} from '@element-plus/icons-vue'
import McpButton from '@/components/mcp-button/index.vue'
import { AccessType, TokenType, type InstanceResult } from '@/types'
import { InstanceAPI } from '@/api/mcp/instance'
import { getToken } from '@/utils/system'
import { useUserStore } from '@/stores'
import { cloneDeep } from 'lodash-es'
import { JsonFormatter } from '@/utils/json'
import { useRouterHooks } from '@/utils/url'

const { jumpToPage } = useRouterHooks()

const disabledTokenStatus = computed(() => {
  return tokenList.value.some((item) => item.enabledTransport) // !'存在一个透传的时候必须开启'
})
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
  tokenType: null as TokenType | null,
  enabledTransport: false,
  expireAt: null as number | null,
  usages: [] as string[],
})
const userDataKey = ref({
  visible: false,
  username: '',
  password: '',
})
const rules = reactive({
  tokenType: [
    {
      required: true,
      validator: (rule: any, value: number, callback: (error?: string | Error) => void) => {
        if (!value) {
          callback(new Error(t('mcp.instance.token.placeholderTokenType')))
        }
        if (value === 4) {
          if (!userDataKey.value.username || !userDataKey.value.password) {
            callback(
              new Error(
                userDataKey.value.visible
                  ? t('mcp.instance.token.Basic')
                  : t('mcp.instance.token.mustToken'),
              ),
            )
          }
        } else {
          if (!formData.value.token) {
            callback(new Error(t('mcp.instance.token.mustToken')))
          }
        }
        callback()
      },
      trigger: 'change',
    },
  ],
  token: [{ required: true, message: t('mcp.instance.token.mustToken'), trigger: 'blur' }],
})
const dialogInfo = ref({
  visible: false,
  title: t('mcp.instance.config'),
  instanceInfo: {} as InstanceResult,
  currentTokenIndex: null as number | null,
  currentEditIndex: null as number | null,
})

const configUrl = computed(() => {
  if (dialogInfo.value.instanceInfo.accessType === AccessType.DIRECT) {
    const mcpServers = JSON.parse(dialogInfo.value.instanceInfo.sourceConfig).mcpServers
    return mcpServers[Object.keys(mcpServers)[0]].url
  }
  return `${window.location.origin}${dialogInfo.value.instanceInfo.publicProxyPath}`
})
const configToken = computed(() => {
  if (dialogInfo.value.instanceInfo.accessType === AccessType.DIRECT) {
    const mcpServers = JSON.parse(dialogInfo.value.instanceInfo.sourceConfig).mcpServers
    return mcpServers[Object.keys(mcpServers)[0]].token || 'No Data'
  }
  return `${
    dialogInfo.value.currentTokenIndex !== null
      ? dialogInfo.value.instanceInfo.tokens[dialogInfo.value.currentTokenIndex].token
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
    let headersString = null
    let tokenType = 0
    if (dialogInfo.value.currentTokenIndex !== null) {
      const headers =
        dialogInfo.value.instanceInfo.tokens[dialogInfo.value.currentTokenIndex].headers || {}
      tokenType = Number(
        dialogInfo.value.instanceInfo.tokens[dialogInfo.value.currentTokenIndex].tokenType,
      )
      headersString = Object.entries(headers)
        .map(([key, value]) => `"${key}": "${value.replace(/"/g, '\\"')}"`)
        .join(',\n                        ')
    }

    return JsonFormatter.format(
      `{
          "mcpServers": {
                "mcp-${dialogInfo.value.instanceInfo.instanceId.slice(0, 8)}": {
                      "url": "${configUrl.value}",
                      "headers": {
                            "${['Unknown', 'Authorization', 'Api-Key', 'X-API-Key', 'Authorization'][tokenType]}": "${
                              dialogInfo.value.currentTokenIndex !== null
                                ? dialogInfo.value.instanceInfo.tokens[
                                    dialogInfo.value.currentTokenIndex
                                  ].token
                                : ''
                            }"${headersString ? ',' + headersString : ''}
                      }
                }
          }
      }`,
      4,
    )
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

const editConfig = computed(() => {
  const headersString = formData.value.headers
    .filter((header) => header.key && header.value)
    .map((header) => `"${header.key}": "${header.value.replace(/"/g, '\\"')}"`)
    .join(',\n                        ')

  return JsonFormatter.format(
    `{
          "mcpServers": {
                "mcp-${dialogInfo.value.instanceInfo.instanceId.slice(0, 8)}": {
                      "url": "${configUrl.value}",
                      "headers": {
                            "${['Unknown', 'Authorization', 'Api-Key', 'X-API-Key', 'Authorization'][formData.value.tokenType || 0]}": "${
                              formData.value.tokenType === 4
                                ? 'Basic ' +
                                  btoa(
                                    userDataKey.value.username + ':' + userDataKey.value.password,
                                  )
                                : formData.value.token || ''
                            }"${headersString ? ',' + headersString : ''}
                      }
                }
          }
      }`,
    4,
  )
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
  formData.value = {
    visible: true,
    token: '',
    headers: [] as { key: string; value: string }[],
    tokenType: null as TokenType | null,
    enabledTransport: false,
    expireAt: null as number | null,
    usages: [] as string[],
  }
  userDataKey.value = {
    visible: false,
    username: '',
    password: '',
  }
  dialogInfo.value.currentEditIndex = null
  formRef.value?.resetFields()
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
  formData.value.headers = Object.entries(token.headers || {}).map(([key, value]) => ({
    key: key,
    value: value,
  }))
  formData.value.tokenType = token.tokenType
  formData.value.enabledTransport = token.enabledTransport
  handleTokenTypeChange()
}

const handleChangeBasic = () => {
  userDataKey.value.visible = !userDataKey.value.visible
  formData.value.token = ''
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
  let roginData = { token: '', tokenType: '' }
  if (dialogInfo.value.currentEditIndex !== null) {
    roginData = dialogInfo.value.instanceInfo.tokens[dialogInfo.value.currentEditIndex] as any
  }
  if (Number(roginData.tokenType) === 4) {
    userDataKey.value.username = atob(roginData.token.split(' ')[1]).split(':')[0]
    userDataKey.value.password = atob(roginData.token.split(' ')[1]).split(':')[1]
    return
  }
  formData.value.token = roginData.token
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

const handleViewLog = (index: number) => {
  jumpToPage({
    url: '/token-log',
    data: {
      instanceId: dialogInfo.value.instanceInfo.instanceId,
      token: dialogInfo.value.instanceInfo.tokens[index].token || '',
    },
  })
}

// handle confirm token
const handleConfirmToken = async () => {
  const result = await formRef.value.validate()
  if (!result) return
  if (Number(formData.value.tokenType) === 4) {
    if (userDataKey.value.visible) {
      const base64Credentials = btoa(`${userDataKey.value.username}:${userDataKey.value.password}`)
      formData.value.token = `Basic ${base64Credentials}`
    }
  }
  if (dialogInfo.value.currentEditIndex) {
    dialogInfo.value.instanceInfo.tokens[dialogInfo.value.currentEditIndex] = {
      token: formData.value.token,
      expireAt: formData.value.expireAt || 0,
      publishAt: dialogInfo.value.instanceInfo.tokens[dialogInfo.value.currentEditIndex].publishAt,
      usages: formData.value.usages,
      enabledTransport: formData.value.enabledTransport,
      headers: Object.fromEntries(
        formData.value.headers.map((header) => [header.key, header.value]),
      ),
      tokenType: formData.value.tokenType,
    }
    dialogInfo.value.currentEditIndex = null
    await handleSaveTokens()
    formData.value = {
      visible: false,
      token: '',
      headers: [],
      tokenType: null,
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
      headers: Object.fromEntries(
        formData.value.headers.map((header) => [header.key, header.value]),
      ),
      tokenType: formData.value.tokenType,
    })
    await handleSaveTokens()
    formData.value = {
      visible: false,
      token: '',
      headers: [],
      tokenType: null,
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

/**
 * Handle init model data
 * @param config - public proxy config
 */
const init = (instanceInfo: InstanceResult) => {
  dialogInfo.value.visible = true
  dialogInfo.value.instanceInfo = cloneDeep(instanceInfo)
  if (instanceInfo.enabledToken) {
    dialogInfo.value.currentTokenIndex = 0
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
