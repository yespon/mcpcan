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
                      <el-dropdown-item
                        v-if="!token.usages.includes('default')"
                        @click="handleEditToken(index)"
                      >
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
                        v-if="!token.usages.includes('default')"
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
                <div class="grid-cols-span-1">
                  {{ '是否启用' }}:
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
  <!-- token config dialog -->
  <el-dialog
    v-model="formData.visible"
    width="580px"
    top="20vh"
    :show-close="false"
    @close="formRef?.resetFields()"
    header-class="token-header-border"
    footer-class="token-footer-border"
  >
    <template #header>
      <div class="center mb-4">{{ t('mcp.instance.token.title') }}</div>
    </template>
    <el-row v-loading="dialogInfo.instanceInfo.loading" class="mt-4">
      <el-col :span="24" class="collapsed">
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
            <el-form-item :label="'网关认证'" prop="tokenType">
              <div class="w-full u-line-1" style="white-space: nowrap">
                Authorization：{{ formData.token }}
              </div>
            </el-form-item>
            <!-- <el-form-item prop="tokenType">
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
            </el-form-item> -->
            <el-form-item prop="enabledTransport" class="enabledTransport">
              <template #label>
                <div class="w-full flex justify-between items-center">
                  <span class="mr-2"> 透传 {{ 'Headers' }} </span>
                  <div class="center">
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
                v-for="(item, index) in formData.headers"
                :key="index"
                class="flex items-center my-2 pr-3"
              >
                <el-row :gutter="12" class="flex-sub">
                  <el-col :span="7">
                    <div class="flex h-full">
                      <el-dropdown
                        v-if="index === 0"
                        trigger="click"
                        class="h-full w-full flex items-center justify-end"
                        :show-arrow="false"
                      >
                        <div class="center cursor-pointer">
                          {{ item.key }}
                          <el-icon class="text-purple ml-1"><Sort /></el-icon>
                        </div>
                        <template #dropdown>
                          <el-dropdown-menu>
                            <el-dropdown-item @click="handleTokenTypeChange(1)">
                              Authorization(Bearer)
                            </el-dropdown-item>
                            <el-dropdown-item @click="handleTokenTypeChange(2)">
                              Api-Key
                            </el-dropdown-item>
                            <el-dropdown-item @click="handleTokenTypeChange(3)">
                              X-API-key
                            </el-dropdown-item>
                            <el-dropdown-item @click="handleTokenTypeChange(4)">
                              Authorization(Basic)
                            </el-dropdown-item>
                          </el-dropdown-menu>
                        </template>
                      </el-dropdown>
                      <el-input
                        v-else
                        v-model="item.key"
                        :placeholder="t('mcp.instance.token.headersKey')"
                        class="flex-sub"
                      >
                      </el-input>
                      <span class="ml-2">:</span>
                    </div>
                  </el-col>
                  <el-col :span="15" class="flex">
                    <el-input
                      v-model="item.value"
                      :placeholder="t('mcp.instance.token.headersValue')"
                      class="flex-sub"
                    ></el-input>
                  </el-col>
                  <el-col :span="2">
                    <div
                      v-if="index === 0"
                      class="text-purple cursor-pointer"
                      @click="handleChangeBasic"
                    >
                      {{ Number(formData.tokenType) === 4 ? '账号' : '  ' }}
                    </div>
                    <div
                      v-else
                      class="cursor-pointer border border-style-solid delete-header border-white px-1 ml-2 center bg-red-100/50 color-white hover-bg-red-400/90 hover-scale-105"
                      @click="formData.headers.splice(index, 1)"
                    >
                      <el-icon><Minus /></el-icon>
                    </div>
                  </el-col>
                </el-row>
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
      <!-- <el-col :span="12" class="expanded">
        <el-scrollbar ref="scrollbarRef" height="50vh" always class="config-info">
          <div class="py-5 px-5">{{ editConfig }}</div>
        </el-scrollbar>
      </el-col> -->
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
  <!-- user accountPassword -->
  <el-dialog v-model="userDataKey.visible" width="400px" top="30vh" :show-close="false">
    <el-form :model="userDataKey" class="p-4" label-width="80px">
      <el-form-item label="用户名" prop="username">
        <el-input v-model="userDataKey.username" placeholder="请输入用户名" />
      </el-form-item>
      <el-form-item label="密码" prop="password">
        <el-input v-model="userDataKey.password" type="password" placeholder="请输入密码" />
      </el-form-item>
    </el-form>
    <template #footer>
      <div class="center">
        <mcp-button @click="handleConfirmAccount" class="w-25">{{ t('common.ok') }}</mcp-button>
      </div>
    </template>
  </el-dialog>
