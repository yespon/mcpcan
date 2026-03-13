import { store } from '@/stores'
import AuthAPI, { type LoginFormData, type EncryptionInfo } from '@/api/auth-api.ts'
import UserAPI, { type UserInfo } from '@/api/user/user-api'
import { RoleAPI } from '@/api/system/index.ts'
import { defineStore } from 'pinia'
import { Storage } from '@/utils/storage'
import { useStorage } from '@vueuse/core'
import { ElLoading, ElMessage } from 'element-plus'
import { KEYUTIL, KJUR, RSAKey, hextob64 } from 'jsrsasign'
import i18n from '@/lang'

export const useUserStore = defineStore('user', () => {
  const userInfo = useStorage('userInfo', {} as UserInfo)
  const encryptionInfo = ref<EncryptionInfo>({} as EncryptionInfo)
  const loginFormData = useStorage('loginFormData', {} as string)
  const allMenus = useStorage('allMenus', [] as any[])
  const allAuths = useStorage('allAuths', {} as any)
  // 按钮权限列表
  const currentBtnAuths = computed(() => {
    return allAuths.value?.menus
      ?.filter((item: any) => Number(item.type) === 2 && item.permission)
      .map((item: any) => item.permission)
  })
  // 菜单权限列表（tree）
  const currentMenuAuths = computed(() => {
    const permissions: string[] = allAuths.value?.permissions || []
    const permissionSet = new Set<string>(permissions)
    const filterMenuTree = (list: any[]): any[] => {
      if (!Array.isArray(list) || list.length === 0) return []
      return list
        .filter((item) => {
          return (
            (Number(item?.type) === 0 || Number(item?.type) === 1) &&
            permissionSet.has(item?.permission)
          )
        })
        .map((item) => {
          return {
            ...item,
            children: filterMenuTree(item.children),
          }
        })
    }
    return filterMenuTree(allMenus.value) || []
  })
  // 授权菜单列表(array list)
  const allAuthMenuList = computed(() => {
    const list: string[] = []
    const walk = (nodes: any[]) => {
      if (!Array.isArray(nodes) || nodes.length === 0) return
      for (const n of nodes) {
        // 仅收集菜单/目录(当前 currentMenuAuths 已过滤 type 0/1)
        const path = n?.path || n?.url
        if (path) list.push(String(path))
        if (n?.children?.length) walk(n.children)
      }
    }

    walk(currentMenuAuths.value as any[])
    return list
  })

  // PEM interchange ArrayBuffer
  function pemToArrayBuffer(pem: string) {
    // remove PEM empty info
    const stripped = pem
      .replace(/-----BEGIN PUBLIC KEY-----/, '')
      .replace(/-----END PUBLIC KEY-----/, '')
      .replace(/\s+/g, '')
    // Base64 interchange Binary string
    const binary = atob(stripped)
    // Uint8Array
    const buffer = new Uint8Array(binary.length)
    for (let i = 0; i < binary.length; i++) {
      buffer[i] = binary.charCodeAt(i)
    }
    return buffer
  }
  /**
   * RSA-OPEN
   * @param plaintext
   * @returns
   */
  const encryptPassword = async (data: string) => {
    try {
      let publicKeyPem
      // 解码publicKeyPem
      if (encryptionInfo.value.publicKey) {
        publicKeyPem = atob(encryptionInfo.value.publicKey?.replace(/-/g, '+').replace(/_/g, '/'))
      } else {
        ElMessage.error(i18n.global.t('request.KeyPem'))
        throw new Error(i18n.global.t('request.KeyPem'))
      }

      // 1. 将PEM公钥转换为ArrayBuffer
      const publicKeyBuffer = pemToArrayBuffer(publicKeyPem)
      // https 协议下加密原生
      if (window.isSecureContext && window.crypto?.subtle) {
        // 2. 导入公钥（指定RSA-OAEP算法）
        const publicKey = await window.crypto.subtle.importKey(
          'spki', // 公钥使用SPKI格式
          publicKeyBuffer,
          {
            name: 'RSA-OAEP',
            hash: 'SHA-256', // 与Node版本保持一致的哈希算法
          },
          false, // 公钥不可提取
          ['encrypt'], // 仅用于加密
        )

        // 3. 将明文转换为ArrayBuffer
        const encoder = new TextEncoder()
        const dataBuffer = encoder.encode(data)
        // 4. 执行RSA-OAEP加密
        const encryptedBuffer = await window.crypto.subtle.encrypt(
          { name: 'RSA-OAEP' },
          publicKey,
          dataBuffer,
        )
        // 5. 将加密结果转换为Base64字符串
        return btoa(String.fromCharCode(...new Uint8Array(encryptedBuffer)))
      }

      // http协议环境下加密
      // 转换公钥为jsrsasign可用的格式
      const publicKey = KEYUTIL.getKey(publicKeyPem) as RSAKey
      const encryptedHex = KJUR.crypto.Cipher.encrypt(
        data,
        publicKey,
        'RSAOAEP256', // 指定SHA-256哈希
      )
      // 转换为Base64
      return hextob64(encryptedHex)
    } catch (error: any) {
      console.error('加密失败:', error)
      throw new Error(`加密错误: ${error?.message}`)
    }
  }
  /**
   * Login
   *
   * @param {LoginFormData}
   * @returns
   */
  async function login(LoginFormData: LoginFormData) {
    loginFormData.value = btoa(JSON.stringify(LoginFormData))
    const encryptedPassword = await encryptPassword(LoginFormData.password || '')
    return new Promise((resolve, reject) => {
      AuthAPI.login({
        encryptedPassword,
        keyId: encryptionInfo.value.keyId,
        timestamp: String(new Date().getTime()),
        username: LoginFormData.username,
      })
        .then((data) => {
          Storage.set('token', data.token)
          Storage.set('refreshToken', data.refreshToken)
          userInfo.value = data.userInfo
          resolve(data)
        })
        .catch((err) => {
          // 登录失败重新获取秘钥
          handleGetEncryptionKey()
          reject(err)
        })
    })
  }

  /**
   * Handle get Encryption info
   * @returns
   */
  function handleGetEncryptionKey() {
    return new Promise<EncryptionInfo>((resolve, reject) => {
      AuthAPI.getEncryptionKey()
        .then((data) => {
          if (!data) {
            reject('无法获取秘钥信息.')
            return
          }
          Object.assign(encryptionInfo.value, { ...data })
          resolve(data)
        })
        .catch((error) => reject(error))
    })
  }

  /**
   * Handle get user info data
   *
   * @returns {UserInfo}
   */
  function getUserInfo() {
    return new Promise<UserInfo>((resolve, reject) => {
      UserAPI.getInfo()
        .then((data) => {
          if (!data) {
            reject('Verification failed, please Login again.')
            return
          }
          Object.assign(userInfo.value, { ...data })
          resolve(data)
        })
        .catch((error) => {
          reject(error)
        })
    })
  }
  /**
   * Handle get user info data
   *
   * @returns {UserInfo}
   */
  function validateInfo() {
    return new Promise<UserInfo>((resolve, reject) => {
      AuthAPI.validate()
        .then((data) => {
          userInfo.value = data.userInfo
          if (!data) {
            reject('Verification failed, please Login again.')
            return
          }
          Object.assign(userInfo.value, { ...data })
          resolve(data)
        })
        .catch((error) => {
          reject(error)
        })
    })
  }

  /**
   * Logout
   */
  function logout() {
    return new Promise<void>((resolve, reject) => {
      const loadingInstance = ElLoading.service({ fullscreen: true, text: '退出登录中' })
      AuthAPI.logout({
        token: Storage.get('token'),
        userId: userInfo.value.userId,
      })
        .then(() => {
          resetUserState()
          window.location.reload()
          resolve()
        })
        .catch((error) => {
          reject(error)
        })
        .finally(() => {
          loadingInstance.close()
        })
    })
  }

  /**
   * reset user status
   *
   */
  function resetUserState() {
    // remove token info
    Storage.remove('token')
    Storage.remove('refresh_token')
    // reset user info
    userInfo.value = {} as UserInfo
  }

  /**
   * refresh token
   */
  function refreshToken() {
    const refreshToken = Storage.get('refreshToken')

    if (!refreshToken) {
      return Promise.reject(new Error('没有有效的刷新令牌'))
    }
    return new Promise<void>((resolve, reject) => {
      AuthAPI.refreshToken(refreshToken as string)
        .then((data) => {
          Storage.set('token', data.token)
          Storage.set('refreshToken', data.refreshToken)
          resolve()
        })
        .catch((error) => {
          reject(error)
        })
    })
  }

  /**
   * get all menu list
   */
  function getMenus() {
    return new Promise<any[]>((resolve, reject) => {
      RoleAPI.getAllMenus()
        .then((data) => {
          allMenus.value = data
          resolve(data)
        })
        .catch((error) => {
          reject(error)
        })
    })
  }

  /**
   * get current user role auth
   */
  function getRoleAuth() {
    return new Promise<any[]>((resolve, reject) => {
      // 不传参数获取当前用户的菜单权限，传参用户id获取指定用户的菜单权限
      // RoleAPI.getRoleMenus({ roleIds: [...(userInfo.value.roleIds || [])] })
      RoleAPI.getRoleMenus()
        .then((data) => {
          allAuths.value = data
          resolve(data)
        })
        .catch((error) => {
          reject(error)
        })
    })
  }

  /**
   * Handle menu authorization
   * @returns menu tree with role auth
   */
  function handleMenuAuth() {
    const appConfig = (window as any).__APP_CONFIG__ || {}
    if (appConfig.CodeMode === 'OpenCode') {
      const openCodeMenus = [
        { path: '/home', permission: 'mcpcan_home', type: 1, title: 'home' },
        {
          path: '/instance-manage',
          permission: 'mcpcan_instance',
          type: 1,
          title: 'instanceManage',
        },
        {
          path: '/market-manage',
          permission: 'mcpcan_market_manage',
          type: 1,
          title: 'marketManage',
        },
        {
          path: '/template-manage',
          permission: 'mcpcan_template',
          type: 1,
          title: 'templateManage',
        },
        {
          path: '/working-environment',
          permission: 'mcpcan_working_environment',
          type: 1,
          title: 'runEnviroment',
        },
        {
          path: '/resource-manage',
          permission: 'mcpcan_resource_manage',
          type: 1,
          title: 'resourceManage',
        },
        { path: '/agent-manage', permission: 'mcpcan_agent_manage', type: 1, title: 'agentManage' },
        { path: '/model-manage', permission: 'mcpcan_model_manage', type: 1, title: 'modelManage' },
        { path: '/ai-chat', permission: 'mcpcan_ai_chat', type: 1, title: 'AI Chat' },
      ]
      allMenus.value = openCodeMenus
      allAuths.value = {
        permissions: openCodeMenus.map((m) => m.permission),
        menus: openCodeMenus.map((m) => ({ ...m, permission: m.permission })),
      }
      return Promise.resolve({ menus: openCodeMenus, auths: allAuths.value })
    }

    // 企业版模式：如果当前缓存中只有开源版菜单（即没有 rbac 相关的菜单），强制刷新
    const hasRbacMenu = (allMenus.value || []).some((m: any) =>
      String(m.permission).includes('rbac_manage'),
    )

    if (hasRbacMenu && allMenus.value.length > 0 && allAuths.value.permissions) {
      return Promise.resolve({ menus: allMenus.value, auths: allAuths.value })
    }

    return Promise.all([getMenus(), getRoleAuth()]).then(([menus, auths]) => {
      return { menus, auths }
    })
  }

  return {
    userInfo,
    isLogin: () => !!Storage.get('token'),
    getUserInfo,
    validateInfo,
    handleMenuAuth,
    currentBtnAuths,
    currentMenuAuths,
    allAuthMenuList,
    login,
    logout,
    // resetAllState,
    resetUserState,
    refreshToken,
    handleGetEncryptionKey,
    loginFormData,
  }
})

/**
 * 在组件外部使用UserStore的钩子函数
 * @see https://pinia.vuejs.org/core-concepts/outside-component-usage.html
 */
export function useUserStoreHook() {
  return useUserStore(store)
}
