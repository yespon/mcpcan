<template>
  <el-dialog
    v-model="dialogInfo.visible"
    :title="dialogInfo.title"
    :show-close="false"
    width="680px"
    top="10vh"
  >
    <el-form ref="formRef" :model="formModel" :rules="rules" label-width="110px">
      <el-form-item label="接入名称" prop="accessName">
        <el-input
          v-model="formModel.accessName"
          placeholder="请输入接入名称"
          maxlength="64"
          show-word-limit
        />
      </el-form-item>

      <el-form-item label="接入类型" prop="accessType">
        <el-select v-model="formModel.accessType" placeholder="请选择接入类型">
          <el-option label="MySQL" value="mysql" />
          <el-option label="PostgreSQL" value="postgres" />
          <el-option label="SQLServer" value="sqlserver" />
          <el-option label="SQLite" value="sqlite" />
        </el-select>
      </el-form-item>

      <el-form-item label="数据库地址" prop="dbHost">
        <el-input v-model="formModel.dbHost" placeholder="例如：127.0.0.1 或 dns" />
      </el-form-item>

      <el-form-item label="端口" prop="dbPort">
        <el-input-number
          v-model="formModel.dbPort"
          :min="1"
          :max="65535"
          :controls="false"
          placeholder="端口"
          style="width: 100%"
        />
      </el-form-item>

      <el-form-item label="用户名" prop="dbUser">
        <el-input v-model="formModel.dbUser" placeholder="数据库用户名" />
      </el-form-item>

      <el-form-item label="密码" prop="dbPassword">
        <el-input
          v-model="formModel.dbPassword"
          type="password"
          show-password
          placeholder="数据库密码"
        />
      </el-form-item>

      <el-form-item label="数据库名" prop="dbName">
        <el-input v-model="formModel.dbName" placeholder="例如：mcp" />
      </el-form-item>
    </el-form>

    <template #footer>
      <div class="center">
        <el-button @click="handleCancel" class="mr-2">取消</el-button>
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
// import {} from ''

const { t } = useI18n()
const dialogInfo = ref<{
  title: string
  visible: boolean
  loading: boolean
}>({
  title: '智能体接入',
  visible: false,
  loading: false,
})

// 表单 model
const formModel = ref<{
  accessName: string
  accessType: string
  dbHost: string
  dbPort: number
  dbUser: string
  dbPassword: string
  dbName: string
}>({
  accessName: '',
  accessType: '',
  dbHost: '',
  dbPort: 3306,
  dbUser: '',
  dbPassword: '',
  dbName: '',
})

const formRef = ref<FormInstance>()

// 校验规则
const rules: FormRules<typeof formModel.value> = {
  accessName: [
    { required: true, message: '请输入接入名称', trigger: 'blur' },
    { min: 2, max: 64, message: '长度 2-64 个字符', trigger: 'blur' },
  ],
  accessType: [{ required: true, message: '请选择接入类型', trigger: 'change' }],
  dbHost: [{ required: true, message: '请输入数据库地址', trigger: 'blur' }],
  dbPort: [
    {
      required: true,
      validator: (_rule, value, callback) => {
        if (!value || value < 1 || value > 65535) return callback(new Error('端口范围 1-65535'))
        callback()
      },
      trigger: 'change',
    },
  ],
  dbUser: [{ required: true, message: '请输入数据库用户名', trigger: 'blur' }],
  dbPassword: [{ required: true, message: '请输入数据库密码', trigger: 'blur' }],
  dbName: [{ required: true, message: '请输入数据库名', trigger: 'blur' }],
}

// 取消
const handleCancel = () => {
  dialogInfo.value.visible = false
}

// 提交
const emit = defineEmits<{
  (e: 'submit', payload: typeof formModel.value): void
}>()

const handleSubmit = async () => {
  if (!formRef.value) return
  await formRef.value.validate((valid) => {
    if (!valid) return
    emit('submit', { ...formModel.value })
    dialogInfo.value.visible = false
  })
}

/**
 * Init and open dialog
 */
const init = (type: string) => {
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
