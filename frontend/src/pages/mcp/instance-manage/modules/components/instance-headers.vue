<template>
  <div class="instance-headers-wrapper w-full">
    <el-form-item prop="headers" label-width="0" class="w-full mb-4">
      <div class="w-full flex-col">
        <!-- 假扮的 Label 区域（避免参与 el-form 的 label-width="auto" 计算） -->
        <div class="w-full mb-1">
          <div class="flex items-center text-sm font-bold text-[var(--el-text-color-regular)]">
            <span class="mr-2">🚀 {{ locale === 'en' ? 'Custom Headers' : '自定义请求头 (Headers)' }}</span>
          </div>
          <div class="text-[12px] text-gray-500 font-normal mt-1" style="line-height: 1.6;">
            {{ locale === 'en' ? 'Used for backend authentication to MCP or OpenAPI services (e.g., API Key, Token). The gateway will attach these automatically; the client does not need to send them.' : '用于后端对 MCP 或 OpenAPI 服务进行隐式鉴权（如 API Key、Token 等）。网关转发时将自动携带，客户端无需感知和传入。' }}
          </div>
        </div>

        <!-- 组件内容区域 -->
        <div class="w-full border border-gray-200 dark:border-gray-700 rounded-md p-3 bg-gray-50/50 dark:bg-gray-800/50 mt-1 box-border">
          <div v-if="localHeaders.length === 0" class="text-center py-2 text-gray-400 text-sm">
            <el-empty :image-size="40" :description="locale === 'en' ? 'No custom headers configured' : '暂未配置自定义请求头'" class="my-0 py-0"></el-empty>
            <el-button plain type="primary" size="small" class="mt-2" @click="handleAddHeader" :icon="Plus">
              {{ locale === 'en' ? 'Add Header' : '添加请求头' }}
            </el-button>
          </div>

          <template v-else>
            <div
              v-for="(item, index) in localHeaders"
              :key="index"
              class="flex items-center mb-3 last:mb-0 w-full"
            >
              <el-row :gutter="12" class="flex-sub align-middle w-full mx-0">
                <el-col :span="9">
                  <el-input
                    v-model="item.key"
                    :placeholder="t('mcp.instance.token.headersKey')"
                    @input="updateHeaders"
                  >
                    <template #prepend>Key</template>
                  </el-input>
                </el-col>
                <el-col :span="13">
                  <el-input
                    v-model="item.value"
                    :placeholder="t('mcp.instance.token.headersValue')"
                    @input="updateHeaders"
                  >
                    <template #prepend>Value</template>
                  </el-input>
                </el-col>
                <el-col :span="2" class="flex items-center justify-center">
                  <el-button type="danger" plain circle size="small" :icon="Minus" @click="handleDeleteHeader(index)"></el-button>
                </el-col>
              </el-row>
            </div>
            <div class="mt-2">
              <el-button plain type="primary" size="small" @click="handleAddHeader" :icon="Plus">
                {{ locale === 'en' ? 'Add Another Header' : '继续添加' }}
              </el-button>
            </div>
          </template>
        </div>
      </div>
    </el-form-item>
  </div>
</template>
<script setup lang="ts">
import { ref, watch } from 'vue'
import { Plus, Minus } from '@element-plus/icons-vue'
import { useI18n } from 'vue-i18n'

const props = defineProps<{ headers: { key: string; value: string }[] | null | undefined }>()
const emit = defineEmits(['update:headers'])
const { t, locale } = useI18n()

// We need a local array to drive the v-for since props.headers might be object or null
const localHeaders = ref<{ key: string; value: string }[]>([])

watch(
  () => props.headers,
  (newVal) => {
    if (!newVal) {
      localHeaders.value = []
    } else if (Array.isArray(newVal)) {
       // if it's already an array of key/value
       localHeaders.value = newVal.map(item => ({...item}))
    } else {
       // if it's an object of {k:v} map
       localHeaders.value = Object.keys(newVal).map(k => ({ key: k, value: (newVal as any)[k] }))
    }
  },
  { immediate: true, deep: true }
)

const updateHeaders = () => {
  emit('update:headers', localHeaders.value)
}

const handleAddHeader = () => {
  localHeaders.value.push({ key: '', value: '' })
  updateHeaders()
}

const handleDeleteHeader = (index: number) => {
  localHeaders.value.splice(index, 1)
  updateHeaders()
}
</script>
<style lang="scss" scoped>
.delete-header {
  width: 24px;
  height: 24px;
  border-radius: 4px;
}
</style>

