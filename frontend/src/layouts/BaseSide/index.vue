<template>
  <div>
    <el-menu
      router
      :default-active="currentRoutePath"
      class="el-menu-vertical-demo overflow-y-auto h-full relative"
      :collapse="isCollapse"
    >
      <el-menu-item>
        <Logo v-if="isCollapse" width="20" height="20" class="is-collapse" />
        <template #title>
          <div class="center" @click="handleToHome">
            <Logo v-if="!isCollapse" />
            <span class="ml-2"> MCP CAN </span>
          </div>
        </template>
      </el-menu-item>
      <el-menu-item disabled v-auth="'mcpcan_instance:create'">
        <AddDropdown v-if="isCollapse">
          <el-icon class="cursor-pointer"><i class="icon iconfont MCP-xinjianziyuan"></i></el-icon>
        </AddDropdown>
        <template #title>
          <AddDropdown
            v-if="!isCollapse"
            :popperOptions="{
              modifiers: [
                {
                  name: 'offset',
                  options: {
                    offset: [100, 0], // 向右偏移100px，避免紧贴按钮
                  },
                },
              ],
            }"
          >
            <GlareHover
              width="100%"
              height="auto"
              background="transparent"
              border-color="#222222"
              border-radius="0px"
              glare-color="#ffffff"
              :glare-opacity="0.3"
              :glare-size="300"
              :transition-duration="800"
              :play-once="false"
            >
              <el-button type="primary" class="w-full base-btn text-left">
                <el-icon><i class="icon iconfont MCP-xinjianziyuan"></i></el-icon>
                {{ t('sideMenu.add') }}
              </el-button>
            </GlareHover>
          </AddDropdown>
          <span v-else>{{ t('sideMenu.add') }}</span>
        </template>
      </el-menu-item>
      <template v-for="menu in currentMenuAuths">
        <el-sub-menu
          v-if="menu.children && menu.children.length > 0"
          :key="menu.path"
          :index="menu.path"
        >
          <template #title>
            <div class="customize-menu">
              <el-icon>
                <i class="icon iconfont" :class="menuMap[menu.permission].icon"></i>
              </el-icon>
              <template v-if="!isCollapse">
                {{ menuMap[menu.permission]?.title || menu.title }}
              </template>
            </div>
          </template>
          <el-menu-item v-for="subMenu in menu.children" :key="subMenu.path" :index="subMenu.path">
            <div v-if="isCollapse" class="is-collapse">
              <el-icon>
                <i class="icon iconfont" :class="menuMap[subMenu.permission].icon"></i>
              </el-icon>
            </div>
            <template #title>
              <div :class="!isCollapse ? 'customize-menu' : ''">
                <el-icon v-if="!isCollapse">
                  <i class="icon iconfont" :class="menuMap[subMenu.permission].icon"></i>
                </el-icon>
                {{ menuMap[subMenu.permission].title }}
              </div>
            </template>
          </el-menu-item>
        </el-sub-menu>
        <el-menu-item v-else :index="menu.path">
          <div v-if="isCollapse" class="is-collapse">
            <el-icon>
              <i class="icon iconfont" :class="menuMap[menu.permission].icon"></i>
            </el-icon>
          </div>
          <template #title>
            <div class="customize-menu">
              <el-icon v-if="!isCollapse"
                ><i class="icon iconfont" :class="menuMap[menu.permission]?.icon"></i
              ></el-icon>
              {{ menuMap[menu.permission]?.title }}
            </div>
          </template>
        </el-menu-item>
      </template>
      <el-menu-item v-if="false" v-for="menu in menuList" :key="menu.index" :index="menu.index">
        <div v-if="isCollapse" class="is-collapse">
          <el-icon>
            <i class="icon iconfont" :class="menu.icon"></i>
          </el-icon>
        </div>
        <template #title>
          <div class="customize-menu">
            <el-icon v-if="!isCollapse"><i class="icon iconfont" :class="menu.icon"></i></el-icon>
            {{ menu.title }}
          </div>
        </template>
      </el-menu-item>
      <div class="setting-card text-right" @click="handleChangeCollapse">
        <el-icon class="cursor-pointer link-hover" v-if="isCollapse"><Expand /></el-icon>
        <el-icon class="cursor-pointer link-hover" v-else><Fold /></el-icon>
      </div>
    </el-menu>
  </div>
</template>

<script lang="ts" setup>
import { Expand, Fold } from '@element-plus/icons-vue'
import { useSystemStoreHook, useUserStore } from '@/stores'
import Logo from './logo.vue'
import AddDropdown from './add-dropdown.vue'
import GlareHover from '@/components/Animation/GlareHover.vue'
import { useRouterHooks } from '@/utils/url'

