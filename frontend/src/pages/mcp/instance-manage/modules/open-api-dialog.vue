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
    <el-splitter v-loading="dialogInfo.loading">
      <el-splitter-panel size="50%" class="p-4">
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
          <el-form-item :label="t('mcp.instance.formData.environmentId')" prop="environmentId">
            <el-select
              v-model="formData.environmentId"
              :placeholder="t('mcp.instance.formData.environmentId')"
            >
              <el-option
                v-for="(env, index) in envList"
                :key="index"
                :label="env.name"
                :value="env.id"
              />
            </el-select>
          </el-form-item>
          <el-form-item :label="t('mcp.instance.openApi.serviceUrl')" prop="openapiBaseUrl">
            <el-input
              v-model="formData.openapiBaseUrl"
              :placeholder="t('mcp.instance.openApi.serviceUrl')"
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
        <div class="mt-8 color-gray text-3">{{ t('mcp.instance.openApi.tips') }}</div>
      </el-splitter-panel>
      <el-splitter-panel size="50%" :min="600" class="p-4">
        <div class="flex-sub link-hover" v-if="!formData.openapiFileID">
          <div class="flex flex-col" style="height: 75vh">
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
                  {{ t('mcp.instance.openApi.support') }}
                </div>
              </div>
            </el-card>
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
                  {{ t('mcp.instance.openApi.localFile') }}
                </div>
              </div>
            </el-upload>
            <div
              class="flex-sub select-api border-rd-1 mt-2 center cursor-pointer"
              @click="handleSelectDocs"
            >
              <div>
                <el-icon class="el-icon--upload" size="67"><Files /></el-icon>
                <div>{{ t('mcp.instance.openApi.selectDocs') }}</div>
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
            :default-checked-keys="defaultCheckedKeys"
            ref="apiTreeRef"
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
    v-model:selected="formData.openapiFileID"
    red="opneAPISelect"
    :title="t('api.pageDesc.apiSelectTitle')"
    @confirm="handleGetAPIDetail"
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
import { useMcpStoreHook, useUserStore } from '@/stores'
import { InstanceAPI } from '@/api/mcp/instance'
import { getToken } from '@/utils/system'
import { SourceType, TokenType } from '@/types'
import { ElTooltip } from 'element-plus'

const { userInfo } = useUserStore()
const { envList } = toRefs(useMcpStoreHook())
const { handleGetEnvList } = useMcpStoreHook()
const { t } = useI18n()
const emit = defineEmits(['on-refresh'])
const dialogInfo = ref<any>({
  visible: false,
  loading: false,
  title: t('mcp.instance.openApi.importTitle'),
})
const selectVisible = ref(false)
const docsList = ref<any>([])
const formData = ref({
  instanceId: '',
  name: '',
  notes: '',
  iconPath: '',
  environmentId: '',
  openapiBaseUrl: '',
  openapiFileID: '',
  chooseOpenapiFileID: '', // 选择的文档库文件ID
  sourceType: SourceType.OPENAPI,
  tokens: [
    {
      enabledTransport: false,
      expireAt: '',
      headers: [],
      publishAt: new Date().getTime(),
      token:
        'Bearer ' +
        getToken(
          JSON.stringify({
            expireAt: Date.now(),
            userId: userInfo.userId,
            username: userInfo.username,
          }),
        ),
      tokenType: TokenType.BEARER,
      usages: ['default'],
    },
  ],
})
const rules = ref({
  name: [{ required: true, message: t('mcp.instance.rules.name'), trigger: 'blur' }],
  openapiBaseUrl: [
    { required: true, message: t('mcp.instance.rules.openapiBaseUrl'), trigger: 'blur' },
  ],
  environmentId: [
    { required: true, message: t('mcp.instance.rules.environmentId'), trigger: 'change' },
  ],
})
const baseInfo = ref<any>(null)
const apiNodeList = ref<any[]>([])
const defaultCheckedKeys = ref<any[]>([])
const apiTreeRef = ref(null)
const docObject = ref<any>(null)
const originFileText = ref<any>(null)

const action = ref(
  baseConfig.SERVER_BASE_URL + baseConfig.baseUrlVersion + '/market/openapi/upload',
)
const headers = ref({
  Authorization: `Bearer ${Storage.get('token')}`,
})

/**
 * current checked keys
 */
