<template>
  <el-dialog
    v-model="dialogInfo.visible"
    :title="dialogInfo.title"
    :show-close="false"
    :close-on-click-modal="false"
    width="680px"
    top="20vh"
  >
    <template #header>
      <div class="flex items-center w-full">
        <span>{{ dialogInfo.title }}</span
        >:
        <McpImage :src="dify" fit="contain" width="80" height="20" />
        <span class="text-purple">{{
          formModel.accessType === 'Dify'
            ? t('agent.action.community')
            : t('agent.action.enterprise')
        }}</span>
      </div>
    </template>
    <div v-loading="dialogInfo.loading" :element-loading-text="t('agent.formData.loadingText')">
      <el-form ref="formRef" :model="formModel" :rules="rules" label-width="110px">
        <el-form-item :label="t('agent.formData.accessName')" prop="accessName">
          <el-input
            v-model="formModel.accessName"
            :placeholder="t('agent.placeholder.accessName')"
            maxlength="64"
            show-word-limit
          />
        </el-form-item>

        <el-form-item :label="t('agent.formData.dbHost')" prop="dbHost">
          <el-input v-model="formModel.dbHost" :placeholder="t('agent.placeholder.dbHost')" />
        </el-form-item>

        <el-form-item :label="t('agent.formData.dbPort')" prop="dbPort">
          <el-input-number
            v-model="formModel.dbPort"
            :min="1"
            :max="65535"
            :controls="false"
            :placeholder="t('agent.placeholder.dbPort')"
            style="width: 100%"
          />
        </el-form-item>

        <el-form-item :label="t('agent.formData.dbUser')" prop="dbUser">
          <el-input v-model="formModel.dbUser" :placeholder="t('agent.placeholder.dbUser')" />
        </el-form-item>

        <el-form-item :label="t('agent.formData.dbPassword')" prop="dbPassword">
          <el-input
            v-model="formModel.dbPassword"
            type="password"
            show-password
            :placeholder="t('agent.placeholder.dbPassword')"
          />
        </el-form-item>

        <el-form-item :label="t('agent.formData.dbName')" prop="dbName">
          <el-input v-model="formModel.dbName" :placeholder="t('agent.placeholder.dbName')" />
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
import { ref } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import McpButton from '@/components/mcp-button/index.vue'
import { AgentAPI } from '@/api/agent/index'
import { ElMessage, ElMessageBox } from 'element-plus'
import McpImage from '@/components/mcp-image/index.vue'
import { dify } from '@/utils/logo.ts'

const { t } = useI18n()
const dialogInfo = ref<{
  title: string
  visible: boolean
  loading: boolean
}>({
  title: t('agent.pageDesc.formTitle'),
  visible: false,
  loading: false,
})

// formData model
const formModel = ref<{
  accessID: string
  accessName: string
  accessType: string
  dbHost: string
  dbPort: number
  dbUser: string
  dbPassword: string
  dbName: string
}>({
  accessID: '',
  accessName: '',
  accessType: '',
  dbHost: '',
  dbPort: 5432,
  dbUser: '',
  dbPassword: '',
  dbName: '',
})

const formRef = ref<FormInstance>()

// validation rules
const rules: FormRules<typeof formModel.value> = {
  accessName: [
    { required: true, message: t('agent.formData.accessName'), trigger: 'blur' },
    { min: 2, max: 64, message: t('agent.formData.accessName'), trigger: 'blur' },
  ],
  accessType: [{ required: true, message: t('agent.formData.accessType'), trigger: 'change' }],
  dbHost: [{ required: true, message: t('agent.formData.dbHost'), trigger: 'blur' }],
  dbPort: [
    {
      required: true,
      validator: (_rule, value, callback) => {
        if (!value || value < 1 || value > 65535)
          return callback(new Error(t('agent.formData.dbPortLength')))
        callback()
      },
      trigger: 'change',
    },
  ],
  dbUser: [{ required: true, message: t('agent.formData.dbUser'), trigger: 'blur' }],
  dbPassword: [{ required: true, message: t('agent.formData.dbPassword'), trigger: 'blur' }],
  dbName: [{ required: true, message: t('agent.formData.dbName'), trigger: 'blur' }],
}

// cancel dialog
const handleCancel = () => {
  dialogInfo.value.visible = false
}

// submit
const emit = defineEmits<{
  (e: 'submit', payload: typeof formModel.value): void
  (e: 'on-refresh'): void
}>()

// handle form submit
const handleSubmit = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    try {
      dialogInfo.value.loading = true
      const params = {
        accessName: formModel.value.accessName,
        accessType: formModel.value.accessType,
        dbHost: formModel.value.dbHost,
        dbPort: formModel.value.dbPort,
        dbUser: formModel.value.dbUser,
        dbPassword: formModel.value.dbPassword,
        dbName: formModel.value.dbName,
      }
      await AgentAPI.connectionTest(params)
        .then(async () => {
          await (formModel.value.accessID
            ? AgentAPI.update({ accessID: formModel.value.accessID, ...params })
            : AgentAPI.create(params))
          ElMessage.success(formModel.value.accessID ? t('action.update') : t('action.create'))
          emit('on-refresh')
          dialogInfo.value.visible = false
          formRef.value?.resetFields()
        })
        .catch(async () => {
          const result = await ElMessageBox.confirm(
            t('agent.action.failConnectionTips'),
            t('common.warn'),
            {
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
            },
          )
          if (result === 'confirm') {
            await (formModel.value.accessID
              ? AgentAPI.update({ accessID: formModel.value.accessID, ...params })
              : AgentAPI.create(params))
            ElMessage.success(formModel.value.accessID ? t('action.update') : t('action.create'))
            emit('on-refresh')
            dialogInfo.value.visible = false
            formRef.value?.resetFields()
          }
        })
    } catch {
    } finally {
      dialogInfo.value.loading = false
    }
  })
}

/**
 * Init and open dialog
 */
const init = (accessType: string, row: any) => {
  dialogInfo.value.visible = true
  if (row) {
    formModel.value = row
  } else {
    formRef.value?.resetFields()
    formModel.value = {
      accessID: '',
      accessName: '',
      accessType,
      dbHost: '',
      dbPort: 5432,
      dbUser: '',
      dbPassword: '',
      dbName: '',
    }
  }
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