const route = useRoute()
const currentRoutePath = computed(() => route.path)
const { jumpToPage } = useRouterHooks()
const { t } = useI18n()
const { isCollapse } = toRefs(useSystemStoreHook())
const { currentMenuAuths } = storeToRefs(useUserStore())
const handleChangeCollapse = () => {
  useSystemStoreHook().changeCollapse(isCollapse.value)
}
const menuMap: Record<string, { title: string; icon: string }> = {
  mcpcan_home: { title: t('sideMenu.home'), icon: 'MCP-shouye1' },
  mcpcan_instance: { title: t('sideMenu.instanceManage'), icon: 'MCP-MCPshili' },
  mcpcan_market_manage: { title: t('sideMenu.marketManage'), icon: 'MCP-shichangcaidan' },
  mcpcan_template: { title: t('sideMenu.templateManage'), icon: 'MCP-MCPmoban' },
  mcpcan_working_environment: { title: t('sideMenu.runEnviroment'), icon: 'MCP-huanjingguanli' },
  codeList: { title: t('sideMenu.codeList'), icon: 'MCP-daimaguanli' },
  apiDocsList: { title: t('sideMenu.apiDocsList'), icon: 'MCP-wenjian' },
  mcpcan_resource_manage: { title: t('sideMenu.resourceManage'), icon: 'MCP-wenjian' },
  mcpcan_agent_manage: { title: t('sideMenu.agentManage'), icon: 'MCP-zhinengti' },
  'mcpcan_rbac_manage:role': { title: t('sideMenu.roleManage'), icon: 'MCP-jiaose' },
  'mcpcan_rbac_manage:user': { title: t('sideMenu.userManage'), icon: 'MCP-yonghu' },
  'mcpcan_rbac_manage:dept': { title: t('sideMenu.departmentManage'), icon: 'MCP-zuzhibumen' },
  mcpcan_rbac_manage: { title: t('sideMenu.rbacManage'), icon: 'MCP-qita' },
  mcpcan_model_manage: { title: '模型管理', icon: 'MCP-damoxing' },
  mcpcan_ai_chat: { title: 'AI 聊天', icon: 'MCP-AIduihua' },
}

const menuList = shallowRef([
  { title: t('sideMenu.home'), icon: 'MCP-shouye1', index: '/home' },
  { title: t('sideMenu.instanceManage'), icon: 'MCP-MCPshili', index: '/instance-manage' },
  { title: t('sideMenu.marketManage'), icon: 'MCP-shichangcaidan', index: '/market-manage' },
  { title: t('sideMenu.templateManage'), icon: 'MCP-MCPmoban', index: '/template-manage' },
  { title: t('sideMenu.runEnviroment'), icon: 'MCP-huanjingguanli', index: '/working-environment' },
  // { title: t('sideMenu.codeList'), icon: 'MCP-daimaguanli', index: '/code-list' },
  // { title: t('sideMenu.apiDocsList'), icon: 'MCP-wenjian', index: '/api-docs-list' },
  { title: t('sideMenu.resourceManage'), icon: 'MCP-wenjian', index: '/resource-manage' },
  { title: t('sideMenu.agentManage'), icon: 'MCP-zhinengti', index: '/agent-manage' },
  { title: t('sideMenu.roleManage'), icon: 'MCP-jiaose', index: '/role-manage' },
  { title: t('sideMenu.userManage'), icon: 'MCP-yonghu', index: '/user-manage' },
  { title: t('sideMenu.departmentManage'), icon: 'MCP-zuzhibumen', index: '/department-manage' },
  { title: '模型管理', icon: 'MCP-damoxing', index: '/model-manage' },
  { title: 'AI 聊天', icon: 'MCP-AIduihua', index: '/ai-chat' },
])

const handleToHome = () => {
  jumpToPage({
    url: '/home',
    data: {},
  })
}
onMounted(() => {
  console.log('currentMenuAuths', currentMenuAuths.value)
})
</script>

<style lang="scss" scoped>
.el-menu-vertical-demo {
  height: 100%;
}
.el-menu-vertical-demo:not(.el-menu--collapse) {
  width: 180px;
  min-height: 400px;
}
.el-menu-item.is-disabled {
  opacity: 1 !important;
  cursor: default !important;
}

.el-menu-item {
  .customize-menu {
    height: 32px;
    flex: 1;
    border: 1px solid transparent;
    border-radius: 8px;
    display: flex;
    align-items: center;
    box-sizing: border-box;
  }
  &:hover {
    background-color: transparent;
    .customize-menu {
      border: 1px solid #a083f7;
      color: var(--ep-color);
      background-color: var(--ep-bg-purple-color);
    }
    &:has(.is-collapse) {
      background-color: var(--ep-bg-purple-color);
    }
  }
  &.is-active {
    color: var(--ep-color);
    .customize-menu {
      background-color: var(--ep-bg-purple-color);
      border: 1px solid rgba(255, 255, 255, 0.5);
    }
    &:has(.is-collapse) {
      background-color: var(--ep-bg-purple-color);
    }
  }
}

.setting-card {
  cursor: pointer;
  position: absolute;
  font-size: 20px;
  bottom: 0;
  width: 100%;
  height: 50px;
  display: flex;
  justify-content: end;
  align-items: center;
  margin-left: -20px;
}
:deep(.el-menu-item *) {
  vertical-align: top;
}
:deep(.el-sub-menu__title:hover) {
  background-color: var(--el-color-primary-hover);
}
</style>
