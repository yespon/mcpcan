<template>
  <div v-loading="pageInfo.loading" :element-loading-text="pageInfo.loadingText">
    <div class="flex justify-between page-header mb-4">
      <div class="header-container">
        {{ t('model.pageDesc.list') }}
        <span class="desc">{{ t('model.pageDesc.desc') }}</span>
      </div>
      <mcp-button type="primary" :icon="Plus" @click="handleCreate">{{
        t('model.create')
      }}</mcp-button>
    </div>

    <TablePlus
      :showOperation="true"
      searchContainer="#modelSearch"
      ref="tablePlus"
      :requestConfig="requestConfig"
      :columns="columns"
      v-model:pageConfig="pageConfig"
      :handlerColumnConfig="{
        width: '180px',
        fixed: 'right',
        align: 'center',
      }"
    >
      <template #action>
        <div class="flex justify-between mb-4">
          <div class="center">
            <span class="desc">{{ t('common.total') }}：{{ pageConfig.total }}</span>
          </div>
          <div id="modelSearch"></div>
        </div>
      </template>
      <template #allowedModels="{ row }">
        <template v-for="value in row.allowedModels || []">
          <el-tag class="m-1">{{ value }}</el-tag>
        </template>
      </template>

      <template #operation="{ row }">
        <el-button type="primary" link size="small" @click="handleTestConnection(row)">
          {{ t('model.testConnection') }}
        </el-button>
        <el-button type="primary" link size="small" @click="handleEdit(row)">
          {{ t('common.edit') }}
        </el-button>
        <el-button type="danger" link size="small" @click="handleDelete(row)">
          {{ t('common.delete') }}
        </el-button>
      </template>
    </TablePlus>

    <ModelDialog v-model="dialogVisible" :current-model="currentModel" @success="handleSuccess" />

    <el-dialog v-model="testDialogVisible" :title="t('model.selectModelToTest')" width="700px">
      <div class="test-dialog-content">
        <div class="test-dialog-left">
          <div v-if="testModelOptions.length === 0" class="text-gray-400 text-center py-4">
            {{ t('model.noModelsAvailable') }}
          </div>
          <div v-for="m in testModelOptions" :key="m" class="model-item-container">
            <el-button
              class="w-full justify-start model-btn"
              :class="{
                'is-success': modelTestStatus[m] === 'success',
                'is-error': modelTestStatus[m] === 'error',
                'is-process': modelTestStatus[m] === 'loading',
              }"
              :loading="testingConnection && selectedTestModel === m"
              :type="getModelButtonType(m)"
              @click="clickModelButton(m)"
              :disabled="testingConnection && selectedTestModel !== m"
            >
              <div class="flex items-center justify-between w-full">
                <span class="truncate" :title="m">{{ m }}</span>
                <el-icon v-if="modelTestStatus[m] === 'success'" class="status-icon success"
                  ><Check
                /></el-icon>
                <el-icon v-if="modelTestStatus[m] === 'error'" class="status-icon error"
                  ><Close
                /></el-icon>
              </div>
            </el-button>
          </div>
        </div>
        <div class="test-dialog-divider"></div>
        <div class="test-dialog-right">
          <!-- Placeholder for potential future details or logs -->
          <div
            v-if="!currentLogMessage"
            class="text-gray-500 text-sm flex items-center justify-center h-full"
          >
            {{ t('model.clickToTest') }}
          </div>
          <div v-else class="log-content">
            {{ currentLogMessage }}
          </div>
        </div>
      </div>
      <template #footer>
        <div class="dialog-footer center-footer">
          <el-button @click="testDialogVisible = false">{{ t('common.close') }}</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, h } from 'vue'
import { Plus, Check, Close } from '@element-plus/icons-vue'
import TablePlus from '@/components/TablePlus/index.vue'
import { useModelTableHooks } from './hooks'
import { ElMessage, ElMessageBox, ElSelect, ElOption } from 'element-plus'
import ModelDialog from './components/ModelDialog.vue'
import { ChatAPI } from '@/api/agent'

const { t, tablePlus, columns, requestConfig, pageConfig, pageInfo } = useModelTableHooks()

const dialogVisible = ref(false)
const currentModel = ref<any>(null)
const testDialogVisible = ref(false)
const selectedTestModel = ref('')
const testingConnection = ref(false)
const testModelOptions = ref<string[]>([])
const currentTestModalId = ref(0)
const modelTestStatus = ref<Record<string, 'success' | 'error' | 'loading' | ''>>({})
const currentLogMessage = ref('')

const handleCreate = () => {
  currentModel.value = null
  dialogVisible.value = true
}

const handleEdit = (row: any) => {
  currentModel.value = { ...row }
  dialogVisible.value = true
}

const getModelButtonType = (model: string) => {
  const status = modelTestStatus.value[model]
  if (status === 'success') return 'success'
  if (status === 'error') return 'danger'
  return 'default'
}

