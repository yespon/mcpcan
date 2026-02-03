import request from '@/utils/request'
import baseConfig from '@/config/base_config.ts'
import { create } from 'lodash-es'

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
  allList() {
    // return request<any, List<any>>({
    //   url: `/authz/roles/list-roles`,
    //   method: 'GET',
    // })
    return new Promise<any>((resolve) => {
      resolve({
        data: [
          {
            id: 1,
            name: `测试角色`,
            level: 1,
            dataScope: '全部数据权限',
            description: '拥有系统所有权限',
            createTime: '2025-02-02 12:00:00',
          },
        ],
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
  removeUser(data: any) {
    return request<any, any>({
      url: `/authz/roles/users/remove`,
      method: 'POST',
      data,
    })
  },
  // 给角色授权菜单
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
export const DeptAPI = {
  list(data: TableData) {
    // return request<TableData, List<any>>({
    //   url: `/authz/depts/page-depts`,
    //   method: 'POST',
    //   data,
    // })
    return new Promise<any>((resolve) => {
      resolve({
        list: [
          {
            id: 1,
            name: `测试部门`,
            deptSort: 1,
            createTime: '2025-02-02 12:00:00',
          },
        ],
        page: 1,
        pageSize: 10,
        total: 10,
      })
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
    // return request<any, any>({
    //   url: `/authz/depts/tree`,
    //   method: 'GET',
    // })
    return new Promise<any>((resolve) => {
      resolve({
        data: [
          {
            id: 1,
            name: `测试部门`,
            children: [
              {
                id: 2,
                name: `子部门`,
              },
            ],
          },
        ],
      })
    })
  },
}

export const UserAPI = {
  list(data: TableData) {
    // return request<TableData, List<any>>({
    //   url: `/authz/users/list-users`,
    //   method: 'POST',
    //   data,
    // })
    return new Promise<any>((resolve) => {
      resolve({
        list: [
          {
            id: '123',
            username: 'string',
            fullName: 'string',
            password: 'string',
            email: 'string',
            phone: 'string',
            avatar: 'string',
            status: 'UserStatusUnspecified',
            source: 'UserSourceUnspecified',
            deptId: 'string',
            deptName: 'string',
            roleIds: ['string'],
            roleNames: ['string'],
            createdAt: 'string',
            updatedAt: 'string',
            createdBy: 'string',
            updatedBy: 'string',
          },
        ],
        page: 1,
        pageSize: 10,
        total: 100,
      })
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
}
// 列表请求
export interface TableData {
  /** 页码 */
  page: string | number
  /** 每页显示数量 */
  pageSize: string | number
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
