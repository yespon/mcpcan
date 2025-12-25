<template>
  <el-card class="mcp-card">
    <div class="flex gap-1 items-center">
      <mcp-image width="60" height="60"></mcp-image>
      <div class="flex-1">
        <div class="mb-1">MCP-NAME</div>
        <div class="font-size-3 font-bold">作者</div>
      </div>
      <el-icon class="cursor-pointer link-hover" title="github">
        <i class="icon iconfont MCP-GitHub"></i>
      </el-icon>
    </div>
    <div class="my-2">
      <el-tag>标签</el-tag>
    </div>
    <div class="ellipsis-three">
      Lorem ipsum dolor sit amet consectetur adipisicing elit. Culpa laudantium esse enim adipisci.
      Recusandae tenetur mollitia obcaecati vero perspiciatis suscipit, qui ex blanditiis at.
      Repellat, pariatur amet. Laudantium, ipsa enim?
    </div>
    <template #footer>
      <div class="flex justify-between items-center">
        <div class="flex gap-4">
          <div>
            <div>
              <el-icon class="cursor-pointer mr-6 link-hover">
                <i class="icon iconfont MCP-fork"></i>
              </el-icon>
            </div>
            <div>25</div>
          </div>
          <div class="ml-4">
            <div>
              <el-icon class="cursor-pointer mr-6 link-hover">
                <i class="icon iconfont MCP-GitHub"></i>
              </el-icon>
            </div>
            <div>25</div>
          </div>
        </div>
        <el-button type="primary" @click="handleJumpToDetail">查看详情</el-button>
      </div>
    </template>
  </el-card>
</template>
<script setup lang="ts">
import McpImage from '@/components/mcp-image/index.vue'
import { useRouterHooks } from '@/utils/url'

const { jumpToPage } = useRouterHooks()

const handleJumpToDetail = () => {
  jumpToPage({ url: '/market-detail', data: {} })
}
</script>

<style lang="scss" scoped>
.mcp-card {
}
</style>
