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
      <el-col :span="24" class="collapsed">
        <el-scrollbar ref="scrollbarRef" height="50vh" always class="pr-4">
          <el-form
            ref="formRef"
            :model="formData"
            :rules="rules"
            label-width="auto"
            label-position="top"
            class="mx-2"
          >
            <el-form-item :label="t('mcp.instance.token.lifespan')" prop="expireAt">
              <template #label>
                <div class="center">
                  <span class="mr-2">{{ t('mcp.instance.token.lifespan') }}</span>
                  <el-button
                    class="base-btn"
                    type="primary"
                    size="small"
                    @click.stop="handleAddExpireAt(7)"
                    >7{{ t('mcp.instance.token.day') }}
                  </el-button>
                  <el-button
                    class="base-btn"
                    type="primary"
                    size="small"
                    @click.stop="handleAddExpireAt(15)"
                    >15{{ t('mcp.instance.token.day') }}
                  </el-button>
                  <el-button
                    class="base-btn"
                    type="primary"
                    size="small"
                    @click.stop="handleAddExpireAt(30)"
                    >30{{ t('mcp.instance.token.day') }}
                  </el-button>
                </div>
              </template>
              <el-date-picker
                ref="datePicker"
                v-model="formData.expireAt"
                type="datetime"
                value-format="x"
                :placeholder="t('mcp.instance.token.placeholderDate')"
                style="width: 100%"
                :disabled-date="(date: Date) => date.getTime() < Date.now()"
              ></el-date-picker>
            </el-form-item>
            <el-form-item :label="t('mcp.token.authentication')" prop="tokenType">
              <div class="w-full u-line-1" style="white-space: nowrap">
                Authorization：{{ formData.token }}
              </div>
            </el-form-item>

            <el-form-item prop="enabledTransport" class="enabledTransport">
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
                + New Tag
              </el-button>
            </el-form-item>
          </el-form>
        </el-scrollbar>
      </el-col>
    </el-row>
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
import { TokenType, type InstanceResult } from '@/types'
import McpButton from '@/components/mcp-button/index.vue'
import { getToken } from '@/utils/system'
import { useUserStore } from '@/stores'
import { Plus, Sort, Minus } from '@element-plus/icons-vue'

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

const { userInfo } = useUserStore()
const { t } = useI18n()
const formRef = ref()

const formData = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value),
})

const rules = reactive({})

const tokenTypeOptions = [
  { label: 'Authorization', value: 1 },
  { label: 'Api-Key', value: 2 },
  { label: 'X-API-key', value: 3 },
  { label: 'Authorization', value: 4 },
]

const userDataKey = ref({
  visible: false,
  username: '',
  password: '',
  index: 0,
})
const showTagInput = ref(false)
const inputValue = ref('')

const showClosAble = computed(() => {
  return (tag: string) => {
    return ![
      'dify_user_id',
      'dify_user_name',
      'dify_space_id',
      'dify_space_name',
      'intelligent_access_id',
      'intelligent_access_name',
      'intelligent_access_type',
      'default',
    ].some((keyword) => tag.includes(keyword))
  }
})

// handle add expire at
const handleAddExpireAt = (days: number) => {
  const expireDate = new Date()
  expireDate.setDate(expireDate.getDate() + days)
  formData.value = {
    ...formData.value,
    expireAt: expireDate.getTime(),
  }
}

// handle add header
const handleAddHeader = () => {
  formData.value = {
    ...formData.value,
    headers: [...formData.value.headers, { key: '', value: '' }],
  }
}

// handle token type change and clear token value
const handleTokenTypeChange = (tokenType: number, index: number) => {
  formData.value.headers[index].tokenType = tokenType
  formData.value.headers[index].key = tokenTypeOptions[tokenType - 1].label
  let roginData = null
  if (props.currentEditIndex !== null) {
    roginData = props.tokenList[props.currentEditIndex] as any
    if (tokenType === 4) {
      userDataKey.value.username = atob(roginData.headers.Authorization.split(' ')[1]).split(':')[0]
      userDataKey.value.password = atob(roginData.headers.Authorization.split(' ')[1]).split(':')[1]
    }
    return
  }
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
    visible: false,
    token: '',
    headers: [],
    tokenType: 1 as TokenType | null,
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
    headers: Object.fromEntries(
      formData.value.headers.map((header: any) => [header.key, header.value]),
    ),
  }

  emit('on-confirm', {
    ...tokenData,
    currentEditIndex: props.currentEditIndex,
  })

  formData.value = {
    visible: false,
    token: '',
    headers: [],
    tokenType: 1 as TokenType | null,
    enabled: true,
    expireAt: null,
    usages: [],
  }
}
// confirm input account interchange token
const handleConfirmAccount = () => {
  handleGetTokenValue(4, userDataKey.value.index)
  userDataKey.value.visible = false
}
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
