<template>
  <div>
    <!-- 头部区域 -->
    <div class="flex justify-between page-header">
      <div class="header-container flex items-center">
        <el-link v-if="layout" link @click="handleBack" class="link-hover mr-4" underline="never">
          <el-icon class="mr-2">
            <i class="icon iconfont MCP-fanhui"></i>
          </el-icon>
          {{ t('common.back') }}
        </el-link>
        {{ query.name }} - {{ t('env.volume.title') }}
        <span class="desc">{{ t('env.volume.desc') }}</span>
      </div>

      <mcp-button :icon="Plus" @click="handleAddVolume">
        {{ t('env.volume.add') }}
      </mcp-button>
    </div>

    <!-- 表格区域 -->
    <TablePlus
      ref="tablePlus"
      search-container="#volumeSearch"
      :showOperation="true"
      :requestConfig="requestConfig"
      :columns="columns"
      v-model:pageConfig="pageConfig"
      :showPage="false"
      :handlerColumnConfig="{
        fixed: 'right',
        width: '120px',
      }"
    >
      <template #action>
        <div class="flex justify-between mb-4">
          <div class="center">
            <el-image :src="pvcLogo" style="width: 20px; height: 20px"></el-image>
            <span class="desc">{{ t('env.volume.total') }}：{{ pageConfig.total }}</span>
          </div>
          <div id="volumeSearch" v-show="false"></div>
        </div>
      </template>
      <template #status="{ row }">
        <el-tag
          :type="
            row.status === 'Pending' ? 'warning' : row.status === 'Bound' ? 'success' : 'danger'
          "
        >
          {{ row.status }}
        </el-tag>
      </template>
      <template #createdAt="{ row }">{{ timestampToDate(row.createdAt) }} </template>
      <template #updatedAt="{ row }">{{ timestampToDate(row.updatedAt) }} </template>
      <template #operation="{ row }">
        <el-button type="danger" link @click="handleDeleteVolume(row)">
          {{ t('env.run.action.delete') }}
        </el-button>
      </template>
    </TablePlus>
    <NewVolumeDialog ref="newVolumeDialog" @on-refresh="init"></NewVolumeDialog>
  </div>
</template>

<script setup lang="ts">
import TablePlus from '@/components/TablePlus/index.vue'
import { Plus } from '@element-plus/icons-vue'
import { timestampToDate } from '@/utils/system'
import { usePvcTableHooks } from './hooks'
import NewVolumeDialog from './modules/new-volume-dialog.vue'
import pvcLogo from '@/assets/logo/pvc.png'
import McpButton from '@/components/mcp-button/index.vue'
import { useRouterHooks } from '@/utils/url'
import { ElMessageBox, ElMessage } from 'element-plus'

const { t } = useI18n()
const layout = useLayout()
const { tablePlus, columns, VolumeAPI, requestConfig, pageConfig, newVolumeDialog, query } =
  usePvcTableHooks()

const { jumpBack } = useRouterHooks()

const handleAddVolume = () => {
  newVolumeDialog.value.init()
}

// back last class page
const handleBack = () => {
  jumpBack()
}

const handleDeleteVolume = (row: any) => {
  ElMessageBox.confirm(t('env.volume.confirm'), t('common.confirm'), {
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
  })
    .then(async () => {
      try {
        await VolumeAPI.delete({
          name: row.name,
          environmentId: query.environmentId,
        })
        ElMessage.success(t('action.delete'))
        tablePlus.value.initData()
      } catch (error: any) {}
    })
    .catch(() => {
      // cancel
    })
}

/**
 * handle init list and storageClass list
 */
const init = () => {
  tablePlus.value.initData()
}

onMounted(init)
</script>
<style lang="scss" scoped>
.page-header {
  margin-bottom: 24px;
  .header-container {
    font-size: 20px;
  }
}
.desc {
  font-size: 16px;
  color: #999999;
  margin-left: 16px;
}
</style>
