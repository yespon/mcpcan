import { useRouterHooks } from '@/utils/url'
import { useMcpStore } from '@/stores/modules/mcp-store'

export const useMarketDetailHooks = () => {
  const { t, locale } = useI18n()
  const { jumpBack, jumpToPage } = useRouterHooks()
  const { currentMCP } = useMcpStore()

  return { t, locale, jumpBack, jumpToPage, currentMCP }
}
