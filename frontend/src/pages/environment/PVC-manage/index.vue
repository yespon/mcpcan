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
        {{ query.name }} - {{ t('env.run.pageDesc.pvcList') }}
        <span class="desc">{{ t('env.run.pageDesc.pvcDesc') }}</span>
      </div>

      <mcp-button :icon="Plus" @click="handleAddPvc">
        {{ t('env.pvc.action.add') }}
      </mcp-button>
    </div>

    <!-- 表格区域 -->
    <TablePlus
      ref="tablePlus"
      search-container="#pvcSearch"
      :showOperation="false"
      :requestConfig="requestConfig"
      :columns="columns"
      v-model:pageConfig="pageConfig"
      :handlerColumnConfig="null"
      :showPage="false"
    >
      <template #action>
        <div class="flex justify-between mb-4">
          <div class="center">
            <el-image :src="pvcLogo" style="width: 20px; height: 20px"></el-image>
            <span class="desc">{{ t('env.run.pageDesc.pvcTotal') }}：{{ pageConfig.total }}</span>
          </div>
          <div id="pvcSearch" v-show="false"></div>
        </div>
      </template>
      <template #accessModes="{ row }">
        <el-tag type="primary" v-for="(accessModes, index) in row.accessModes" :key="index">
          {{ accessModes }}
        </el-tag>
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
      <template #pods="{ row }">
        <el-tag type="info" v-if="!row.pods?.length"> {{ t('env.pvc.unbound') }} </el-tag>
        <el-tag type="primary" v-for="(pod, index) in row.pods" :key="index" class="mr-1">
          {{ pod }}
        </el-tag>
      </template>
      <template #createdAt="{ row }">{{ timestampToDate(row.createdAt) }} </template>
      <template #updatedAt="{ row }">{{ timestampToDate(row.updatedAt) }} </template>
    </TablePlus>
    <NewPvcDialog ref="newPvcDialog" @on-refresh="init"></NewPvcDialog>
  </div>
</template>

<script setup lang="ts">
import TablePlus from '@/components/TablePlus/index.vue'
import { Plus } from '@element-plus/icons-vue'
import { timestampToDate } from '@/utils/system'
import { usePvcTableHooks } from './hooks'
import NewPvcDialog from './modules/new-pvc-dialog.vue'
import pvcLogo from '@/assets/logo/pvc.png'
import McpButton from '@/components/mcp-button/index.vue'
import { useRouterHooks } from '@/utils/url'

const { t } = useI18n()
const layout = useLayout()
const {
  PvcAPI,
  tablePlus,
  columns,
  storageClassOptions,
  requestConfig,
  pageConfig,
  newPvcDialog,
  query,
} = usePvcTableHooks()

const { jumpBack } = useRouterHooks()

const handleAddPvc = () => {
  newPvcDialog.value.init()
}

/**
 * Handle get storageClass list
 */
const handleGetStorageClassList = async () => {
  const data = await PvcAPI.storageList({ environmentId: query.environmentId })
  storageClassOptions.value = data.list.map((storage) => {
    return { label: storage.name, value: storage.name }
  })
}

// back last class page
const handleBack = () => {
  jumpBack()
}

/**
 * handle init list and storageClass list
 */
const init = () => {
  tablePlus.value.initData()
  handleGetStorageClassList()
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
