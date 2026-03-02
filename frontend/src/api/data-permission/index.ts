import request from '@/utils/request'

export interface GetDataPermissionParams {
  dataType: string
  dataId: string
}

export interface DataPermissionResult {
  dataType: string
  dataId: string
  isAllPersonnel: boolean
  deptIds: number[]
  roleIds: number[]
  userIds: number[]
  blacklistUserIds: number[]
}

export interface SaveDataPermissionParams {
  dataType: string
  dataId: string
  isAllPersonnel: boolean
  deptIds: number[]
  roleIds: number[]
  userIds: number[]
  blacklistUserIds: number[]
}

export const DataPermissionAPI = {
  /**
   * 获取数据权限
   */
  get(params: GetDataPermissionParams) {
    return request<GetDataPermissionParams, DataPermissionResult>({
      url: `/market/data_permission`,
      method: 'GET',
      params,
    })
  },

  /**
   * 保存数据权限
   */
  save(data: SaveDataPermissionParams) {
    return request<SaveDataPermissionParams, any>({
      url: `/market/data_permission`,
      method: 'POST',
      data,
    })
  },
}
