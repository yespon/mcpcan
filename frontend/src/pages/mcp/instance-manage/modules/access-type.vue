<template>
  <div>
    <el-dialog
      v-model="dialogInfo.visible"
      :title="'快速开始'"
      append-to-body
      width="1200px"
      class="access-type-dialog"
    >
      <div class="flex justify-center gap-8 py-8">
        <div
          v-for="item in accessOptions"
          :key="item.value"
          class="access-card flex flex-col items-center p-6 cursor-pointer transition-all duration-300 rounded-xl border bg-card"
          @click="handleSelect(item)"
          @mouseenter="currentHover = item.value"
          @mouseleave="currentHover = null"
        >
          <GradientText
            v-if="currentHover === item.value"
            text=""
            :colors="['#ffaa40', '#9c40ff', '#ffaa40']"
            :animation-speed="10"
            :show-border="false"
          >
            <div class="w-full h-full flex flex-col items-center">
              <el-icon :size="48" class="mb-4 !text-[48px]" :color="item.color">
                <i class="icon iconfont !text-[48px]" :class="item.icon"></i>
              </el-icon>
              <div class="text-lg font-bold mb-2">{{ item.label }}</div>
              <div class="text-left tip tip-primary">
                {{ item.description }}
              </div>
              <div class="mt-4 text-size-sm w-full">
                <div class="font-bold text-left">支持协议类型</div>
                <div v-if="item.supportTypes.includes(McpProtocol.SSE)" class="item-control">
                  SSE协议
                </div>
                <div
                  v-if="item.supportTypes.includes(McpProtocol.STEAMABLE_HTTP)"
                  class="item-control"
                >
                  STEAMABLE_HTTP协议
                </div>
                <div v-if="item.supportTypes.includes(McpProtocol.STDIO)" class="item-control">
                  标准输入输出STDIO协议（对外暴露SSE/STEAMABLE_HTTP）
                </div>
              </div>
            </div>
          </GradientText>
          <div v-else class="w-full h-full flex flex-col items-center">
            <el-icon :size="48" class="mb-4 !text-[48px]" :color="item.color">
              <i class="icon iconfont !text-[48px]" :class="item.icon"></i>
            </el-icon>
            <div class="text-lg font-bold mb-2">{{ item.label }}</div>
            <div class="text-left tip tip-primary">
              {{ item.description }}
            </div>
            <div class="mt-4 text-size-sm w-full">
              <div class="font-bold text-left">支持协议类型</div>
              <div v-if="item.supportTypes.includes(McpProtocol.SSE)" class="item-control">
                SSE协议
              </div>
              <div
                v-if="item.supportTypes.includes(McpProtocol.STEAMABLE_HTTP)"
                class="item-control"
              >
                STEAMABLE_HTTP协议
              </div>
              <div v-if="item.supportTypes.includes(McpProtocol.STDIO)" class="item-control">
                标准输入输出STDIO协议（对外暴露SSE/STEAMABLE_HTTP）
              </div>
            </div>
          </div>
        </div>
      </div>
    </el-dialog>
    <component ref="formComponent" :is="currentModal?.formComponent"></component>
  </div>
</template>

<script setup lang="ts">
import { Coin, Connection, Link, DocumentAdd } from '@element-plus/icons-vue'
import GradientText from '@/components/Animation/GradientText.vue'
import { useInstanceFormHooks } from '../hooks/form-instance.ts'
import { AccessType, McpProtocol } from '@/types/instance.ts'
import HostingDialog from './hosting-dialog.vue'
import ProxyDialog from './proxy-dialog.vue'
import DirectDialog from './direct-dialog.vue'
import OpenApiDialog from './open-api-dialog.vue'
import { type InstanceResult } from '@/types/instance.ts'
import { create } from 'lodash-es'

