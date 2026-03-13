import { DeptAPI } from '@/api/system/index.ts'
import { useRouterHooks } from '@/utils/url'
import { timestampToDate } from '@/utils/system'

export const useDeptTableHooks = () => {
  const { t } = useI18n()
  const pageInfo = ref({
    loading: false,
  })
  const requestConfig = ref({
    api: DeptAPI.list,
    searchQuery: {
      model: {
        parentId: null,
      },
    },
  })
  const { jumpToPage, reload } = useRouterHooks()
  const columns = ref([
    {
      label: t('system.department.columns.name'),
      dataIndex: 'name',
      searchConfig: {
        component: 'el-input',
        label: t('system.department.columns.name'),
        props: {
          placeholder: t('system.department.placeholder.name'),
        },
      },
    },
    {
      label: t('system.department.columns.deptSort'),
      dataIndex: 'sort',
    },
    {
      label: t('system.department.columns.status'),
      dataIndex: 'status',
      searchConfig: {
        component: 'el-select',
        label: t('system.department.columns.status'),
        props: {
          placeholder: t('system.department.placeholder.status'),
          options: [
            { label: t('system.user.status.enabled'), value: 1 },
            { label: t('system.user.status.disabled'), value: 2 },
          ],
        },
      },
    },
    {
      label: t('system.department.columns.createTime'),
      dataIndex: 'createdAt',
      customRender: ({ row }: { row: any }) => {
        return timestampToDate(row.createdAt)
      },
    },
  ])
  const pageConfig = ref({
    total: 0,
    page: 1,
    pageSize: 10,
  })

  return {
    t,
    pageInfo,
    requestConfig,
    pageConfig,
    columns,
    jumpToPage,
    reload,
  }
}
