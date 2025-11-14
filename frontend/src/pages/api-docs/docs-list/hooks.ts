import { DocsAPI } from '@/api/api-docs/index'
import type { Code } from '@/types'
import { timestampToDate, formatFileSize } from '@/utils/system'

export const useCodeTableHooks = () => {
  const { t } = useI18n()
  const tablePlus = ref()
  const pageInfo = ref({
    loading: false,
    loadingText: t('api.action.loadingText'),
  })
  const columns = ref<any>([
    {
      label: t('api.columns.name'),
      dataIndex: 'name',
      searchConfig: {
        component: 'el-input',
        label: t('api.columns.name'),
        props: {
          placeholder: t('api.columns.name'),
        },
      },
    },
    {
      label: t('api.columns.size'),
      dataIndex: 'size',
      customRender: ({ row }: { row: Code }) => {
        return formatFileSize(row.size)
      },
    },
    {
      label: t('api.columns.type'),
      dataIndex: 'types',
      searchConfig: {
        component: 'el-select',
        label: t('api.columns.type'),
        props: {
          placeholder: t('api.columns.type'),
          multiple: true,
          options: [
            { label: t('api.columns.json'), value: 1 },
            { label: t('api.columns.yaml'), value: 2 },
          ],
        },
      },
      customRender: ({ row }: { row: Code }) => {
        return [t('api.columns.unspecified'), t('api.columns.json'), t('api.columns.yaml')][
          row.type
        ]
      },
    },
    {
      dataIndex: 'createdAt',
      label: t('api.columns.createdAt'),
      customRender: ({ row }: { row: Code }) => {
        return timestampToDate(row.createdAt)
      },
    },
    {
      dataIndex: 'updatedAt',
      label: t('api.columns.updatedAt'),
      customRender: ({ row }: { row: Code }) => {
        return timestampToDate(row.updatedAt)
      },
    },
  ])
  const requestConfig = {
    api: DocsAPI.list,
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
