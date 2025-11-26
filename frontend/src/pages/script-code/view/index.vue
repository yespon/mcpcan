<template>
  <div>
    <div class="title flex justify-between">
      <div class="font-bold text-5">{{ t('code.detail') }} - {{ query.name }}</div>
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
            {{ t('code.action.download') }}</el-button
          >
        </GlareHover>
      </div>
    </div>
    <div class="mt-4">
      <el-row :gutter="32" :style="{ height: 'calc(100vh - 150px)' }">
        <el-col :span="4">
          <el-card :style="{ height: 'calc(100vh - 150px)' }">
            <template #header>
              <span>{{ t('code.fileTree') }}</span>
            </template>
            <el-tree
              :style="{ overflow: 'auto' }"
              :data="fileTree"
              :props="defaultProps"
              @node-click="handleNodeClick"
              default-expand-all
              :expand-on-click-node="false"
            >
              <!-- 自定义节点内容 -->
              <template #default="{ data }">
                <div class="flex items-center">
                  <!-- 根据节点层级显示不同图标 -->
                  <component :is="getNodeIcon(data)" class="ml-2" :size="4" />
                  <span class="ml-2">{{ data.label }}</span>
                </div>
              </template>
            </el-tree>
          </el-card>
        </el-col>
        <el-col :span="20" v-loading="contentLoading">
          <el-card :style="{ height: '100%' }">
            <template #header>
              <el-space>
                <span>
                  {{ t('code.codeEditor') }}<span class="ml-4">{{ currentDir.label }}</span>
                </span>
              </el-space>
            </template>
            <el-scrollbar ref="scrollbarRef" height="calc(100vh - 260px)" always>
              <div v-if="!currentContent">{{ t('code.empty') }}</div>
              <!-- 图片预览 -->
              <div v-else-if="isImageFile" class="image-preview">
                <img :src="imageSrc" :alt="currentDir.label" class="preview-image" />
              </div>
              <!-- 文本内容显示 -->
              <div class="text-conetnt" v-else>
                {{ currentContent }}
              </div>
            </el-scrollbar>
          </el-card>
        </el-col>
      </el-row>
    </div>
  </div>
</template>

<script setup lang="ts">
import {
  Notebook,
  Document,
  Picture,
  VideoPlay,
  Headset,
  Box as Collection,
} from '@element-plus/icons-vue'
import { CodeAPI } from '@/api/code/index'
import Files from '../modules/files.vue'
import GlareHover from '@/components/Animation/GlareHover.vue'

defineOptions({
  name: 'CodePackageView',
})

const { t } = useI18n()
const $route = useRoute()
const { query } = $route
const contentLoading = ref(false)
const currentContent = ref('')
const currentDir = ref<any>({})
const fileTree = ref<any>([])
const imageSrc = ref('')

/**
 * file tree config
 */
const defaultProps = {
  children: 'children',
  label: 'label',
  isLeaf: 'isLeaf',
}

/**
 * Handle icon by fileName
 * @param fileName
 * @returns - return file icon
 */
const getNodeIcon = (nodeData: any) => {
  if (!nodeData.isLeaf) return Files
  const fileName = nodeData.label
  const ext = fileName.split('.').pop()?.toLowerCase() || ''
  switch (ext) {
    case 'js':
    case 'ts':
    case 'jsx':
    case 'tsx':
    case 'py':
    case 'txt':
    case 'xml':
    case 'yml':
    case 'yaml':
    case 'json':
    case 'dxt':
    case 'sh':
      return Document
    case 'png':
    case 'jpg':
    case 'jpeg':
    case 'gif':
    case 'svg':
      return Picture
    case 'mp4':
    case 'mov':
    case 'avi':
      return VideoPlay
    case 'mp3':
    case 'wav':
    case 'flac':
      return Headset
    case 'zip':
    case 'rar':
    case 'tar':
    case 'gz':
      return Collection
    case 'txt':
    case 'md':
    case 'markdown':
      return Notebook
    default:
      return Files // default
  }
}

/**
 * interchange tree data
 */
const convertToTreeData = (nodes: any) => {
  return nodes.map((node: any) => ({
    label: node.name,
    isLeaf: !node.isDir,
    originalData: node, // save origin data
    children: node.children ? convertToTreeData(node.children) : undefined,
  }))
}

/**
 * Handle download code package
 */
