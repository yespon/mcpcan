<template>
  <el-dialog
    v-model="dialogInfo.visible"
    width="1200px"
    align-center
    :show-close="false"
    header-class="header-border"
    footer-class="footer-border"
  >
    <template #title>{{ dialogInfo.title }}</template>
    <el-splitter>
      <el-splitter-panel size="50%" class="p-4">
        <el-card class="mb-3">
          <div class="flex items-center">
            <mcp-image
              :src="openapi"
              width="28"
              height="28"
              border-radius="4"
              class="mr-3"
            ></mcp-image>
            <div class="flex-sub">
              支持导入 OpenAPI 3.0、3.1 或 Swagger 2.0 数据格式的 JSON 或 YAML 文件。
            </div>
          </div>
        </el-card>
        <el-form
          ref="baseInfo"
          :model="formData"
          :rules="rules"
          label-width="auto"
          label-position="top"
        >
          <el-form-item :label="t('mcp.instance.formData.instanceName')" prop="name">
            <el-input
              v-model="formData.name"
              :placeholder="t('mcp.instance.formData.instanceName')"
            />
          </el-form-item>
          <el-form-item :label="t('mcp.template.formData.notes')" prop="notes">
            <el-input
              v-model="formData.notes"
              :rows="4"
              type="textarea"
              :placeholder="t('mcp.template.formData.notes')"
            />
          </el-form-item>
          <el-form-item :label="t('mcp.template.formData.icon')" prop="iconPath">
            <Upload v-model="formData.iconPath"></Upload>
          </el-form-item>
        </el-form>
        <div class="mt-8 color-gray text-3">注：通过OpenAPI文档导入的MCP服务将以STDIO协议访问</div>
      </el-splitter-panel>
      <el-splitter-panel size="50%" :min="600" class="p-4">
        <div class="flex-sub link-hover" v-if="!formData.openapiRaw">
          <div class="flex flex-col" style="height: 75vh">
            <el-upload
              class="upload-demo flex-sub"
              drag
              :action="action"
              :on-success="handleSuccess"
              :before-upload="handleBeforeUpload"
              :headers="headers"
              accept=".yaml, .JSON, application/yaml, application/JSON"
              :auto-upload="true"
              :show-file-list="false"
            >
              <div>
                <el-icon class="el-icon--upload"><upload-filled /></el-icon>
                <div class="el-upload__text">
                  {{ '本地上传' }}
                </div>
              </div>
            </el-upload>
            <div
              class="flex-sub select-api border-rd-1 mt-2 center cursor-pointer"
              @click="handleSelectDocs"
            >
              <div>
                <el-icon class="el-icon--upload" size="67"><Files /></el-icon>
                <div>从文档库选择</div>
              </div>
            </div>
          </div>
        </div>

        <div v-else class="p-3">
          <el-tree
            style="height: 75vh"
            :data="apiNodeList"
            show-checkbox
            node-key="id"
            :default-expand-all="true"
            :default-checked-keys="checkedKeys"
            :render-content="renderContent"
          />
        </div>
      </el-splitter-panel>
    </el-splitter>
    <template #footer>
      <div class="center">
        <el-button class="mr-4 w-25" @click="dialogInfo.visible = false">{{
          t('common.cancel')
        }}</el-button>
        <mcp-button class="w-25" @click="handleConfirm">{{ t('common.ok') }}</mcp-button>
      </div>
    </template>
  </el-dialog>
  <Select
    v-model="selectVisible"
    v-model:selected="formData.openapiFileId"
    red="opneAPISelect"
    :title="t('api.pageDesc.apiSelectTitle')"
    :options="docsList"
  ></Select>
</template>

<script lang="ts" setup>
import baseConfig from '@/config/base_config.ts'
import { Storage } from '@/utils/storage'
import { ElMessage } from 'element-plus'
import { UploadFilled, Files } from '@element-plus/icons-vue'
import McpImage from '@/components/mcp-image/index.vue'
import { openapi } from '@/utils/logo.ts'
import McpButton from '@/components/mcp-button/index.vue'
import Select from '@/components/mcp-select/index.vue'
import Upload from '@/components/upload/index.vue'
import yaml from 'js-yaml'
import { buildApiTree } from '@/utils/json.ts'
import { DocsAPI } from '@/api/api-docs'

const { t } = useI18n()
const dialogInfo = ref<any>({
  visible: false,
  loading: false,
  title: '导入数据',
})
const selectVisible = ref(false)
const docsList = ref<any>([])
const formData = ref({
  name: '',
  notes: '',
  iconPath: '',
  openapiRaw: '',
  openapiFileId: '',
})
const rules = ref({
  name: [{ required: true, message: t('mcp.instance.rules.name'), trigger: 'blur' }],
})
const apiNodeList = ref<any[]>([])
const checkedKeys = ref<any[]>([])

const action = ref(baseConfig.SERVER_BASE_URL + baseConfig.baseUrlVersion + '/market/code/upload')
const headers = ref({
  Authorization: `Bearer ${Storage.get('token')}`,
})

/**
 * handle Setect a openApi docs
 */
const handleSelectDocs = () => {
  selectVisible.value = true
}

