<template>
  <el-form
    ref="formRef"
    :model="localFormData"
    label-width="auto"
    label-position="top"
    class="mx-2"
  >
    <el-form-item :label="'网关认证'" prop="tokenType">
      <div class="w-full u-line-1" style="white-space: nowrap">
        Authorization：{{ localFormData.token }}
      </div>
    </el-form-item>
    <el-form-item prop="enabledTransport" class="enabledTransport">
      <template #label>
        <div class="w-full flex justify-between items-center">
          <span class="mr-2"> 透传 {{ 'Headers' }} </span>
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
        v-for="(item, index) in localFormData.headers"
        :key="index"
        class="flex items-center my-2 pr-3"
      >
        <el-row :gutter="12" class="flex-sub align-center">
          <el-col :span="7">
            <div class="flex h-full items-center justify-end">
              <el-dropdown
                v-if="index === 0"
                trigger="click"
                class="h-full w-full flex justify-end"
                :show-arrow="false"
              >
                <div class="center cursor-pointer">
                  {{ item.key }}
                  <el-icon class="text-purple ml-1"><Sort /></el-icon>
                </div>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item @click="handleTokenTypeChange(1)">
                      Authorization(Bearer)
                    </el-dropdown-item>
                    <el-dropdown-item @click="handleTokenTypeChange(2)"> Api-Key </el-dropdown-item>
                    <el-dropdown-item @click="handleTokenTypeChange(3)">
                      X-API-key
                    </el-dropdown-item>
                    <el-dropdown-item @click="handleTokenTypeChange(4)">
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
          <el-col :span="15" class="flex">
            <el-input
              v-model="item.value"
              :placeholder="t('mcp.instance.token.headersValue')"
              class="flex-sub"
            ></el-input>
          </el-col>
          <el-col :span="2">
            <div v-if="index === 0" class="text-purple cursor-pointer" @click="handleChangeBasic">
              {{ Number(localFormData.tokenType) === 4 ? '账号' : '  ' }}
            </div>
            <div
              v-else
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
      <el-input-tag
        v-model="localFormData.usages"
        collapse-tags
        collapse-tags-tooltip
        :max-collapse-tags="3"
        clearable
        draggable
        tag-type="primary"
        tag-effect="plain"
        :placeholder="t('mcp.instance.token.placeholderTag')"
        class="tag-input"
      >
        <template #tag="{ value }">
          <div class="flex items-center">
            <span>{{ value }}</span>
          </div>
        </template>
      </el-input-tag>
    </el-form-item>
  </el-form>
  <el-dialog v-model="userDataKey.visible" width="400px" top="30vh" :show-close="false">
    <el-form :model="userDataKey" class="p-4" label-width="80px">
      <el-form-item label="用户名" prop="username">
        <el-input v-model="userDataKey.username" placeholder="请输入用户名" />
      </el-form-item>
      <el-form-item label="密码" prop="password">
        <el-input v-model="userDataKey.password" type="password" placeholder="请输入密码" />
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
import { Plus, Sort, Minus } from '@element-plus/icons-vue'
import { getToken } from '@/utils/system'
import McpButton from '@/components/mcp-button/index.vue'
import { useUserStore } from '@/stores'

const props = defineProps<{ formData: any }>()
const emit = defineEmits(['update:formData'])
const { t } = useI18n()
const { userInfo } = useUserStore()

const localFormData = computed({
  get: () => {
    return props.formData
  },
  set: (val) => {
    const emitData = { ...val }
    emit('update:formData', emitData)
  },
})
const tokenTypeOptions = [
  { label: 'Authorization', value: 1 },
  { label: 'Api-Key', value: 2 },
  { label: 'X-API-key', value: 3 },
  { label: 'Authorization', value: 4 },
]
// 账号密码弹窗数据
const userDataKey = ref({ visible: false, username: '', password: '' })

// handle add header
const handleAddHeader = () => {
  localFormData.value.headers.push({ key: '', value: '' })
  emit('update:formData', { ...localFormData.value })
}
const handleDeleteHeader = (index: number) => {
  localFormData.value.headers.splice(index, 1)
  emit('update:formData', { ...localFormData.value })
}
// handle token type change and clear token value
const handleTokenTypeChange = (tokenType: number) => {
  localFormData.value.tokenType = tokenType
  localFormData.value.headers[0].key = tokenTypeOptions[tokenType - 1].label
  if (localFormData.value.headers && localFormData.value.headers[0]) {
    localFormData.value.headers[0].key = tokenTypeOptions[tokenType - 1].label
  }
  emit('update:formData', { ...localFormData.value })
  handleGetTokenValue()
}

// handle get token value
const handleGetTokenValue = () => {
  if (Number(localFormData.value.tokenType) === 1) {
    localFormData.value.headers[0].value =
      'Bearer ' +
      getToken(
        JSON.stringify({
          expireAt: localFormData.value.expireAt,
          userId: userInfo.userId,
          username: userInfo.username,
        }),
      )
  } else if (Number(localFormData.value.tokenType) === 2) {
    localFormData.value.headers[0].value = getToken(
      JSON.stringify({
        expireAt: localFormData.value.expireAt,
        userId: userInfo.userId,
        username: userInfo.username,
      }),
    )
  } else if (Number(localFormData.value.tokenType) === 3) {
    localFormData.value.headers[0].value = getToken(
      JSON.stringify({
        expireAt: localFormData.value.expireAt,
        userId: userInfo.userId,
        username: userInfo.username,
      }),
    )
  } else if (Number(localFormData.value.tokenType) === 4) {
    localFormData.value.headers[0].value =
      'Basic ' + btoa(`${userDataKey.value.username}:${userDataKey.value.password}`) // Base64 编码
  }
  emit('update:formData', { ...localFormData.value })
}

const handleChangeBasic = () => {
  userDataKey.value.visible = !userDataKey.value.visible
}

const handleConfirmAccount = () => {
  handleGetTokenValue()
  userDataKey.value.visible = false
}
const init = () => {
  if (!localFormData.value.headers[0]) {
    localFormData.value.headers[0] = { key: 'Authorization', value: localFormData.value.token }
  }
}
onMounted(init)
</script>
<style lang="scss" scoped>
:deep(.el-form-item__label) {
  width: 100%;
}
:deep(.el-form-item__content) {
  display: block;
  width: 100%;
}
.tag-input {
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
}
.delete-header {
  width: 24px;
  height: 24px;
  border-radius: 4px;
}
</style>