const handleDownload = async () => {
  const response = await CodeAPI.download({ ...query })
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

/**
 * Handle node click
 */
const handleNodeClick = (data: any) => {
  if (data.originalData.isDir) {
    return
  } else {
    currentDir.value = data
  }
  handleGetContent(data.originalData.path)
}

/**
 * Check if current file is an image
 */
const isImageFile = computed(() => {
  if (!currentDir.value?.label) return false
  const fileName = currentDir.value.label
  const ext = fileName.split('.').pop()?.toLowerCase() || ''
  const imageExts = ['png', 'jpg', 'jpeg', 'gif', 'svg', 'webp', 'bmp', 'ico']
  return imageExts.includes(ext)
})

/**
 * Convert string content to base64 image data URL
 * Note: Backend now returns base64-encoded string for image files to avoid data corruption
 * For backward compatibility, this function also handles corrupted string data
 */
const convertToImageSrc = (content: string, fileName: string): string => {
  try {
    if (!content) return ''

    // 获取文件扩展名以确定MIME类型
    const ext = fileName.split('.').pop()?.toLowerCase() || ''
    const mimeTypes: Record<string, string> = {
      png: 'image/png',
      jpg: 'image/jpeg',
      jpeg: 'image/jpeg',
      gif: 'image/gif',
      svg: 'image/svg+xml',
      webp: 'image/webp',
      bmp: 'image/bmp',
      ico: 'image/x-icon',
    }
    const mimeType = mimeTypes[ext] || 'image/png'

    // 如果content已经是base64字符串或data URL，直接使用
    if (content.startsWith('data:image/')) {
      return content
    }

    // 如果content已经是base64字符串（没有data:前缀），添加前缀
    // 后端现在对图片文件返回base64编码，所以直接使用
    const trimmedContent = content.trim()
    if (/^[A-Za-z0-9+/=]+$/.test(trimmedContent) && trimmedContent.length > 100) {
      return `data:${mimeType};base64,${trimmedContent}`
    }

    // 后端使用 string([]byte) 转换二进制数据
    // 在Go中，当[]byte转换为string时，如果字节不是有效的UTF-8，会被替换为替换字符(U+FFFD)或转义
    // 在前端，我们需要将字符串转换回字节数组
    // 使用 TextEncoder 和 TextDecoder 可能不够准确，因为数据可能已经损坏

    // 尝试直接使用 charCodeAt 转换（适用于单字节字符）
    // 但需要注意：如果字符串包含多字节UTF-8字符，这种方法会失败
    const bytes = new Uint8Array(content.length)
    let validBytes = 0
    for (let i = 0; i < content.length; i++) {
      const charCode = content.charCodeAt(i)
      // 只处理单字节字符（0-255）
      if (charCode <= 255) {
        bytes[validBytes++] = charCode
      } else {
        // 如果是多字节字符，尝试提取字节
        // 这通常意味着数据已经损坏，但我们可以尝试恢复
      }
    }

    // 如果有效字节数少于原始长度，说明数据可能已损坏
    if (validBytes < content.length * 0.9) {
      console.warn('Significant data loss detected during conversion')
    }

    // 截取有效字节
    const validBytesArray = bytes.slice(0, validBytes)

    // 将字节数组转换为base64
    let binaryString = ''
    const chunkSize = 8192 // 分块处理，避免栈溢出
    for (let i = 0; i < validBytesArray.length; i += chunkSize) {
      const chunk = validBytesArray.slice(i, i + chunkSize)
      binaryString += String.fromCharCode.apply(null, Array.from(chunk))
    }
    const base64 = btoa(binaryString)
    return `data:${mimeType};base64,${base64}`
  } catch (error) {
    console.error('Failed to convert image content:', error)
    return ''
  }
}

/**
 * Handle get file content by API
 */
const handleGetContent = async (filePath: string) => {
  try {
    contentLoading.value = true
    const data = await CodeAPI.fileContent({
      packageId: query.id,
      filePath,
    })
    currentContent.value = data.content

    // 如果是图片文件，转换为可显示的格式
    // 后端现在对图片文件返回base64编码，前端直接转换为data URL
    if (isImageFile.value && data.content) {
      imageSrc.value = convertToImageSrc(data.content, currentDir.value.label)
      // 如果转换失败，显示错误提示
      if (!imageSrc.value) {
        console.error('Failed to convert image content')
        currentContent.value = `[图片文件无法显示：数据格式错误]`
      }
    } else {
      imageSrc.value = ''
    }
  } catch (error) {
    console.error('Failed to get file content:', error)
    currentContent.value = ''
    imageSrc.value = ''
  } finally {
    contentLoading.value = false
  }
}

/**
 * Handle get code package file tree by API
 */
const handleGetFileTree = async () => {
  const data = await CodeAPI.fileTree({ packageId: query.id })
  fileTree.value = convertToTreeData([data.fileStructure])
}

/**
 * Handle init page data
 */
const init = () => {
  handleGetFileTree()
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
  background-color: var(--ep-bg-color-deep);
  padding: 10px;
  border-radius: 8px;
}
.image-preview {
  display: flex;
  justify-content: center;
  align-items: center;
  height: calc(100vh - 260px);
  padding: 20px;
  background-color: var(--ep-bg-color-deep);
  border-radius: 8px;
  overflow: auto;
}
.preview-image {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
  border-radius: 4px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}
:deep(.el-card__body) {
  height: calc(100%);
  overflow-y: auto;
}
/* 图标样式：区分文件夹和文件 */
.folder-icon {
  color: #f59e0b; // 文件夹橙色
  margin-right: 4px;
}
.file-icon {
  color: #60a5fa; // 文件蓝色（可根据需要调整）
  margin-right: 4px;
}
/* 修复树节点文字与图标对齐 */
:deep(.el-tree-node__content) {
  align-items: center !important;
  padding: 4px 0 !important;
}
</style>
