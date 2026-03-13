import { RoleAPI, UserAPI } from '@/api/system/index.ts'

export const useUserListRoleTableHooks = () => {
  const { t } = useI18n()
  const { query } = useRoute()
  const requestConfig = ref({
    api: UserAPI.list,
    searchQuery: {
      model: {
        roleId: query.roleId,
      },
    },
  })
  const columns = ref([
    {
      label: t('system.user.columns.deptName'),
      dataIndex: 'deptName',
    },
    {
      label: t('system.user.columns.username'),
      dataIndex: 'username',
      searchConfig: {
        component: 'el-input',
        label: t('system.user.columns.username'),
        props: {
          placeholder: t('system.user.placeholder.userName'),
        },
      },
    },
    {
      label: t('system.user.columns.nickName'),
      dataIndex: 'nickName',
      searchConfig: {
        component: 'el-input',
        label: t('system.user.columns.nickName'),
        props: {
          placeholder: t('system.user.placeholder.nickName'),
        },
      },
    },
    {
      label: t('system.user.columns.email'),
      dataIndex: 'email',
    },
    {
      label: t('system.user.columns.enabled'),
      dataIndex: 'status',
    },
  ])
  const pageConfig = ref({
    total: 0,
    page: 1,
    pageSize: 10,
  })

  return {
    t,
    requestConfig,
    columns,
    pageConfig,
    query,
  }
}
