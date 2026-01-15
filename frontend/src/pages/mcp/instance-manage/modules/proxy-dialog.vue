<template>
  <div>
    <el-dialog
      v-model="dialogInfo.visible"
      :title="'代理模式'"
      :show-close="false"
      :close-on-click-modal="false"
      class="access-type-dialog"
      width="800px"
      top="10vh"
    >
      <el-scrollbar height="70vh">
        <p v-for="item in 20" :key="item" class="scrollbar-demo-item">{{ item }}</p>
        <p v-for="item in 20" :key="item" class="scrollbar-demo-item">{{ item }}</p>
      </el-scrollbar>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
const dialogInfo = ref({
  visible: false,
})
const init = () => {
  dialogInfo.value.visible = true
}
defineExpose({
  init,
})
</script>
