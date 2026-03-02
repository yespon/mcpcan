<template>
  <el-dialog
    v-model="dialogVisible"
    :title="t('dataPermission.title')"
    width="600px"
    align-center
    :show-close="true"
    destroy-on-close
    @closed="handleClosed"
  >
    <div v-loading="loading">
      <div class="mb-4">
        <div class="text-sm color-gray mb-2">
          {{ t('dataPermission.desc') }}
        </div>
        <div class="flex align-center mb-4">
          <span class="mr-2 font-bold">{{ t('dataPermission.resourceName') }}：</span>
          <span>{{ resourceName }}</span>
        </div>
        <div class="flex align-center mb-4">
          <span class="mr-2 font-bold">{{ t('dataPermission.resourceType') }}：</span>
          <el-tag>{{ resourceTypeLabel }}</el-tag>
        </div>
      </div>

      <!-- 预留：数据权限配置区域 -->
      <div class="data-permission-content">
        <el-empty :description="t('dataPermission.empty')"></el-empty>
      </div>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="dialogVisible = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">
          {{ t('common.ok') }}
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
/**
 * 数据权限弹窗组件（公共）
 *
 * 资源类型 resourceType 枚举：
 *   - instance:            MCP实例
 *   - openapi_package:     OpenAPI 文件
 *   - code_package:        代码包
 *   - instance_tokens:     令牌
 *   - intelligent_access:  智能体平台
 */

const { t } = useI18n()

const dialogVisible = ref(false)
const loading = ref(false)
const submitLoading = ref(false)
const resourceName = ref('')
const resourceId = ref('')
const resourceType = ref('')

const RESOURCE_TYPE_MAP: Record<string, string> = {
  instance: 'dataPermission.type.instance',
  openapi_package: 'dataPermission.type.openapiPackage',
  code_package: 'dataPermission.type.codePackage',
  instance_tokens: 'dataPermission.type.instanceTokens',
  intelligent_access: 'dataPermission.type.intelligentAccess',
}

const resourceTypeLabel = computed(() => {
  const key = RESOURCE_TYPE_MAP[resourceType.value]
  return key ? t(key) : resourceType.value
})

const emit = defineEmits<{
  (e: 'on-refresh'): void
}>()

export interface DataPermissionInitParams {
  /** 资源唯一标识 */
  id: string
  /** 资源显示名称 */
  name: string
  /** 资源类型: instance | openapi_package | code_package | instance_tokens | intelligent_access */
  type: string
}

/**
 * 初始化弹窗
 * @param params - 资源信息
 */
const init = (params: DataPermissionInitParams) => {
  resourceId.value = params.id
  resourceName.value = params.name
  resourceType.value = params.type
  dialogVisible.value = true
  // TODO: 根据 resourceId 和 resourceType 加载数据权限数据
}

/**
 * 提交数据权限配置
 */
const handleSubmit = async () => {
  try {
    submitLoading.value = true
    // TODO: 调用数据权限保存接口，传入 resourceId、resourceType
    dialogVisible.value = false
    emit('on-refresh')
  } finally {
    submitLoading.value = false
  }
}

/**
 * 弹窗关闭回调
 */
const handleClosed = () => {
  resourceName.value = ''
  resourceId.value = ''
  resourceType.value = ''
}

defineExpose({
  init,
})
</script>

<style scoped lang="scss">
.data-permission-content {
  min-height: 200px;
}
</style>
