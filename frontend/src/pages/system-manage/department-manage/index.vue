<template>
  <div>
    <component v-if="DepartmentManage" :is="DepartmentManage" />
    <el-empty v-else description="DepartmentManage 组件未安装或路径不存在" />
  </div>
</template>

<script setup lang="ts">
const DepartmentManage = shallowRef<any>(null)
DepartmentManage.value = defineAsyncComponent({
  loader: () => import('@/components/mcpcan-tools/mcpcan-business/web/department-manage/index.vue'),
  // 如果组件不存在/构建时未包含，会走到这里
  onError(_err, _retry, fail) {
    DepartmentManage.value = null
    fail()
  },
})
</script>
