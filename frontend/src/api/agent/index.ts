import request from '@/utils/request'
import baseConfig from '@/config/base_config.ts'
import { update } from 'lodash-es'

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
  // connection test
  connectionTest(data: CreateAgentRequest) {
    return request<any, any>({
      url: `${baseConfig.baseUrlVersion}/market/intelligent_access/test-connection`,
      method: 'POST',
      data,
    })
  },
  // delete agent platform
  delete(accessID: string) {
    return request<any, any>({
      url: `${baseConfig.baseUrlVersion}/market/intelligent_access/delete`,
      method: 'DELETE',
      data: { accessID },
    })
  },
  // update agent platform
  update(data: CreateAgentRequest) {
    return request<any, any>({
      url: `${baseConfig.baseUrlVersion}/market/intelligent_access/edit`,
      method: 'put',
      data,
    })
  },
  // get namespaces by platform
  getNamespaces(data: any) {
    return request<any, any>({
      url: `${baseConfig.baseUrlVersion}/market/intelligent_access/list-dify-user-space`,
      method: 'POST',
      data,
    })
  },
  // create a sync task
  createSyncTask(data: any) {
    return request<any, any>({
      url: `${baseConfig.baseUrlVersion}/market/mcp_to_intelligent_task`,
      method: 'POST',
      data,
    })
  },
  // get task list
  taskList(params: any) {
    return request<any, any>({
      url: `${baseConfig.baseUrlVersion}/market/mcp_to_intelligent_task/list`,
      method: 'GET',
      params,
    })
  },
  // get task detail
  taskDetail(id: string) {
    return request<any, any>({
      url: `${baseConfig.baseUrlVersion}/market/mcp_to_intelligent_task/${id}`,
      method: 'GET',
    })
  },
  // cancel task
  cancelTask(id: string) {
    return request<any, any>({
      url: `${baseConfig.baseUrlVersion}/market/mcp_to_intelligent_task/${id}/cancel`,
      method: 'PUT',
    })
  },
}

export interface CreateAgentRequest {
  accessID?: string
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
