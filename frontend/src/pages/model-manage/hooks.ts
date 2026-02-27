import { ChatAPI } from '@/api/agent'
import { timestampToDate } from '@/utils/system'
import { useI18n } from 'vue-i18n'

export const useModelTableHooks = () => {
  const { t } = useI18n()
  const tablePlus = ref()
  const pageInfo = ref({
    loading: false,
    loadingText: t('common.loading'),
  })

  const columns = ref<any>([
    {
      label: t('model.name'),
      dataIndex: 'name',
      searchConfig: {
        component: 'el-input',
        label: t('model.name'),
        props: {
          placeholder: t('model.namePlaceholder'),
        },
      },
    },
    {
      label: t('model.provider'),
      dataIndex: 'provider',
      searchConfig: {
        component: 'el-input',
        label: t('model.provider'),
        props: {
          placeholder: t('model.providerPlaceholder'),
        },
      },
    },
    {
      label: t('model.modelName'),
      dataIndex: 'modelName',
    },
    {
      label: t('model.baseUrl'),
      dataIndex: 'baseUrl',
    },
    {
      dataIndex: 'updatedAt',
      label: t('common.updatedAt'),
      customRender: ({ row }: { row: any }) => {
        return timestampToDate(row.updatedAt)
      },
    },
  ])

  const requestConfig = {
    api: ChatAPI.listModelAccess,
    searchQuery: {
      model: {},
    },
  }

  const pageConfig = ref({
    total: 0,
    page: 1,
    pageSize: 10,
  })

  return { t, columns, requestConfig, tablePlus, pageConfig, pageInfo }
}
