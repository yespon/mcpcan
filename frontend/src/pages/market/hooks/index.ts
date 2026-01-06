import { useStorage } from '@vueuse/core'

export const useMarketListHooks = () => {
  const { t } = useI18n()
  const loading = ref(false)
  const searchInputRef = ref()
  const keyword = useStorage('market-list-keyword', '')
  const categoryName = useStorage('market-list-category', '')
  const typeMap = computed(() => [
    { label: t('market.type.test'), value: 'test', icon: 'MCP-ceshi' },
    {
      label: t('market.type.browser-automation'),
      value: 'browser-automation',
      icon: 'MCP-liulanqi',
    },
    { label: t('market.type.search'), value: 'search', icon: 'MCP-xinxijiansuo' },
    { label: t('market.type.communication'), value: 'communication', icon: 'MCP-jiaoliu' },
    { label: t('market.type.developer-tools'), value: 'developer-tools', icon: 'MCP-kaifagongju' },
    {
      label: t('market.type.entertainment-and-media'),
      value: 'entertainment-and-media',
      icon: 'MCP-meiti',
    },
    { label: t('market.type.file-systems'), value: 'file-systems', icon: 'MCP-wenjian1' },
    { label: t('market.type.finance'), value: 'finance', icon: 'MCP-financialStatements' },
    {
      label: t('market.type.knowledge-and-memory'),
      value: 'knowledge-and-memory',
      icon: 'MCP-zhishi',
    },
    { label: t('market.type.location-services'), value: 'location-services', icon: 'MCP-position' },
    { label: t('market.type.art-and-culture'), value: 'art-and-culture', icon: 'MCP-yishu' },
    {
      label: t('market.type.research-and-data'),
      value: 'research-and-data',
      icon: 'MCP-kexuejiaoyu',
    },
    {
      label: t('market.type.calendar-management'),
      value: 'calendar-management',
      icon: 'MCP-richeng',
    },
    {
      label: t('market.type.efficiency-tools'),
      value: 'efficiency-tools',
      icon: 'MCP-xiaoshuaigongju',
    },
    { label: t('market.type.other'), value: 'other', icon: 'MCP-qita' },
  ])
  const pagerConfig = useStorage('market-list-pager', {
    total: 0,
    page: 1,
    pageSize: 12,
  })

  return { t, loading, typeMap, categoryName, keyword, searchInputRef, pagerConfig }
}
