<template>
  <div class="role-page" v-loading="pageInfo.loading">
    <div class="flex justify-between page-header">
      <div class="header-container flex-1">
        {{ t('system.department.title') }}
        <span class="desc">{{ t('system.department.pageDesc') }}</span>
      </div>
      <div v-auth="'mcpcan_rbac_manage:dept:create'">
        <mcp-button :icon="Plus" @click="handleNewRole">{{
          t('system.department.operation.add')
        }}</mcp-button>
      </div>
    </div>
    <div>
      <TablePlus
        :showOperation="true"
        searchContainer="#roleSearch"
        ref="tablePlus"
        :requestConfig="requestConfig"
        :columns="columns"
        v-model:pageConfig="pageConfig"
        :gridConfig="{ xs: 24, sm: 12, md: 12, lg: 8, xl: 6 }"
        :showPage="false"
        lazy
        :load="load"
        :handlerColumnConfig="{
          width: '160px',
          fixed: 'right',
          align: 'center',
        }"
      >
        <template #action>
          <div class="flex justify-between mb-4">
            <div class="center"></div>
            <div id="roleSearch"></div>
          </div>
        </template>
        <template #status="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'danger'">{{
            row.status === 1 ? t('system.user.status.enabled') : t('system.user.status.disabled')
          }}</el-tag>
        </template>
        <template #operation="{ row }">
          <div class="flex justify-center">
            <el-button
              v-auth="'mcpcan_rbac_manage:dept:update'"
              type="text"
              size="small"
              link
              class="base-btn-link"
              @click="handleEdit(row)"
            >
              {{ t('system.department.operation.edit') }}
            </el-button>
            <el-button
              v-auth="'mcpcan_rbac_manage:dept:delete'"
              type="danger"
              size="small"
              link
              @click="handleDelete(row)"
            >
              {{ t('system.department.operation.delete') }}
            </el-button>
          </div>
        </template>
      </TablePlus>
    </div>
    <!-- department form -->
    <FormDataDialog ref="formDataDialog" @on-refresh="handlePageRefresh"></FormDataDialog>
  </div>
</template>

<script setup lang="ts">
import { Plus } from '@element-plus/icons-vue'
import { useDeptTableHooks } from './hooks/index.ts'
import TablePlus from '@/components/TablePlus/index.vue'
import FormDataDialog from './modules/form-data-dialog.vue'
import { DeptAPI } from '@/api/system/index.ts'

const { t, pageInfo, requestConfig, columns, pageConfig, jumpToPage, reload } = useDeptTableHooks()
const formDataDialog = ref()
const tablePlus = ref()

/**
 * handle add a new role
 */
const handleNewRole = () => {
  formDataDialog.value.init()
}
// handle edit department
const handleEdit = (form: any) => {
  formDataDialog.value.init(form)
}

// handle delete department
const handleDelete = (row: any) => {
  ElMessageBox.confirm(t('system.department.desc.delete'), t('common.warn'), {
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
    await DeptAPI.delete(row.id)
    ElMessage({
      type: 'success',
      message: t('status.delete') + t('status.success'),
    })
    handlePageRefresh()
  })
}
// 异步加载树形结构部门信息
const load = async (row: any, treeNode: unknown, resolve: (data: any[]) => void) => {
  const { list } = await DeptAPI.list({
    parentId: row.id,
  })
  resolve(list || [])
}

// reload page data after create or edit department
const handlePageRefresh = () => {
  reload()
}

const init = () => {
  tablePlus.value?.initData()
}
onMounted(() => {
  init()
})
</script>

<style scoped lang="scss">
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
