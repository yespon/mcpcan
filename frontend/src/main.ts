// import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import App from './App.vue'
import router from './router'
import i18n from './lang'
import 'uno.css'
import 'element-plus/dist/index.css'
import '~/styles/index.scss'
// If you want to use ElMessage, import it.
import 'element-plus/theme-chalk/dark/css-vars.css'
import '~/assets/icon/iconfont.css'
import countTo from './directives/numberScrollDirective.ts'
import auth from './directives/auth.ts'
import { store } from './stores'
import { initAuthInfo } from '@/utils/system'
import McpButton from '@/components/mcp-button/index.vue'
import McpImage from '@/components/mcp-image/index.vue'

const app = createApp(App)
// 全局指令注册
app.directive('countTo', countTo)
app.directive('auth', auth)
// 全局组件注册
app.component('McpButton', McpButton)
app.component('McpImage', McpImage)
// app.use(createPinia())
app.use(i18n)
app.use(router)
app.use(ElementPlus)
app.use(store)

async function bootstrap() {
  await initAuthInfo()
  app.mount('#app')
}

bootstrap()
