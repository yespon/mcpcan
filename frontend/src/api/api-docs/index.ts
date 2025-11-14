import request from '@/utils/request'
import baseConfig from '@/config/base_config.ts'

export const DocsAPI = {
  // docs list
  list(params: TableData | null) {
    return request<any, List<any>>({
      url: `${baseConfig.baseUrlVersion}/market/openapi/files`,
      method: 'GET',
      params,
    })
  },
  // download docs
  download(code: any) {
    return request<any, any>({
      url: `${baseConfig.baseUrlVersion}/market/openapi/download/${code.id}`,
      method: 'GET',
      responseType: 'blob',
    })
  },
  // delete Docs
  delete(id: string) {
    return request<string>({
      url: `${baseConfig.baseUrlVersion}/market/openapi/files/${id}`,
      method: 'DELETE',
    })
  },
  // get Docs COntent
  fileContent(params: any) {
    return request<any, any>({
      url: `${baseConfig.baseUrlVersion}/market/openapi/content`,
      method: 'GET',
      params,
    })
  },
  // edit docs
  editDocs(params: any) {
    return request<any, any>({
      url: `${baseConfig.baseUrlVersion}/market/openapi/edit`,
      method: 'POST',
      params,
    })
  },
}
// list params
export interface TableData {
  page: number
  pageSize: number
  [key: string]: any
}
export interface List<T = any> {
  list: T[]
  page: number
  pageSize: number
  total: number
}
