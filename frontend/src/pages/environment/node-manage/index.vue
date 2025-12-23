<template>
  <div>
    <!-- 头部区域 -->
    <div class="flex justify-between page-header">
      <div class="header-container flex items-center">
        <el-link v-if="layout" link @click="handleBack" class="link-hover mr-4" underline="never">
          <el-icon class="mr-2">
            <i class="icon iconfont MCP-fanhui"></i>
          </el-icon>
          {{ t('common.back') }}
        </el-link>
        {{ t('env.run.pageDesc.nodeList') }}-{{ query.name }}
        <span class="desc"></span>
      </div>
    </div>

    <!-- 表格区域 -->

    <TablePlus
      ref="tablePlus"
      search-container="#nodeSearch"
      :showOperation="false"
      :requestConfig="requestConfig"
      :columns="columns"
      v-model:pageConfig="pageConfig"
      :handlerColumnConfig="null"
      :showPage="false"
    >
      <template #action>
        <div class="flex justify-between mb-4">
          <div class="center">
            <el-image :src="nodeLogo" style="width: 20px; height: 20px"></el-image>
            <span class="desc">{{ t('env.run.pageDesc.nodeTotal') }}：{{ pageConfig.total }}</span>
          </div>
          <div id="nodeSearch" v-show="false"></div>
        </div>
      </template>
      <template #status="{ row }">
        <el-tag :type="statusOptions[row.status as keyof typeof statusOptions]">
          {{ row.status }}
        </el-tag>
      </template>
      <template #roles="{ row }">
        <el-tag type="info" v-for="(role, index) in row.roles" :key="index" class="mr-1">
          {{ role }}
        </el-tag>
      </template>
      <template #createdAt="{ row }">{{ timestampToDate(row.createdAt) }} </template>
      <template #updatedAt="{ row }">{{ timestampToDate(row.updatedAt) }} </template>
    </TablePlus>
  </div>
</template>

<script setup lang="ts">
import TablePlus from '@/components/TablePlus/index.vue'
import { timestampToDate } from '@/utils/system'
import { usePvcTableHooks } from './hooks'
import nodeLogo from '@/assets/logo/node.png'
import { useRouterHooks } from '@/utils/url'

const { t } = useI18n()
const { tablePlus, columns, requestConfig, pageConfig, statusOptions, query, meta } =
  usePvcTableHooks()
const { jumpBack } = useRouterHooks()
const layout = useLayout()

// back last class page
const handleBack = () => {
  jumpBack()
}

/**
 * Handle init list
 */
const init = () => {
  tablePlus.value.initData()
}

onMounted(init)
</script>
<style lang="scss" scoped>
.page-header {
  margin-bottom: 24px;
  .header-container {
    font-size: 20px;
  }
}
.desc {
  font-size: 16px;
  color: #999999;
  margin-left: 16px;
}
</style>
