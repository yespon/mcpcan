import { UserAPI } from '@/api/system/index.ts'
import { useRouterHooks } from '@/utils/url'
import { timestampToDate } from '@/utils/system'

export const useUserListTableHooks = () => {
  const { t } = useI18n()
  const pageInfo = ref({
    loading: false,
  })
  const requestConfig = ref<any>({
    api: UserAPI.list,
    searchQuery: {
      model: {},
    },
  })
  const { jumpToPage } = useRouterHooks()
  const pageConfig = ref({
    total: 0,
    page: 1,
    pageSize: 10,
  })
  const columns = ref([
    {
      label: t('system.user.columns.username'),
      dataIndex: 'username',
      searchConfig: {
        component: 'el-input',
        label: t('system.user.search.name'),
        props: {
          placeholder: t('system.user.search.name'),
        },
      },
    },
    {
      label: t('system.user.columns.nickName'),
      dataIndex: 'nickName',
      // searchConfig: {
      //   component: 'el-input',
      //   label: t('system.user.columns.nickName'),
      //   props: {
      //     placeholder: t('system.user.placeholder.nickName'),
      //   },
      // },
    },

    {
      label: t('system.user.columns.phone'),
      dataIndex: 'phone',
    },
    {
      label: t('system.user.columns.email'),
      dataIndex: 'email',
    },
    {
      label: t('system.user.columns.department'),
      dataIndex: 'deptName',
    },
    {
      label: t('system.user.columns.roles'),
      dataIndex: 'roleNames',
      customRender: ({ row }: { row: any }) => {
        return row.roleNames?.join(', ')
      },
    },
    {
      label: t('system.user.columns.enabled'),
      dataIndex: 'status',
      searchConfig: {
        component: 'el-select',
        label: t('system.user.columns.enabled'),
        props: {
          placeholder: t('system.user.columns.enabled'),
          options: [
            { label: t('system.user.status.enabled'), value: 1 },
            { label: t('system.user.status.disabled'), value: 2 },
          ],
        },
      },
    },
    {
      label: t('system.user.columns.createTime'),
      dataIndex: 'createdAt',
      customRender: ({ row }: { row: any }) => {
        return timestampToDate(row.createdAt)
      },
    },
  ])
  return { t, pageInfo, requestConfig, jumpToPage, columns, pageConfig }
}
