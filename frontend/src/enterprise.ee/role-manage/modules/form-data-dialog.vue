<template>
  <el-dialog
    v-model="dialogInfo.visible"
    :title="dialogInfo.title"
    :show-close="false"
    :close-on-click-modal="false"
    width="660px"
    top="20vh"
  >
    <template #header>
      <div class="flex items-center w-full">
        <span>{{ dialogInfo.title }}</span>
      </div>
    </template>
    <div v-loading="dialogInfo.loading" :element-loading-text="dialogInfo.loadingText">
      <el-form ref="formRef" :model="formModel" :rules="rules" label-width="100px">
        <el-form-item :label="t('system.role.columns.name')" prop="name">
          <el-input
            v-model="formModel.name"
            :placeholder="t('system.role.placeholder.name')"
            maxlength="30"
            show-word-limit
          />
        </el-form-item>
        <el-form-item :label="t('system.role.columns.level')" prop="level">
          <el-input-number v-model="formModel.level" :min="1" />
        </el-form-item>
        <!-- <el-form-item :label="t('system.role.columns.dataScope')" prop="dataScope">
          全部
        </el-form-item> -->
        <el-form-item :label="t('system.role.columns.description')" prop="description">
          <el-input
            v-model="formModel.description"
            :rows="3"
            type="textarea"
            maxlength="100"
            show-word-limit
            :placeholder="t('system.role.placeholder.description')"
          />
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
import McpButton from '@/components/mcp-button/index.vue'
import { RoleAPI } from '@/api/system'

const { t, locale } = useI18n()
const dialogInfo = ref<{
  title: string
  visible: boolean
  loading: boolean
  loadingText: string
}>({
  title: t('system.role.add'),
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
  name: string
  level: number
  description: string
  [key: string]: any
}>({
  name: '',
  level: 1,
  description: '',
  dataScope: 'all',
})

const formRef = ref<FormInstance>()
// validation rules
const rules: FormRules<typeof formModel.value> = {
  name: [
    { required: true, message: t('system.role.placeholder.name'), trigger: 'blur' },
    { max: 30, message: t('system.role.placeholder.nameMaxLength'), trigger: 'blur' },
  ],
  level: [{ required: true, message: t('system.role.placeholder.level'), trigger: 'blur' }],
}

// cancel dialog
const handleCancel = () => {
  dialogInfo.value.visible = false
}
const handleSubmit = () => {
  formRef.value?.validate(async (valid: boolean) => {
    if (valid) {
      const { data } = await (formModel.value.id
        ? RoleAPI.edit(formModel.value)
        : RoleAPI.create(formModel.value))
      emit('on-refresh')
      dialogInfo.value.visible = false
      ElMessage({
        type: 'success',
        message: formModel.value.id ? t('action.edit') : t('action.create'),
      })
    }
  })
}
/**
 * Init and open dialog
 */
const init = (roleInfo: any | null) => {
  formModel.value = roleInfo || {
    id: undefined,
    name: '',
    level: 1,
    description: '',
    dataScope: 'all',
  }
  dialogInfo.value.title = roleInfo ? t('system.role.edit') : t('system.role.add')
  dialogInfo.value.visible = true
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