const performTestConnection = async (id: number, modelName: string) => {
  testingConnection.value = true
  selectedTestModel.value = modelName
  modelTestStatus.value[modelName] = 'loading'
  currentLogMessage.value = t('model.testing', { model: modelName }) || `Testing ${modelName}...`
  try {
    const { success, message } = await ChatAPI.testConnectionNew({ id, modelName })
    if (success) {
      modelTestStatus.value[modelName] = 'success'
      currentLogMessage.value =
        t('model.testSuccess', { model: modelName }) || `${modelName} Connection Successful!`
      ElMessage.success(t('model.testConnectionSuccess'))
    } else {
      modelTestStatus.value[modelName] = 'error'
      // Use the specific error message from the nested data object if available
      const errorMsg = message || 'Unknown error'
      currentLogMessage.value = t('model.testFailed', { model: modelName }) + '\nError: ' + errorMsg
      ElMessage.error(t('model.testConnectionFailed'))
    }
  } catch (error: any) {
    console.error(error)
    modelTestStatus.value[modelName] = 'error'
    currentLogMessage.value = `${modelName} Error: ${error.message || 'Unknown error'}`
    ElMessage.error(t('model.testConnectionFailed'))
  } finally {
    testingConnection.value = false
    selectedTestModel.value = ''
  }
}

const clickModelButton = async (model: string) => {
  if (testingConnection.value) return
  await performTestConnection(currentTestModalId.value, model)
}

const handleTestConnection = async (row: any) => {
  modelTestStatus.value = {}
  currentLogMessage.value = ''
  let models: string[] = []
  if (Array.isArray(row.allowedModels)) {
    models = row.allowedModels
  } else if (typeof row.allowedModels === 'string') {
    try {
      // Try parsing as JSON first
      const parsed = JSON.parse(row.allowedModels)
      if (Array.isArray(parsed)) {
        models = parsed
      } else {
        // If it's a simple string, treat as single item array
        models = [parsed]
      }
    } catch {
      // Fallback to comma separation
      models = row.allowedModels
        .split(',')
        .map((s: string) => s.trim())
        .filter(Boolean)
    }
  }

  if (models.length === 0) {
    ElMessage.warning(t('model.noModelsAvailable'))
    return
  }

  if (models.length === 1) {
    await performTestConnection(row.id, models[0])
  } else {
    currentTestModalId.value = row.id
    testModelOptions.value = [...models, 'doubao-seed-code-preview-251028']
    selectedTestModel.value = ''
    testDialogVisible.value = true
  }
}

const submitTestConnection = async () => {
  if (!selectedTestModel.value) return
  await performTestConnection(currentTestModalId.value, selectedTestModel.value)
}

const handleDelete = (row: any) => {
  ElMessageBox.confirm(t('model.deleteConfirm'), t('common.warn'), {
    confirmButtonText: t('common.ok'),
    cancelButtonText: t('common.cancel'),
    type: 'warning',
  }).then(async () => {
    try {
      await ChatAPI.deleteModelAccess(row.id)
      ElMessage.success(t('model.deleteSuccess'))
      tablePlus.value.initData()
    } catch (error) {
      // Error handling
    }
  })
}

const handleSuccess = () => {
  tablePlus.value.initData()
}

onMounted(() => {
  // Initial load is handled by TablePlus usually, or call initData
  tablePlus.value.initData()
})
</script>

<style lang="scss" scoped>
.page-header {
  .header-container {
    font-size: 20px;
    font-weight: 500;
  }
}
.desc {
  font-size: 14px;
  color: #999999;
  margin-left: 10px;
}

.test-dialog-content {
  display: flex;
  height: 400px;
  gap: 16px;
}

.test-dialog-left {
  flex: 1;
  overflow-y: auto;
  padding-right: 8px;
}

.model-btn {
  margin-bottom: 8px;
  transition: all 0.3s;

  &.is-success {
    --el-button-bg-color: var(--el-color-success-light-9);
    --el-button-border-color: var(--el-color-success-light-5);
    --el-button-text-color: var(--el-color-success);

    &:hover {
      --el-button-bg-color: var(--el-color-success-light-8);
      --el-button-border-color: var(--el-color-success-light-3);
    }
  }

  &.is-error {
    --el-button-bg-color: var(--el-color-danger-light-9);
    --el-button-border-color: var(--el-color-danger-light-5);
    --el-button-text-color: var(--el-color-danger);

    &:hover {
      --el-button-bg-color: var(--el-color-danger-light-8);
      --el-button-border-color: var(--el-color-danger-light-3);
    }
  }
}

.test-dialog-divider {
  width: 1px;
  background-color: var(--el-border-color-lighter);
}

.test-dialog-right {
  flex: 1;
  padding-left: 12px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.log-content {
  width: 100%;
  height: 100%;
  overflow-y: auto;
  overflow-x: hidden;
  font-family: monospace;
  white-space: pre-wrap;
  word-break: break-all;
  color: var(--el-text-color-regular);
  font-size: 13px;
  line-height: 1.5;
  padding-right: 4px; /* Space for scrollbar */
}

.center-footer {
  display: flex;
  justify-content: center;
}
</style>
