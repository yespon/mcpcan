<template>
  <div>
    <component v-if="UserManage" :is="UserManage" />
    <el-empty v-else description="UserManage 组件未安装或路径不存在" />
  </div>
</template>

<script setup lang="ts">
const UserManage = shallowRef<any>(null)
const modules = import.meta.glob(
  '@/components/mcpcan-tools/mcpcan-business/web/user-manage/user-with-role.vue',
)
const loader = Object.values(modules)[0]

if (loader) {
  UserManage.value = defineAsyncComponent({
    loader: loader as any,
    // 如果组件不存在/构建时未包含，会走到这里
    onError(_err, _retry, fail) {
      UserManage.value = null
      fail()
    },
  })
}
</script>
