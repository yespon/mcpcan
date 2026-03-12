<template>
  <div class="role-page" v-loading="pageInfo.loading">
    <div class="flex justify-between page-header">
      <div class="header-container">
        {{ t('system.role.title') }}
        <span class="desc">{{ t('system.role.pageDesc') }}</span>
      </div>
      <div v-auth="'mcpcan_rbac_manage:role:create'">
        <mcp-button :icon="Plus" @click="handleNewRole">{{ t('system.role.add') }}</mcp-button>
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
        :queryFormatter="handleFormatter"
        :gridConfig="{ xs: 24, sm: 12, md: 12, lg: 8, xl: 6 }"
        :handlerColumnConfig="{
          width: '260px',
          fixed: 'right',
          align: 'center',
        }"
      >
        <template #action>
          <div class="flex justify-between mb-4">
            <div class="center">
              <el-icon>
                <i class="icon iconfont MCP-jiaose"></i>
              </el-icon>
              <span class="desc">{{ t('system.role.total') }}：{{ pageConfig.total }}</span>
            </div>
            <div id="roleSearch"></div>
          </div>
        </template>
        <template #operation="{ row }">
          <div v-if="row.description !== 'admin'" class="flex justify-start">
            <el-button
              type="text"
              size="small"
              link
              class="base-btn-link"
              @click="handleViewUser(row)"
            >
              {{ t('system.role.operation.view') }}
            </el-button>
            <el-button
              type="text"
              size="small"
              link
              class="base-btn-link"
              v-auth="'mcpcan_rbac_manage:role:menus'"
              @click="handleAuth(row)"
            >
              {{ t('system.role.operation.auth') }}
            </el-button>
            <el-button
              type="text"
              size="small"
              link
              class="base-btn-link"
              v-auth="'mcpcan_rbac_manage:role:update'"
              @click="handleEditRole(row)"
            >
              {{ t('system.role.operation.edit') }}
            </el-button>
            <el-button
              v-auth="'mcpcan_rbac_manage:role:delete'"
              type="danger"
              size="small"
              link
              @click="handleDelete(row)"
            >
              {{ t('system.role.operation.delete') }}
            </el-button>
          </div>
        </template>
      </TablePlus>
    </div>
    <!-- 角色表单 -->
    <FormDataDialog ref="formDataDialog" @on-refresh="init"></FormDataDialog>
    <!-- 授权菜单 -->
    <MenuAuthDrawer ref="menuAuthDrawer" @on-refresh="init"></MenuAuthDrawer>
  </div>
</template>

<script setup lang="ts">
import { Plus } from '@element-plus/icons-vue'
import { useRoleTableHooks } from './hooks/index.ts'
import TablePlus from '@/components/TablePlus/index.vue'
import FormDataDialog from './modules/form-data-dialog.vue'
import MenuAuthDrawer from './modules/menu-auth-drawer.vue'

const { t, pageInfo, requestConfig, columns, pageConfig, jumpToPage, RoleAPI } = useRoleTableHooks()
const formDataDialog = ref()
const menuAuthDrawer = ref()
const tablePlus = ref()

/**
 * handle add a new role
 */
const handleNewRole = () => {
  formDataDialog.value.init()
}
// 参数格式化
const handleFormatter = (queryInfo: any) => {
  return {
    blurry: queryInfo.name,
  }
}
const handleViewUser = (roleInfo: any) => {
  jumpToPage({
    url: '/user-with-role',
    data: {
      roleId: roleInfo.id,
      roleName: roleInfo.name,
    },
  })
}

// handle delete role
const handleDelete = (roleInfo: any) => {
  ElMessageBox.confirm(t('system.role.desc.delete'), t('common.warn'), {
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
    await RoleAPI.delete(roleInfo.id)
    ElMessage({
      type: 'success',
      message: t('status.delete') + t('status.success'),
    })
    init()
  })
}

/**
 * handle edit role
 */
const handleEditRole = (roleInfo: any) => {
  formDataDialog.value.init(roleInfo)
}
/**
 * handle menu auth with role
 */
const handleAuth = (roleInfo: any) => {
  menuAuthDrawer.value.init(roleInfo)
}

const init = () => {
  tablePlus.value.initData()
}

onMounted(init)
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
