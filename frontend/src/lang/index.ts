import { createI18n } from 'vue-i18n'
import { Storage } from '@/utils/storage'

// global
import enGlobal from './global/en.json'
import zhCnGlobal from './global/zh-cn.json'

// home
import enHome from './home/en.json'
import zhCnHome from './home/zh-cn.json'
// mcp instance and template
import enMcpManage from './mcp/en.json'
import zhMcpManage from './mcp/zh-cn.json'
// env manage
import enEnvManage from './env/en.json'
import zhEnvManage from './env/zh-cn.json'
// code page
import enCodePackage from './code-package/en.json'
import zhCodePackage from './code-package/zh-cn.json'
// api docs
import enApiDocs from './api-docs/en.json'
import zhApiDocs from './api-docs/zh-cn.json'
// agent
import enAgent from './agent/en.json'
import zhCnAgent from './agent/zh-cn.json'
// market
import enMarket from './market/en.json'
import zhCnMarket from './market/zh-cn.json'
// system
import enSystem from './system/en.json'
import zhCnSystem from './system/zh-cn.json'
// model
import enModel from './model/en.json'
import zhCnModel from './model/zh-cn.json'
// ai chat
import enAiChat from './ai-chat/en.json'
import zhCnAiChat from './ai-chat/zh-cn.json'

const messages = {
  'zh-cn': {
    ...zhCnGlobal,
    ...zhCnHome,
    ...zhCnAgent,
    ...zhMcpManage,
    ...zhEnvManage,
    ...zhCodePackage,
    ...zhApiDocs,
    ...zhCnMarket,
    ...zhCnSystem,
    ...zhCnModel,
    ...zhCnAiChat,
  },
  en: {
    ...enGlobal,
    ...enHome,
    ...enAgent,
    ...enMcpManage,
    ...enEnvManage,
    ...enCodePackage,
    ...enApiDocs,
    ...enMarket,
    ...enSystem,
    ...enModel,
    ...enAiChat,
  },
}

export const i18n = createI18n({
  legacy: false,
  locale: Storage.get('language'),
  messages,
  globalInjection: true,
})

export default i18n