const emit = defineEmits(['select'])
const { t } = useI18n()
const { pageInfo, jumpToPage, originForm } = useInstanceFormHooks()
const dialogInfo = ref({
  visible: false,
})
const currentHover = ref<AccessType | null>()
const formComponent = ref()
const accessOptions = [
  {
    label: '托管 (Hosting)',
    value: AccessType.HOSTING,
    icon: 'MCP-anquan',
    formComponent: HostingDialog,
    color: '#67C23A',
    supportTypes: [McpProtocol.SSE, McpProtocol.STEAMABLE_HTTP, McpProtocol.STDIO],
    description:
      '托管模式是指平台利用自身容器管理能力运行 MCP 服务，并通过系统内置网关和适配器解决流量代理监控和协议不兼容问题',
  },
  {
    label: '代理 (Proxy)',
    value: AccessType.PROXY,
    icon: 'MCP-daili',
    formComponent: ProxyDialog,
    color: '#E6A23C',
    supportTypes: [McpProtocol.SSE, McpProtocol.STEAMABLE_HTTP],
    description:
      '代理模式将平台转化为 MCP 服务的统一访问网关。而是通过平台提供的代理地址进行交互;平台在转发请求的过程中附加安全防护与审计能力，实现屏蔽后端、统一入口的目标。',
  },
  {
    label: '直连 (Direct)',
    value: AccessType.DIRECT,
    icon: 'MCP-zhilian',
    formComponent: DirectDialog,
    color: '#409EFF',
    supportTypes: [McpProtocol.SSE, McpProtocol.STEAMABLE_HTTP],
    description:
      '直连模式是平台最轻量级的接入方式，平台仅承担配置注册中心角色，不代理任何业务流量，不参与健康探测与运行监控。客户端按照平台存储的配置，直接与外部 MCP 服务通信。',
  },
  {
    label: 'OpenAPI',
    value: 4,
    icon: 'MCP-wenjian1',
    formComponent: OpenApiDialog,
    color: '#ff8eb9',
    supportTypes: [McpProtocol.STEAMABLE_HTTP],
    description:
      '将标准的 OpenAPI 接口文档自动转化为 MCP 服务。平台解析接口文档并生成适配器，使得传统的 HTTP 接口能够通过 MCP 协议进行流式访问，从而快速实现现有业务接口到 MCP 生态的无缝集成。',
  },
]
const currentModal = computed(() => {
  return accessOptions.find((option) => option.value === pageInfo.value.accessType)
})

const handleSelect = (item: any) => {
  pageInfo.value.accessType = item.value
  dialogInfo.value.visible = false
  // dialog view
  if (item.value === 4) {
    nextTick(() => {
      formComponent.value?.init()
    })
    return
  }
  // page view with create
  jumpToPage({
    url: '/new-instance',
    data: {
      type: item.value,
    },
  })
}

const init = (instance: InstanceResult | null) => {
  if (instance) {
    pageInfo.value.accessType = instance.accessType
    if (instance.instanceId) {
      nextTick(() => {
        formComponent.value?.init(instance)
      })
      return
    }
  }
  dialogInfo.value.visible = true
}
defineExpose({
  init,
})
</script>

<style lang="scss" scoped>
.access-card {
  width: 240px;
  // height: 280px;
  border-color: var(--ep-border-color-lighter);
  color: var(--ep-text-color-secondary);
  font-family: 'Monaco, Menlo, "Ubuntu Mono", monospace';
  font-size: 12px;
  word-break: normal;
  &:hover {
    transform: scale(1.1);
    box-shadow: 0 12px 32px rgba(0, 0, 0, 0.1);
    border: 1px solid var(--el-color-primary);
    z-index: 10;
  }
}
.bg-card {
  background-color: var(--ep-home-glass);
}
.tip {
  padding: 10px;
  border-radius: 4px;
  font-size: 12px !important;
  &.tip-warning {
    background-color: #fff1f0;
    border-left: 5px solid var(--el-color-danger);
  }
  &.tip-primary {
    background-color: #409eff1a;
    border-left: 5px solid var(--el-color-primary);
  }
}
.item-control {
  font-size: 12px;
  padding-left: 12px;
  &::before {
    content: '• ';
    color: var(--el-color-primary);
  }
}
</style>
