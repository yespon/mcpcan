import request from '@/utils/request'

export const MarketAPI = {
  // agent list
  list(params: TableData | null) {
    return request<any, any>({
      url: `/market/platform/list`,
      method: 'GET',
      params,
    })
  },
}
export interface TableData {
  page: number
  pageSize: number
  [key: string]: any
}
