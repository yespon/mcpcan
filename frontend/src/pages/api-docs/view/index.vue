<template>
  <div>
    <div class="title flex justify-between">
      <div class="font-bold text-5">{{ t('api.pageDesc.detail') }} - {{ query.name }}</div>
      <div>
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
      <el-card :style="{ height: 'calc(100vh - 150px)' }">
        <template #header>
          <el-space>
            <span> {{ t('code.codeEditor') }}<span class="ml-4"></span> </span>
          </el-space>
        </template>
        <el-scrollbar ref="scrollbarRef" height="calc(100vh - 260px)" always>
          <div v-if="!currentContent">{{ t('api.pageDesc.empty') }}</div>
          <div class="text-conetnt" v-else>
            <!-- 图片显示 -->
            {{ currentContent }}
          </div>
        </el-scrollbar>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { CodeAPI } from '@/api/code/index'
import GlareHover from '@/components/Animation/GlareHover.vue'

const { t } = useI18n()
const $route = useRoute()
const { query } = $route
const currentContent = ref('')

/**
 * Handle download code package
 */
const handleDownload = async () => {
  const data = await CodeAPI.download({ ...query })
  const blobUrl = URL.createObjectURL(new Blob([data], { type: 'application/yaml' }))
  const link = document.createElement('a')
  link.href = blobUrl
  link.download = `${query.name}.yaml`
  document.body.appendChild(link)
  link.click()
}

const handleGetContent = async () => {
  const { data } = await CodeAPI.fileContent({ openapiFileId: query.id })
  currentContent.value = data
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
  height: calc(100vh - 260px);
  font-family: 'Monaco, Menlo, "Ubuntu Mono", monospace';
  font-size: 12px;
  line-height: 1.5;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-all;
  background-color: #000;
  padding: 10px;
  border-radius: 8px;
}
:deep(.el-card__body) {
  height: calc(100%);
  overflow-y: auto;
}
</style>
