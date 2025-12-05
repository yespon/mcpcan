<template>
  <div class="flex">
    <el-input
      ref="searchInputRef"
      v-model="localFormData[formConfig[0].key]"
      v-bind="{ ...formConfig[0].props }"
      @keyup.enter="onQuery"
      :suffix-icon="Search"
    ></el-input>
    <el-button style="width: 32px" class="ml-3" @click="onQuery">
      <el-icon><Refresh /></el-icon>
    </el-button>
    <el-popover
      v-model:visible="showMoreSearch"
      placement="bottom"
      :width="464"
      trigger="click"
      :show-arrow="false"
      append-to-body
      style="padding: 24px"
    >
      <template #reference>
        <el-button style="width: 32px">
          <el-icon><i class="icon iconfont MCP-shaixuan1"></i></el-icon>
        </el-button>
      </template>
      <FormPlus
        ref="searchFromRef"
        :form-config="formConfig"
        :form-data="localFormData"
        name="formPlus"
        label-width="108"
        label-position="left"
        @update:formData="onFormDataUpdate"
      >
        <template #handler>
          <el-form-item class="search-buttons flex-sub text-right">
            <div class="flex-sub flex justify-end">
              <el-button @click="onReset" class="mr-2">{{ t('common.reseat') }}</el-button>
              <GlareHover
                width="auto"
                height="auto"
                background="transparent"
                border-color="#222222"
                border-radius="4px"
                glare-color="#ffffff"
                :glare-opacity="0.3"
                :glare-size="300"
                :transition-duration="800"
                :play-once="false"
              >
                <el-button type="primary" @click="onQuery" class="base-btn">{{
                  t('common.ok')
                }}</el-button>
              </GlareHover>
            </div>
          </el-form-item>
        </template>
      </FormPlus>
    </el-popover>
  </div>
</template>

<script setup lang="ts">
import { Refresh, Search } from '@element-plus/icons-vue'
import FormPlus from '../FormPlus/index.vue'
import GlareHover from '../Animation/GlareHover.vue'

interface FormConfigItem {
  key: string
  props?: Record<string, any>
}
interface FormData {
  [key: string]: any
}
const { t } = useI18n()
const props = defineProps<{
  formConfig: FormConfigItem[]
  formData: FormData
}>()
const showMoreSearch = ref(false)
const emit = defineEmits(['update:formData', 'reset-fields', 'handle-query'])
const localFormData = ref<FormData>({ ...props.formData })

// 双向数据流：监听外部formData变化，更新本地localFormData
watch(
  () => props.formData,
  (newVal) => {
    localFormData.value = { ...newVal }
  },
)

// 监听本地localFormData变化，向父组件同步
watch(
  () => localFormData.value,
  (newVal) => {
    emit('update:formData', { ...newVal })
  },
  { deep: true },
)

const onFormDataUpdate = (newVal: FormData) => {
  localFormData.value = { ...newVal }
}
const searchFromRef = ref()
const onReset = () => {
  showMoreSearch.value = false
  localFormData.value = { ...props.formData }
  searchFromRef.value.resetFields()
  emit('reset-fields')
}
const onQuery = () => {
  showMoreSearch.value = false
  emit('handle-query', localFormData.value)
}
</script>
