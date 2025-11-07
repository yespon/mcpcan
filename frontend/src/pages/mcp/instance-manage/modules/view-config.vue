<template>
  <el-dialog v-model="dialogInfo.visible" width="1000px" top="10vh" :show-close="false">
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
      <span class="ml-2">开启令牌认证；访问该MCP服务时将进行令牌认证校验</span>
    </div>
    <el-row :gutter="12" class="vc-row">
      <el-col
        :span="12"
        :class="['config-col left', { collapsed: !dialogInfo.instanceInfo.enabledToken }]"
      >
        <el-scrollbar ref="scrollbarRef" max-height="590px" always>
          <div class="token-list">
            <div class="font-bold flex justify-between items-center">
              <mcp-button :icon="Plus" size="small" @click="handleAddToken">添加令牌</mcp-button>
              <div>
                <span class="mr-4 color-green">
                  有效:
                  {{ tokenList.filter((token) => !token.expire).length }}个
                </span>
                <span class="color-red">
                  过期:{{ tokenList.filter((token) => token.expire).length }}个
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
                  trigger="hover"
                  class="ml-2"
                  style="cursor: pointer !important"
                  @click.stop
                  :show-arrow="false"
                >
                  <el-icon size="18"><Operation /></el-icon>
                  <template #dropdown>
                    <el-dropdown-menu>
                      <el-dropdown-item command="handleViewLog" @click="handleEditToken(index)">
                        {{ '编辑' }}
                      </el-dropdown-item>
                      <el-dropdown-item command="handleViewLog" @click="handleDeleteToken(index)">
                        <el-button type="danger" link>
                          {{ '删除' }}
                        </el-button>
                      </el-dropdown-item>
                    </el-dropdown-menu>
                  </template>
                </el-dropdown>
              </div>

              <div class="ellipsis-one">
                有效时间：<span
                  :class="
                    !token.expireAt ? 'color-green' : token.expireAt < Date.now() ? 'color-red' : ''
                  "
                >
                  {{
                    !token.expireAt
                      ? '永久有效'
                      : timestampToDate(token.expireAt) +
                        (token.expireAt < Date.now() ? '（已过期）' : '')
                  }}
                </span>
              </div>
              <div class="ellipsis-one">创建时间：{{ timestampToDate(token.publishAt) }}</div>
            </div>
          </div>
        </el-scrollbar>
      </el-col>
      <el-col
        :span="12"
        :class="['config-col right', { expanded: !dialogInfo.instanceInfo.enabledToken }]"
      >
        <el-scrollbar ref="scrollbarRef" max-height="590px" always>
          <div class="config-info">{{ config }}</div>
          <el-icon class="base-btn-link copy-icon" @click="handleCopy"><CopyDocument /></el-icon>
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
  <el-dialog v-model="formData.visible" width="360px" top="30vh" :show-close="false">
    <template #header>
      <div class="center mb-4">安全认证令牌</div>
      <el-form
        ref="formRef"
        :model="formData"
        :rules="rules"
        label-width="auto"
        label-position="top"
      >
        <el-form-item label="有效期" prop="token">
          <template #label>
            <div class="center">
              <span class="mr-2">有效期</span>
              <mcp-button
                class="mr-2 base-btn cursor-pointer"
                size="small"
                @click.stop="handleAddExpireAt(7)"
                >7天
              </mcp-button>
              <mcp-button
                class="mr-2 base-btn cursor-pointer"
                size="small"
                @click.stop="handleAddExpireAt(15)"
                >15天
              </mcp-button>
              <mcp-button
                class="mr-2 base-btn cursor-pointer"
                size="small"
                @click.stop="handleAddExpireAt(30)"
                >30天
              </mcp-button>
            </div>
          </template>
          <el-date-picker
            ref="datePicker"
            v-model="formData.expireAt"
            type="datetime"
            value-format="x"
            placeholder="请选择过期时间，不选择则永久有效"
            style="width: 100%"
            :disabled-date="(date: Date) => date.getTime() < Date.now()"
          ></el-date-picker>
        </el-form-item>
        <el-form-item label="Token" prop="token">
          <template #label>
            Token值
            <el-tag class="ml-2 base-btn cursor-pointer" effect="dark" @click="handleRandomToken"
              >随机生成</el-tag
            >
          </template>
          <el-input
            v-model="formData.token"
            :rows="4"
            type="textarea"
            placeholder="Authorization:请输入Token值或自动生成"
          />
        </el-form-item>
      </el-form>
    </template>
    <template #footer>
      <div class="center">
        <mcp-button @click="handleConfirmToken" class="w-25">{{ t('common.ok') }}</mcp-button>
      </div>
    </template>
  </el-dialog>
</template>
<script setup lang="ts">
import { Plus, Operation } from '@element-plus/icons-vue'
import { setClipboardData, timestampToDate } from '@/utils/system'
import { JsonFormatter } from '@/utils/json.ts'
import { ElMessage } from 'element-plus'
import { CopyDocument } from '@element-plus/icons-vue'
import McpButton from '@/components/mcp-button/index.vue'
import { type InstanceResult } from '@/types'
import { InstanceAPI } from '@/api/mcp/instance'
import { getToken } from '@/utils/system'
import { useUserStore } from '@/stores'
import { cloneDeep } from 'lodash-es'
import { de } from 'element-plus/es/locales.mjs'

