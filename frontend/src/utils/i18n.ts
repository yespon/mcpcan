import i18n from '@/lang'

// 提供一个在模块中安全使用的 t 函数包装器（基于 i18n 实例）
export function t(key: string, value?: any): string {
  return i18n.global.t(key, value) as string
}
