<template>
  <div class="wh-full flex justify-center align-center">
    <img class="min-w-[23.4375rem] sm:w-150" src="@/assets/images/401.svg" alt="401" />
    <div class="w-75">
      <div class="oops mb-5 text-[2rem] font-bold">OOPS！</div>
      <div class="info text-gray mb-7 text-[0.8125rem]">
        您没有权限访问该页面。
        <el-link type="primary" target="_blank"> 联系管理员 </el-link>
      </div>
      <div class="headline mb-2.5 text-xl font-bold text-[#222]">403 Forbidden</div>
      <div class="info text-gray mb-7 text-[0.8125rem]">
        请确认您的账号已被授予相应权限，或切换账号后重试。您也可以点击下方按钮返回首页。
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
defineOptions({ name: 'Page403' })
</script>

<style lang="scss" scoped>
.wh-full {
  width: 100%;
  height: 100vh;
}
@mixin slide-up-animation($delay: 0s) {
  opacity: 0;
  animation-name: slideUp;
  animation-duration: 0.5s;
  animation-delay: $delay;
  animation-fill-mode: forwards;
}

@keyframes slideUp {
  0% {
    opacity: 0;
    transform: translateY(60px);
  }

  100% {
    opacity: 1;
    transform: translateY(0);
  }
}

.oops {
  color: var(--el-color-primary);
  @include slide-up-animation(0s);
}

.headline {
  @include slide-up-animation(0.1s);
}

.info {
  @include slide-up-animation(0.2s);
}

.btn {
  @include slide-up-animation(0.3s);
}
</style>
