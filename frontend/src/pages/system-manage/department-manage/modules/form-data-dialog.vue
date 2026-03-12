<template>
  <el-dialog
    v-model="dialogInfo.visible"
    :title="dialogInfo.title"
    :show-close="false"
    :close-on-click-modal="false"
    width="480px"
    top="20vh"
  >
    <template #header>
      <div class="flex items-center w-full">
        <span>{{ dialogInfo.title }}</span>
      </div>
    </template>
    <div v-loading="dialogInfo.loading" :element-loading-text="dialogInfo.loadingText">
      <el-form ref="formRef" :model="formModel" :rules="rules" label-width="100px">
        <el-form-item :label="t('system.department.formData.name')" prop="name">
          <el-input
            v-model="formModel.name"
            :placeholder="t('system.department.placeholder.name')"
            maxlength="20"
            show-word-limit
          />
        </el-form-item>
        <el-form-item :label="t('system.department.formData.imageUrl')" prop="imageUrl">
          <Upload v-model="formModel.imageUrl"></Upload>
        </el-form-item>
        <el-form-item :label="t('system.department.formData.deptSort')" prop="sort">
          <el-input-number v-model="formModel.sort" :min="1" />
        </el-form-item>
        <el-form-item :label="t('system.department.formData.pid')" prop="parentId">
          <el-tree-select
            v-model="formModel.parentId"
            :data="deptTree"
            node-key="id"
            filterable
            check-strictly
            :render-after-expand="false"
            :placeholder="t('system.department.placeholder.pid')"
          >
          </el-tree-select>
        </el-form-item>
        <el-form-item :label="t('system.department.formData.status')" prop="status">
          <el-radio-group v-model="formModel.status">
            <el-radio :value="1">{{ t('system.user.status.enabled') }}</el-radio>
            <el-radio :value="2">{{ t('system.user.status.disabled') }}</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
    </div>

    <template #footer>
      <div class="center">
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
import Upload from '@/components/upload/index.vue'
import { DeptAPI } from '@/api/system/index.ts'

const { t, locale } = useI18n()
const dialogInfo = ref<{
  title: string
  visible: boolean
  loading: boolean
  loadingText: string
}>({
  title: t('system.department.formData.title'),
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
  id?: string
  name: string
  imageUrl: string
  sort: number
  parentId: string | number | null
  status: number
}>({
  name: '',
  imageUrl: '',
  sort: 1,
  parentId: 0,
  status: 1,
})
const deptTree = ref<any>([])
const formRef = ref<FormInstance>()
// validation rules
const rules: FormRules<typeof formModel.value> = {
  name: [
    { required: true, message: t('system.department.placeholder.name'), trigger: 'blur' },
    { max: 20, message: t('system.department.placeholder.nameMaxLength'), trigger: 'blur' },
  ],
  sort: [{ required: true, message: t('system.department.placeholder.deptSort'), trigger: 'blur' }],
  parentId: [
    { required: true, message: t('system.department.placeholder.pid'), trigger: 'change' },
  ],
}

// cancel dialog
const handleCancel = () => {
  dialogInfo.value.visible = false
}
// handle Submit deptData
const handleSubmit = () => {
  formRef.value?.validate(async (valid) => {
    if (valid) {
      dialogInfo.value.loading = true
      const params = { ...formModel.value, parentId: formModel.value.parentId || null }
      try {
        await (formModel.value.id ? DeptAPI.edit(params) : DeptAPI.create(params))
        dialogInfo.value.visible = false
        emit('on-refresh')
        ElMessage({
          type: 'success',
          message: formModel.value.id ? t('action.edit') : t('action.create'),
        })
      } catch (error) {
        console.error('Error creating department:', error)
      } finally {
        dialogInfo.value.loading = false
      }
    }
  })
}

const handleGetDeptTree = async (id?: string) => {
  const data = await DeptAPI.deptTree()
  const traverse = (list: any[]) => {
    list.forEach((item) => {
      item.label = item.name
      item.disabled = item.id === id
      if (item.children?.length) {
        traverse(item.children)
      }
    })
  }
  traverse(data || [])
  if (data) {
    deptTree.value = [{ id: 0, label: t('system.user.desc.topDept') }, ...data]
  } else {
    deptTree.value = [{ id: 0, label: t('system.user.desc.topDept') }]
  }
}

/**
 * Init and open dialog
 */
const init = (row: any) => {
  if (row) {
    formModel.value = {
      id: row.id,
      name: row.name,
      sort: row.sort,
      parentId: row.parentId || 0,
      imageUrl: row.imageUrl,
      status: row.status,
    }
  } else {
    formModel.value = {
      name: '',
      sort: 1,
      parentId: 0,
      imageUrl: '',
      status: 1,
    }
  }

  dialogInfo.value.title = row?.id
    ? t('system.department.formData.edit')
    : t('system.department.formData.title')
  dialogInfo.value.visible = true
  handleGetDeptTree(row?.id)
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
