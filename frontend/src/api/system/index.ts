import request from '@/utils/request'

export const RoleAPI = {
  list(data: TableData) {
    return request<TableData, List<any>>({
      url: `/authz/roles/list-roles`,
      method: 'GET',
      data,
    })
  },
  allList() {
    return request<any, List<any>>({
      url: `/authz/roles/list-roles`,
      method: 'GET',
      params: {
        page: 1,
        pageSize: 9999,
      },
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
  // 给角色授权菜单
  saveRoleMenus(data: any) {
    return request<any, any>({
      url: `/authz/roles/save-menus`,
      method: 'POST',
      data,
    })
  },
  // 获取所有菜单列表
  getAllMenus() {
    return request<any, any>({
      url: `/authz/menus/tree`,
      method: 'GET',
    })
  },
  // 获取当前角色的菜单权限
  getRoleMenus(data: any) {
    return request<any, any>({
      url: `/authz/roles/find-menus`,
      method: 'POST',
      data,
    })
  },
}
export const DeptAPI = {
  list(data: TableData) {
    return request<TableData, List<any>>({
      url: `/authz/depts/list-depts`,
      method: 'GET',
      params: data,
    })
  },
  create(data: any) {
    return request<any, any>({
      url: `/authz/depts`,
      method: 'POST',
      data,
    })
  },
  edit(data: any) {
    return request<any, any>({
      url: `/authz/depts`,
      method: 'PUT',
      data,
    })
  },
  delete(id: string) {
    return request<any, any>({
      url: `/authz/depts/${id}`,
      method: 'DELETE',
    })
  },
  deptTree() {
    return request<any, any>({
      url: `/authz/depts/tree`,
      method: 'GET',
    })
  },
}

export const UserAPI = {
  list(data: TableData) {
    return request<TableData, List<any>>({
      url: `/authz/users/list-users`,
      method: 'GET',
      params: data,
    })
  },
  create(data: any) {
    return request<any, any>({
      url: `/authz/users`,
      method: 'POST',
      data,
    })
  },
  edit(data: any) {
    return request<any, any>({
      url: `/authz/users/${data.id}`,
      method: 'PUT',
      data,
    })
  },
  delete(id: string) {
    return request<any, any>({
      url: `/authz/users/${id}`,
      method: 'DELETE',
    })
  },
  // 给用户添加角色
  addUserRoles(data: any) {
    return request<any, any>({
      url: `/authz/users/add-role`,
      method: 'POST',
      data,
    })
  },
  // 移除用户角色
  removeUserRoles(data: any) {
    return request<any, any>({
      url: `/authz/users/remove-role`,
      method: 'POST',
      data,
    })
  },
}
// 列表请求
export interface TableData {
  /** 页码 */
  page?: string | number
  /** 每页显示数量 */
  pageSize?: string | number
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