const { t } = useI18n()
const { userInfo } = useUserStore()
const emit = defineEmits<{
  (e: 'on-refresh'): void
}>()
const formData = ref({
  visible: false,
  token: '',
  expireAt: null as number | null,
})
const rules = reactive({
  name: [{ required: true, message: '请输入令牌名称', trigger: 'blur' }],
})
const dialogInfo = ref({
  visible: false,
  title: t('mcp.instance.config'),
  instanceInfo: {} as InstanceResult,
  currentTokenIndex: null as number | null,
  currentEditIndex: null as number | null,
})

// config Info
const config = computed(() => {
  // "type": "${Object.keys(McpProtocol).filter((key) => isNaN(Number(key)))[dialogInfo.value.instanceInfo.proxyProtocol]}",
  if (dialogInfo.value.instanceInfo.enabledToken) {
    return JsonFormatter.format(
      `{
      "mcpServers": {
          "mcp-${dialogInfo.value.instanceInfo.instanceId.slice(0, 8)}": {
              "url": "${window.location.origin}${dialogInfo.value.instanceInfo.publicProxyPath}",
              "headers": {
                "Authorization": "${
                  dialogInfo.value.currentTokenIndex !== null
                    ? dialogInfo.value.instanceInfo.tokens[dialogInfo.value.currentTokenIndex].token
                    : ''
                }"
              }
          }
      }
  }`,
      4,
    )
  } else {
    return JsonFormatter.format(
      `{
        "mcpServers": {
            "mcp-${dialogInfo.value.instanceInfo.instanceId.slice(0, 8)}": {
                "url": "${window.location.origin}${dialogInfo.value.instanceInfo.publicProxyPath}"
            }
        }
    }`,
      4,
    )
  }
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
  formData.value.visible = true
}

// handle enabled token switch
const handleEabledToken = async () => {
  try {
    dialogInfo.value.instanceInfo.loading = true
    await InstanceAPI.updateTokenStatus({
      instanceId: dialogInfo.value.instanceInfo.instanceId,
      enabledToken: dialogInfo.value.instanceInfo.enabledToken,
    })
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
  formData.value.token =
    'Bearer ' +
    getToken(
      JSON.stringify({
        userId: userInfo.userId,
        username: userInfo.username,
        expireAt: formData.value.expireAt,
      }),
    )
}

// handle add expire at
const datePicker = ref()
const handleAddExpireAt = (days: number) => {
  const expireDate = new Date()
  expireDate.setDate(expireDate.getDate() + days)
  formData.value.expireAt = expireDate.getTime()
  console.log(datePicker.value)
  nextTick(() => {
    datePicker.value.blur()
  })
}
// handle edit token
const handleEditToken = (index: number) => {
  formData.value.visible = true
  dialogInfo.value.currentEditIndex = index
  const token = dialogInfo.value.instanceInfo.tokens[index]
  formData.value.token = token.token
  formData.value.expireAt = token.expireAt
}

// handle delete token
const handleDeleteToken = (index: number) => {
  dialogInfo.value.instanceInfo.tokens.splice(index, 1)
  if (dialogInfo.value.currentTokenIndex === index) {
    dialogInfo.value.currentTokenIndex = null
  } else if (dialogInfo.value.currentTokenIndex && dialogInfo.value.currentTokenIndex > index) {
    dialogInfo.value.currentTokenIndex!--
  }
  handleSaveTokens()
}

// handle confirm token
const handleConfirmToken = () => {
  if (dialogInfo.value.currentEditIndex) {
    dialogInfo.value.instanceInfo.tokens[dialogInfo.value.currentEditIndex] = {
      token: formData.value.token,
      expireAt: formData.value.expireAt || 0,
      publishAt: dialogInfo.value.instanceInfo.tokens[dialogInfo.value.currentEditIndex].publishAt,
      usages: dialogInfo.value.instanceInfo.tokens[dialogInfo.value.currentEditIndex].usages,
    }
    dialogInfo.value.currentEditIndex = null
    formData.value.visible = false
    formData.value.token = ''
    formData.value.expireAt = null
    handleSaveTokens()
    return
  }
  dialogInfo.value.instanceInfo.tokens.push({
    token: formData.value.token,
    expireAt: formData.value.expireAt || 0,
    publishAt: Date.now(),
    usages: [],
  })
  formData.value.visible = false
  formData.value.token = ''
  formData.value.expireAt = null
  handleSaveTokens()
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
  } catch (error) {
    throw error
  } finally {
    dialogInfo.value.instanceInfo.loading = false
  }
}

/**
 * Handle copy config info
 */
const handleCopy = async () => {
  await setClipboardData(config.value)
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
  background: #000000;
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
  }
}
.config-info {
  min-height: 590px;
  font-family: 'Monaco, Menlo, "Ubuntu Mono", monospace';
  font-size: 12px;
  line-height: 1.8;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-all;
  border-radius: 8px;
  background: #000000;
  border-radius: 8px;
  padding: 24px;
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
</style>
