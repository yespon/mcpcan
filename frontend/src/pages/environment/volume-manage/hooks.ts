import { VolumeAPI } from '@/api/env'

export const usePvcTableHooks = () => {
  const { t } = useI18n()
  const { query, meta } = useRoute()
  const tablePlus = ref()
  const columns = computed<any>(() => {
    return [
      {
        dataIndex: 'name',
        label: t('env.volume.name'),
        searchConfig: {
          component: 'el-input',
          label: t('env.volume.name'),
          props: {
            placeholder: t('env.volume.name'),
          },
        },
      },
      {
        dataIndex: 'driver',
        label: t('env.volume.driver'),
      },
      {
        dataIndex: 'mountpoint',
        label: t('env.volume.mountpoint'),
      },
      {
        dataIndex: 'scope',
        label: t('env.volume.scope'),
      },
      { dataIndex: 'createdAt', label: t('env.pvc.creationTime') },
    ]
  })
  // ref()
  const requestConfig = {
    api: VolumeAPI.list,
    searchQuery: {
      model: {
        environmentId: query.environmentId,
      },
    },
  }

  const pageConfig = ref({
    total: 0,
    page: 1,
    pageSize: 10,
  })
  const newVolumeDialog = ref()

  return {
    VolumeAPI,
    tablePlus,
    columns,
    requestConfig,
    pageConfig,
    newVolumeDialog,
    query,
    meta,
  }
}
