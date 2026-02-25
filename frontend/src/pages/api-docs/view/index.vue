<template>
  <div>
    <div class="title flex justify-between">
      <div class="font-bold text-5">
        <el-link v-if="layout" link @click="handleBack" class="link-hover mr-4" underline="never">
          <el-icon class="mr-2">
            <i class="icon iconfont MCP-fanhui"></i>
          </el-icon>
          {{ t('common.back') }} </el-link
        >{{ t('api.pageDesc.detail') }} - {{ query.name }}
      </div>
      <div v-auth="'mcpcan_resource_manage:download_openapi'">
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
          <el-button type="primary" @click="handleDownload" class="base-btn">
            <el-icon class="mr-2"><i class="icon iconfont MCP-xiazai"></i></el-icon>
            {{ t('api.action.download') }}</el-button
          >
        </GlareHover>
      </div>
    </div>
    <div class="mt-4">
      <el-card>
        <template #header>
          <el-space>
            <span> {{ t('code.codeEditor') }}<span class="ml-4"></span> </span>
          </el-space>
        </template>
        <el-scrollbar ref="scrollbarRef" height="calc(100vh - 260px)" class="border-rd-1" always>
          <div v-if="!currentContent">{{ t('api.pageDesc.empty') }}</div>
          <div class="text-conetnt" v-else>
            {{ currentContent }}
          </div>
        </el-scrollbar>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { DocsAPI } from '@/api/api-docs/index'
import GlareHover from '@/components/Animation/GlareHover.vue'
import { useRouterHooks } from '@/utils/url'

const { t } = useI18n()
const $route = useRoute()
const layout = useLayout()
const { query } = $route
const currentContent = ref('')
const { jumpBack } = useRouterHooks()

/**
 * Handle download code package
 */
const handleDownload = async () => {
  const response = await DocsAPI.download({ ...query })
  const blobUrl = URL.createObjectURL(
    new Blob([response.data], { type: response.headers['content-type'] }),
  )
  const link = document.createElement('a')
  link.href = blobUrl
  link.download =
    response.headers['content-disposition']?.split('filename=')[1]?.match(/filename=("?)(.*?)\1/) ||
    query.name
  document.body.appendChild(link)
  link.click()
}

const handleGetContent = async () => {
  const { content } = await DocsAPI.fileContent({ openapiFileId: query.id })
  currentContent.value = content
}
// back last class page
const handleBack = () => {
  jumpBack()
}

/**
 * Handle init page data
 */
const init = () => {
  handleGetContent()
}

onMounted(init)
</script>

<style lang="scss" scoped>
.text-conetnt {
  font-family: 'Monaco, Menlo, "Ubuntu Mono", monospace';
  font-size: 12px;
  line-height: 1.5;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-all;
  background-color: var(--ep-bg-color-deep);
  padding: 10px;
  border-radius: 8px;
}
:deep(.el-card__body) {
  height: calc(100%);
  overflow-y: auto;
}
</style>
