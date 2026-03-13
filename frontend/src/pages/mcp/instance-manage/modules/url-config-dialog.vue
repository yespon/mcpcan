<template>
  <el-dialog v-model="dialogInfo.visible" width="480px" align-center :show-close="false">
    <template #header>
      <div class="center">{{ dialogInfo.title }}</div>
    </template>
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
      <div class="custom-style my-4 px-5">
        <el-segmented
          v-model="dialogInfo.pathType"
          :options="pathTypeOptions"
          :disabled="disabled"
        />
      </div>
    </el-scrollbar>
    <template #footer>
      <div class="center">
        <el-button class="w-25 mr-2" @click="dialogInfo.visible = false">{{
          t('common.close')
        }}</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { CopyDocument, Link, Key } from '@element-plus/icons-vue'
import { setClipboardData } from '@/utils/system'
import { ElMessage } from 'element-plus'
import { AccessType, McpProtocol } from '@/types'
import { JsonFormatter } from '@/utils/json'
import { TokenAPI } from '@/api/mcp/instance'

const { t } = useI18n()
const dialogInfo = ref<any>({
  visible: false,
  loading: false,
  title: '访问配置',
  instanceInfo: {
    instanceId: '',
  },
  pathType: 'streamable_http',
})
const pathTypeOptions = [
  { label: 'SSE', value: 'sse' },
  { label: 'STREAMABLE_HTTP', value: 'streamable_http' },
]
// token list
const tokenList = ref<Array<any>>([])
const disabled = computed(() => {
  return !(
    dialogInfo.value.instanceInfo.accessType === AccessType.HOSTING &&
    dialogInfo.value.instanceInfo.mcpProtocol === McpProtocol.STDIO
  )
})
const configUrl = computed(() => {
  if (dialogInfo.value.instanceInfo.accessType === AccessType.DIRECT) {
    try {
      const mcpServers = JSON.parse(dialogInfo.value.instanceInfo.sourceConfig).mcpServers
      return mcpServers[Object.keys(mcpServers)[0]].url
    } catch {
      return ''
    }
  }
  let publicPath = dialogInfo.value.instanceInfo.publicProxyPath || ''
  // 如果当前 pathType 与 publicProxyPath 末尾不一致（HOSTING 模式下 STDIO 实例可切换），则替换末尾路径段
  if (
    dialogInfo.value.instanceInfo.accessType === AccessType.HOSTING &&
    dialogInfo.value.instanceInfo.mcpProtocol === McpProtocol.STDIO
  ) {
    const lastSlashIndex = publicPath.lastIndexOf('/')
    if (lastSlashIndex !== -1) {
      publicPath =
        publicPath.substring(0, lastSlashIndex + 1) +
        (dialogInfo.value.pathType === 'sse' ? 'sse' : 'mcp')
    }
  }
  return `${window.location.origin}${(window as any).__APP_CONFIG__?.PUBLIC_PATH || ''}${publicPath}`
})
const configToken = computed(() => {
  if (dialogInfo.value.instanceInfo.accessType === AccessType.DIRECT) {
    const mcpServers = JSON.parse(dialogInfo.value.instanceInfo.sourceConfig).mcpServers
    return mcpServers[Object.keys(mcpServers)[0]].token || 'No Data'
  }
  return `${tokenList.value[0].token}`
})
const config = computed(() => {
  if (dialogInfo.value.instanceInfo.accessType === AccessType.DIRECT) {
    return JsonFormatter.format(dialogInfo.value.instanceInfo.sourceConfig, 4)
  }
  // 根据实际 URL 末尾推断协议类型
  const urlPath = configUrl.value || ''
  const inferredType = urlPath.endsWith('/sse') ? 'sse' : 'streamable_http'
  if (dialogInfo.value.instanceInfo.enabledToken) {
    if (!tokenList.value) return JsonFormatter.format(`{}`, 4)
    if (tokenList.value.length) {
      return JsonFormatter.format(
        `{
          "mcpServers": {
                "mcp-${dialogInfo.value.instanceInfo.instanceId.slice(0, 8)}": {
                      "url": "${configUrl.value}",
                      "type": "${inferredType}",
                      "headers": {
                            "Authorization": "${tokenList.value[0]?.token}"
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
              "url": "${configUrl.value}",
              "type": "${inferredType}"
          }
      }
  }`,
    4,
  )
})

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

const handleTokenList = async () => {
  dialogInfo.value.instanceInfo.loading = true
  try {
    const res = await TokenAPI.list({
      instanceId: dialogInfo.value.instanceInfo.instanceId,
    })
    const tokens = res.tokens || res.list || []
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

const init = (instanceInfo: any) => {
  dialogInfo.value.instanceInfo = instanceInfo
  // 根据 proxyProtocol 或 mcpProtocol 决定默认显示的协议类型
  const proxyProto = instanceInfo.proxyProtocol || instanceInfo.mcpProtocol
  if (proxyProto === McpProtocol.SSE || proxyProto === 1) {
    dialogInfo.value.pathType = 'sse'
  } else {
    dialogInfo.value.pathType = 'streamable_http'
  }
  // 如果 publicProxyPath 明确以 /sse 结尾，强制设为 sse
  const proxyPath = instanceInfo.publicProxyPath || ''
  if (proxyPath.endsWith('/sse')) {
    dialogInfo.value.pathType = 'sse'
  } else if (proxyPath.endsWith('/mcp')) {
    dialogInfo.value.pathType = 'streamable_http'
  }
  handleTokenList()
  dialogInfo.value.visible = true
}

defineExpose({
  init,
})
</script>

<style lang="scss" scoped>
.custom-style {
  position: absolute;
  bottom: 0;
  right: 0;
  .el-segmented {
    --el-segmented-item-selected-color: var(--el-text-color-primary);
    --el-border-radius-base: 16px;
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
</style>
