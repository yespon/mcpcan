<template>
  <el-dialog v-model="dialogInfo.visible" width="1000px" top="10vh">
    <template #header>
      <div class="center">{{ dialogInfo.title }}</div>
    </template>
    <el-row :gutter="12">
      <el-col :span="12">
        <el-scrollbar ref="scrollbarRef" max-height="590px" always>
          <div class="token-list">
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
            <div class="font-bold">MCP 服务令牌</div>
            <div
              class="token-card border-rounded-2 mt-4 p-4 cursor-pointer line-height-8"
              :class="dialogInfo.currentTokenIndex === index ? 'active' : ''"
              v-for="(token, index) in dialogInfo.instanceInfo.tokens"
              :key="index"
              @click="dialogInfo.currentTokenIndex = index"
            >
              <div class="ellipsis-one">
                {{ token.token }}
              </div>
              <div class="ellipsis-one">
                有效时间：<span :class="!token.expireAt ? 'color-green' : ''">{{
                  token.expireAt ? timestampToDate(token.expireAt) : '永久有效'
                }}</span>
              </div>
              <div class="ellipsis-one">创建时间：{{ timestampToDate(token.publishAt) }}</div>
            </div>
          </div>
        </el-scrollbar>
      </el-col>
      <el-col :span="12">
        <el-scrollbar ref="scrollbarRef" max-height="590px" always>
          <div class="config-info">{{ config }}</div>
          <el-icon class="base-btn-link copy-icon" @click="handleCopy"><CopyDocument /></el-icon>
        </el-scrollbar>
      </el-col>
    </el-row>
    <template #footer>
      <div class="center">
        <mcp-button @click="handleCopy" class="w100">{{
          t('mcp.instance.action.copy')
        }}</mcp-button>
      </div>
    </template>
  </el-dialog>
</template>
<script setup lang="ts">
import { setClipboardData, timestampToDate } from '@/utils/system'
import { JsonFormatter } from '@/utils/json.ts'
import { ElMessage } from 'element-plus'
import { CopyDocument } from '@element-plus/icons-vue'
import McpButton from '@/components/mcp-button/index.vue'
import type { InstanceResult } from '@/types'
import { InstanceAPI } from '@/api/mcp/instance'

const { t } = useI18n()
const dialogInfo = ref({
  visible: false,
  title: t('mcp.instance.config'),
  instanceInfo: {} as InstanceResult,
  currentTokenIndex: null as number | null,
})

const config = computed(() => {
  return JsonFormatter.format(dialogInfo.value.instanceInfo.publicProxyConfig)
})

// handle enabled token switch
const handleEabledToken = async () => {
  try {
    dialogInfo.value.instanceInfo.loading = true
    await InstanceAPI.updateTokenStatus({
      instanceId: dialogInfo.value.instanceInfo.instanceId,
      enabledToken: dialogInfo.value.instanceInfo.enabledToken,
    })
  } catch (error) {
    dialogInfo.value.instanceInfo.enabledToken = !dialogInfo.value.instanceInfo.enabledToken
  } finally {
    dialogInfo.value.instanceInfo.loading = false
  }
}
/**
 * Handle copy config info
 */
const handleCopy = async () => {
  await setClipboardData(config)
  ElMessage.success(t('action.copy'))
}

/**
 * Handle init model data
 * @param config - public proxy config
 */
const init = (instanceInfo: InstanceResult) => {
  dialogInfo.value.visible = true
  dialogInfo.value.instanceInfo = instanceInfo
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
      background-color: var(--ep-bg-purple-color);
    }
    &:hover {
      border-color: var(--ep-purple-color);
    }
  }
}
.config-info {
  min-height: 590px;
  font-family: 'Monaco, Menlo, "Ubuntu Mono", monospace';
  font-size: 12px;
  line-height: 1.5;
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
</style>
