import { Storage } from '@/utils/storage'
import { useSystemStoreHook } from '@/stores/modules/system-store'

// 时间戳转换为标准显示时间
export const timestampToDate = (time: number | string, format: string = 'YYYY-MM-DD HH:mm:ss') => {
  let date: Date
  if (typeof time === 'string') {
    // 若为字符串，先尝试解析为时间戳（如"1761201587000"），失败则直接作为字符串时间解析
    const timestamp = Number(time)
    date = isNaN(timestamp) ? new Date(time.replace(' CST', '')) : new Date(timestamp)
  } else {
    // 若为数字，直接作为时间戳解析
    date = new Date(time)
  }

  // 2. 验证时间有效性（兼容Invalid Date）
  if (isNaN(date.getTime())) {
    // console.warn(`无效的时间格式：${time}`)
    return '--'
  }

  // 提取时间各部分（月份从0开始，需+1）
  const year = date.getFullYear()
  const month = date.getMonth() + 1
  const day = date.getDate()
  const hour = date.getHours()
  const minute = date.getMinutes()
  const second = date.getSeconds()
  // 补零工具函数（确保两位数）
  const padZero = (num: number, length: number = 2): string => num.toString().padStart(length, '0')

  // 替换格式字符串
  return format
    .replace('YYYY', year.toString())
    .replace('MM', padZero(month))
    .replace('DD', padZero(day))
    .replace('HH', padZero(hour))
    .replace('mm', padZero(minute)) // 注意：分钟用mm避免与月份MM冲突
    .replace('ss', padZero(second))
}

// github 数据转换
export const githubNumber = (num: number | string) => {
  const n = Number(num)
  if (!n) return 0
  return n >= 1000 ? (n / 1000).toFixed(1) + 'k' : n
}

/**
 * 文件大小转换（字节转 KB/MB/GB/TB）
 * @param bytes 原始文件大小（单位：B，支持数字或字符串类型）
 * @param decimalPlaces 保留的小数位数（默认 2 位，可选 0-10）
 * @returns 格式化后的大小字符串（如 "1.23 MB"），无效输入返回 "0 B"
 */
export function formatFileSize(bytes: number | string, decimalPlaces: number = 2): string {
  // 1. 处理输入：转换为数字并校验有效性
  let byteNum: number
  if (typeof bytes === 'string') {
    byteNum = Number(bytes.trim())
    // 字符串需能转换为有效数字，且非负数
    if (isNaN(byteNum) || byteNum < 0) {
      return '0 B'
    }
  } else {
    // 数字需为非负且有限（排除 Infinity/NaN）
    if (bytes < 0 || !isFinite(bytes)) {
      return '0 B'
    }
    byteNum = bytes
  }

  // 2. 处理 0 字节的特殊情况
  if (byteNum === 0) {
    return '0 B'
  }

  // 3. 定义单位层级（B → KB → MB → GB → TB，1024 进制）
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const unitStep = 1024 // 1 KB = 1024 B

  // 4. 计算适配的单位层级（如 1024 B → 层级 1 → KB）
  const unitIndex = Math.floor(Math.log(byteNum) / Math.log(unitStep))
  // 防止超出最大单位（超过 TB 仍用 TB 显示）
  const safeIndex = Math.min(unitIndex, units.length - 1)

  // 5. 计算转换后的值并保留指定小数位数
  const convertedSize = byteNum / Math.pow(unitStep, safeIndex)
  // 限制小数位数在 0-10 之间（避免无意义的精度）
  const safeDecimal = Math.max(0, Math.min(decimalPlaces, 10))
  const formattedSize = convertedSize.toFixed(safeDecimal)

  // 6. 拼接结果（移除末尾无意义的 ".00"，如 "2.00 MB" → "2 MB"）
  const finalSize = formattedSize.replace(/\.?0*$/, '') || '0'
  return `${finalSize} ${units[safeIndex]}`
}

/**
 * 复制内容
 */
export const setClipboardData = (data: any) => {
  return new Promise((success) => {
    const textarea: any = document.createElement('textarea')
    textarea.value = data
    textarea.readOnly = 'readOnly'
    document.body.appendChild(textarea)
    textarea.select()
    textarea.setSelectionRange(0, data.length)
    document.execCommand('copy')
    textarea.remove()
    success(data)
  })
}

// 生成 uuid
function genUUID() {
  // 简单 uuid v4 生成
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function (c) {
    const r = (Math.random() * 16) | 0,
      v = c === 'x' ? r : (r & 0x3) | 0x8
    return v.toString(16)
  })
}

// 生成token
export const getToken = (baseInfo: any) => {
  const uuid = genUUID()
  return btoa(uuid + baseInfo)
}

