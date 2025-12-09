import { defineStore } from 'pinia'
import { store } from '@/stores'
import { AgentAPI } from '@/api/agent'

export const useBusinessStore = defineStore('business', () => {
  const taskInfo = ref({
    visible: false,
    list: [] as any[],
  })
  const taskTimer = ref(0)
  watch(
    () => taskInfo.value.visible,
    (newVal) => {
      // 当任务列表展开时开启定时任务请求
      if (newVal) {
        taskTimer.value = setInterval(() => {
          handleGetTaskList()
        }, 30 * 1000)
      } else {
        clearInterval(taskTimer.value)
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
