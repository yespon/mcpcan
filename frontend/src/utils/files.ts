/**
 * 文件数据下载至本地
 * @param fileInfo
 * @param suffix
 */
export const downloadData = async (
  fileInfo: {
    fileName: string
    data: string
  },
  suffix: string = 'txt',
) => {
  return new Promise<void>((resolve) => {
    const blob = new Blob([fileInfo.data], { type: 'text/plain;charset=utf-8' })
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `${fileInfo.fileName}.${suffix}`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)
    resolve()
  })
}
