import request from '@/utils/request'
import baseConfig from '@/config/base_config.ts'
import { type Code } from '@/types/index'

export const CodeAPI = {
  // code package list
  list(params: TableData | null) {
    return request<any, List<Code>>({
      url: `${baseConfig.baseUrlVersion}/market/code/packages`,
      method: 'GET',
      params,
    })
  },
  // download code package
  download(code: any) {
    return request<any, any>({
      url: `${baseConfig.baseUrlVersion}/market/code/download/${code.id}`,
      method: 'GET',
      responseType: 'blob',
    })
  },
  // delete package
  delete(id: string) {
    return request<string>({
      url: `${baseConfig.baseUrlVersion}/market/code/packages/${id}`,
      method: 'DELETE',
    })
  },
  // get files tree list
  fileTree(params: any) {
    return request<any, FileTree>({
      url: `${baseConfig.baseUrlVersion}/market/code/tree`,
      method: 'GET',
      params,
    })
  },
  // get file content
  fileContent(params: any) {
    return request<any, any>({
      url: `${baseConfig.baseUrlVersion}/market/code/get`,
      method: 'GET',
      params,
    })
  },
  // get file content as blob (for binary files like images)
  fileContentBlob(params: any) {
    return request<any, any>({
      url: `${baseConfig.baseUrlVersion}/market/code/get`,
      method: 'GET',
      params,
      responseType: 'blob',
    })
  },
}

// list params
export interface TableData {
  page: number
  pageSize: number
  [key: string]: any
}
// reponse list
export interface List<T = any> {
  list: T[]
  page: number
  pageSize: number
  total: number
}

export interface FileTree {
  fileStructure: FileTree[]
  isDir: boolean
  name: string
  path: string
}
