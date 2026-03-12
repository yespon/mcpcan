<template>
  <!-- token config dialog -->
  <el-dialog
    v-model="formData.visible"
    width="580px"
    top="20vh"
    :show-close="false"
    @close="formRef?.resetFields()"
    header-class="token-header-border"
    footer-class="token-footer-border"
  >
    <template #header>
      <div class="center mb-4">{{ t('mcp.instance.token.title') }}</div>
    </template>
    <el-row v-loading="instanceInfo.loading" class="mt-4">
      <el-col :span="24">
        <el-form
          ref="formRef"
          :model="formData"
          :rules="rules"
          label-width="auto"
          label-position="top"
          class="mx-2"
        >
          <!-- 有效期 -->
          <el-form-item :label="t('mcp.instance.token.lifespan')" prop="expireAt">
            <template #label>
              <div class="center">
                <span class="mr-2">{{ t('mcp.instance.token.lifespan') }}</span>
                <el-button class="base-btn" type="primary" size="small" @click.stop="handleAddExpireAt(7)">
                  7{{ t('mcp.instance.token.day') }}
                </el-button>
                <el-button class="base-btn" type="primary" size="small" @click.stop="handleAddExpireAt(15)">
                  15{{ t('mcp.instance.token.day') }}
                </el-button>
                <el-button class="base-btn" type="primary" size="small" @click.stop="handleAddExpireAt(30)">
                  30{{ t('mcp.instance.token.day') }}
                </el-button>
              </div>
            </template>
            <el-date-picker
              v-model="formData.expireAt"
              type="datetime"
              value-format="x"
              :placeholder="t('mcp.instance.token.placeholderDate')"
              style="width: 100%"
              :disabled-date="(date: Date) => date.getTime() < Date.now()"
            />
          </el-form-item>
          <!-- 网关认证 token 值展示 -->
          <el-form-item :label="t('mcp.token.authentication')" prop="tokenType">
            <div class="w-full u-line-1" style="white-space: nowrap">
              Authorization：{{ formData.token }}
            </div>
          </el-form-item>
          <!-- 标签管理 -->
          <el-form-item :label="t('mcp.instance.token.tag')" prop="usages">
            <el-tag
              v-for="(tag, num) in formData.usages"
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
      </el-col>
    </el-row>
    <template #footer>
      <div class="center">
        <el-button @click="handleCancelToken" class="mr-4 w-25">{{ t('common.cancel') }}</el-button>
        <mcp-button @click="handleConfirmToken" class="w-25">{{ t('common.ok') }}</mcp-button>
      </div>
    </template>
  </el-dialog>

</template>
<script setup lang="ts">
import type { TokenType, InstanceResult } from '@/types'
import McpButton from '@/components/mcp-button/index.vue'

interface Props {
  modelValue: {
    visible: boolean
    token: string
    headers: Array<{ key: string; value: string; tokenType?: number }>
    tokenType: TokenType | null
    enabled: boolean
    expireAt: number | null
    usages: string[]
  }
  instanceInfo: InstanceResult
  tokenList: Array<any>
  currentEditIndex: number | null
}

const props = defineProps<Props>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: Props['modelValue']): void
  (e: 'on-refresh'): void
  (
    e: 'on-confirm',
    data: {
      token: string
      expireAt: number
      usages: string[]
      headers: Record<string, string>
      currentEditIndex: number | null
    },
  ): void
  (e: 'on-cancel'): void
}>()

const { t } = useI18n()
const formRef = ref()

const formData = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value),
})

const rules = reactive({})
const showTagInput = ref(false)
const inputValue = ref('')

// handle add expire at
const handleAddExpireAt = (days: number) => {
  const expireDate = new Date()
  expireDate.setDate(expireDate.getDate() + days)
  formData.value = {
    ...formData.value,
    expireAt: expireDate.getTime(),
  }
}

const showClosAble = computed(() => {
  return (tag: string) => {
    return ![
      'user_id',
      'user_name',
      'space_id',
      'space_name',
      'intelligent_access_id',
      'intelligent_access_name',
      'intelligent_access_type',
      'default',
    ].some((keyword) => tag.includes(keyword))
  }
})

// handle tag delete
const handleCloseTag = (num: number) => {
  formData.value.usages.splice(num, 1)
}

const handleTagConfirm = () => {
  const value = inputValue.value.trim()
  if (value && !formData.value.usages.includes(value)) {
    formData.value.usages.push(value)
  }
  inputValue.value = ''
  showTagInput.value = false
}

// handle cancel token
const handleCancelToken = () => {
  formData.value = {
    ...formData.value,
    visible: false,
    token: '',
    headers: [],
    enabled: true,
    expireAt: null,
    usages: [],
  }
  emit('on-cancel')
}

// handle confirm token
const handleConfirmToken = async () => {
  const result = await formRef.value.validate()
  if (!result) return

  const tokenData = {
    token: formData.value.token,
    expireAt: formData.value.expireAt || 0,
    usages: formData.value.usages,
    headers: {},
  }

  emit('on-confirm', {
    ...tokenData,
    currentEditIndex: props.currentEditIndex,
  })

  formData.value = {
    ...formData.value,
    visible: false,
    token: '',
    headers: [],
    enabled: true,
    expireAt: null,
    usages: [],
  }
}
</script>

<style lang="scss" scoped>
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
.enabledTransport {
  :deep(.el-form-item__label) {
    width: 100%;
  }
  :deep(.el-form-item__content) {
    display: block;
    width: 100%;
  }
}
.delete-header {
  width: 24px;
  height: 24px;
  border-radius: 4px;
}
</style>
<style lang="scss">
.el-dialog__header.token-header-border {
  background-color: transparent !important;
  border-bottom: 1px solid var(--el-border-color-light) !important;
}
.el-dialog__footer.token-footer-border {
  background-color: transparent !important;
  border-top: 1px solid var(--el-border-color-light) !important;
}
</style>
