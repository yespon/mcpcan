<template>
  <el-dialog
    v-model="dialogInfo.visible"
    :title="dialogInfo.title"
    :show-close="false"
    :close-on-click-modal="false"
    width="480px"
    top="15vh"
  >
    <template #header>
      <div class="flex items-center w-full">
        <span>{{ dialogInfo.title }}</span>
      </div>
    </template>
    <div v-loading="dialogInfo.loading" :element-loading-text="dialogInfo.loadingText">
      <el-form ref="formRef" :model="formModel" :rules="rules" label-width="100px">
        <el-form-item :label="t('system.user.columns.username')" prop="username">
          <el-input
            v-model="formModel.username"
            :placeholder="t('system.user.placeholder.userName')"
            maxlength="30"
            show-word-limit
            :disabled="!!formModel.id"
          />
        </el-form-item>
        <el-form-item :label="t('system.user.columns.phone')" prop="phone">
          <el-input
            v-model="formModel.phone"
            :placeholder="t('system.user.placeholder.phone')"
            maxlength="11"
            show-word-limit
          />
        </el-form-item>
        <el-form-item :label="t('system.user.columns.nickName')" prop="nickName">
          <el-input
            v-model="formModel.nickName"
            :placeholder="t('system.user.placeholder.nickName')"
            maxlength="30"
            show-word-limit
          />
        </el-form-item>
        <el-form-item :label="t('system.user.columns.email')" prop="email">
          <el-input v-model="formModel.email" :placeholder="t('system.user.placeholder.email')" />
        </el-form-item>
        <el-form-item :label="t('system.user.columns.department')" prop="deptId">
          <el-tree-select
            v-model="formModel.deptId"
            :data="deptTree"
            node-key="id"
            filterable
            check-strictly
            :render-after-expand="false"
            :placeholder="t('system.user.placeholder.department')"
          />
        </el-form-item>
        <!-- <el-form-item :label="t('system.user.columns.gender')" prop="gender">
          <el-radio-group v-model="formModel.gender">
            <el-radio :value="1">{{ t('system.user.gender.male') }}</el-radio>
            <el-radio :value="0">{{ t('system.user.gender.female') }}</el-radio>
          </el-radio-group>
        </el-form-item> -->
        <el-form-item :label="t('system.user.columns.enabled')" prop="status">
          <el-radio-group v-model="formModel.status">
            <el-radio :value="1">{{ t('system.user.status.enabled') }}</el-radio>
            <el-radio :value="2">{{ t('system.user.status.disabled') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="t('system.user.columns.roles')" prop="roleIds">
          <el-select
            v-model="formModel.roleIds"
            multiple
            collapse-tags
            collapse-tags-tooltip
            :options="allRoleList"
            :props="{
              label: 'name',
              value: 'id',
            }"
            :placeholder="t('system.user.placeholder.roles')"
          >
          </el-select>
        </el-form-item>
      </el-form>
    </div>

    <template #footer>
      <div class="center mt-[100px]">
        <el-button @click="handleCancel" class="mr-2">{{ t('common.cancel') }}</el-button>
        <mcp-button @click="handleSubmit" :loading="dialogInfo.loading">{{
          t('common.save')
        }}</mcp-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import type { FormInstance, FormRules } from 'element-plus'
import { RoleAPI, DeptAPI, UserAPI } from '@/api/system/index.ts'

const { t, locale } = useI18n()
const dialogInfo = ref<{
  title: string
  visible: boolean
  loading: boolean
  loadingText: string
}>({
  title: t('system.user.operation.new'),
  visible: false,
  loading: false,
  loadingText: t('agent.formData.loadingText'),
})
// submit
const emit = defineEmits<{
  (e: 'submit', payload: typeof formModel.value): void
  (e: 'on-refresh'): void
}>()
// formData model
const formModel = ref<{
  id?: number | string
  username: string
  phone: string
  nickName: string
  email: string
  deptId: string | null
  // gender: number
  status: number
  roleIds: string[] | string
}>({
  username: '',
  phone: '',
  nickName: '',
  email: '',
  deptId: '',
  // gender: 1,
  status: 1,
  roleIds: '',
})
const deptTree = ref<any[]>([])
const allRoleList = ref<any[]>([])
const formRef = ref<FormInstance>()
// validation rules
const rules: FormRules<typeof formModel.value> = {
  username: [
    { required: true, message: t('system.user.placeholder.userName'), trigger: 'blur' },
    {
      validator: (_rule, value, callback) => {
        const v = String(value ?? '')
        // 禁止中文/中日韩统一表意符号（也覆盖大部分日文汉字）
        // 如需放开部分字符集，可在此调整正则。
        const hasCJK = /[\u3400-\u9FFF\uF900-\uFAFF]/.test(v)
        if (hasCJK) {
          callback(new Error(t('system.user.placeholder.usernameRule') as string))
          return
        }
        callback()
      },
      trigger: ['blur', 'change'],
    },
    { max: 30, message: t('system.user.placeholder.nameMaxLength'), trigger: 'blur' },
  ],
  nickName: [{ required: true, message: t('system.user.placeholder.nickName'), trigger: 'blur' }],
  email: [
    { required: true, message: t('system.user.placeholder.email'), trigger: 'blur' },
    {
      type: 'email',
      message: t('system.user.placeholder.emailInvalid'),
      trigger: ['blur', 'change'],
    },
  ],
  roleIds: [{ required: true, message: t('system.user.placeholder.roles'), trigger: 'blur' }],
}

// cancel dialog
const handleCancel = () => {
  dialogInfo.value.visible = false
}
const handleSubmit = () => {
  formRef.value?.validate(async (valid) => {
    if (valid) {
      try {
        dialogInfo.value.loading = true
        const { data } = await (formModel.value.id
          ? UserAPI.edit(formModel.value)
          : UserAPI.create(formModel.value))
        emit('on-refresh')
        dialogInfo.value.visible = false
        ElMessage({
          type: 'success',
          message: formModel.value.id ? t('action.edit') : t('action.create'),
        })
        if (!formModel.value.id) {
          ElMessageBox.confirm(t('system.user.desc.initPassword'), t('common.warn'), {
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
        }
      } finally {
        dialogInfo.value.loading = false
      }
    }
  })
}

// handle get dept tree
const handleGetDeptTree = async () => {
  const data = await DeptAPI.deptTree()
  const traverse = (list: any[]) => {
    list.forEach((item) => {
      item.label = item.name
      if (item.children?.length) {
        traverse(item.children)
      }
    })
  }
  traverse(data || [])
  deptTree.value = data || []
}

// handle get all role List
const handleGetRoleList = async () => {
  const { list } = await RoleAPI.allList()
  allRoleList.value = (list || []).map((item: any) => ({
    ...item,
    disabled: item.name === 'admin', // 禁止选择 admin 角色
  }))
}

/**
 * Init and open dialog
 */
const init = (userInfo?: any, deptId?: string) => {
  if (userInfo) {
    formModel.value = { ...userInfo, deptId: userInfo.deptId || null }
  } else {
    formModel.value = {
      username: '',
      phone: '',
      nickName: '',
      email: '',
      deptId: deptId || null,
      // gender: 1,
      status: 1,
      roleIds: '',
    }
  }
  dialogInfo.value.title = userInfo?.id
    ? t('system.user.operation.update')
    : t('system.user.operation.new')
  dialogInfo.value.visible = true
  handleGetDeptTree()
  handleGetRoleList()
}
defineExpose({
  init,
})
</script>

<style scoped>
.dialog-footer {
  display: inline-flex;
  gap: 8px;
}
</style>
