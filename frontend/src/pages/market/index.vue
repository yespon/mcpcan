<template>
  <div>
    <div><el-card></el-card></div>
    <div v-loading="loading" class="flex gap-4 mt-4">
      <div class="w-75 type-sticky">
        <el-input
          ref="searchInputRef"
          v-model="keyword"
          placeholder="搜索MCP服务名称或是关键词...."
          class="mb-4"
          :suffix-icon="Search"
          @keyup.enter="handleQuery"
        />
        <el-card>
          <div class="flex justify-between items-center">
            <span class="my-1">筛选</span>
            <el-button v-if="activeType" type="primary" size="small" @click="clearType">
              清除
            </el-button>
          </div>
          <!-- 分类菜单（左侧菜单样式） -->
          <div class="mt-3 flex flex-col gap-2">
            <div
              v-for="type in typeMap"
              :key="type.value"
              class="w-full flex items-center gap-3 rounded-md px-3 py-2 text-left type-item"
              :class="{ 'active-type': activeType === type.value }"
              @click="selectType(type.value)"
            >
              <el-icon>
                <i class="icon iconfont" :class="type.icon"></i>
              </el-icon>
              <span class="text-sm">{{ type.label }}</span>
            </div>
          </div>
        </el-card>
      </div>
      <div class="flex-1">
        <template v-if="marketList.length || true">
          <div class="grid grid-cols-12 gap-4">
            <McpCard
              class="col-span-3"
              v-for="(card, index) in marketList"
              :card="card"
              :key="index"
            ></McpCard>
            <McpCard class="col-span-3"></McpCard>
          </div>
          <div class="mt-8 flex justify-end">
            <el-pagination
              background
              :total="pagerConfig.total"
              :page="pagerConfig.page"
              :limit="pagerConfig.pageSize"
              @current-change="handlePageChange"
            />
          </div>
        </template>
        <el-empty v-else class="mt-40"></el-empty>
      </div>
    </div>
  </div>
</template>
<script setup lang="ts">
import { Search } from '@element-plus/icons-vue'
import { useMarketListHooks } from './hooks/index.ts'
import McpCard from './modules/mcp-card.vue'
import { MarketAPI } from '@/api/market/index.ts'

const { t, loading, typeMap, activeType, keyword, searchInputRef, pagerConfig } =
  useMarketListHooks()

const marketList = ref([])

// handle selected type
const selectType = (value: string) => {
  activeType.value = value
  handleGetMarketList()
}

// handle clear type
const clearType = () => {
  activeType.value = ''
  keyword.value = ''
  handleGetMarketList()
}

const handleQuery = () => {
  handleGetMarketList()
}

const handlePageChange = (newPage: number) => {
  pagerConfig.page = newPage
  handleGetMarketList()
}

// handle get market list data
const handleGetMarketList = async () => {
  const { data } = await MarketAPI.list({
    page: pagerConfig.page,
    pageSize: pagerConfig.pageSize,
    name: keyword.value,
    categoryName: activeType.value,
  })
  marketList.value = data || []
}

onMounted(() => {
  handleGetMarketList()
  searchInputRef.value.$el.getElementsByClassName('el-input__suffix')[0].onclick = handleQuery
})
</script>

<style lang="scss" scoped>
.type-sticky {
  position: sticky;
  top: 0px;
  align-self: flex-start;
  height: fit-content;
  z-index: 1;
}
:deep(.el-input__suffix) {
  cursor: pointer;
}
.type-item {
  transition: all 0.3s ease;
  overflow: hidden;
  background-color: var(--ep-bg-purple-color);
  border: 1px solid transparent;
  &.active-type {
    background-color: var(--ep-bg-purple-color-deep);
  }
  &:hover {
    scale: 1.02;
    cursor: pointer;
    background-color: var(--ep-bg-purple-color-deep);
    border-color: var(--ep-btn-color-top);
  }
}
</style>
