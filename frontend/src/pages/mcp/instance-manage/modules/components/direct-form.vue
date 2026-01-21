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
      <el-form-item prop="mcpServers">
        <el-input
          v-model="pageInfo.formData.mcpServers"
          :rows="14"
          type="textarea"
          :placeholder="placeholderServer"
          @blur="handleFormat"
        />
      </el-form-item>
      <div
        class="mt-2 p-3 rounded border border-[var(--ep-border-color-lighter)] bg-[var(--ep-fill-color-lighter)] text-[var(--ep-text-color-secondary)] text-xs leading-6 tracking-wide"
      >
        MCP服务 SSE / STEAMABLE HTTP 协议配置当前为直连模式，主要是填写外部 MCP
        访问配置，平台仅承担「配置注册中心」角色。如果需要代理业务流量或者参与健康探测与运行监控请切换为代理模式。
      </div>
    </el-form>
  </div>
</template>
<script setup lang="ts">
import { useInstanceFormHooks } from '../../hooks/form-instance.ts'
import Upload from '@/components/upload/index.vue'
import { JsonFormatter } from '@/utils/json'
import { ElLoading, ElMessage, ElMessageBox } from 'element-plus'
import { InstanceAPI } from '@/api/mcp/instance'
import { AccessType } from '@/types/instance.ts'
import { TemplateAPI } from '@/api/mcp/template'

const { t } = useI18n()
const { pageInfo, placeholderServer } = useInstanceFormHooks()
const baseInfo = ref()
const protocolOptions = [
  { label: 'SSE', value: 1 },
  { label: 'STEAMABLE_HTTP', value: 2 },
]
const handleFormat = () => {
  pageInfo.value.formData.mcpServers = JsonFormatter.format(pageInfo.value.formData.mcpServers)
}
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
        await (pageInfo.value.formData.instanceId ? InstanceAPI.edit : InstanceAPI.create)({
          ...pageInfo.value.formData,
          environmentVariables: pageInfo.value.formData.environmentVariables?.reduce(
            (obj: any, item: any) => ({ ...obj, [item.key]: item.value }),
            {},
          ),
        })
        ElMessage.success(
          pageInfo.value.formData.instanceId ? t('action.edit') : t('action.create'),
        )
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
        pageInfo.value.loading = true
        await TemplateAPI.create({
          ...pageInfo.value.formData,
          environmentVariables: pageInfo.value.formData.environmentVariables?.reduce(
            (obj: any, item: any) => ({ ...obj, [item.key]: item.value }),
            {},
          ),
        })
        ElMessage.success(t('action.create'))
        pageInfo.value.loading = false
        dialogInfo.value.visible = false
      }
    })
  } finally {
    pageInfo.value.loading = false
  }
}
const init = () => {
  nextTick(() => {
    pageInfo.value.formData.accessType = AccessType.DIRECT
    pageInfo.value.formData.mcpProtocol = 1
  })
}
defineExpose({
  init,
  handleConfirm,
  handleSaveAsTemplate,
})
</script>
