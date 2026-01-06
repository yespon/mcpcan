# Composables

这个目录包含可复用的 composable 函数。

## useLayout

全局 layout 状态 composable，提供响应式的 `layout` 变量，值为 `route.meta.hideLayout`。

### 使用方法

在任何组件中直接使用：

```vue
<script setup lang="ts">
// 自动导入，无需手动 import
const layout = useLayout()

// 在模板中使用
// <div v-if="layout">隐藏布局时显示的内容</div>
</script>

<template>
  <div v-if="layout">
    <!-- 当 hideLayout 为 true 时显示 -->
  </div>
</template>
```

### 说明

- `layout` 是一个响应式的 `ComputedRef<boolean>`
- 当路由变化时，`layout` 会自动更新
- 值为 `route.meta.hideLayout`，表示是否隐藏布局
- 由于配置了自动导入，无需手动 `import`

### 示例

```vue
<script setup lang="ts">
const layout = useLayout()

// 可以直接使用 .value 访问值
console.log(layout.value) // true 或 false

// 在模板中会自动解包，无需 .value
</script>

<template>
  <div>
    <el-link v-if="layout" @click="handleBack">返回</el-link>
    <div v-else>正常布局内容</div>
  </div>
</template>
```
