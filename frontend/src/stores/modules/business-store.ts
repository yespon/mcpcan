import { defineStore } from 'pinia'
import { store } from '@/stores'
import { AgentAPI } from '@/api/agent'

export const useBusinessStore = defineStore('business', () => {
  const taskInfo = ref({
    visible: false,
    list: [] as any[],
  })
  watch(
    () => taskInfo.value.visible,
    (newVal) => {
      if (newVal) {
        handleGetTaskList()
      }
    },
  )
  /**
   * Handle get task list
   */
  const handleGetTaskList = async (keyword?: string) => {
    const { list } = await AgentAPI.taskList({ keyword })
    taskInfo.value.list = list
  }
  return {
    taskInfo,
    handleGetTaskList,
  }
})
export function useBusinessStoreHook() {
  return useBusinessStore(store)
}
