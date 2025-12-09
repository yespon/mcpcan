<template>
  <el-dialog
    v-model="dialogData.visible"
    width="580px"
    top="20vh"
    :show-close="false"
    @close="formRef?.resetFields()"
    header-class="token-header-border"
    footer-class="token-footer-border"
  >
    <template #header>
      <div class="center mb-4">{{ t('mcp.instance.token.batchChangeTitle') }}</div>
    </template>
    <el-scrollbar
      v-loading="dialogData.loading"
      ref="scrollbarRef"
      height="50vh"
      always
      class="pr-4"
    >
      <el-form
        ref="formRef"
        :model="formData"
        :rules="rules"
        label-width="auto"
        label-position="top"
        class="mx-2"
      >
        <el-form-item prop="enabledTransport" class="enabledTransport mt-4">
          <template #label>
            <div class="w-full flex justify-between items-center">
              <span class="mr-2"> {{ t('mcp.token.passthrough') }} {{ 'Headers' }} </span>
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
            v-for="(item, index) in formData.headers"
            :key="index"
            class="flex items-center my-2 pr-3"
          >
            <el-row :gutter="12" class="flex-sub align-center">
              <el-col :span="7">
                <div class="flex h-full items-center justify-end">
                  <el-dropdown
                    v-if="tokenTypeOptions.some((tokenType) => item.key === tokenType.label)"
                    trigger="click"
                    class="h-full w-full flex items-center justify-end"
                    :show-arrow="false"
                  >
                    <div class="center cursor-pointer">
                      {{ item.key }}
                      <el-icon class="text-purple ml-1"><Sort /></el-icon>
                    </div>
                    <template #dropdown>
                      <el-dropdown-menu>
                        <el-dropdown-item @click="handleTokenTypeChange(1, index)">
                          Authorization(Bearer)
                        </el-dropdown-item>
                        <el-dropdown-item @click="handleTokenTypeChange(2, index)">
                          Api-Key
                        </el-dropdown-item>
                        <el-dropdown-item @click="handleTokenTypeChange(3, index)">
                          X-API-key
                        </el-dropdown-item>
                        <el-dropdown-item @click="handleTokenTypeChange(4, index)">
                          Authorization(Basic)
                        </el-dropdown-item>
                      </el-dropdown-menu>
                    </template>
                  </el-dropdown>
                  <el-input
                    v-else
                    v-model="item.key"
                    :placeholder="t('mcp.instance.token.headersKey')"
                    class="flex-sub"
                  >
                  </el-input>
                  <span class="ml-2">:</span>
                </div>
              </el-col>
              <el-col :span="15">
                <div class="flex">
                  <el-input
                    v-model="item.value"
                    :placeholder="t('mcp.instance.token.headersValue')"
                    class="flex-sub"
                  ></el-input>
                  <div
                    v-if="tokenTypeOptions.some((tokenType) => item.key === tokenType.label)"
                    class="text-purple cursor-pointer ml-2"
                    @click="handleChangeBasic(index)"
                  >
                    {{ Number(item.tokenType) === 4 ? t('mcp.token.account') : '  ' }}
                  </div>
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
        <el-form-item :label="t('mcp.instance.token.tag')" prop="usages">
          <el-tag
            v-for="(tag, num) in formData.usages"
            :key="num"
            :closable="true"
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
    </el-scrollbar>
    <template #footer>
      <div class="center">
        <el-button @click="handleCancelToken" class="mr-4 w-25">{{ t('common.cancel') }}</el-button>
        <mcp-button @click="handleConfirmToken" class="w-25">{{ t('common.ok') }}</mcp-button>
      </div>
    </template>
  </el-dialog>
  <!-- user accountPassword -->
  <el-dialog v-model="userDataKey.visible" width="400px" top="30vh" :show-close="false">
    <el-form :model="userDataKey" class="p-4" label-width="80px">
      <el-form-item :label="t('login.username')" prop="username">
        <el-input
          v-model="userDataKey.username"
          :placeholder="t('login.message.username.required')"
        />
      </el-form-item>
      <el-form-item :label="t('login.password')" prop="password">
        <el-input
          v-model="userDataKey.password"
          type="password"
          :placeholder="t('login.message.password.required')"
        />
      </el-form-item>
    </el-form>
    <template #footer>
      <div class="center">
        <mcp-button @click="handleConfirmAccount" class="w-25">{{ t('common.ok') }}</mcp-button>
      </div>
    </template>
  </el-dialog>
