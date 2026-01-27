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
            <template #label></template>
            <MonacoEditor v-model="pageInfo.formData.mcpServers" language="json" height="200px" />
          </el-form-item>
        </el-col>
        <el-col :span="6">
          <div
            class="pl-3 rounded border border-[var(--ep-border-color-lighter)] text-[var(--ep-text-color-secondary)] text-xs leading-6 tracking-wide"
          >
            MCP服务SSE/STEAMABLE_HTTP协议配置当前为代理模式，流量会通过此平台网关转发到此配置提供的
            MCP 配置中，保存后会在列表页显示网关访问配置。也可以查看
            <a href="#/template-manage">部署模板</a> 提供了多个启动示例。
          </div>
        </el-col>
      </el-row>
      <el-collapse :expand-icon-position="'left'">
        <el-collapse-item v-if="!pageInfo.formData.instanceId" name="1">
          <template #title>
            <div>
              <span class="mr-1 font-bold">Header 透传配置</span>
              <span
                class="rounded border border-[var(--ep-border-color-lighter)] text-[var(--ep-text-color-secondary)] text-xs leading-6 tracking-wide"
              >
                此项不配置时：默认透传来自客户端的Headers到MCP服务中。
                如果MCPServers配置中存在Header参数，优先级为：MCPServers配置中Header >
                客户端的Headers
              </span>
            </div>
          </template>
          <div>
            <TokenForm ref="tokenForm" :formData="pageInfo.formData.tokens[0]"></TokenForm>
          </div>
          <div class="tip tip-primary">
            注意事项：1.自定义 Header 无需客户端主动提交，网关在转发流量时，会自动将其携带至 MCP
            服务的请求中。 2.若客户端请求 Header、MCP 配置 Header、自定义透传 Header
            三者存在同名项，优先级顺序为：自定义透传 Header > MCP 配置 Header > 客户端请求 Header。
            即同名 Header 取值时，优先采用自定义透传配置，其次为 MCP
            配置，最后为客户端请求传入的值。
          </div>
        </el-collapse-item>
      </el-collapse>
    </el-form>
  </div>
</template>

<script setup lang="ts">
import { useInstanceFormHooks } from '../../hooks/form-instance.ts'
import Upload from '@/components/upload/index.vue'
import { AccessType } from '@/types/instance.ts'
import { JsonFormatter } from '@/utils/json'
import TokenForm from './token-form.vue'
import { ElLoading, ElMessage, ElMessageBox } from 'element-plus'
import { TemplateAPI } from '@/api/mcp/template'
import { InstanceAPI } from '@/api/mcp/instance'
import MonacoEditor from '@/components/MonacoEditor/index.vue'
import { type InstanceResult } from '@/types/index.ts'
import { cloneDeep } from 'lodash-es'

const { t } = useI18n()
const { query, pageInfo, placeholderServer, jumpToPage } = useInstanceFormHooks()
const baseInfo = ref()
const protocolOptions = [
  { label: 'SSE', value: 1 },
  { label: 'STEAMABLE_HTTP', value: 2 },
]
/**
 * Handle McpProtocol Changed
 */
const handleMcpProtocolChange = () => {}

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
</style>
