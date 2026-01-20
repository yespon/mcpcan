<template>
  <el-dialog
    v-model="dialogInfo.visible"
    :title="dialogInfo.title"
    width="60vw"
    top="10vh"
    :close-on-click-modal="false"
  >
    <el-scrollbar
      style="height: 70vh; padding-bottom: 20px"
      always
      class="config-info"
      v-loading="dialogInfo.loading"
    >
      <el-button-group class="action-group my-4 px-5">
        <el-tooltip class="box-item" effect="dark" :content="t('common.download')" placement="top">
          <el-button :icon="Download" @click="handleDownload" />
        </el-tooltip>
        <el-tooltip class="box-item" effect="dark" :content="t('common.refresh')" placement="top">
          <el-button :icon="Refresh" @click="handleRefresh" />
        </el-tooltip>
      </el-button-group>
      <div class="custom-style my-4 px-5">
        <el-segmented v-model="dialogInfo.logType" :options="logTypeOptions" />
      </div>
      <div v-if="dialogInfo.logType === 'container'" class="px-5 py-2 log-info">
        {{ dialogInfo.logContent.logs || '暂无日志内容' }}
      </div>
      <div v-else class="px-5 py-2">
        <template v-if="dialogInfo.accessLogs.length">
          <div v-for="(log, index) in dialogInfo.accessLogs" :key="index" class="mb-4">
            <div><strong>时间：</strong>{{ log.timestamp }}</div>
            <div><strong>方法：</strong>{{ log.method }}</div>
            <div><strong>路径：</strong>{{ log.path }}</div>
            <div><strong>状态码：</strong>{{ log.statusCode }}</div>
            <div><strong>客户端IP：</strong>{{ log.clientIp }}</div>
            <div><strong>用户代理：</strong>{{ log.userAgent }}</div>
            <el-divider></el-divider>
          </div>
        </template>
        <div v-else>暂无访问日志内容</div>
      </div>
    </el-scrollbar>
  </el-dialog>
</template>

<script setup lang="ts">
import { InstanceAPI } from '@/api/mcp/instance'
import { type InstanceResult } from '@/types/instance.ts'
import { Download, Refresh } from '@element-plus/icons-vue'
import { downloadData } from '@/utils/files'
import { ElMessage } from 'element-plus'

const { t } = useI18n()
const logTypeOptions = [
  { label: '容器日志', value: 'container' },
  { label: '访问日志', value: 'access' },
]
const dialogInfo = ref<any>({
  visible: false,
  loading: false,
  title: '日志详情',
  logContent: '',
  accessLogs: [],
  logType: 'container',
  instanceId: '',
})

/**
 * Handle refresh page
 */
const handleRefresh = () => {
  handleGetContainerLogs()
}

/**
 * Handle download logs
 */
const handleDownload = async () => {
  try {
    const { instanceName, instanceId, logs } = dialogInfo.value.logContent
    await downloadData({
      fileName: `${instanceName}_${instanceId}_logs_${new Date().toISOString().slice(0, 19).replace(/:/g, '-')}`,
      data: logs,
    })
    ElMessage.success(t('action.download'))
  } finally {
  }
}
/**
 * Handle get logs API （container logs）
 */
const handleGetContainerLogs = async () => {
  try {
    dialogInfo.value.loading = true
    const data = await InstanceAPI.logs({
      instanceId: dialogInfo.value.instanceId,
      lines: 100,
    })
    dialogInfo.value.logContent = data
  } finally {
    dialogInfo.value.loading = false
  }
}

// get logs based on filters
const handleGetAccessLogs = async () => {
  try {
    dialogInfo.value.loading = true
    const params: any = {
      instanceId: dialogInfo.value.instanceId,
      pageNum: 1,
      pageSize: 999,
    }
    const { logs: logData } = await InstanceAPI.logsByToken(params)
    dialogInfo.value.accessLogs = logData || []
  } catch (error) {
    dialogInfo.value.accessLogs = []
  } finally {
    dialogInfo.value.loading = false
  }
}

// handle init data
const init = (instanceInfo: InstanceResult) => {
  dialogInfo.value.instanceId = instanceInfo.instanceId
  handleGetContainerLogs()
  handleGetAccessLogs()
  dialogInfo.value.visible = true
}

defineExpose({
  init,
})
</script>
<style lang="scss" scoped>
.action-group {
  position: absolute;
  top: 0;
  right: 0;
}
.custom-style {
  position: absolute;
  bottom: 0;
  right: 0;
  .el-segmented {
    --el-segmented-item-selected-color: var(--el-text-color-primary);
    --el-border-radius-base: 16px;
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
</style>