// 将十六进制颜色按 RGB 分量偏移并返回 6 位十六进制字符串（带 '#')
function adjustHexByDeltas(hex: string, dR: number, dG: number, dB: number) {
  if (!hex) return hex
  let h = String(hex).trim()
  if (h.startsWith('#')) h = h.slice(1)
  if (h.length === 3) {
    h = h
      .split('')
      .map((c) => c + c)
      .join('')
  }
  if (h.length !== 6) return `#${h}`

  const clamp = (v: number) => Math.max(0, Math.min(255, Math.round(v)))
  const r = clamp(parseInt(h.slice(0, 2), 16) + dR)
  const g = clamp(parseInt(h.slice(2, 4), 16) + dG)
  const b = clamp(parseInt(h.slice(4, 6), 16) + dB)

  const to2 = (n: number) => n.toString(16).padStart(2, '0')
  return `#${to2(r)}${to2(g)}${to2(b)}`
}

// 获取父级项目缓存信息
// 当前项目作为子项目嵌入父级项目时，尝试访问父级项目的 localStorage
export const getParentLocalStorageItem = (
  key: string,
  timeout = 1000,
): Promise<any | null> | null | string => {
  try {
    // const parentWindow = window.parent || window.top
    // if (!parentWindow) return null
    // 直接访问
    return window.localStorage.getItem(key)
  } catch (err) {
    // 可能因为跨域或 sandbox 抛错
    console.error('无法访问父 localStorage：', err)
    return null
  }
}

// 判断当前项目是否作为子项目嵌入父级项目
export const isEmbeddedInParent = () => {
  try {
    return window.parent && window.parent !== window
  } catch {
    return false
  }
}

//初始化国际化信息
export const initUseI18n = async () => {
  try {
    const systemStore = useSystemStoreHook()
    const locale = await getParentLocalStorageItem('responsive-locale')
    systemStore.language = JSON.parse(locale).locale === 'zh' ? 'zh-cn' : 'en'
  } catch {}
}

//初始化主题信息
export const initThemeInfo = async () => {
  try {
    const systemStore = useSystemStoreHook()
    const theme = await getParentLocalStorageItem('responsive-layout')
    let themeObj = JSON.parse(theme) || {}
    systemStore.themeType = themeObj.overallStyle || Storage.get('theme')
    document.documentElement.style.setProperty(
      '--el-color-primary',
      themeObj.epThemeColor || '#cdbdff',
    )
    document.documentElement.style.setProperty(
      '--el-color-primary-hover',
      (themeObj.epThemeColor || '#cdbdff') + '80',
    )

    document.documentElement.style.setProperty(
      '--ep-bg-purple-color',
      adjustHexByDeltas(themeObj.epThemeColor || '#ccbbff', 1, 2, 0) + '29',
    )
    // 背景色
    document.documentElement.style.setProperty(
      '--ep-bg-purple-color-deep',
      adjustHexByDeltas(themeObj.epThemeColor || '#ccbbff', 1, 2, 0) + '80',
    )
    document.documentElement.style.setProperty(
      '--ep-pager-border',
      adjustHexByDeltas(themeObj.epThemeColor || '#ccbbff', 1, 2, 0),
    )

    // 按钮颜色
    document.documentElement.style.setProperty(
      '--ep-btn-color-top',
      adjustHexByDeltas(themeObj.epThemeColor || ' #a083f7', 1, 2, 0),
    )
    document.documentElement.style.setProperty(
      '--ep-btn-color-bottom',
      adjustHexByDeltas(themeObj.epThemeColor || ' #2a029f', 1, 2, 0),
    )
    document.documentElement.style.setProperty(
      '--ep-btn-color-disabled-top',
      adjustHexByDeltas(themeObj.epThemeColor || ' #8d6fe6', 1, 2, 0),
    )
    document.documentElement.style.setProperty(
      '--ep-btn-color-disabled-bottom',
      adjustHexByDeltas(themeObj.epThemeColor || ' #8d6fe6', 1, 2, 0),
    )
  } catch {}
}

// 初始化用户鉴权信息
export const initAuthInfo = async () => {
  try {
    // normalize possible sync/async return from getParentLocalStorageItem
    const token = await getParentLocalStorageItem('ELADMIN-TOEKN')
    const userInfo = await getParentLocalStorageItem('user-info')
    // 清洗并解析 token，处理可能带引号或其他包装情况
    function cleanToken(raw: any) {
      if (!raw && raw !== 0) return ''
      let s = String(raw)
      s = s.trim()
      // 移除外层双引号或单引号
      if ((s.startsWith('\"') && s.endsWith('\"')) || (s.startsWith("'") && s.endsWith("'"))) {
        s = s.slice(1, -1)
      }
      // 常见格式：Bearer <token> 或 "Bearer <token>"
      if (s.toLowerCase().startsWith('bearer ')) {
        return s.split(' ')[1] || ''
      }
      // 可能直接就是 token
      return s
    }

    if (typeof token === 'string' && token) {
      const parsed = cleanToken(token)
      Storage.set('token', parsed || '')
    }
    if (userInfo) {
      Storage.set('userInfo', userInfo)
    }
  } catch (err) {
    console.warn('init parent token failed', err)
  }
}

// 监听localStorage变化
export const watchStorage = (key: string, callback: (value: any) => void) => {
  window.addEventListener('storage', (event) => {
    if (event.key === key) {
      callback(event.newValue)
    }
  })
}
