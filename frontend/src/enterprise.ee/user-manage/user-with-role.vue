<template>
  <div class="user-with-role">
    <div class="flex justify-between page-header">
      <div class="header-container">
        {{ t('system.role.currentRole') }}
        <span class="desc">{{ query.roleName }}</span>
      </div>
      <div v-auth="'mcpcan_rbac_manage:role:add_user'">
        <mcp-button :icon="Plus" @click="handleAddUser">{{
          t('system.user.operation.add')
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
        :handlerColumnConfig="{
          width: '80px',
          fixed: 'right',
          align: 'center',
        }"
      >
        <template #action>
          <div class="flex justify-between mb-4">
            <div></div>
            <div id="roleSearch"></div>
          </div>
        </template>
        <template #status="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'danger'">{{
            row.status === 1 ? t('system.user.status.enabled') : t('system.user.status.disabled')
          }}</el-tag>
        </template>
        <template #operation="{ row }">
          <el-button
            type="danger"
            link
            @click="handleRemove(row)"
            v-auth="'mcpcan_rbac_manage:role:remove_user'"
            v-if="!row.roleIds?.some((roleId: number) => roleId === 1)"
            >{{ t('common.delete') }}</el-button
          >
        </template>
      </TablePlus>
    </div>
    <AddUserDeptDialog ref="addUserDeptDialog" @on-refresh="init"></AddUserDeptDialog>
  </div>
</template>

<script setup lang="ts">
import { Plus } from '@element-plus/icons-vue'
import { useUserListRoleTableHooks } from './hooks/user-with-role.ts'
import TablePlus from '@/components/TablePlus/index.vue'
import AddUserDeptDialog from './modules/add-user-dept-dialog.vue'
import { UserAPI } from '@/api/system'

const { t, requestConfig, columns, pageConfig, query } = useUserListRoleTableHooks()
const addUserDeptDialog = ref()
const tablePlus = ref()

// 添加用户
const handleAddUser = () => {
  addUserDeptDialog.value.init()
}

// 移除用户
const handleRemove = (row: any) => {
  ElMessageBox.confirm(`${t('system.user.desc.remove')}"${row.username}"?`, t('common.warn'), {
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
    await UserAPI.removeUserRoles({ userIds: [row.id], roleId: Number(query.roleId) })
    ElMessage({
      type: 'success',
      message: t('status.delete') + t('status.success'),
    })
    init()
  })
}

const init = () => {
  tablePlus.value?.initData()
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
