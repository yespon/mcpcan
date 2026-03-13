<template>
  <div>
    <el-dialog
      v-model="dialogInfo.visible"
      :title="dialogInfo.title"
      :show-close="true"
      width="1000px"
      top="10vh"
      header-class="user-header-border"
      footer-class="user-footer-border"
      destroy-on-close
    >
      <div
        class="dialog-content flex gap-4 my-4"
        style="height: 60vh"
        v-loading="dialogInfo.loading"
      >
        <!-- 左侧：部门选择 -->
        <div
          class="left-panel flex-1 flex flex-col border border-solid border-[var(--el-border-color-light)] rounded-md p-3"
        >
          <div class="panel-header mb-3 flex items-center text-[var(--el-text-color-regular)]">
            <span class="font-medium">{{ t('system.user.desc.all') }} ({{ deptTotal }})</span>
          </div>
          <el-input
            v-model="deptSearchText"
            :placeholder="t('system.user.placeholder.deptSearch')"
            :prefix-icon="Search"
            clearable
            class="mb-3"
          />
          <div class="flex-1 overflow-hidden">
            <el-scrollbar>
              <el-tree
                ref="deptTreeRef"
                :data="deptTreeData"
                show-checkbox
                node-key="id"
                default-expand-all
                :filter-node-method="filterNode"
                :props="{
                  children: 'children',
                  label: 'name',
                }"
                @check="handleDeptCheck"
              />
            </el-scrollbar>
          </div>
        </div>

        <!-- 右侧：人员选择 -->
        <div
          class="right-panel flex-[2] flex flex-col border border-solid border-[var(--el-border-color-light)] rounded-md p-3"
        >
          <div class="panel-header mb-3 flex items-center justify-between">
            <span class="text-[var(--el-text-color-regular)]">
              {{ t('system.user.desc.select') }} ({{ selectedUsers.length }})
            </span>
            <el-button type="primary" link @click="clearSelection">{{
              t('system.user.desc.clear')
            }}</el-button>
          </div>
          <el-input
            v-model="userSearchText"
            :placeholder="t('system.user.placeholder.blurry')"
            :suffix-icon="Search"
            clearable
            class="mb-3"
            @keyup.enter="handleUserSearch"
          />

          <!-- 表头 -->
          <div
            class="user-list-header grid grid-cols-[40px_1fr_1fr_1fr_80px] gap-2 px-2 py-2 bg-[var(--el-fill-color-light)] text-[var(--el-text-color-secondary)] text-sm rounded-t-md"
          >
            <div class="flex items-center justify-center">
              <el-checkbox
                v-model="isAllSelected"
                :indeterminate="isIndeterminate"
                @change="handleSelectAllChange"
                :disabled="userList.length === 0"
              />
            </div>
            <div>{{ t('system.user.columns.nickName') }}</div>
            <div>{{ t('system.user.columns.username') }}</div>
            <div>{{ t('system.user.columns.deptName') }}</div>
            <div>{{ t('system.user.columns.enabled') }}</div>
          </div>

          <!-- 用户列表 (无限滚动) -->
          <div
            class="flex-1 user-list-body overflow-auto border-x border-b border-[var(--el-border-color-light)] rounded-b-md"
          >
            <ul
              v-infinite-scroll="loadMoreUsers"
              :infinite-scroll-disabled="loading || noMore"
              class="m-0 p-0 list-none"
            >
              <li
                v-for="user in userList"
                :key="user.id"
                class="user-item grid grid-cols-[40px_1fr_1fr_1fr_80px] gap-2 px-2 py-3 border-b border-[var(--el-border-color-lighter)] items-center hover:bg-[var(--el-fill-color-lighter)] transition-colors text-sm text-[var(--el-text-color-regular)]"
              >
                <div class="flex items-center justify-center">
                  <el-checkbox
                    v-model="user.selected"
                    @change="(val: string | number | boolean) => handleUserSelect(val, user)"
                  />
                </div>
                <div class="truncate" :title="user.nickName">{{ user.nickName }}</div>
                <div class="truncate" :title="user.username">{{ user.username }}</div>
                <div class="truncate" :title="user.departmentName">{{ user.departmentName }}</div>
                <div>
                  <el-tag v-if="user.status === 1" type="success" size="small" effect="plain">
                    {{ t('system.user.status.enabled') }}
                  </el-tag>
                  <el-tag v-else type="danger" size="small" effect="plain">
                    {{ t('system.user.status.disabled') }}
                  </el-tag>
                </div>
              </li>
              <li v-if="loading" class="py-2 text-center text-[var(--el-text-color-secondary)]">
                {{ t('system.user.status.loading') }}
              </li>
              <li
                v-if="noMore && userList.length > 0"
                class="py-2 text-center text-[var(--el-text-color-secondary)]"
              >
                {{ t('system.user.status.noMore') }}
              </li>
              <li
                v-if="userList.length === 0 && !loading"
                class="py-10 text-center text-[var(--el-text-color-secondary)]"
              >
                <el-empty :image-size="60" />
              </li>
            </ul>
          </div>
        </div>
      </div>

      <template #footer>
        <div class="flex justify-center gap-4">
          <el-button @click="dialogInfo.visible = false">{{ t('common.cancel') }}</el-button>
          <mcp-button type="primary" @click="handleConfirm">{{ t('common.confirm') }}</mcp-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, watch, computed } from 'vue'
