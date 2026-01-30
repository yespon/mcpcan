<template>
  <div class="px-5 py-2" v-loading="pageInfo.loading">
    <el-form
      ref="baseInfo"
      :model="pageInfo.formData"
      :rules="pageInfo.rules"
      label-width="auto"
      label-position="left"
    >
      <el-row :gutter="24">
        <el-col :span="18">
          <el-form-item :label="t('mcp.instance.formData.instanceName')" prop="name">
            <el-input
              v-model="pageInfo.formData.name"
              :placeholder="t('mcp.instance.formData.instanceName')"
            />
          </el-form-item>
          <el-form-item :label="t('mcp.template.formData.notes')" prop="notes">
            <el-input
              v-model="pageInfo.formData.notes"
              :rows="2"
              type="textarea"
              :placeholder="t('mcp.template.formData.notes')"
            />
          </el-form-item>
        </el-col>
        <el-col :span="6">
          <el-form-item prop="iconPath">
            <Upload v-model="pageInfo.formData.iconPath"></Upload>
          </el-form-item>
        </el-col>
      </el-row>
      <el-form-item :label="t('mcp.instance.form.mcpProtocol')" prop="mcpProtocol">
        <el-segmented
          v-model="pageInfo.formData.mcpProtocol"
          :options="protocolOptions"
          @change="handleMcpProtocolChange"
        />
      </el-form-item>
      <el-row>
        <el-col :span="18">
          <el-form-item prop="mcpServers">
            <template #label>{{ t('mcp.instance.formData.mcpServers') }}</template>
            <MonacoEditor
              v-model="pageInfo.formData.mcpServers"
              language="json"
              :height="locale === 'en' ? '600px' : '300px'"
            />
          </el-form-item>
        </el-col>
        <el-col :span="6">
          <div
            class="pl-3 rounded border border-[var(--ep-border-color-lighter)] text-[var(--ep-text-color-secondary)] text-xs leading-6 tracking-wide"
            v-html="mcpServersTips"
          ></div>
        </el-col>
      </el-row>
      <el-collapse v-model="currentOpenCollapse" :expand-icon-position="'left'" accordion>
        <el-collapse-item v-if="!pageInfo.formData.instanceId" name="1">
          <template #title>
            <Transition name="tip-fade">
              <div>
                <div class="mr-1 font-bold">{{ t('mcp.instance.hostingForm.headerTitle') }}</div>
                <div
                  v-show="currentOpenCollapse !== '1'"
                  class="flex-1 tip tip-primary my-2 line-height-[18px]"
                  style="margin-left: -16px"
                  v-html="headerTitleTips"
                ></div>
              </div>
            </Transition>
          </template>
          <div>
            <TokenForm ref="tokenForm" :formData="pageInfo.formData.tokens[0]"></TokenForm>
          </div>

          <div
            v-show="currentOpenCollapse === '1'"
            class="flex-1 tip tip-primary my-2 line-height-[18px]"
            v-html="headerContentTips"
          ></div>
        </el-collapse-item>
      </el-collapse>
    </el-form>
  </div>
  <!-- probe instance dialog model -->
  <ProbeStatus ref="probe"></ProbeStatus>
  <ConfigDialog ref="config"></ConfigDialog>
  <LogDialog ref="log"></LogDialog>
</template>

<script setup lang="ts">
import { useInstanceFormHooks } from '../../hooks/form-instance.ts'
import Upload from '@/components/upload/index.vue'
import { AccessType } from '@/types/instance.ts'
import TokenForm from './token-form.vue'
import { ElMessage } from 'element-plus'
import { TemplateAPI } from '@/api/mcp/template'
import { InstanceAPI } from '@/api/mcp/instance'
import MonacoEditor from '@/components/MonacoEditor/index.vue'
import { type InstanceResult } from '@/types/index.ts'
import { cloneDeep } from 'lodash-es'
import ProbeStatus from '../probe-dialog.vue'
import ConfigDialog from '../url-config-dialog.vue'
import LogDialog from '../log-dialog.vue'
import { useMcpStoreHook } from '@/stores'

const { t, locale } = useI18n()
const { query, pageInfo, placeholderServer, jumpToPage } = useInstanceFormHooks()
const { currentInstance } = toRefs(useMcpStoreHook())
const baseInfo = ref()
const currentOpenCollapse = ref()
const protocolOptions = [
  { label: 'SSE', value: 1 },
  { label: 'STREAMABLE_HTTP', value: 2 },
]

const mcpServersTips = computed(() => {
  return locale.value === 'en'
    ? `If you deploy by uploading a code package, set the working directory in the configuration: <code>"cwd": "/app/codepkg/[codezip_name]"</code>.
When the instance starts, the platform will automatically create a default runtime container for the current MCP configuration. The container includes the <code>mcp-hosting</code> hosting component and uses a protocol adapter to expose this configuration as SSE/Streamable HTTP. It also binds to <code>0.0.0.0:8080</code> to provide the service externally.
The MCPCAN gateway will automatically probe and detect the container's network address and complete the routing setup.
For more startup configuration examples of different types, see <a href="#/template-manage">Deploy Templates</a>.`
    : `若需上传代码包部署，需在配置信息中指定工作目录："cwd": "/app/codepkg/[codezip_name]"。
实例启动时，平台将为当前 MCP 配置自动创建默认运行容器，容器内置 mcp-hosting 托管组件，通过协议适配器将本配置转换为 SSE/Streamable HTTP 协议，同时绑定至0.0.0.0:8080端口对外提供服务；MCPCAN 网关服务将自动探测并识别该容器的网络地址，完成对接。
更多不同类型的启动配置示例，可参考<a href="#/template-manage">部署模板</a>。`
})
const headerTitleTips = computed(() => {
  return locale.value === 'en'
    ? `If this field is left empty, headers from the client will be passed through to the MCP service by default. If the MCPServers configuration defines header parameters, precedence is: MCPServers configuration headers > client request headers.`
    : `此项不配置时：默认透传来自客户端的Headers到MCP服务中。如果MCPServers配置中存在Header参数，优先级为：MCPServers配置中Header > 客户端的Headers`
})

