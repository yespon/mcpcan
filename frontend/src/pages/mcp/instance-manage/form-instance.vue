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
    <div class="flex justify-start">
      <div class="form-body">
        <component ref="formComponent" :is="currentComponent"></component>
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

// 初始化当前表单
const init = () => {
  nextTick(() => {
    formComponent.value.init()
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
</style>
