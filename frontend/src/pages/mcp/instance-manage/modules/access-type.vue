<template>
  <div>
    <el-dialog
      v-model="dialogInfo.visible"
      :title="'访问模式'"
      :show-close="false"
      width="900px"
      class="access-type-dialog"
    >
      <div class="flex justify-center gap-8 py-8">
        <div
          v-for="item in accessOptions"
          :key="item.value"
          class="access-card flex flex-col items-center justify-center p-6 cursor-pointer transition-all duration-300 rounded-xl border bg-card"
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
            class-name="w-full h-full flex flex-col items-center justify-center"
          >
            <el-icon :size="48" class="mb-4" :color="item.color">
              <component :is="item.icon" />
            </el-icon>
            <div class="text-lg font-bold mb-2">{{ item.label }}</div>
            <div class="text-sm text-left">
              {{ item.description }}
            </div>
          </GradientText>
          <div v-else class="w-full h-full flex flex-col items-center justify-center">
            <el-icon :size="48" class="mb-4" :color="item.color">
              <component :is="item.icon" />
            </el-icon>
            <div class="text-lg font-bold mb-2">{{ item.label }}</div>
            <div class="text-sm text-left">
              {{ item.description }}
            </div>
          </div>
        </div>
      </div>
    </el-dialog>
    <component ref="formComponent" :is="currentModal?.formComponent"></component>
  </div>
</template>

<script setup lang="ts">
import { Coin, Connection, Link } from '@element-plus/icons-vue'
import GradientText from '@/components/Animation/GradientText.vue'
import { useInstanceFormHooks } from '../hooks/form-instance.ts'
import { AccessType } from '@/types/instance.ts'
import HostingDialog from './hosting-dialog.vue'
import ProxyDialog from './proxy-dialog.vue'
import DirectDialog from './direct-dialog.vue'

const emit = defineEmits(['select'])
const { t } = useI18n()
const { pageInfo } = useInstanceFormHooks()
const dialogInfo = ref({
  visible: false,
})
const currentHover = ref<AccessType | null>()
const formComponent = ref()
const accessOptions = [
  {
    label: '托管 (Hosting)',
    value: AccessType.HOSTING,
    icon: Coin,
    formComponent: HostingDialog,
    color: '#409EFF',
    description:
      '托管模式是指平台利用自身容器管理能力运行 MCP 服务，并通过系统内置网关和适配器解决流量代理监控和“协议不兼容”问题-SSE协议部署-STEAMABLE HTTP协议部署标准输入输出STDIO(转SEE/STEAMABLE HTTP协议后暴露访问入口)部署对外统一暴露SSE/STEAMABLE HTTP端点',
  },
  {
    label: '代理 (Proxy)',
    value: AccessType.PROXY,
    icon: Connection,
    formComponent: ProxyDialog,
    color: '#67C23A',
    description:
      '代理模式将平台转化为 MCP 服务的统一访问网关。而是通过平台提供的代理地址进行交互;平台在转发请求的过程中附加安全防护与审计能力，实现“屏蔽后端、统一入口”的目标。支持两种 MCP 协议:SEE(Server-Sent Events)与STEAMABLE HTTP。',
  },
  {
    label: '直连 (Direct)',
    value: AccessType.DIRECT,
    icon: Link,
    formComponent: DirectDialog,
    color: '#E6A23C',
    description:
      '直连模式是平台最轻量级的接入方式，平台仅承担配置注册中心角色，不代理任何业务流量，不参与健康探测与运行监控。客户端按照平台存储的配置，直接与外部 MCP 服务通信。支持两种 MCP 协议:SEE(Server-Sent Events)与STEAMABLE HTTP。',
  },
]
const currentModal = computed(() => {
  return accessOptions.find((option) => option.value === pageInfo.value.accessType)
})

const handleSelect = (item: any) => {
  // emit('select', item.value)
  pageInfo.value.accessType = item.value
  dialogInfo.value.visible = false
  nextTick(() => {
    formComponent.value?.init()
  })
}

const init = () => {
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
</style>
