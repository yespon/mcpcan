<template>
  <div>
    <!-- Header Info -->
    <div class="flex justify-between items-start">
      <div class="flex gap-6">
        <!-- Logo -->
        <div
          class="w-24 h-24 rounded-2xl border border-gray-200 flex items-center justify-center bg-white shadow-sm overflow-hidden p-2"
        >
          <img src="@/assets/logo.png" alt="Logo" class="w-full h-full object-contain" />
        </div>

        <!-- Content -->
        <div class="flex flex-col justify-between py-1">
          <h1 class="text-3xl font-bold leading-tight">MCP-Auth 服务</h1>

          <div class="flex items-center gap-3 text-gray-500 text-sm mt-1">
            <span class="font-bold text-base">@upstash</span>
            <span class="w-1 h-1 rounded-full bg-gray-400"></span>
            <span>最新发布时间: 12月24, 2025</span>
          </div>
          <div class="flex items-center gap-8 mt-3">
            <el-tag type="primary" effect="dark" class="!border-none px-5 py-1.5 text-sm">
              开发工具
            </el-tag>
            <div class="flex items-center gap-2 text-gray-600">
              <el-icon class="cursor-pointer">
                <i class="icon iconfont MCP-fork"></i>
              </el-icon>
              <span class="text-base font-medium">302</span>
            </div>
            <div class="flex items-center gap-2 text-gray-600">
              <el-icon class="cursor-pointer">
                <i class="icon iconfont MCP-GitHub"></i>
              </el-icon>
              <span class="text-base font-medium">2.3K</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Github Button -->
      <div>
        <div v-if="layout" class="mb-4 text-right">
          <el-button @click="handleBack" class="link-hover">
            <el-icon>
              <i class="icon iconfont MCP-fanhui"></i>
            </el-icon>
            {{ t('common.back') }}
          </el-button>
        </div>
        <div>
          <el-button round size="large" @click="goGithub">
            <template #icon>
              <el-icon class="cursor-pointer link-hover">
                <i class="icon iconfont MCP-GitHub"></i>
              </el-icon>
            </template>
            GitHub
          </el-button>
        </div>
      </div>
    </div>

    <el-divider />

    <div class="grid grid-cols-1 lg:grid-cols-3 gap-4">
      <!-- Left Column -->
      <div class="lg:col-span-2 flex flex-col gap-6">
        <!-- Description Card -->
        <el-card>
          <h2 class="text-lg font-bold mb-4">你可以使用该 MCP Server 做什么？</h2>
          <p class="leading-relaxed text-sm">
            MCP Lite Wrappers MCP
            服务器是一个轻量级平台，将多种工具整合为单操作命令，以优化效率。您可以执行与 Supabase
            相关的所有操作，包括管理项目、组织和存储；执行 SQL
            查询；处理迁移和扩展；部署边缘函数；管理分支；以及检索日志和项目详细信息。
          </p>
        </el-card>

        <!-- Readme Card -->
        <el-card>
          <h2 class="text-3xl font-bold mb-8">MCP-Auth （readme.md原文信息）</h2>
          <div class="max-w-none">
            <p class="leading-relaxed">
              MCP-Auth 是一个高性能的身份认证与授权服务，为整个 MCP
              平台提供安全可靠的身份验证机制。该服务支持多种认证方式，包括 OAuth 2.0、JWT Token、API
              Key
              等，并提供细粒度的权限控制。通过分布式架构设计，确保服务的高可用性和可扩展性，能够处理大规模并发请求。
            </p>
            <!-- Placeholder for more markdown content -->
          </div>
        </el-card>
      </div>

      <!-- Right Column -->
      <div class="lg:col-span-1 flex flex-col gap-6">
        <!-- Config Card -->
        <el-card>
          <h2 class="text-lg font-bold mb-4">配置</h2>
          <div class="bg-gray-100 rounded-lg p-4 border border-gray-200 overflow-hidden">
            <pre
              class="text-xs font-mono text-gray-600 overflow-x-auto whitespace-pre-wrap break-all"
            >
{
  "mcpServers": {
    "time": {
      "command": "uvx",
      "args": [
        "mcp-server-time",
        "--local-timezone=America/New_York"
      ]
    }
  }
}
</pre
            >
          </div>
        </el-card>

        <!-- Action Button -->
        <el-button
          type="primary"
          size="large"
          class="!w-full !h-12 !text-base !rounded-lg !font-medium"
          @click="handleCreate"
        >
          快速创建
        </el-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useMarketDetailHooks } from './hooks/market-detail.ts'

const layout = useLayout()
const { t, jumpBack } = useMarketDetailHooks()

// back last class page
const handleBack = () => {
  jumpBack()
}

const goGithub = () => {
  // TODO: Implement GitHub navigation logic
  // const url = 'https://github.com/...'
  // window.open(url, '_blank')
  console.log('Navigate to GitHub')
}

const handleCreate = () => {
  console.log('Quick create')
}
</script>

<style lang="scss" scoped>
:deep(.el-button) {
  --el-button-hover-bg-color: #fff;
  --el-button-hover-text-color: var(--el-color-primary);
  --el-button-hover-border-color: var(--el-color-primary-light-7);
}
</style>
