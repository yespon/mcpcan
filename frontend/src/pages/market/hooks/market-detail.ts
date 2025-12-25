import { useRouterHooks } from '@/utils/url'

export const useMarketDetailHooks = () => {
  const { t } = useI18n()
  const { jumpBack, jumpToPage } = useRouterHooks()

  return { t, jumpBack }
}
