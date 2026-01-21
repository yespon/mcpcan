<template>
  <div>
    <el-dialog
      v-model="dialogInfo.visible"
      :title="'代理模式'"
      :show-close="false"
      :close-on-click-modal="false"
      class="access-type-dialog"
      width="620px"
      top="10vh"
    >
      <el-scrollbar height="70vh"> </el-scrollbar>
      <template #footer>
        <div class="text-center">
          <el-button @click="handleConfirm">保存并运行</el-button>
          <el-button @click="handleSaveAsTemplate">另存为模板</el-button>
          <el-button @click="handleClose">退出</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
const dialogInfo = ref({
  visible: false,
})

/**
 * Handle Close Dialog
 */
const handleClose = () => {
  dialogInfo.value.visible = false
}
</script>

<style lang="scss" scoped>
.tip {
  padding: 10px;
  border-radius: 4px;
  font-size: 12px;
  &.tip-warning {
    background-color: #fff1f0;
    border-left: 5px solid var(--el-color-danger);
  }
  &.tip-primary {
    background-color: #409eff1a;
    border-left: 5px solid var(--el-color-primary);
  }
}
</style>
