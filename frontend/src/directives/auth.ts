import type { Directive } from 'vue'
import { useUserStore } from '@/stores/modules/user-store.ts'

const applyAuth = (el: HTMLElement, value?: string) => {
  if (!value) return

  const { currentBtnAuths } = storeToRefs(useUserStore())
  const permissions = value
    .split(',')
    .map((s) => s.trim())
    .filter(Boolean)

  // 这里以后要从 store 中取出用户的权限 进行判断
  const hasPermission = permissions.some((permission: string) => {
    return currentBtnAuths.value.includes(permission)
  })

  if (!hasPermission) {
    el.parentNode && el.parentNode.removeChild(el)
  }
}

export default {
  mounted(el, binding) {
    applyAuth(el as HTMLElement, binding.value)
  },
  updated(el, binding) {
    // 当用户权限在页面打开后才加载/刷新时，让指令能重新生效
    if (binding.value !== binding.oldValue) {
      applyAuth(el as HTMLElement, binding.value)
    }
  },
} as Directive
