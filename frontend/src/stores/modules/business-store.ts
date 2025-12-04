import { defineStore } from 'pinia'
import { store } from '@/stores'

export const useBusinessStore = defineStore('business', () => {
  const taskInfo = ref({
    visible: false,
    list: [
      {
        id: 1,
        title: '同步MCP服务_2025/11/14 14:30:22',
        status: 'doing',
        detail: '正在同步服务详细内容……',
        expanded: false,
      },
      {
        id: 2,
        title: '同步MCP服务_2025/11/14 14:30:22',
        status: 'done',
        detail: '同步已完成，所有服务均已同步成功。',
        expanded: false,
      },
      {
        id: 3,
        title: '同步MCP服务_2025/11/14 14:30:22',
        status: 'done',
        detail: '同步已完成，所有服务均已同步成功。',
        expanded: false,
      },
      {
        id: 4,
        title: '同步MCP服务_2025/11/14 14:30:22',
        status: 'error',
        detail: '同步过程中发生异常，请检查网络或服务状态。',
        expanded: false,
      },
    ],
  })
  return {
    taskInfo,
  }
})
export function useBusinessStoreHook() {
  return useBusinessStore(store)
}
