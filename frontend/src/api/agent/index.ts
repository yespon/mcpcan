import request from '@/utils/request'
import baseConfig from '@/config/base_config.ts'

export const AgentAPI = {
  // agent list
  list(params: TableData | null) {
    return request<any, any>({
      url: `${baseConfig.baseUrlVersion}/market/intelligent_access/list`,
      method: 'GET',
      params,
    })
  },
  // create agent
  create(data: CreateAgentRequest) {
    return request<any, any>({
      url: `${baseConfig.baseUrlVersion}/market/intelligent_access`,
      method: 'POST',
      data,
    })
  },
}

export interface CreateAgentRequest {
  accessName: string
  accessType: string
  dbHost: string
  dbPort: number
  dbUser: string
  dbPassword: string
  dbName: string
}
export interface TableData {
  page: number
  pageSize: number
  [key: string]: any
}
