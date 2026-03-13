<template>
  <el-form-item prop="headers" class="enabledTransport">
    <template #label>
      <div class="w-full flex justify-between items-center">
        <span class="mr-2 pr-4"> 🚀 {{ t('mcp.token.passthrough') }} Headers </span>
        <div class="center">
          <div
            class="cursor-pointer border border-style-solid border-rd-md border-white ml-2 p-1 center bg-gray-600 color-white hover-scale-110"
            @click="handleAddHeader"
          >
            <el-icon>
              <Plus />
            </el-icon>
          </div>
        </div>
      </div>
    </template>
    <div
      v-for="(item, index) in localHeaders"
      :key="index"
      class="flex items-center my-2 pr-3 w-full"
    >
      <el-row :gutter="12" class="flex-sub align-center w-full">
        <el-col :span="7">
          <div class="flex h-full items-center justify-end">
            <el-input
              v-model="item.key"
              :placeholder="t('mcp.instance.token.headersKey')"
              class="flex-sub"
              @input="updateHeaders"
            >
            </el-input>
            <span class="ml-2">:</span>
          </div>
        </el-col>
        <el-col :span="15" class="flex">
          <div class="flex w-full">
            <el-input
              v-model="item.value"
              :placeholder="t('mcp.instance.token.headersValue')"
              class="flex-sub w-full"
              @input="updateHeaders"
            ></el-input>
          </div>
        </el-col>
        <el-col :span="2">
          <div
            class="cursor-pointer border border-style-solid delete-header border-white px-1 ml-2 center bg-red-100/50 color-white hover-bg-red-400/90 hover-scale-105"
            @click="handleDeleteHeader(index)"
          >
            <el-icon><Minus /></el-icon>
          </div>
        </el-col>
      </el-row>
    </div>
  </el-form-item>
</template>
<script setup lang="ts">
import { ref, watch } from 'vue'
import { Plus, Minus } from '@element-plus/icons-vue'
import { useI18n } from 'vue-i18n'

const props = defineProps<{ headers: { key: string; value: string }[] | null | undefined }>()
const emit = defineEmits(['update:headers'])
const { t } = useI18n()

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
:deep(.el-form-item__label) {
  width: 100% !important;
}
:deep(.el-form-item__content) {
  display: block;
  width: 100%;
}
.delete-header {
  width: 24px;
  height: 24px;
  border-radius: 4px;
}
</style>

