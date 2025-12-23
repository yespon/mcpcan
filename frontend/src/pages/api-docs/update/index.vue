<template>
  <div class="common-layout">
    <el-container class="h-full">
      <el-header class="p-0 flex align-center justify-between">
        <div class="title">
          <div>{{ t('api.action.upload') }}</div>
          <div class="flex desc align-end">
            <div class="ml-4">JSON {{ t('api.pageDesc.openAI') }} (.json)</div>
            <div class="ml-2">YAML {{ t('api.pageDesc.openAI') }} (.yaml)</div>
          </div>
        </div>
        <el-button v-if="layout" @click="handleBack" class="link-hover">
          <el-icon class="mr-2">
            <i class="icon iconfont MCP-fanhui"></i>
          </el-icon>
          {{ t('common.back') }}
        </el-button>
      </el-header>
      <el-main>
        <div class="flex flex-direction h-full">
          <div class="flex-sub center mt-8 link-hover">
            <el-upload
              class="upload-demo"
              drag
              :action="action"
              multiple
              :on-success="handleSuccess"
              :headers="headers"
              accept=".json, .yaml, application/json, application/yaml"
            >
              <el-icon class="el-icon--upload"><upload-filled /></el-icon>
              <div class="el-upload__text">
                {{ t('code.desc.suport') }}
              </div>
            </el-upload>
          </div>

          <div class="footer">
            {{ t('code.desc.describe') }}
            <div class="desc">
              <div class="ml-8 mt-2">{{ t('api.pageDesc.text1') }}</div>
              <div class="ml-8 mt-2">{{ t('api.pageDesc.text2') }}</div>
              <div class="ml-8 mt-2">{{ t('api.pageDesc.text3') }}</div>
              <div class="ml-8 mt-2">{{ t('api.pageDesc.text4') }}</div>
              <div class="ml-8 mt-2">{{ t('api.pageDesc.text5') }}</div>
            </div>
          </div>
        </div>
      </el-main>
    </el-container>
  </div>
</template>

<script setup lang="ts">
import { UploadFilled } from '@element-plus/icons-vue'
import baseConfig from '@/config/base_config.ts'
import { ElMessage } from 'element-plus'
import { Storage } from '@/utils/storage'
import { useRouterHooks } from '@/utils/url'

const { t } = useI18n()
const layout = useLayout()
const { jumpBack } = useRouterHooks()
const action = ref(
  baseConfig.SERVER_BASE_URL + (window as any).__APP_CONFIG__?.API_BASE + '/market/openapi/upload',
)
const headers = ref({
  Authorization: `Bearer ${Storage.get('token')}`,
})
const handleSuccess = (response: { code: number; data: { path: string } }) => {
  if (response.code !== 0) {
    return
  }
  ElMessage.success(t('action.upload'))
}
// back last class page
const handleBack = () => {
  jumpBack()
}
</script>

<style lang="scss" scoped>
.common-layout {
  width: 100vm;
  height: 100%;
  .el-header {
    padding: 0;
  }
  .el-main {
    padding: 0;
  }
  .el-footer {
    padding: 0;
  }
}
.title {
  display: flex;
  font-size: 20px;
  font-weight: 600;
  .desc {
    font-size: 16px;
    color: #999999;
    font-weight: 400;
  }
}
.upload-demo {
  width: 540px;
  color: var(--el-color-primary);
  :deep(.el-upload-dragger) {
    border: 1px dashed var(--el-color-primary);
    &:hover {
      border-color: var(--el-color-primary-hover);
      .el-icon--upload {
        color: var(--el-color-primary-hover);
      }
      .el-upload__text {
        color: var(--el-color-primary-hover);
      }
    }
  }
  .el-icon--upload {
    color: var(--el-color-primary);
  }
  .el-upload__text {
    color: var(--el-color-primary);
  }
}
.footer {
  font-family:
    PingFangSC,
    PingFang SC;
  font-weight: 600;
  font-size: 20px;
  // color: #cccccc;
  line-height: 28px;
  .desc {
    font-family:
      PingFangSC,
      PingFang SC;
    font-weight: 400;
    font-size: 14px;
    color: #999999;
    line-height: 20px;
    text-align: left;
    font-style: normal;
  }
}
</style>
