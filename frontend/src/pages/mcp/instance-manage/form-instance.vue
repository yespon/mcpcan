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
          <div :class="query.instanceId ? 'flex justify-between items-center' : 'text-center'">
            <div v-if="query.instanceId" class="flex">
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
import { AccessType } from '@/types/instance.ts'
import McpButton from '@/components/mcp-button/index.vue'

const { t } = useI18n()
const layout = useLayout()
const formComponent = ref()
const { query, pageInfo, jumpBack } = useInstanceFormHooks()
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
// 初始化当前表单
const init = () => {
  nextTick(() => {
    formComponent.value.init(query)
  })
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
