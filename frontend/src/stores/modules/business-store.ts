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
        handleGetTaskList()
        taskTimer.value = setInterval(() => {
          handleGetTaskList()
        }, 10 * 1000)
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
    const oldList = taskInfo.value.list || []

    const incomingMap = new Map<string | number, any>()
    list?.forEach((item: any) => {
      if (item.id !== undefined) incomingMap.set(item.id, item)
    })
    // 1) 用新数据覆盖旧数据的属性；旧中不存在于新数据的元素保留
    const merged = oldList.map((old: any) => {
      if (old.id !== undefined && incomingMap.has(old.id)) {
        // 以新数据覆盖属性（新优先）
        return { ...old, ...incomingMap.get(old.id) }
      }
      return old
    })

    // 2) 将新数据中不存在于旧数据的元素追加到末尾
    const oldKeys = new Set(merged.map((o: any) => o.id))
    const additions = (list || []).filter((item: any) => !oldKeys.has(item.id))
    const mergedAll = merged.concat(additions)
    // 根据创建时间降序排序，最新在前
    taskInfo.value.list = mergedAll.sort((a: any, b: any) => {
      const ta = new Date(a?.createdAt || 0).getTime()
      const tb = new Date(b?.createdAt || 0).getTime()
      return tb - ta
    })
  }
  return {
    taskInfo,
    handleGetTaskList,
  }
})
export function useBusinessStoreHook() {
  return useBusinessStore(store)
}
