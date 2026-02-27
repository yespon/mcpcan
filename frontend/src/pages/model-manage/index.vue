<template>
  <div v-loading="pageInfo.loading" :element-loading-text="pageInfo.loadingText">
    <div class="flex justify-between page-header mb-4">
      <div class="header-container">
        {{ t('model.pageDesc.list') }}
        <span class="desc">{{ t('model.pageDesc.desc') }}</span>
      </div>
      <mcp-button type="primary" :icon="Plus" @click="handleCreate">{{
        t('model.create')
      }}</mcp-button>
    </div>

    <TablePlus
      :showOperation="true"
      searchContainer="#modelSearch"
      ref="tablePlus"
      :requestConfig="requestConfig"
      :columns="columns"
      v-model:pageConfig="pageConfig"
      :handlerColumnConfig="{
        width: '180px',
        fixed: 'right',
        align: 'center',
      }"
    >
      <template #action>
        <div class="flex justify-between mb-4">
          <div class="center">
            <span class="desc">{{ t('common.total') }}：{{ pageConfig.total }}</span>
          </div>
          <div id="modelSearch"></div>
        </div>
      </template>

      <template #operation="{ row }">
        <el-button type="primary" link size="small" @click="handleTestConnection(row)">
          {{ t('model.testConnection') }}
        </el-button>
        <el-button type="primary" link size="small" @click="handleEdit(row)">
          {{ t('common.edit') }}
        </el-button>
        <el-button type="danger" link size="small" @click="handleDelete(row)">
          {{ t('common.delete') }}
        </el-button>
      </template>
    </TablePlus>

    <ModelDialog v-model="dialogVisible" :current-model="currentModel" @success="handleSuccess" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Plus } from '@element-plus/icons-vue'
import TablePlus from '@/components/TablePlus/index.vue'
import { useModelTableHooks } from './hooks'
import { ElMessage, ElMessageBox } from 'element-plus'
import ModelDialog from './components/ModelDialog.vue'
import { ChatAPI } from '@/api/agent'

const { t, tablePlus, columns, requestConfig, pageConfig, pageInfo } = useModelTableHooks()

const dialogVisible = ref(false)
const currentModel = ref<any>(null)

const handleCreate = () => {
  currentModel.value = null
  dialogVisible.value = true
}

const handleEdit = (row: any) => {
  currentModel.value = { ...row }
  dialogVisible.value = true
}

const handleTestConnection = async (row: any) => {
  try {
    pageInfo.value.loading = true
    const res = await ChatAPI.testConnectionById(row.id)
    if (res.code === 0) {
      ElMessage.success(t('model.testConnectionSuccess'))
    } else {
      ElMessage.error(t('model.testConnectionFailed'))
    }
  } catch (error) {
    ElMessage.error(t('model.testConnectionFailed'))
  } finally {
    pageInfo.value.loading = false
  }
}

const handleDelete = (row: any) => {
  ElMessageBox.confirm(t('model.deleteConfirm'), t('common.warn'), {
    confirmButtonText: t('common.ok'),
    cancelButtonText: t('common.cancel'),
    type: 'warning',
  }).then(async () => {
    try {
      await ChatAPI.deleteModelAccess(row.id)
      ElMessage.success(t('model.deleteSuccess'))
      tablePlus.value.initData()
    } catch (error) {
      // Error handling
    }
  })
}

const handleSuccess = () => {
  tablePlus.value.initData()
}

onMounted(() => {
  // Initial load is handled by TablePlus usually, or call initData
  tablePlus.value.initData()
})
</script>

<style lang="scss" scoped>
.page-header {
  .header-container {
    font-size: 20px;
    font-weight: 500;
  }
}
.desc {
  font-size: 14px;
  color: #999999;
  margin-left: 10px;
}
</style>
