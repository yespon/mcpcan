export const useMarketListHooks = () => {
  const { t } = useI18n()
  const loading = ref(false)
  const searchInputRef = ref()
  const keyword = ref('')
  const activeType = ref<string>('')
  const typeMap = ref([
    { label: '开发工具', value: 'dev-tools', icon: 'MCP-kaifagongju' },
    { label: '效率工具', value: 'productivity-tools', icon: 'MCP-xiaoshuaigongju' },
    { label: '实用工具', value: 'utilities', icon: 'MCP-shiyonggongjuji' },
    { label: '信息检索', value: 'information-search', icon: 'MCP-xinxijiansuo' },
    { label: '媒体生成', value: 'media-generation', icon: 'MCP-meiti' },
    { label: '商业服务', value: 'business-services', icon: 'MCP-shangyefuwu' },
    { label: '科学教育', value: 'science-education', icon: 'MCP-kexuejiaoyu' },
    { label: '股票金融', value: 'finance', icon: 'MCP-financialStatements' },
  ])
  const pagerConfig = reactive({
    total: 0,
    page: 1,
    pageSize: 8,
  })

  return { t, loading, typeMap, activeType, keyword, searchInputRef, pagerConfig }
}