const currentCheckedKeys = computed(() => {
  return (apiTreeRef as any).value.getCheckedKeys().filter((key: string) => key)
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
const renderContent = (h: any, params: any) => {
  const { data } = params as { node: any; data?: any }
  // build tooltip content
  const buildTooltipContent = () => {
    if (!data || !data.method) return ''
    const parts = []
    if (data.id) parts.push(`ID: <span class="font-500">${data.id}</span>`)
    if (data.method)
      parts.push(
        `Method: <span class="${
          {
            GET: 'color-green',
            POST: 'color-orange',
            PUT: 'color-blue',
            DELETE: 'color-red',
          }[data.method as string]
        }">${data.method.toUpperCase()}</span>`,
      )
    if (data.path) parts.push(`Path: <span class="font-500">${data.path}</span>`)
    if (data.summary) parts.push(`Summary: <span class="font-500">${data.summary}</span>`)
    return `<div class="font-bold">${parts.join('\n')}</div>`
  }

  const tooltipContent = buildTooltipContent()

  return h(
    'div',
    {
      class: ' w-full grid grid-cols-10',
    },
    [
      h(
        'div',
        { class: 'ellipsis-one col-span-9' },
        h('div', { class: 'flex' }, [
          h(
            'span',
            {
              class:
                {
                  GET: 'color-green',
                  POST: 'color-orange',
                  PUT: 'color-blue',
                  DELETE: 'color-red',
                }[data.method as string] + ' font-bold',
            },
            data.method,
          ),
          h('div', { class: 'flex-sub ml-2 u-line-1' }, data.label),
        ]),
      ),
      tooltipContent &&
        h(
          'div',
          { class: 'col-span-1 text-right pr-2' },
          h(
            ElTooltip,
            {
              content: tooltipContent,
              placement: 'left',
              rawContent: true,
              effect: 'dark',
              popperClass: 'api-detail-tooltip',
              style: { width: '400px' },
            },
            {
              default: () =>
                h(
                  'span',
                  { class: 'cursor-pointer', style: { color: 'var(--ep-purple-color)' } },
                  t('common.more'),
                ),
            },
          ),
        ),
    ].filter(Boolean),
  )
}

const handleValidFile = (rawText: string) => {
  return new Promise((resolve, reject) => {
    try {
      docObject.value = JSON.parse(rawText)
      if (!docObject.value.openapi || docObject.value.openapi < '3.0.0') {
        ElMessage.error(t('mcp.instance.openApi.support3'))
        reject(false)
      }
    } catch (e) {
      try {
        docObject.value = yaml.load(rawText)
        if (!docObject.value.openapi || docObject.value.openapi < '3.0.0') {
          ElMessage.error(t('mcp.instance.openApi.support3'))
          reject(false)
        }
      } catch (yamlErr) {
        console.warn('Error', yamlErr)
        reject(false)
      }
    }
    resolve(true)
  })
}

// handle default checked keys and server url、 version check
const handleDefaultCheckedKeys = (rawText: string) => {
  try {
    // keep original file text
    originFileText.value = rawText
    // try to parse as JSON/YAML (fault tolerance)
    try {
      docObject.value = JSON.parse(rawText)
      formData.value.openapiBaseUrl = docObject.value.servers?.length
        ? docObject.value.servers[0]?.url
        : ''
      defaultCheckedKeys.value = []
      const collectIds = (nodes: any[]) => {
        nodes.forEach((n) => {
          if (n.id) defaultCheckedKeys.value.push(n.id)
          if (n.children && n.children.length) collectIds(n.children)
        })
      }
      collectIds(buildApiTree(docObject.value))
    } catch (e) {
      try {
        docObject.value = yaml.load(rawText)
        formData.value.openapiBaseUrl = docObject.value.servers?.length
          ? docObject.value.servers[0]?.url
          : ''
        defaultCheckedKeys.value = []
        const collectIds = (nodes: any[]) => {
          nodes.forEach((n) => {
            if (n.id) defaultCheckedKeys.value.push(n.id)
            if (n.children && n.children.length) collectIds(n.children)
          })
        }
        collectIds(buildApiTree(docObject.value))
        console.log(buildApiTree(docObject.value), defaultCheckedKeys.value)
      } catch (yamlErr) {
        console.warn('Error', yamlErr)
      }
    }
  } catch (err) {
    console.error('Error', err)
  }
}

// Handle default node api list
const handleDefaultNodeAPIlist = (rawText: string) => {
  try {
    // keep original text for later use
    originFileText.value = rawText
    // try to parse as JSON/YAML (fault tolerance)
    try {
      docObject.value = JSON.parse(rawText)
      apiNodeList.value = [
        {
          id: 'root',
          label: t('mcp.instance.openApi.interface'),
          children: buildApiTree(docObject.value),
        },
      ]
      const collectIds = (nodes: any[]) => {
        nodes.forEach((n) => {
          if (n.children && n.children.length) collectIds(n.children)
        })
      }
      collectIds(apiNodeList.value)
    } catch (e) {
      // Not JSON, try to parse as YAML
      try {
        docObject.value = yaml.load(rawText)
        apiNodeList.value = [
          {
            id: 'root',
            label: t('mcp.instance.openApi.interface'),
            children: buildApiTree(docObject.value),
          },
        ]
        const collectIds = (nodes: any[]) => {
          nodes.forEach((n) => {
            if (n.children && n.children.length) collectIds(n.children)
          })
        }
        collectIds(apiNodeList.value)
      } catch (yamlErr) {
        console.warn('Error', yamlErr)
      }
    }
  } catch (err) {
    console.error('Error', err)
  }
}

/**
 * get build api tree
 * @param file Handle before upload
 */
const handleBeforeUpload = async (file: File) => {
  const rawText = await file.text()
  // validate file content
  await handleValidFile(rawText)
  // keep original text for later use
  handleDefaultCheckedKeys(rawText)
  handleDefaultNodeAPIlist(rawText)
}

/**
 *  Handle update success
 * @param response
 */
const handleSuccess = (response: { code: number; data: { openapiFileId: string } }) => {
  if (response.code !== 0) {
    return
  }
  formData.value.openapiFileID = response.data.openapiFileId
  ElMessage.success(t('action.upload'))
}

// get openapi file detail
const handleGetAPIDetail = async (id: string) => {
  try {
    dialogInfo.value.loading = true
    const { content } = await DocsAPI.fileContent({ openapiFileId: id })
    // validate file content
    await handleValidFile(content)
    handleDefaultCheckedKeys(content)
    handleDefaultNodeAPIlist(content)
  } catch {
    formData.value.openapiFileID = ''
  } finally {
    dialogInfo.value.loading = false
  }
}

/**
 * Handle file upload again
 */
const handleUploadAgain = async () => {
  // Reset form data
  const newDoc = JSON.parse(JSON.stringify(docObject.value))
  for (const path in newDoc.paths) {
    const obj = newDoc.paths[path] || {}
    for (const method in obj) {
      const mKey = method as unknown as string
      if (!currentCheckedKeys.value.includes(obj[mKey].operationId)) {
        delete obj[mKey]
      }
    }
    if (Object.keys(newDoc.paths[path]).length === 0) {
      delete newDoc.paths[path]
    }
  }

  // Convert to YAML and upload
  const yamlText = yaml.dump(newDoc)
  const fd = new FormData()
  const blob = new Blob([yamlText], { type: 'application/x-yaml' })
  fd.append('file', blob, 'openapi.yaml')
  fd.append('baseOpenapiFileID', formData.value.openapiFileID)
  try {
    const res = await fetch(action.value, {
      method: 'POST',
      headers: Object.assign({}, headers.value),
      body: fd,
    })
    const body = await res.json()
    if (body?.code === 0 && body.data?.openapiFileId) {
      formData.value.chooseOpenapiFileID = body.data.openapiFileId
      ElMessage.success(t('action.upload'))
    } else {
      ElMessage.error(t('action.uploadFail'))
    }
  } catch (err) {
    console.error('upload error', err)
    ElMessage.error(t('action.uploadFail'))
  }
}

/**
 * Handle confirm and submit data
 */
const handleConfirm = async () => {
  try {
    dialogInfo.value.loading = true
    // handle selected APIs; assemble the selected paths into a new OpenAPI document and upload
    if (originFileText.value) {
      await handleUploadAgain()
    } else {
      ElMessage.error(t('mcp.instance.openApi.validFileFail'))
    }
    baseInfo.value.validate(async (valid: boolean) => {
      if (valid) {
        // 提交数据
        await (formData.value.instanceId
          ? InstanceAPI.editByOpenAPI(formData.value)
          : InstanceAPI.createByOpenAPI(formData.value))
        dialogInfo.value.visible = false
        ElMessage.success(formData.value.instanceId ? t('action.edit') : t('action.create'))
        emit('on-refresh')
      }
    })
  } finally {
    dialogInfo.value.loading = false
  }
}

/**
 * Handle instance detail and openAPI data
 * @param id instance id
 */
const handleGetDetail = async (id: string) => {
  try {
    dialogInfo.value.loading = true
    const data = await InstanceAPI.detail({ instanceId: id })
    formData.value = {
      ...data,
      chooseOpenapiFileID: data.packageId,
      openapiFileID: '',
    }

    // get openapi file content
    const { baseOpenapiFileID, content } = await DocsAPI.fileContent({
      openapiFileId: data.packageId,
    })
    // validate file content
    await handleValidFile(content)
    // handle default checked keys
    handleDefaultCheckedKeys(content)
    formData.value.openapiFileID = baseOpenapiFileID
    // get original api document content
    const res = await DocsAPI.fileContent({
      openapiFileId: baseOpenapiFileID,
    })
    handleDefaultNodeAPIlist(res.content)
  } finally {
    dialogInfo.value.loading = false
  }
}

/**
 * Init dialog data
 */
const init = (id: string | null) => {
  dialogInfo.value.visible = true
  baseInfo.value?.resetFields()
  if (id) {
    // get detail
    handleGetDetail(id)
  } else {
    formData.value = {
      instanceId: '',
      name: '',
      notes: '',
      iconPath: '',
      openapiBaseUrl: '',
      environmentId: '',
      openapiFileID: '',
      chooseOpenapiFileID: '',
      sourceType: SourceType.OPENAPI,
      tokens: [
        {
          enabledTransport: false,
          expireAt: '',
          headers: [],
          publishAt: new Date().getTime(),
          token:
            'Bearer ' +
            getToken(
              JSON.stringify({
                expireAt: Date.now(),
                userId: userInfo.userId,
                username: userInfo.username,
              }),
            ),
          tokenType: TokenType.BEARER,
          usages: ['default'],
        },
      ],
    }
  }

  handleGetAPIlist()
  handleGetEnvList()
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
.api-detail-tooltip {
  max-width: 400px !important;
  white-space: pre-wrap;
  word-break: break-all;
}
</style>
