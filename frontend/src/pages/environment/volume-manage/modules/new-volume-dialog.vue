<template>
  <el-dialog v-model="dialogInfo.visible" :show-close="false" width="480px" top="10vh">
    <template #header>
      <div class="center mt-4 mb-4">{{ dialogInfo.title }}</div>
    </template>
    <el-scrollbar ref="scrollbarRef" max-height="75vh" always>
      <el-form
        ref="formRef"
        :model="dialogInfo.formData"
        :rules="dialogInfo.rules"
        label-width="auto"
        label-position="top"
        class="mr-2 ml-2"
      >
        <el-form-item :label="t('env.volume.name')" prop="name">
          <el-input v-model="dialogInfo.formData.name" :placeholder="t('env.volume.name')" />
        </el-form-item>
        <el-form-item :label="t('mcp.instance.token.tag')" prop="labels">
          <el-row class="w-full" justify="space-between">
            <el-col :span="24" v-for="(label, index) in dialogInfo.formData.labels" :key="index">
              <div class="flex align-center mb-2">
                <el-input
                  v-model="label.key"
                  :placeholder="t('mcp.instance.formData.key')"
                  class="mr-2"
                />
                <el-input
                  v-model="label.value"
                  :placeholder="t('mcp.instance.formData.value')"
                  class="mr-2"
                />
                <el-icon
                  class="cursor-pointer"
                  color="#F56C6C"
                  @click="handleDeleteEnvVariable(index)"
                >
                  <Remove />
                </el-icon>
              </div>
            </el-col>
            <el-col :span="24">
              <el-button class="add-env" :icon="Plus" plain @click="handleAddEnvVariable">
                {{ t('env.volume.addLabel') }}
              </el-button>
            </el-col>
          </el-row>
        </el-form-item>
      </el-form>
    </el-scrollbar>

    <template #footer>
      <div class="center text-center">
        <el-button @click="handleCancel" class="mr-2">{{ t('common.cancel') }}</el-button>
        <mcp-button @click="handleConfirm" :loading="dialogInfo.loading">{{
          t('common.save')
        }}</mcp-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { Plus, Remove } from '@element-plus/icons-vue'
import { VolumeAPI } from '@/api/env'
import { ElMessage } from 'element-plus'
import McpButton from '@/components/mcp-button/index.vue'

const { t } = useI18n()
const $route = useRoute()
const { query } = $route
const formRef = ref()
const emit = defineEmits(['onRefresh'])
const dialogInfo = ref<any>({
  visible: false,
  loading: false,
  title: t('env.volume.add'),
  formData: {
    name: '',
    labels: [],
  },
  rules: {
    name: [{ required: true, message: t('env.volume.rules.name'), trigger: 'blur' }],
  },
})

/**
 * Handle delete environment variables
 * @param index - Index of environment variables
 */
const handleDeleteEnvVariable = (index: number | string) => {
  dialogInfo.value.formData.labels.splice(index, 1)
}
/**
 * Handle add an environment variables
 */
const handleAddEnvVariable = () => {
  dialogInfo.value.formData.labels.push({ key: '', value: '' })
}
/**
 * Handle cancel
 */
const handleCancel = () => {
  dialogInfo.value.visible = false
}
/**
 * Handle confirm
 */
const handleConfirm = () => {
  formRef.value.validate(async (valid: boolean) => {
    if (valid) {
      try {
        dialogInfo.value.loading = true
        await VolumeAPI.createVolume({
          environmentId: Number(query.environmentId),
          name: dialogInfo.value.formData.name,
          labels: Object.fromEntries(
            dialogInfo.value.formData.labels.map((label: any) => [label.key, label.value]),
          ),
        })
        ElMessage.success(t('action.create'))
        dialogInfo.value.visible = false
        emit('onRefresh')
      } finally {
        dialogInfo.value.loading = false
      }
    }
  })
}

/**
 * Handle init form data
 * @param form - form data
 */
const init = () => {
  dialogInfo.value.visible = true
}

defineExpose({
  init,
})
</script>

<style lang="scss" scoped>
.add-env {
  width: 100%;
  border: 1px dashed var(--el-border-color);
}
</style>
