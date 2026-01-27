<template>
  <div v-loading="pageInfo.loading">
    <div class="page-title flex justify-between items-center">
      <el-button v-if="layout" @click="handleBack" class="link-hover">
        <el-icon class="mr-2">
          <i class="icon iconfont MCP-fanhui"></i>
        </el-icon>
        {{ t('common.back') }}
      </el-button>
    </div>
    <div class="flex justify-center">
      <div class="form-body position-relative">
        <component ref="formComponent" :is="currentComponent"></component>
        <div class="footer-action">
          <div
            :class="
              query.instanceId && Number(query.type) !== AccessType.DIRECT
                ? 'flex justify-between items-center'
                : 'text-center'
            "
          >
            <div v-if="query.instanceId && Number(query.type) !== AccessType.DIRECT" class="flex">
              <el-button link type="primary" @click="handleConfig"> 访问配置 </el-button>
              <el-divider direction="vertical" class="!h-4 !my-auto" />
              <el-button link type="warning" @click="handleViewStatus"> 状态探测 </el-button>
              <el-divider direction="vertical" class="!h-4 !my-auto" />
              <el-button link type="success" @click="handleViewLog"> 查看日志 </el-button>
            </div>
            <div class="flex justify-center">
              <mcp-button @click="handleConfirm" class="mr-4"> 保存并运行 </mcp-button>
              <mcp-button plain @click="handleSaveAsTemplate" class="mr-4"> 另存为模板 </mcp-button>
              <el-button @click="handleClose">返回列表</el-button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useInstanceFormHooks } from './hooks/form-instance.ts'
import HostForm from './modules/components/host-form.vue'
import ProxyForm from './modules/components/proxy-form.vue'
import DirectForm from './modules/components/direct-form.vue'
import { AccessType, McpProtocol, SourceType, InstanceData, NodeVisible } from '@/types/instance'
import McpButton from '@/components/mcp-button/index.vue'
import { useMcpStoreHook } from '@/stores'
import { InstanceAPI } from '@/api/mcp/instance'
import { TemplateAPI } from '@/api/mcp/template'
import { JsonFormatter } from '@/utils/json'
import { formatFileSize, timestampToDate, getToken } from '@/utils/system'

const { t, locale } = useI18n()
const layout = useLayout()
const formComponent = ref()
const { query, pageInfo, userInfo, jumpBack, currentMCP } = useInstanceFormHooks()
const currentComponent = computed(() => {
  switch (Number(query.type)) {
    case AccessType.HOSTING:
      return HostForm
    case AccessType.PROXY:
      return ProxyForm
    case AccessType.DIRECT:
      return DirectForm
    default:
      return null
  }
})
const { currentInstance } = toRefs(useMcpStoreHook())
// back last class page
const handleBack = () => {
  jumpBack()
}
const handleConfig = () => {
  formComponent.value.handleConfig()
}
const handleClose = () => {
  jumpBack()
}
const handleViewStatus = () => {
  formComponent.value.handleViewStatus()
}
const handleViewLog = () => {
  formComponent.value.handleViewLog()
}
const handleConfirm = () => {
  formComponent.value.handleConfirm()
}
const handleSaveAsTemplate = () => {
  formComponent.value.handleSaveAsTemplate()
}
// instance details
const handleGetDetail = async () => {
  let formData: any = {}
  const data = await InstanceAPI.detail({
    instanceId: query.instanceId,
  })
  formData = data
  formData.accessType = data.accessType
  formData.mcpServers = JsonFormatter.format(data.mcpServers, 2)
  formData.environmentVariables = data.environmentVariables
    ? Object.keys(data.environmentVariables)?.map((key) => ({
        key,
        value: data.environmentVariables[key],
      }))
    : []
  formData.volumeMounts = data.volumeMounts || []
  return formData
}
// template details
const handleGetTemplateDetail = async () => {
  let formData: any = {}
  const data = await TemplateAPI.detail({
    id: query.templateId,
  })
  formData = data
  formData.mcpServers = JsonFormatter.format(data.mcpServers)
  formData.environmentVariables = data.environmentVariables
    ? Object.keys(data.environmentVariables)?.map((key) => ({
        key,
        value: data.environmentVariables[key],
      }))
    : []
  formData.volumeMounts = data.volumeMounts || []
  formData.sourceType = SourceType.TEMPLATE
  // default open token
  let tokenValue =
    'Bearer ' +
    getToken(
      JSON.stringify({
        expireAt: Date.now(),
        userId: userInfo.userId,
        username: userInfo.username,
      }),
    )
  formData.enabledToken = true
  formData.tokens = [
    {
      expireAt: '',
      enabled: true,
      publishAt: new Date().getTime(),
      headers: [{ key: 'Authorization', value: tokenValue }],
      token: tokenValue,
      usages: ['default'],
    },
  ]
  return formData
}
//instacen  market details
const handleInitMarketInstance = async () => {
  let tokenValue =
    'Bearer ' +
    getToken(
      JSON.stringify({
        expireAt: Date.now(),
        userId: userInfo.userId,
        username: userInfo.username,
      }),
    )
  let formData = {
    sourceType: SourceType.MARKET,
    name: locale.value === 'zh-cn' ? currentMCP.name : currentMCP.nameEn,
    accessType: AccessType.HOSTING,
    mcpProtocol: McpProtocol.STDIO,
    imgAddress: InstanceData.value.IMGADDRESS,
    notes: locale.value === 'zh-cn' ? currentMCP.description : currentMCP.descriptionEn,
    mcpServers: JsonFormatter.format(currentMCP.configTemplate),
    iconPath: currentMCP.githubOwnerAvatarUrl,
    packageId: '',
    environmentId: '',
    port: InstanceData.value.PORT,
    environmentVariables: [],
    volumeMounts: [],
    initScript: InstanceData.value.INITSCRIPT,
    command: '',
    enabledToken: true,
    tokens: [
      {
        enabled: true,
        expireAt: '',
        publishAt: new Date().getTime(),
        headers: [{ key: 'Authorization', value: tokenValue }],
        token: tokenValue,
        usages: ['default'],
      },
    ],
  }
  return formData
}
// handle init components formdata
const init = () => {
  if (query.instanceId) {
    // edit instance
    handleGetDetail().then((formData) => {
      nextTick(() => {
        formComponent.value.init(formData)
      })
    })
  } else if (query.templateId) {
    // create instance by template
    handleGetTemplateDetail().then((formData) => {
      nextTick(() => {
        formComponent.value.init(formData)
      })
    })
  } else if (query.from === 'market') {
    // create instance by market
    handleInitMarketInstance().then((formData) => {
      nextTick(() => {
        formComponent.value.init(formData)
      })
    })
  } else {
    // create instance by custom
    nextTick(() => {
      formComponent.value.init(null)
    })
  }
}
onMounted(() => {
  init()
})
</script>

<style lang="scss" scoped>
.page-title {
  font-family:
    PingFangSC,
    PingFang SC;
  font-weight: 600;
  font-size: 20px;
  line-height: 28px;
  &.base-info {
    margin-top: 40px;
    margin-bottom: 16px;
  }
}
.form-body {
  width: 850px;
}
.footer-action {
  position: sticky;
  bottom: -20px;
  z-index: 1000;
  border-radius: 6px;
  background: var(--ep-bg-color);
  padding: 16px;
  box-shadow: 0 -2px 8px rgba(0, 0, 0, 0.1);
}
</style>
