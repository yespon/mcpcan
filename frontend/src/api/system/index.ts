import request from '@/utils/request'
import baseConfig from '@/config/base_config.ts'

export const RoleAPI = {
  list(data: TableData) {
    // return request<TableData, List<any>>({
    //   url: `/authz/roles/page-roles`,
    //   method: 'POST',
    //   data,
    // })
    return new Promise<any>((resolve) => {
      resolve({
        list: [
          {
            id: 1,
            name: `测试角色`,
            level: 1,
            dataScope: '全部数据权限',
            description: '拥有系统所有权限',
            createTime: '2025-02-02 12:00:00',
          },
        ],
        page: 1,
        pageSize: 10,
        total: 100,
      })
    })
  },
  create(form: any) {
    return request<any, any>({
      url: `/authz/roles`,
      method: 'POST',
      data: form,
    })
  },
  edit(form: any) {
    return request<any, any>({
      url: `/authz/roles`,
      method: 'PUT',
      data: form,
    })
  },
  delete(id: string) {
    return request<any, any>({
      url: `/authz/roles/${id}`,
      method: 'DELETE',
    })
  },
  saveRoleMenus(roleId: string, menuIds: string[]) {
    return request<any, any>({
      url: `/authz/roles/menus`,
      method: 'POST',
      data: {
        roleId,
        menuIds,
      },
    })
  },
}
// 列表请求
export interface TableData {
  /** 页码 */
  page: string
  /** 每页显示数量 */
  pageSize: string
  /** 允许传入其他任意类型的参数 */
  [key: string]: any
}
// 列表返回
export interface List<T = any> {
  list: T[]
  page: number
  pageSize: number
  total: number
}

export interface RoleInfo {
  createdAt?: string
  createdBy?: string
  dataScope?: string
  description?: string
  id?: string
  level?: string
  name?: string
  updatedAt?: string
  updatedBy?: string
  [property: string]: any
}
