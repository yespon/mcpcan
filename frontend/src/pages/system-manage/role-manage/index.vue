<template>
  <div>
    <component v-if="RoleManage" :is="RoleManage" />
    <el-empty v-else description="RoleManage 组件未安装或路径不存在" />
  </div>
</template>

<script setup lang="ts">
const RoleManage = shallowRef<any>(null)
// 使用 import.meta.glob 避免构建时因文件不存在而报错
const modules = import.meta.glob(
  '@/components/mcpcan-tools/mcpcan-business/web/role-manage/index.vue',
)
const loader = Object.values(modules)[0]

if (loader) {
  RoleManage.value = defineAsyncComponent({
    loader: loader as any,
    // 如果组件加载失败
    onError(_err, _retry, fail) {
      RoleManage.value = null
      fail()
    },
  })
}
</script>
