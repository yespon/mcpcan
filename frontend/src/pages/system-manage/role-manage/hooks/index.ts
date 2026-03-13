import { RoleAPI } from '@/api/system/index.ts'
import { useRouterHooks } from '@/utils/url'
import { timestampToDate } from '@/utils/system'

export const useRoleTableHooks = () => {
  const { t } = useI18n()
  const pageInfo = ref({
    loading: false,
  })
  const requestConfig = ref({
    api: RoleAPI.list,
    searchQuery: {
      model: {},
    },
  })
  const { jumpToPage } = useRouterHooks()
  const columns = ref([
    {
      label: t('system.role.columns.name'),
      dataIndex: 'name',
      searchConfig: {
        component: 'el-input',
        label: t('system.role.columns.name'),
        props: {
          placeholder: t('system.role.placeholder.name'),
        },
      },
    },
    {
      label: t('system.role.columns.level'),
      dataIndex: 'level',
    },
    {
      label: t('system.role.columns.dataScope'),
      dataIndex: 'dataScope',
    },
    {
      label: t('system.role.columns.description'),
      dataIndex: 'description',
      // searchConfig: {
      //   component: 'el-input',
      //   label: t('system.role.columns.description'),
      //   props: {
      //     placeholder: t('system.role.placeholder.description'),
      //   },
      // },
    },
    {
      label: t('system.role.columns.createTime'),
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
    RoleAPI,
  }
}