</template>
<script setup lang="ts">
import { setClipboardData, timestampToDate } from '@/utils/system'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Operation, CopyDocument, Link, Key, Sort, Minus } from '@element-plus/icons-vue'
import McpButton from '@/components/mcp-button/index.vue'
import { AccessType, TokenType, type InstanceResult } from '@/types'
import { InstanceAPI, TokenAPI } from '@/api/mcp/instance'
import { getToken } from '@/utils/system'
import { useUserStore } from '@/stores'
import { cloneDeep } from 'lodash-es'
import { JsonFormatter } from '@/utils/json'
import { useRouterHooks } from '@/utils/url'

const { jumpToPage } = useRouterHooks()

const { t } = useI18n()
const { userInfo } = useUserStore()
const emit = defineEmits<{
  (e: 'on-refresh'): void
}>()
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
const userDataKey = ref({
  visible: false,
  username: '',
  password: '',
})
const tokenTypeOptions = [
  { label: 'Authorization', value: 1 },
  { label: 'Api-Key', value: 2 },
  { label: 'X-API-key', value: 3 },
  { label: 'Authorization', value: 4 },
]
const rules = reactive({
  // tokenType: [
  //   {
  //     required: true,
  //     validator: (rule: any, value: number, callback: (error?: string | Error) => void) => {
  //       if (!value) {
  //         callback(new Error(t('mcp.instance.token.placeholderTokenType')))
  //       }
  //       if (value === 4) {
  //         if (!userDataKey.value.username || !userDataKey.value.password) {
  //           callback(
  //             new Error(
  //               userDataKey.value.visible
  //                 ? t('mcp.instance.token.Basic')
  //                 : t('mcp.instance.token.mustToken'),
  //             ),
  //           )
  //         }
  //       } else {
  //         if (!formData.value.token) {
  //           callback(new Error(t('mcp.instance.token.mustToken')))
  //         }
  //       }
  //       callback()
  //     },
  //     trigger: 'change',
  //   },
  // ],
  // token: [{ required: true, message: t('mcp.instance.token.mustToken'), trigger: 'blur' }],
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

// token list
const tokenList = ref<Array<any>>([])

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
  if(formData.value.headers[0].value.startsWith('Basic')) {
    handleTokenTypeChange(4)
  }
}

const handleChangeBasic = () => {
  userDataKey.value.visible = !userDataKey.value.visible
  // formData.value.token = ''
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
const handleTokenTypeChange = (tokenType: number) => {
  formData.value.tokenType = tokenType
  formData.value.headers[0].key = tokenTypeOptions[tokenType - 1].label
  let roginData = null
  if (dialogInfo.value.currentEditIndex) {
    roginData = tokenList.value[dialogInfo.value.currentEditIndex] as any
    if ( tokenType === 4) {
      userDataKey.value.username = atob(roginData.headers.Authorization.split(' ')[1]).split(':')[0]
      userDataKey.value.password = atob(roginData.headers.Authorization.split(' ')[1]).split(':')[1]
    }
    return
  }
  handleRandomToken()
}

// handle random token
const handleRandomToken = () => {
  handleGetTokenValue()
}

// handle get token value
const handleGetTokenValue = () => {
  if (Number(formData.value.tokenType) === 1) {
    formData.value.headers[0].value =
      'Bearer ' +
      getToken(
        JSON.stringify({
          expireAt: formData.value.expireAt,
          userId: userInfo.userId,
          username: userInfo.username,
        }),
      )
  } else if (Number(formData.value.tokenType) === 2) {
    formData.value.headers[0].value = getToken(
      JSON.stringify({
        expireAt: formData.value.expireAt,
        userId: userInfo.userId,
        username: userInfo.username,
      }),
    )
  } else if (Number(formData.value.tokenType) === 3) {
    formData.value.headers[0].value = getToken(
      JSON.stringify({
        expireAt: formData.value.expireAt,
        userId: userInfo.userId,
        username: userInfo.username,
      }),
    )
  } else if (Number(formData.value.tokenType) === 4) {
    formData.value.headers[0].value =
      'Basic ' + btoa(`${userDataKey.value.username}:${userDataKey.value.password}`) // Base64 编码
  }
  console.log(formData.value.token, formData.value.headers[0]?.value)
}

const handleConfirmAccount = () => {
  handleGetTokenValue()
  userDataKey.value.visible = false
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
const handleConfirmToken = async () => {
  const result = await formRef.value.validate()
  if (!result) return
  if (dialogInfo.value.currentEditIndex) {
    tokenList.value[dialogInfo.value.currentEditIndex] = {
      token: formData.value.token,
      expireAt: formData.value.expireAt || 0,
      publishAt: tokenList.value[dialogInfo.value.currentEditIndex].publishAt,
      usages: formData.value.usages,
      enabled: tokenList.value[dialogInfo.value.currentEditIndex].enabled,
      headers: Object.fromEntries(
        formData.value.headers.map((header: any) => [header.key, header.value]),
      ),
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
      token: formData.value.token,
      expireAt: formData.value.expireAt || 0,
      publishAt: Date.now(),
      usages: formData.value.usages,
      headers: Object.fromEntries(
        formData.value.headers.map((header: any) => [header.key, header.value]),
      ),
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
    // 翻转一次；使默认token 在最前面
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
