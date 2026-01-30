<template>
  <div>
    <component v-if="RoleManage" :is="RoleManage" />
    <el-empty v-else description="RoleManage 组件未安装或路径不存在" />
  </div>
</template>

<script setup lang="ts">
const RoleManage = shallowRef<any>(null)
RoleManage.value = defineAsyncComponent({
  loader: () => import('@/components/mcpcan-tools/mcpcan-business/web/role-manage/index.vue'),
  // 如果组件不存在/构建时未包含，会走到这里
  onError(_err, _retry, fail) {
    RoleManage.value = null
    fail()
  },
})
</script>