import { Search } from '@element-plus/icons-vue'
import { useI18n } from 'vue-i18n'
import type { ElTree } from 'element-plus'
import { DeptAPI, UserAPI } from '@/api/system/index.ts'

const { t } = useI18n()
// --- 接口定义 (预留) ---
interface DeptNode {
  id: string
  name?: string
  label?: string
  children?: DeptNode[]
}

interface UserItem {
  id: string
  username: string
  nickName: string
  departmentName: string
  status: number // 1: 启用, 0: 禁用
  selected?: boolean
  roleIds?: (string | number)[]
}

const emit = defineEmits<{
  (e: 'confirm', payload: { users: UserItem[]; deptIds: string[] }): void
  (e: 'on-refresh'): void
}>()
const { query } = useRoute()
// --- Dialog State ---
const dialogInfo = reactive({
  visible: false,
  loading: false,
  title: t('system.user.operation.add'),
})

// --- Left Panel: Dept Tree ---
const deptSearchText = ref('')
const deptTreeData = ref<DeptNode[]>([])
const deptTreeRef = ref<InstanceType<typeof ElTree>>()
const deptTotal = ref(0)
const selectedDeptIds = ref<string[]>([])

watch(deptSearchText, (val) => {
  deptTreeRef.value!.filter(val)
})

const filterNode = (value: string, data: DeptNode) => {
  if (!value) return true
  const label = (data as any)?.label ?? (data as any)?.name ?? ''
  return String(label).includes(value)
}

const handleDeptCheck = (data: DeptNode, checkedState: any) => {
  // 使用勾选的部门ID筛选右侧用户
  const checkedKeys: string[] = checkedState?.checkedKeys || []
  selectedDeptIds.value = checkedKeys

  // 重置用户列表并重新加载
  resetUserListAndLoad()
}

// --- Right Panel: User List ---
const userSearchText = ref('')
const userList = ref<UserItem[]>([])
const selectedUsers = ref<UserItem[]>([])
const loading = ref(false)
const noMore = ref(false)

const pageConfig = reactive({
  page: 1,
  pageSize: 20,
})

const isAllSelected = computed({
  get: () => userList.value.length > 0 && userList.value.every((u) => u.selected),
  set: (val) => {
    handleSelectAllChange(val)
  },
})

const isIndeterminate = computed(() => {
  const selectedCount = userList.value.filter((u) => u.selected).length
  return selectedCount > 0 && selectedCount < userList.value.length
})