const headerContentTips = computed(() => {
  return locale.value === 'en'
    ? `
      Notes:
      <br/>
      1. Custom headers do not need to be actively submitted by the client; the gateway will automatically include them in requests forwarded to the MCP service.
      <br/>
      2. If there are duplicate header names among client request headers, MCP configuration headers, and custom pass-through headers, the precedence order is: custom pass-through headers > MCP configuration headers > client request headers. That is, when retrieving values for duplicate header names, the custom pass-through configuration is used first, followed by the MCP configuration, and finally the value provided by the client request.`
    : `注意事项：
      <br/>
      1.自定义 Header 无需客户端主动提交，网关在转发流量时，会自动将其携带至 MCP 服务的请求中。
      <br/>
      2.若客户端请求 Header、MCP 配置 Header、自定义透传 Header 三者存在同名项，优先级顺序为：自定义透传 Header > MCP 配置 Header > 客户端请求 Header。即同名 Header 取值时，优先采用自定义透传配置，其次为 MCP 配置，最后为客户端请求传入的值。`
})

/**
 * Handle McpProtocol Changed
 */
const handleMcpProtocolChange = () => {}
const config = ref()
const handleConfig = () => {
  config.value.init(Object.assign(currentInstance.value, pageInfo.value.formData))
}
const probe = ref()
const handleViewStatus = () => {
  probe.value.init(pageInfo.value.formData)
}
const log = ref()
const handleViewLog = () => {
  log.value.init(pageInfo.value.formData)
}
// Handle confirm save
const handleConfirm = async () => {
  baseInfo.value.validate(async (valid: boolean) => {
    if (valid) {
      try {
        pageInfo.value.loading = true
        if (!pageInfo.value.formData.instanceId) {
          if (Array.isArray(pageInfo.value.formData.tokens[0].headers)) {
            pageInfo.value.formData.tokens[0].headers = Object.fromEntries(
              pageInfo.value.formData.tokens[0].headers?.map((header: any) => [
                header.key,
                header.value,
              ]),
            )
          }
        }
        const { instanceId } = await (
          pageInfo.value.formData.instanceId ? InstanceAPI.edit : InstanceAPI.create
        )({
          ...pageInfo.value.formData,
          environmentVariables: pageInfo.value.formData.environmentVariables?.reduce(
            (obj: any, item: any) => ({ ...obj, [item.key]: item.value }),
            {},
          ),
        })
        ElMessage.success(
          pageInfo.value.formData.instanceId ? t('action.edit') : t('action.create'),
        )
        pageInfo.value.formData.instanceId = instanceId
        jumpToPage({
          url: '/new-instance',
          data: {
            instanceId,
            type: query.type,
          },
        })
      } finally {
        pageInfo.value.loading = false
      }
    }
  })
}
/**
 * save as a template
 */
const handleSaveAsTemplate = async () => {
  try {
    pageInfo.value.loading = true
    baseInfo.value.validate(async (valid: boolean) => {
      if (valid) {
        try {
          pageInfo.value.loading = true
          await TemplateAPI.create({
            ...pageInfo.value.formData,
            environmentVariables: pageInfo.value.formData.environmentVariables?.reduce(
              (obj: any, item: any) => ({ ...obj, [item.key]: item.value }),
              {},
            ),
          })
          ElMessage.success(t('action.create'))
        } finally {
          pageInfo.value.loading = false
        }
      }
    })
  } finally {
    pageInfo.value.loading = false
  }
}
const init = (instance: InstanceResult | null) => {
  if (instance) {
    pageInfo.value.formData = cloneDeep(instance)
  } else {
    pageInfo.value.formData.mcpProtocol = 1
  }
  nextTick(() => {
    pageInfo.value.formData.accessType = AccessType.PROXY
  })
}
defineExpose({
  init,
  handleConfig,
  handleViewStatus,
  handleViewLog,
  handleConfirm,
  handleSaveAsTemplate,
})
</script>
<style lang="scss" scoped>
.tip {
  padding: 10px;
  border-radius: 4px;
  font-size: 12px;
  &.tip-warning {
    background-color: #fff1f0;
    border-left: 5px solid var(--el-color-danger);
  }
  &.tip-primary {
    background-color: #409eff1a;
    border-left: 5px solid var(--el-color-primary);
  }
}

/* 让 el-collapse-item 左侧图标与标题内容顶部对齐 */
:deep(.el-collapse-item__header) {
  align-items: flex-start;
}

/* header 内部容器通常也是 flex，保持一致 */
:deep(.el-collapse-item__header .el-collapse-item__title) {
  align-items: flex-start;
}

/* 图标在顶部附近，避免贴边太紧 */
:deep(.el-collapse-item__header .el-collapse-item__arrow) {
  margin-top: 16px;
}
</style>
