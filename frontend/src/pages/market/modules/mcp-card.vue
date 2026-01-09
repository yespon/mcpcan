<template>
  <el-card class="mcp-card">
    <div class="flex gap-1 items-center">
      <mcp-image :src="iconUrl(card)" :key="card.id" width="60" height="60"></mcp-image>
      <div class="flex-1">
        <div class="mb-1">{{ locale === 'zh-cn' ? card.name : card.nameEn }}</div>
        <div class="font-size-3 font-bold">{{ card.githubOwner }}</div>
      </div>
      <el-icon class="cursor-pointer link-hover" title="github" @click="handleJumpToGithub">
        <i class="icon iconfont MCP-GitHub"></i>
      </el-icon>
    </div>
    <div class="my-2 ellipsis-one">
      <el-tag v-for="(tag, index) in card.categoryIds" :key="index" class="mr-1">{{
        translationTag(tag.code)
      }}</el-tag>
    </div>
    <div class="ellipsis-three">
      {{ locale === 'zh-cn' ? card.description : card.descriptionEn }}
    </div>
    <template #footer>
      <div class="flex justify-between items-center">
        <div class="flex gap-4">
          <div class="flex flex-col items-center">
            <el-icon class="mb-1">
              <i class="icon iconfont MCP-fork"></i>
            </el-icon>
            <div>{{ githubNumber(card.githubForksCount) }}</div>
          </div>
          <div class="flex flex-col items-center ml-4">
            <el-icon class="mb-1">
              <i class="icon iconfont MCP-GitHub"></i>
            </el-icon>
            <div>{{ githubNumber(card.githubStargazersCount) }}</div>
          </div>
        </div>
        <mcp-button @click="handleJumpToDetail">{{ t('market.action.viewDetail') }}</mcp-button>
      </div>
    </template>
  </el-card>
</template>
<script setup lang="ts">
import McpImage from '@/components/mcp-image/index.vue'
import { useRouterHooks } from '@/utils/url'
import { useMcpStore } from '@/stores/modules/mcp-store'
import { githubNumber } from '@/utils/system'
import baseConfig from '@/config/base_config.ts'
import McpButton from '@/components/mcp-button/index.vue'

const { jumpToPage } = useRouterHooks()
const { t, locale } = useI18n()
const mcpStore = useMcpStore()

const props = defineProps({
  card: {
    type: Object,
    default: () => ({}),
  },
})
const iconUrl = computed(() => {
  return (card: any) => {
    if (card.iconUrl) {
      return baseConfig.MAIN_SITE_URL + card.iconUrl
    } else if (card.githubOwnerAvatarUrl) {
      return card.githubOwnerAvatarUrl
    } else {
      return ''
    }
  }
})

const translationTag = (code: string) => {
  return t('market.type.' + code)
}

/**
 * handle jump to github
 */
const handleJumpToGithub = () => {
  window.open(props.card.githubRepoUrl, '_blank')
}

const handleJumpToDetail = () => {
  mcpStore.currentMCP = props.card
  jumpToPage({ url: '/market-detail', data: {} })
}
</script>

<style lang="scss" scoped>
.mcp-card {
  // min-width: 230px;
}
</style>
