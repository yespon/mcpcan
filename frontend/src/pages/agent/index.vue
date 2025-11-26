<template>
  <div class="agent-page">
    <!-- 头部工具栏 -->
    <div class="flex justify-between page-header">
      <div class="header-container">
        {{ t('agent.pageDesc.title') }}
        <span class="desc">{{ t('agent.pageDesc.desc') }}</span>
      </div>
      <el-dropdown
        trigger="click"
        class="ml-4"
        @click.stop
        :show-arrow="false"
        @command="handleNewAgent"
      >
        <mcp-button :icon="Plus">{{ t('agent.action.create') }}</mcp-button>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="community">
              <div class="flex align-center">
                <McpImage :src="dify" fit="contain" width="80" height="20" />
                {{ '社区版' }}
              </div>
            </el-dropdown-item>
            <el-dropdown-item command="business">
              <McpImage :src="dify" fit="contain" width="80" height="20" />
              {{ '商业版' }}
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>

    <!-- 内容区域：列表/卡片切换 -->
    <div v-if="viewMode === 'table'" v-loading="loading">
      <TablePlus
        :showOperation="true"
        searchContainer="#agentSearch"
        ref="tablePlus"
        :requestConfig="requestConfig"
        :columns="columns"
        show-view-mode
        v-model:pageConfig="pageConfig"
        :handlerColumnConfig="{
          width: '120px',
          fixed: 'right',
        }"
      >
        <template #action>
          <div class="flex justify-between mb-4">
            <div class="center">
              <el-image :src="agentLogo" style="width: 20px; height: 20px"></el-image>
              <span class="desc">{{ t('agent.pageDesc.total') }}：{{ pageConfig.total }}</span>
            </div>
            <div id="agentSearch"></div>
          </div>
        </template>
        <template #name="{ row }">
          <div class="flex align-center">
            <el-tooltip effect="dark" placement="top" class="flex-sub" :raw-content="true">
              <div class="flex-sub ml-2 ellipsis-two">{{ row.name }}</div>
              <template #content>
                <div style="width: 300px">
                  {{ row.name }}
                </div>
              </template>
            </el-tooltip>
          </div>
        </template>
        <template #operation="{ row }">
          <el-button type="text" size="small" link class="base-btn-link">{{
            t('common.view')
          }}</el-button>

          <el-dropdown
            trigger="click"
            class="ml-4"
            @click.stop
            :show-arrow="false"
            @command="(cmd: string) => handleCommand(cmd, row)"
          >
            <el-icon class="link-hover cursor-pointer"><More /></el-icon>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="handleDownload">
                  {{ t('common.download') }}
                </el-dropdown-item>
                <el-dropdown-item command="handleDelete">
                  <el-button type="danger" size="small" link>{{ t('common.delete') }}</el-button>
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </template>
        <template #slotCard="{ row }">
          <el-card>
            <template #header>
              <span>{{ t('agent.pageDesc.total') }}：{{ pageConfig.total }}</span>
            </template>
            <div class="center">
              <McpImage :src="kymo" width="20" height="20" />
              <div class="flex-sub ml-2 ellipsis-two">{{ row.name }}</div>
            </div>
          </el-card>
        </template>
      </TablePlus>
    </div>
    <FormAgent ref="formAgent"></FormAgent>
  </div>
</template>

<script setup lang="ts">
import { Plus, More } from '@element-plus/icons-vue'
import TablePlus from '@/components/TablePlus/index.vue'
import McpButton from '@/components/mcp-button/index.vue'
import McpImage from '@/components/mcp-image/index.vue'
import FormAgent from './modules/form-dialog.vue'
import { kymo, dify } from '@/utils/logo.ts'
import agentLogo from '@/assets/logo/instance.png'
import { useAgentTableHooks } from './index.ts'

const { t, tablePlus, columns, requestConfig, pageConfig } = useAgentTableHooks()
// 视图模式：'card' 或 'table'
const viewMode = ref<'card' | 'table'>('table')
const loading = ref(false)
const formAgent = ref()

const handleNewAgent = (type: string) => {
  formAgent.value.init(type)
}

const handleEdit = (row: any) => {
  // TODO: 跳转到编辑页或打开弹窗
  console.debug('edit agent', row?.agentId)
}

const handleDelete = (row: any) => {
  // TODO: 删除智能体
  console.debug('delete agent', row?.agentId)
}

const handleCommand = (cmd: string, row: any) => {
  if (cmd === 'handleEdit') return handleEdit(row)
  if (cmd === 'handleDelete') return handleDelete(row)
}

/**
 * Handle init page list data
 */
const init = () => {
  tablePlus.value.initData()
}

onMounted(init)
</script>

<style scoped lang="scss">
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
