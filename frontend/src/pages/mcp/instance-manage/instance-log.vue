<template>
  <div class="common-layout">
    <el-container>
      <el-header class="p-0 flex align-center">
        <div class="log-title flex text-grey mb-2">
          <el-link v-if="layout" link @click="handleBack" class="link-hover mr-4" underline="never">
            <el-icon class="mr-2">
              <i class="icon iconfont MCP-fanhui"></i>
            </el-icon>
            {{ t('common.back') }}
          </el-link>
          <div>{{ t('mcp.instance.log.instanceId') }}:{{ pageInfo.loginfo.instanceId }}</div>
        </div>
      </el-header>
      <el-main v-loading="pageInfo.loading">
        <div class="log-info" v-if="pageInfo.loginfo.logs">{{ pageInfo.loginfo.logs }}</div>
        <div v-else class="center" style="width: 100%; height: 100%">
          <el-empty></el-empty>
        </div>
      </el-main>
      <el-footer class="center">
        <el-button @click="handleRefresh" :icon="RefreshRight">{{ t('common.refresh') }}</el-button>
        <el-button @click="handleDownload" :icon="Download">
          {{ t('common.download') }}
        </el-button>
        <el-button @click="handleClose"> {{ t('common.close') }} </el-button>
      </el-footer>
    </el-container>
  </div>
</template>

<script setup lang="ts">
import { RefreshRight, Download } from '@element-plus/icons-vue'
import { InstanceAPI } from '@/api/mcp/instance'
import { downloadData } from '@/utils/files'
import { ElMessage } from 'element-plus'
import { useRouterHooks } from '@/utils/url'
// import { timestampToDate } from '@/utils/system'

const { t } = useI18n()
const { query } = useRoute()
const layout = useLayout()
const { jumpBack } = useRouterHooks()

const pageInfo = ref({
  title: t('mcp.instance.log.title'),
  visible: false,
  loading: false,
  instance: {
    instanceId: '',
    updatedAt: '',
  },
  loginfo: {
    id: '',
    logs: '',
    instanceName: '',
    instanceId: '',
  },
})

/**
 * Handle refresh page
 */
const handleRefresh = () => {
  handleGetLogs()
}

/**
 * Handle download logs
 */
const handleDownload = async () => {
  try {
    const { instanceName, instanceId, logs } = pageInfo.value.loginfo
    await downloadData({
      fileName: `${instanceName}_${instanceId}_logs_${new Date().toISOString().slice(0, 19).replace(/:/g, '-')}`,
      data: logs,
    })
    ElMessage.success(t('action.download'))
  } finally {
  }
}

// back last class page
const handleBack = () => {
  jumpBack()
}

/**
 * Handle close page
 */
const handleClose = () => {
  jumpBack()
  // window.close()
}

/**
 * Handle get logs API
 */
const handleGetLogs = async () => {
  try {
    pageInfo.value.loading = true
    const data = await InstanceAPI.logs({
      instanceId: query.instanceId,
      lines: 100,
    })
    pageInfo.value.loginfo = data
  } finally {
    pageInfo.value.loading = false
  }
}

/**
 * Handle init data
 *
 */
const init = () => {
  handleGetLogs()
}

onMounted(init)
</script>

<style lang="scss" scoped>
.common-layout {
  width: 100vm;
  height: 100vh;
  .el-header {
    border-bottom: 1px solid var(--el-menu-border-color);
  }
  .el-footer {
    border-top: 1px solid var(--el-menu-border-color);
  }
  .el-container {
    height: 100%;
  }
  .el-main {
    height: calc(100vh - 120px);
    &.p-0 {
      padding: 0 !important;
    }
  }
}
.log-info {
  font-family: 'Monaco, Menlo, "Ubuntu Mono", monospace';
  font-size: 12px;
  line-height: 1.5;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-all;
  border-radius: 8px;
}
</style>
