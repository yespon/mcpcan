<template>
  <el-dialog
    v-model="dialogInfo.visible"
    width="720px"
    top="10vh"
    :show-close="false"
    @close="emit('on-refresh')"
  >
    <template #header>
      <div class="center">{{ dialogInfo.title }}</div>
    </template>
    <el-scrollbar ref="scrollbarRef" max-height="75vh" always class="pb-4">
      <el-row :gutter="12" class="pl-2 pb-4 w-full" align="middle">
        <el-col :span="4" class="mb-4">Header</el-col>
        <el-col :span="10" class="mb-4">{{ t('mcp.instance.token.tokenKey') }}</el-col>
        <el-col :span="8" class="mb-4">{{ t('mcp.instance.token.expireAt') }}</el-col>
        <template v-for="(token, index) in dialogInfo.instanceInfo.tokens" :key="token.token">
          <el-col :span="4" class="mb-4">
            <span>Authorization</span>
          </el-col>
          <el-col :span="10" class="mb-4">
            <el-tooltip effect="dark" placement="top" class="ml-6" :raw-content="true" width="300">
              <div class="flex">
                <div class="flex-sub ml-2 ellipsis-one">
                  {{ token.token || t('mcp.instance.form.tokenPlaceholder') }}
                </div>
                <el-icon class="cursor-pointer base-btn-link copy-icon"><CopyDocument /></el-icon>
              </div>

              <template #content>
                <div>{{ token.token || t('mcp.instance.form.tokenPlaceholder') }}</div>
              </template>
            </el-tooltip>
            <span></span>
          </el-col>
          <el-col :span="8" class="mb-4">
            <el-date-picker
              v-model="token.expireAt"
              type="datetime"
              value-format="x"
              :placeholder="
                t(
                  index
                    ? 'mcp.instance.token.placeholderDate'
                    : 'mcp.instance.token.placeholderAlways',
                )
              "
              :disabled="!index"
              :disabled-date="(date: Date) => date.getTime() < Date.now()"
              @change="handleConfirmDate($event, index)"
            />
          </el-col>
          <el-col :span="2" class="mb-4">
            <el-button v-if="index" type="danger" link @click="handleDeleteToken(index)">
              {{ t('mcp.instance.action.delete') }}
            </el-button>
          </el-col>
        </template>
      </el-row>
      <el-button class="add-token" :icon="Plus" plain @click="handleAddToken">
        {{ t('mcp.instance.formData.addToken') }}
      </el-button>
    </el-scrollbar>
  </el-dialog>
</template>

<script setup lang="ts">
import { Plus, CopyDocument } from '@element-plus/icons-vue'
import { InstanceAPI } from '@/api/mcp/instance'
import type { InstanceResult } from '@/types'
import { getToken } from '@/utils/system'
import { useUserStore } from '@/stores'

const { t } = useI18n()
const { userInfo } = useUserStore()
const dialogInfo = ref({
  loading: false,
  visible: false,
  title: t('mcp.instance.token.title'),
  instanceInfo: {
    instanceId: '',
    tokens: [] as Array<{
      token: string
      expireAt: number
      publishAt: number
      usages: string[]
    }>,
  },
})

const emit = defineEmits<{
  (e: 'on-refresh'): void
}>()

/**
 * Handle add a new token row
 */
const handleAddToken = () => {
  dialogInfo.value.instanceInfo.tokens.push({
    token: '',
    expireAt: 0,
    publishAt: new Date().getTime(),
    usages: [],
  })
}

/**
 * @param index - token index
 * Handle delete a token row
 */
const handleDeleteToken = (index: number) => {
  dialogInfo.value.instanceInfo.tokens.splice(index, 1)
  handleSaveTokens()
}

/**
 * Handle confirm date change
 */
const handleConfirmDate = (date: Date, index: number) => {
  dialogInfo.value.instanceInfo.tokens[index].token =
    'Bearer ' +
    getToken(
      JSON.stringify({
        userId: userInfo.userId,
        username: userInfo.username,
        expireAt: date,
      }),
    )
  handleSaveTokens()
}

/**
 * Handle save tokens
 */
const handleSaveTokens = async () => {
  try {
    dialogInfo.value.loading = true
    await InstanceAPI.updateInstanceTokens({
      instanceId: dialogInfo.value.instanceInfo.instanceId,
      tokens: dialogInfo.value.instanceInfo.tokens,
    })
  } catch (error) {
    throw error
  } finally {
    dialogInfo.value.loading = false
  }
}

/**
 * Init the dialog with instance row data
 * @param row - instance row data
 */
const init = (row: InstanceResult) => {
  dialogInfo.value.instanceInfo = row
  dialogInfo.value.visible = true
}

defineExpose({
  init,
})
</script>

<style lang="scss" scoped>
.add-token {
  width: 100%;
  border: 1px dashed var(--el-border-color);
}
</style>
