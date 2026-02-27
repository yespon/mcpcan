<template>
  <el-dialog
    v-model="visible"
    :title="isEdit ? t('model.edit') : t('model.create')"
    width="520px"
    @close="handleClose"
  >
    <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
      <el-form-item :label="t('model.name')" prop="name">
        <el-input v-model="form.name" :placeholder="t('model.namePlaceholder')" />
      </el-form-item>
      <el-form-item :label="t('model.provider')" prop="provider">
        <el-select
          v-model="form.provider"
          :placeholder="t('model.providerPlaceholder')"
          filterable
          class="w-full"
          @change="handleProviderChange"
        >
          <el-option v-for="p in supportedProviders" :key="p.id" :label="p.name" :value="p.id">
            <span class="float-left">{{ p.name }}</span>
            <span class="float-right text-gray-400 text-xs ml-2">{{ p.id }}</span>
          </el-option>
        </el-select>
      </el-form-item>
      <el-form-item :label="t('model.apiKey')" prop="apiKey">
        <el-input
          v-model="form.apiKey"
          :placeholder="t('model.apiKeyPlaceholder')"
          :disabled="isEdit"
        />
      </el-form-item>
      <el-form-item :label="t('model.baseUrl')" prop="baseUrl">
        <el-input v-model="form.baseUrl" :placeholder="t('model.baseUrlPlaceholder')" />
        <div
          class="text-xs text-[var(--ep-text-color-secondary)] mt-1"
          v-if="form.provider && getProviderBaseUrl(form.provider)"
        >
          {{ t('model.baseUrlHelper', { url: getProviderBaseUrl(form.provider) }) }}
          <el-button
            type="primary"
            link
            size="small"
            @click="form.baseUrl = getProviderBaseUrl(form.provider)"
          >
            {{ t('model.use') }}
          </el-button>
        </div>
      </el-form-item>
      <el-form-item :label="t('model.allowedModels')" prop="allowedModels">
        <el-select
          v-model="form.allowedModels"
          multiple
          filterable
          allow-create
          default-first-option
          :placeholder="t('model.allowedModelsPlaceholder')"
          class="w-full"
        >
          <el-option v-for="m in selectedProviderModels" :key="m" :label="m" :value="m" />
        </el-select>
        <div class="text-xs text-[var(--ep-text-color-secondary)] mt-1">
          {{ t('model.allowedModelsDesc') }}
        </div>
      </el-form-item>
    </el-form>
    <template #footer>
      <span class="dialog-footer">
        <el-button @click="visible = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="loading">
          {{ t('common.confirm') }}
        </el-button>
      </span>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, watch, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ChatAPI } from '@/api/agent'
import { ElMessage } from 'element-plus'
import type { Model } from '@/types/model'

interface SupportedProvider {
  id: string
  name: string
  models: string[]
  baseUrl: string
}

const props = defineProps<{
  modelValue: boolean
  currentModel?: Model
}>()

const emit = defineEmits(['update:modelValue', 'success'])

const { t } = useI18n()
const visible = ref(false)
const loading = ref(false)
const isEdit = ref(false)
const formRef = ref()
const supportedProviders = ref<SupportedProvider[]>([])

const form = reactive({
  name: '',
  provider: '',
  modelName: '',
  apiKey: '',
  baseUrl: '',
  allowedModels: [] as string[],
})

const rules = {
  name: [{ required: true, message: t('model.namePlaceholder'), trigger: 'blur' }],
  provider: [{ required: true, message: t('model.providerPlaceholder'), trigger: 'blur' }],
  apiKey: [{ required: true, message: t('model.apiKeyPlaceholder'), trigger: 'blur' }],
}

const selectedProviderModels = computed(() => {
  if (!form.provider || !supportedProviders.value) return []
  const p = supportedProviders.value.find((x) => x.id === form.provider)
  return p ? p.models : []
})

const getProviderBaseUrl = (pid: string) => {
  if (!supportedProviders.value) return ''
  const p = supportedProviders.value.find((x) => x.id === pid)
  return p ? p.baseUrl : ''
}

const handleProviderChange = () => {
  // Logic from ChatInput.vue
}

const fetchSupportedProviders = async () => {
  try {
    const { providers: list } = await ChatAPI.getSupportedModels()
    supportedProviders.value = list || []
  } catch (error) {
    console.error('Failed to fetch supported providers', error)
  }
}

watch(
  () => props.modelValue,
  async (val) => {
    visible.value = val
    if (val) {
      if (props.currentModel?.id) {
        isEdit.value = true
        loading.value = true
        try {
          const { access } = await ChatAPI.getModelAccess(props.currentModel.id)
          const data = access
          // Ensure allowedModels is split if it's a string, or defaults to empty array
          let allowedModels: string[] = []
          if (data.allowedModels) {
            if (Array.isArray(data.allowedModels)) {
              allowedModels = data.allowedModels
            } else if (typeof data.allowedModels === 'string') {
              try {
                // Try to parse as JSON first (backend sends JSON string)
                const parsed = JSON.parse(data.allowedModels)
                if (Array.isArray(parsed)) {
                  allowedModels = parsed
                } else {
                  // Fallback for simple string or comma-separated
                  allowedModels = data.allowedModels.split(',')
                }
              } catch {
                // Not JSON, assume comma-separated
                allowedModels = data.allowedModels.split(',')
              }
            }
          }

          Object.assign(form, {
            ...data,
            allowedModels,
          })
        } catch (error) {
          console.error(error)
        } finally {
          loading.value = false
        }
      } else {
        isEdit.value = false
        // Reset form for create mode
        formRef.value?.resetFields()
        Object.assign(form, {
          name: '',
          provider: '',
          modelName: '',
          apiKey: '',
          baseUrl: '',
          allowedModels: [],
        })
      }
    }
  },
)

watch(visible, (val) => {
  emit('update:modelValue', val)
})

const handleClose = () => {
  formRef.value?.resetFields()
}

const handleSubmit = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid: boolean) => {
    if (valid) {
      loading.value = true
      // Infer modelName (Target Model) from allowedModels if available
      let finalModelName = form.modelName
      if (form.allowedModels && form.allowedModels.length > 0) {
        finalModelName = form.allowedModels[0]
      }

      try {
        if (isEdit.value && props.currentModel?.id) {
          await ChatAPI.updateModelAccess({
            id: props.currentModel.id,
            name: form.name,
            provider: form.provider,
            baseUrl: form.baseUrl,
            modelName: finalModelName,
            allowedModels: form.allowedModels || [],
          })
          ElMessage.success(t('model.updateSuccess'))
        } else {
          await ChatAPI.createModelAccess({
            name: form.name,
            provider: form.provider,
            apiKey: form.apiKey,
            baseUrl: form.baseUrl,
            modelName: finalModelName,
            allowedModels: form.allowedModels || [],
          })
          ElMessage.success(t('model.createSuccess'))
        }
        visible.value = false
        emit('success')
      } catch (error) {
        // Error handling
      } finally {
        loading.value = false
      }
    }
  })
}

onMounted(() => {
  fetchSupportedProviders()
})
</script>