</template>
<script setup lang="ts">
import { getToken } from '@/utils/system'
import { useUserStore } from '@/stores'
import { Plus, Sort, Minus } from '@element-plus/icons-vue'
import McpButton from '@/components/mcp-button/index.vue'

const { t } = useI18n()
const { userInfo } = useUserStore()
const dialogData = ref({
  visible: false,
  loading: false,
})
const formRef = ref()
const formData = ref({
  expireAt: '',
  headers: [{ key: 'Authorization', value: '' }] as Array<{
    key: string
    value: string
    tokenType?: number
  }>,
  usages: [] as string[],
})
const rules = ref({})
const tokenTypeOptions = [
  { label: 'Authorization', value: 1 },
  { label: 'Api-Key', value: 2 },
  { label: 'X-API-key', value: 3 },
  { label: 'Authorization', value: 4 },
]
const emit = defineEmits<{
  (
    e: 'on-confirm',
    data: {
      expireAt: string
      headers: Array<{ key: string; value: string }>
    },
  ): Promise<void>
  (e: 'on-cancel'): void
}>()
const userDataKey = ref({
  visible: false,
  username: '',
  password: '',
  index: 0,
})
const showTagInput = ref(false)
const inputValue = ref('')
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

// handle add header
const handleAddHeader = () => {
  formData.value.headers.push({ key: '', value: '' })
}
// handle token type change and clear token value
const handleTokenTypeChange = (tokenType: number, index: number) => {
  formData.value.headers[index].tokenType = tokenType
  formData.value.headers[index].key = tokenTypeOptions[tokenType - 1].label
  handleGetTokenValue(tokenType, index)
}
// handle get token value handle random token
const handleGetTokenValue = (tokenType: number, index: number) => {
  let tokenValue = ''
  if (tokenType === 1) {
    tokenValue =
      'Bearer ' +
      getToken(
        JSON.stringify({
          expireAt: formData.value.expireAt,
          userId: userInfo.userId,
          username: userInfo.username,
        }),
      )
  } else if (tokenType === 2) {
    tokenValue = getToken(
      JSON.stringify({
        expireAt: formData.value.expireAt,
        userId: userInfo.userId,
        username: userInfo.username,
      }),
    )
  } else if (tokenType === 3) {
    tokenValue = getToken(
      JSON.stringify({
        expireAt: formData.value.expireAt,
        userId: userInfo.userId,
        username: userInfo.username,
      }),
    )
  } else if (tokenType === 4) {
    tokenValue = 'Basic ' + btoa(`${userDataKey.value.username}:${userDataKey.value.password}`) // Base64 编码
  }
  formData.value.headers[index].value = tokenValue
}

const handleChangeBasic = (index: number) => {
  userDataKey.value.visible = !userDataKey.value.visible
  userDataKey.value.index = index
  // formData.value.token = ''
}

// handle delete header
const handleDeleteHeader = (index: number) => {
  formData.value.headers.splice(index, 1)
}

// confirm input account interchange token
const handleConfirmAccount = () => {
  handleGetTokenValue(4, userDataKey.value.index)
  userDataKey.value.visible = false
}
// handle cancel token
const handleCancelToken = () => {
  formData.value = {
    expireAt: '',
    headers: [],
    usages: [],
  }
  dialogData.value.visible = false
  emit('on-cancel')
}
// handle confirm token
const handleConfirmToken = () => {
  try {
    emit('on-confirm', formData.value)
    dialogData.value.loading = true
  } finally {
    dialogData.value.loading = false
  }
}
const finish = () => {
  dialogData.value.visible = false
  dialogData.value.loading = false
  formData.value = {
    expireAt: '',
    headers: [],
    usages: [],
  }
}

// init data
const init = () => {
  dialogData.value.visible = true
}

defineExpose({
  init,
  finish,
})
</script>
<style lang="scss" scoped>
:deep(.el-tag) {
  color: var(--ep-purple-color);
  border-color: var(--ep-border-color);
  .el-tag__close {
    color: var(--ep-purple-color);
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