/**
 * Get api list
 */
const handleGetAPIlist = async () => {
  const data = await DocsAPI.list({ page: 1, pageSize: 999 })
  docsList.value = data.list
}

/**
 * api list render
 * @param h
 * @param param1
 */
const renderContent = (h, { node, data }) => {
  return h(
    'div',
    {
      class: ' w-full grid grid-cols-10 ',
    },
    [
      h('div', { class: 'ellipsis-one col-span-8' }, node.method || node.label),
      h('div', { class: 'col-span-2 text-right pr-2' }, '其他信息'),
    ],
  )
}

const handleBeforeUpload = async (file: File) => {
  try {
    // 现代浏览器支持直接调用 file.text()
    const rawText = await file.text() // 原始文本内容
    // 保存原始文本以供后续使用
    formData.value.openapiRaw = rawText
    // 尝试解析为 JSON/YAML（容错）
    let doc: any = null
    try {
      doc = JSON.parse(rawText)
      apiNodeList.value = [{ id: 'root', label: '接口', children: buildApiTree(doc) }]
      // set checked keys to all node ids
      checkedKeys.value = []
      const collectIds = (nodes: any[]) => {
        nodes.forEach((n) => {
          if (n.id) checkedKeys.value.push(n.id)
          if (n.children && n.children.length) collectIds(n.children)
        })
      }
      collectIds(apiNodeList.value)
      console.log('解析为 JSON 成功parse')
    } catch (e) {
      console.log(e)
      // 不是 JSON，就当做 YAML 尝试解析
      try {
        doc = yaml.load(rawText)
        console.log('解析为 YAML 成功load', doc, Object.keys(doc.paths).length)
        apiNodeList.value = [{ id: 'root', label: '接口', children: buildApiTree(doc) }]
        checkedKeys.value = []
        const collectIds = (nodes: any[]) => {
          nodes.forEach((n) => {
            if (n.id) checkedKeys.value.push(n.id)
            if (n.children && n.children.length) collectIds(n.children)
          })
        }
        collectIds(apiNodeList.value)
        console.log(apiNodeList.value)
      } catch (yamlErr) {
        console.warn('无法解析为 JSON 或 YAML，可当作纯文本处理', yamlErr)
      }
    }

    // 如果你想在客户端先处理并阻止 Element Plus 直接上传，返回 false
    // return false

    // 若不阻止上传，返回 true 或不返回（或者 return file）
    return true
  } catch (err) {
    console.error('读取文件出错', err)
    // 阻止上传
    return false
  }
}

/**
 *  Handle update success
 * @param response
 */
const handleSuccess = (response: { code: number; data: { path: string } }) => {
  if (response.code !== 0) {
    return
  }
  ElMessage.success(t('action.upload'))
}

/**
 * Handle confirm and submit data
 */
const handleConfirm = () => {
  console.log(checkedKeys.value)
}

/**
 * Init dialog data
 */
const init = () => {
  dialogInfo.value.visible = true
  handleGetAPIlist()
}

defineExpose({
  init,
})
</script>

<style lang="scss" scoped>
.api-info {
  border: 1px dashed var(--ep-purple-color);
}
.upload-demo {
  width: 100%;
  color: var(--ep-purple-color);
  :deep(.el-upload.is-drag) {
    height: 100%;
  }
  :deep(.el-upload-dragger) {
    height: 100%;
    border: 1px dashed var(--ep-purple-color);
    display: flex;
    align-items: center;
    justify-content: center;
    &:hover {
      border-color: var(--ep-purple-color-hover);
      .el-icon--upload {
        color: var(--ep-purple-color-hover);
      }
      .el-upload__text {
        color: var(--ep-purple-color-hover);
      }
    }
  }
  .el-icon--upload {
    color: var(--ep-purple-color);
  }
  .el-upload__text {
    color: var(--ep-purple-color);
  }
}
.select-api {
  border: 1px dashed var(--ep-purple-color);
  color: var(--ep-purple-color);
  &:hover {
    border-color: var(--ep-purple-color-hover);
    color: var(--ep-purple-color-hover);
  }
}
:deep(.el-tree-node__content) {
  padding: 0 0 1px 4px;
  margin: 2px 0;
  &:hover {
    background-color: var(--ep-bg-purple-color);
    border-radius: 4px;
  }
}
:deep(.el-checkbox__input.is-checked .el-checkbox__inner) {
  background-color: var(--ep-purple-color);
  border-color: var(--ep-purple-color);
  border-radius: 4px;
}
:deep(.el-checkbox__input.is-indeterminate .el-checkbox__inner) {
  background-color: var(--ep-purple-color);
}
:deep(.el-checkbox__input.is-indeterminate .el-checkbox__inner) {
  border-color: var(--ep-purple-color);
}
:deep(.el-checkbox__inner) {
  &:hover {
    border-color: var(--ep-purple-color);
  }
}
</style>

<style lang="scss">
.el-dialog__header.header-border {
  background-color: transparent !important;
  border-bottom: 1px solid var(--el-border-color-light) !important;
}
.el-dialog__footer.footer-border {
  background-color: transparent !important;
  border-top: 1px solid var(--el-border-color-light) !important;
}
</style>
