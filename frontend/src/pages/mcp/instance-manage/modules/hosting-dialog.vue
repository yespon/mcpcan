<template>
  <div>
    <el-dialog
      v-model="dialogInfo.visible"
      :title="'托管模式'"
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
      <template #footer>
        <div class="text-left">
          <el-button>保存并运行</el-button>
          <el-button>另存为模板</el-button>
          <el-button @click="handleClose">退出</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { useInstanceFormHooks } from '../hooks/form-instance.ts'

const { pageInfo } = useInstanceFormHooks()
const dialogInfo = ref({
  visible: false,
})

const handleClose = () => {
  dialogInfo.value.visible = false
}

const init = () => {
  dialogInfo.value.visible = true
}

defineExpose({
  init,
})
</script>
