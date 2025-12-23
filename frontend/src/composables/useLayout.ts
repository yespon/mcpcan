/**
 * 全局 layout 状态 composable
 * 提供响应式的 layout 变量，值为 route.meta.hideLayout
 * 任何组件内都可以直接使用 const layout = useLayout() 来访问
 */
export function useLayout() {
  const route = useRoute()

  // 响应式计算属性，当路由变化时自动更新
  const layout = computed(() => route.meta.hideLayout)

  return layout
}