/** Load more users for infinite scroll */
const loadMoreUsers = async () => {
  if (loading.value || noMore.value) return
  loading.value = true
  try {
    await handleGetUserList()
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

const resetUserListAndLoad = () => {
  pageConfig.page = 1
  userList.value = []
  noMore.value = false
  loadMoreUsers()
}

/** Handle user search */
const handleUserSearch = () => {
  resetUserListAndLoad()
}

/** Select all change handler */
const handleSelectAllChange = (val: string | number | boolean) => {
  const checked = val as boolean
  userList.value.forEach((user) => {
    user.selected = checked
    handleUserSelect(checked, user)
  })
}

/**
 * Handle user selection
 * @param val - Selection state
 * @param user - User item
 */
const handleUserSelect = (val: string | number | boolean, user: UserItem) => {
  const checked = val as boolean
  if (checked) {
    if (!selectedUsers.value.some((u) => u.id === user.id)) {
      selectedUsers.value.push(user)
    }
  } else {
    const index = selectedUsers.value.findIndex((u) => u.id === user.id)
    if (index > -1) {
      selectedUsers.value.splice(index, 1)
    }
  }
}

const clearSelection = () => {
  selectedUsers.value = []
  userList.value.forEach((u) => (u.selected = false))
}

/**
 * handle Get Dept Tree
 */
const handleGetDeptTree = async () => {
  try {
    dialogInfo.loading = true
    const raw = await DeptAPI.deptTree()
    // 兼容两种返回：直接 data[] 或 { data: [] }
    const data = Array.isArray(raw) ? raw : raw?.data
    const traverse = (list: any[]) => {
      list.forEach((item) => {
        item.label = item.name
        if (item.children?.length) traverse(item.children)
      })
    }

    traverse(data || [])
    deptTreeData.value = data || []
    deptTotal.value = Array.isArray(data) ? data.length : 0
  } finally {
    dialogInfo.loading = false
  }
}
/**
 * handle Get User List
 */
const handleGetUserList = async () => {
  const { list, total } = await UserAPI.list({
    page: pageConfig.page,
    pageSize: pageConfig.pageSize,
    blurry: userSearchText.value,
    deptIds: selectedDeptIds.value,
  })

  const next = (list || []).map((u: any) => ({
    ...u,
    // 回显勾选状态
    selected: selectedUsers.value.some((s) => s.id === u.id),
  }))

  if (pageConfig.page === 1) userList.value = []
  userList.value.push(...next)
  // 分页推进 & 结束判断
  if (!next.length || next.length < Number(pageConfig.pageSize || 20)) {
    noMore.value = true
  } else {
    pageConfig.page++
  }
}

/**
 * handle Confirm select
 */
const handleConfirm = async () => {
  try {
    dialogInfo.loading = true
    await UserAPI.addUserRoles({
      userIds: selectedUsers.value
        .filter((u) => !u.roleIds?.find((id) => id === Number(query.roleId)))
        .map((u) => u.id),
      roleId: Number(query.roleId),
    })
    dialogInfo.visible = false
    emit('on-refresh')
  } finally {
    dialogInfo.loading = false
  }
}

// --- Init ---
const init = async () => {
  dialogInfo.visible = true
  dialogInfo.title = t('system.user.operation.add')
  // 重置状态
  deptSearchText.value = ''
  userSearchText.value = ''
  userList.value = []
  selectedUsers.value = []
  pageConfig.page = 1
  noMore.value = false
  selectedDeptIds.value = []
  // 加载数据
  await handleGetDeptTree()
  // 初始加载用户
  loadMoreUsers()
}
defineExpose({
  init,
})
</script>
<style lang="scss" scoped>
:deep(.user-header-border) {
  background-color: transparent !important;
  border-bottom: 1px solid var(--el-border-color-light) !important;
  margin-right: 0 !important;
  padding-bottom: 15px;
}
:deep(.user-footer-border) {
  background-color: transparent !important;
  border-top: 1px solid var(--el-border-color-light) !important;
  padding-top: 15px;
}
/* 自定义滚动条样式(可选) */
:deep(.el-scrollbar__wrap) {
  overflow-x: hidden;
}
.user-item:last-child {
  border-bottom: none;
}
</style>
