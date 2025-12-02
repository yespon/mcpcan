import { AgentAPI } from '@/api/agent/index'
import { timestampToDate } from '@/utils/system'

export const useAgentTableHooks = () => {
  const { t } = useI18n()
  const tablePlus = ref()
  const pageInfo = ref({
    loading: false,
    loadingText: t('api.action.loadingText'),
  })
  const columns = ref<any>([
    {
      label: t('agent.columns.accessName'),
      dataIndex: 'accessName',
      searchConfig: {
        component: 'el-input',
        label: t('agent.columns.accessName'),
        props: {
          placeholder: t('agent.columns.accessName'),
        },
      },
    },
    {
      label: t('agent.columns.accessType'),
      dataIndex: 'accessType',
      searchConfig: {
        component: 'el-select',
        label: t('agent.columns.accessType'),
        props: {
          placeholder: t('agent.columns.accessType'),
          options: [
            { label: 'MySQL', value: 'mysql' },
            { label: 'PostgreSQL', value: 'postgres' },
            { label: 'SQLServer', value: 'sqlserver' },
            { label: 'SQLite', value: 'sqlite' },
          ],
        },
      },
    },
    {
      label: t('agent.columns.dbHost'),
      dataIndex: 'dbHost',
    },
    {
      dataIndex: 'createdAt',
      label: t('api.columns.createdAt'),
      customRender: ({ row }: { row: any }) => {
        return timestampToDate(row.createdAt)
      },
    },
    {
      dataIndex: 'updatedAt',
      label: t('api.columns.updatedAt'),
      customRender: ({ row }: { row: any }) => {
        return timestampToDate(row.updatedAt)
      },
    },
  ])
  const requestConfig = {
    api: AgentAPI.list,
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
