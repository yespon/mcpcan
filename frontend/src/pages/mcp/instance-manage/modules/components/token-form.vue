<template>
  <el-form
    ref="formRef"
    :model="localFormData"
    label-width="auto"
    label-position="top"
    class="mx-2"
  >
    <el-form-item prop="usages">
      <el-tag
        v-for="(tag, num) in localFormData.usages"
        :key="num"
        :closable="showClosAble(tag)"
        class="mx-2 my-1"
        :disable-transitions="false"
        @close="handleCloseTag(num)"
        color="var(--ep-bg-purple-color)"
      >
        {{ tag }}
      </el-tag>
      <el-input
        v-if="showTagInput"
        ref="InputRef"
        v-model="inputValue"
        class="w-5 mx-2"
        style="width: 100px"
        size="small"
        @keyup.enter="handleTagConfirm"
        @blur="handleTagConfirm"
      />
      <el-button v-else class="mx-2" size="small" @click="showTagInput = true">
        + {{ t('mcp.instance.token.newTag') }}
      </el-button>
    </el-form-item>
  </el-form>
</template>
<script setup lang="ts">
import McpButton from '@/components/mcp-button/index.vue'

// suppress unused warning – McpButton referenced in parent context
void McpButton

const props = defineProps<{ formData: any }>()
const emit = defineEmits(['update:formData'])
const { t } = useI18n()
const showTagInput = ref(false)
const inputValue = ref('')
const localFormData = computed({
  get: () => {
    return props.formData
  },
  set: (val) => {
    const emitData = { ...val }
    emit('update:formData', emitData)
  },
})
const showClosAble = computed(() => {
  return (tag: string) => {
    return [
      'user_id',
      'user_name',
      'space_id',
      'space_name',
      'intelligent_access_id',
      'intelligent_access_name',
      'intelligent_access_type',
      'default',
    ].some((keyword) => tag.includes(keyword))
      ? false
      : true
  }
})
// handle tag delete
const handleCloseTag = (num: number) => {
  localFormData.value.usages.splice(num, 1)
}
const handleTagConfirm = () => {
  const value = inputValue.value.trim()
  if (value && !localFormData.value.usages.includes(value)) {
    localFormData.value.usages.push(value)
  }
  inputValue.value = ''
  showTagInput.value = false
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
:deep(.el-tag) {
  color: var(--el-color-primary);
  border-color: var(--ep-border-color);
  .el-tag__close {
    color: var(--el-color-primary);
    &:hover {
      background-color: var(--ep-bg-purple-color-deep);
    }
  }
}
</style>
