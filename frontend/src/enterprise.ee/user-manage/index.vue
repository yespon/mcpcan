<template>
  <div class="h-full">
    <el-splitter>
      <el-splitter-panel size="300px" class="pr-4" :resizable="false">
        <el-card class="h-full">
          <div class="font-bold mb-4">{{ t('system.user.subTitle') }}</div>
          <el-input
            v-model="deptSearchText"
            :placeholder="t('system.user.placeholder.deptSearch')"
            clearable
            class="mb-3"
          />
          <el-tree
            style="height: 75vh"
            :data="deptTree"
            node-key="id"
            :default-expand-all="false"
            :default-checked-keys="[]"
            :highlight-current="true"
            ref="apiTreeRef"
            :props="{
              children: 'children',
              label: 'label',
            }"
            :filter-node-method="filterNode"
            @current-change="handleDeptChange"
          />
        </el-card>
      </el-splitter-panel>
      <el-splitter-panel class="pl-4">
        <el-card class="h-full">
          <div class="flex justify-between items-center page-header mb-4">
            <div class="header-container flex-1 mr-2 font-bold">
              {{ t('system.user.title') }}
              <span class="desc font-normal">{{ t('system.user.pageDesc') }}</span>
            </div>
            <mcp-button
              v-auth="'mcpcan_rbac_manage:user:create'"
              :icon="Plus"
              @click="handleCreateUser"
              >{{ t('system.user.operation.new') }}</mcp-button
            >
          </div>
          <TablePlus
            :showOperation="true"
            searchContainer="#roleSearch"
            ref="tablePlus"
            :requestConfig="requestConfig"
            :columns="columns"
            v-model:pageConfig="pageConfig"
            :queryFormatter="handleFormatter"
            :handlerColumnConfig="{
              width: '160px',
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
                row.status === 1
                  ? t('system.user.status.enabled')
                  : t('system.user.status.disabled')
              }}</el-tag>
            </template>
            <template #operation="{ row }">
              <template v-if="!row.roleIds?.some((roleId: number) => roleId === 1)">
                <el-button
                  type="text"
                  size="small"
                  link
                  class="base-btn-link"
                  v-auth="'mcpcan_rbac_manage:user:update'"
                  @click="handleEditUser(row)"
                >
                  {{ t('system.user.operation.edit') }}
                </el-button>
                <el-button
                  v-auth="'mcpcan_rbac_manage:user:delete'"
                  type="danger"
                  size="small"
                  link
                  @click="handleDelete(row)"
                >
                  {{ t('system.user.operation.delete') }}
                </el-button>
              </template>
            </template>
          </TablePlus>
        </el-card>
      </el-splitter-panel>
    </el-splitter>
    <FormDataDialog ref="formDataDialog" @on-refresh="init"></FormDataDialog>
  </div>
</template>

<script setup lang="ts">
import { Plus } from '@element-plus/icons-vue'
import { useUserListTableHooks } from './hooks/index.ts'
import TablePlus from '@/components/TablePlus/index.vue'
import FormDataDialog from './modules/form-data-dialog.vue'
import { UserAPI, DeptAPI } from '@/api/system/index.ts'

const formDataDialog = ref()
const tablePlus = ref()
const { t, pageInfo, requestConfig, jumpToPage, columns, pageConfig } = useUserListTableHooks()
const deptTree = ref<any[]>([])
const deptSearchText = ref('')
const apiTreeRef = ref<any>()
const currentTreeNode = ref<any>(null)

watch(deptSearchText, (val) => {
  apiTreeRef.value?.filter?.(val)
})

const filterNode = (value: string, data: any) => {
  if (!value) return true
  const label = (data?.label ?? data?.name ?? '') as string
  return String(label).toLowerCase().includes(String(value).toLowerCase())
}
// 参数格式化
const handleFormatter = (queryInfo: any) => {
  return {
    blurry: queryInfo.username,
  }
}
// handle create a new User
const handleCreateUser = () => {
  formDataDialog.value?.init(null, currentTreeNode.value?.id)
}
// handle edit user
const handleEditUser = (userInfo: any) => {
  formDataDialog.value?.init(userInfo, '')
}
// handle delete User
const handleDelete = (row: any) => {
  ElMessageBox.confirm(t('system.user.desc.delete'), t('common.warn'), {
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
    await UserAPI.delete(row.id)
    ElMessage({
      type: 'success',
      message: t('status.delete') + t('status.success'),
    })
    init()
  })
}
// handle get dept tree
const handleGetDeptTree = async () => {
  const data = (await DeptAPI.deptTree()) as any[]
  const traverse = (list: any[]) => {
    list.forEach((item) => {
      item.label = item.name
      if (item.children?.length) {
        traverse(item.children)
      }
    })
  }
  const list = Array.isArray(data) ? data : []
  traverse(list)
  deptTree.value = [{ id: '', label: t('system.user.desc.all') }, ...list]
}

const handleDeptChange = (currentNode: any) => {
  currentTreeNode.value = currentNode
  // 如果选中部门存在 children，则将其子孙节点 id 拍平后传给 deptIds
  // 约定：id 为空字符串代表“全部”，此时不传 deptIds 以返回全量数据
  const flattenDeptIds = (node: any): any[] => {
    if (!node) return []
    if (!node.children || node.children.length === 0) {
      return node.id ? [node.id] : []
    }
    return node.children.reduce(
      (ids: any[], child: any) => {
        return ids.concat(flattenDeptIds(child))
      },
      node.id ? [node.id] : [],
    )
  }
  if (!currentNode?.id) {
    requestConfig.value.searchQuery.model.deptIds = []
  } else {
    requestConfig.value.searchQuery.model.deptIds = flattenDeptIds(currentNode)
  }
  tablePlus.value.initData()
}
/**
 * handle init table data
 */
const init = () => {
  handleGetDeptTree()
  tablePlus.value.initData()
}

onMounted(init)
</script>

<style lang="scss" scoped>
.header-container {
  font-size: 20px;
}
.desc {
  font-size: 16px;
  color: #999999;
  margin-left: 16px;
}
:deep(.el-tree .el-tree-node__content) {
  height: 33px;
}
:deep(.el-tree--highlight-current .el-tree-node.is-current > .el-tree-node__content) {
  background-color: var(--ep-bg-purple-color-deep);
  border-radius: 4px;
}
</style>
