<template>
  <el-dialog v-model="dialogInfo.visible" width="1000px" top="10vh">
    <template #header>
      <div class="center">{{ dialogInfo.title }}</div>
    </template>
    <div class="mb-2 flex items-center">
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
    <el-row :gutter="12" class="vc-row" v-loading="dialogInfo.instanceInfo.loading">
      <el-col
        :span="12"
        :class="['config-col left', { collapsed: !dialogInfo.instanceInfo.enabledToken }]"
      >
        <el-scrollbar ref="scrollbarRef" max-height="590px" always>
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

              <div class="ellipsis-one">
                {{ t('mcp.instance.token.expireAt') }}：<span
                  :class="
                    !token.expireAt ? 'color-green' : token.expireAt < Date.now() ? 'color-red' : ''
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
              <div class="ellipsis-one">
                {{ t('mcp.instance.createTime') }}：{{ timestampToDate(token.publishAt) }}
              </div>
              <div class="ellipsis-one">
                {{ t('mcp.instance.token.tag') }}：<el-tag
                  v-for="(tag, num) in token.usages"
                  :key="num"
                  effect="plain"
                  class="mr-2"
                >
                  {{ tag }}
                </el-tag>
              </div>
            </div>
          </div>
        </el-scrollbar>
      </el-col>
      <el-col
        :span="12"
        :class="['config-col right', { expanded: !dialogInfo.instanceInfo.enabledToken }]"
      >
        <el-scrollbar ref="scrollbarRef" max-height="590px" always class="config-info">
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
    width="360px"
    top="20vh"
    :show-close="false"
    @close="formRef?.resetFields()"
  >
    <template #header>
      <div class="center mb-4">{{ t('mcp.instance.token.title') }}</div>
      <el-form
        ref="formRef"
        :model="formData"
        :rules="rules"
        label-width="auto"
        label-position="top"
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
        <el-form-item label="Token" prop="token">
          <template #label>
            Token
            <el-tag class="ml-2 base-btn cursor-pointer" effect="dark" @click="handleRandomToken">{{
              t('mcp.instance.token.random')
            }}</el-tag>
          </template>
          <el-input
            v-model="formData.token"
            :rows="4"
            type="textarea"
            :placeholder="t('mcp.instance.token.placeholderToken')"
          />
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
    </template>
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
import { Plus, Operation, CopyDocument, Link, Key } from '@element-plus/icons-vue'
import McpButton from '@/components/mcp-button/index.vue'
import { type InstanceResult } from '@/types'
import { InstanceAPI } from '@/api/mcp/instance'
import { getToken } from '@/utils/system'
import { useUserStore } from '@/stores'
import { cloneDeep } from 'lodash-es'

const { t } = useI18n()
const { userInfo } = useUserStore()
const emit = defineEmits<{
  (e: 'on-refresh'): void
}>()
const formRef = ref()
const formData = ref({
  visible: false,
  token: '',
  expireAt: null as number | null,
  usages: [] as string[],
})
const rules = reactive({
  token: [{ required: true, message: t('mcp.instance.token.mustToken'), trigger: 'blur' }],
})
const dialogInfo = ref({
  visible: false,
  title: t('mcp.instance.config'),
  instanceInfo: {} as InstanceResult,
  currentTokenIndex: null as number | null,
  currentEditIndex: null as number | null,
})

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

// handle add token
const handleAddToken = () => {
  formData.value.token = ''
  formData.value.expireAt = null
  formData.value.usages = []
  dialogInfo.value.currentEditIndex = null
  formData.value.visible = true
  formRef.value?.resetFields()
}

// handle enabled token switch
const handleEabledToken = async () => {
  try {
    dialogInfo.value.instanceInfo.loading = true
    await InstanceAPI.updateTokenStatus({
      instanceId: dialogInfo.value.instanceInfo.instanceId,
      enabledToken: dialogInfo.value.instanceInfo.enabledToken,
    })
    dialogInfo.value.currentTokenIndex = null
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
      formData.value.token =
        'Bearer ' +
        getToken(
          JSON.stringify({
            expireAt: formData.value.expireAt,
            userId: userInfo.userId,
            username: userInfo.username,
          }),
        )
    })
    return
  }

  formData.value.token =
    'Bearer ' +
    getToken(
      JSON.stringify({
        expireAt: formData.value.expireAt,
        userId: userInfo.userId,
        username: userInfo.username,
      }),
    )
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

  if (dialogInfo.value.currentEditIndex) {
    dialogInfo.value.instanceInfo.tokens[dialogInfo.value.currentEditIndex] = {
      token: formData.value.token,
      expireAt: formData.value.expireAt || 0,
      publishAt: dialogInfo.value.instanceInfo.tokens[dialogInfo.value.currentEditIndex].publishAt,
      usages: formData.value.usages,
    }
    dialogInfo.value.currentEditIndex = null
    formData.value.visible = false
    formData.value.token = ''
    formData.value.expireAt = null
    formData.value.usages = []
    handleSaveTokens()
    return
  }
  try {
    dialogInfo.value.instanceInfo.tokens.push({
      token: formData.value.token,
      expireAt: formData.value.expireAt || 0,
      publishAt: Date.now(),
      usages: formData.value.usages,
    })
    formData.value.visible = false
    formData.value.token = ''
    formData.value.expireAt = null
    await handleSaveTokens()
  } catch (error) {
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
  min-height: 590px;
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

/* layout row fix to allow flex transitions */
.vc-row {
  display: flex;
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
</style>
