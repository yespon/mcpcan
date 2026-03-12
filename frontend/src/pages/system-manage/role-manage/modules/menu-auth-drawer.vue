<template>
  <el-drawer
    v-model="drawerInfo.visible"
    :title="t('system.role.desc.authTitle') + ':' + roleInfo?.name"
    :with-header="true"
    destroy-on-close
  >
    <div v-loading="drawerInfo.loading">
      <el-tree
        style="height: 450px"
        :data="menuNodeList"
        show-checkbox
        node-key="id"
        :default-expand-all="false"
        :default-checked-keys="defaultCheckedKeys"
        :props="{
          children: 'children',
          label: locale === 'en' ? 'engTitle' : 'title',
        }"
        ref="apiTreeRef"
      >
        <template #default="{ node, data }">
          <div class="custom-tree-node">
            <el-tag :type="['primary', 'success', 'warning'][data.type]">{{
              [t('system.auth.directory'), t('system.auth.menu'), t('system.auth.button')][
                data.type
              ]
            }}</el-tag>
            <span class="ml-2">{{ node.label }}</span>
          </div>
        </template>
      </el-tree>
    </div>
    <template #footer>
      <div class="center">
        <mcp-button
          v-auth="'mcpcan_rbac_manage:role:save_menus'"
          type="primary"
          class="w-25 mr-2"
          @click="handleConfirm"
          >{{ t('common.save') }}</mcp-button
        >
        <el-button @click="drawerInfo.visible = false" class="w-25">{{
          t('common.cancel')
        }}</el-button>
      </div>
    </template>
  </el-drawer>
</template>

<script setup lang="ts">
import { RoleAPI } from '@/api/system/index'

const { t, locale } = useI18n()
const drawerInfo = ref({
  loading: false,
  visible: false,
})
const apiTreeRef = ref()
const roleInfo = ref<any>({})
const menuNodeList = ref<any[]>([])
const defaultCheckedKeys = ref<any[]>([])
// submit
const emit = defineEmits<{
  (e: 'on-refresh'): void
}>()

// 基于完整菜单树(menuNodeList)构建 id -> hasChildren 映射
const buildHasChildrenMap = (list: any[]) => {
  const map = new Map<any, boolean>()
  const traverse = (nodes: any[]) => {
    if (!Array.isArray(nodes) || nodes.length === 0) return
    for (const n of nodes) {
      const hasChildren = Array.isArray(n?.children) && n.children.length > 0
      map.set(n?.id, hasChildren)
      if (hasChildren) traverse(n.children)
    }
  }
  traverse(list)
  return map
}
// handle get all menus
const handleGetAllMenus = async () => {
  const data = await RoleAPI.getAllMenus()
  menuNodeList.value = data || []
}
// handle get authorized menus
const handleGetAuthMenus = async () => {
  const { menus, permissions } = await RoleAPI.getRoleMenus({ roleIds: [roleInfo.value.id] })
  const permissionSet = new Set<string>(permissions || [])
  const hasChildrenMap = buildHasChildrenMap(menuNodeList.value)
  const checked: any[] = []
  const traverse = (list: any[]) => {
    if (!Array.isArray(list) || list.length === 0) return
    for (const item of list) {
      // 仅勾选叶子节点：是否叶子以完整树(menuNodeList)为准，而非授权树(menus)
      const hasChildrenInAll = !!hasChildrenMap.get(item?.id)
      if (!hasChildrenInAll && item?.permission && permissionSet.has(item.permission)) {
        checked.push(item.id)
      }
      if (item?.children?.length) {
        traverse(item.children)
      }
    }
  }
  traverse(menus || [])
  defaultCheckedKeys.value = checked
}

// confirm authorization
const handleConfirm = async () => {
  try {
    drawerInfo.value.loading = true
    await RoleAPI.saveRoleMenus({
      roleId: roleInfo.value.id,
      menuIds: [...apiTreeRef.value.getHalfCheckedKeys(), ...apiTreeRef.value.getCheckedKeys()],
    })
    ElMessage.success(t('action.auth'))
    emit('on-refresh')
    drawerInfo.value.visible = false
  } finally {
    drawerInfo.value.loading = false
  }
}

// init drawer
const init = async (data: any) => {
  drawerInfo.value.visible = true
  roleInfo.value = data
  defaultCheckedKeys.value = []
  try {
    drawerInfo.value.loading = true
    // 先拉取完整菜单树，再根据完整树判定哪些节点是叶子节点
    await handleGetAllMenus()
    await handleGetAuthMenus()
  } finally {
    drawerInfo.value.loading = false
  }
}

defineExpose({
  init,
})
</script>

<style lang="scss" scoped>
:deep(.el-tree .el-tree-node__content) {
  height: 33px;
}
</style>
