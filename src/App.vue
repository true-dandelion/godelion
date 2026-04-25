<template>
  <el-config-provider :builtin-theme="isDark ? 'dark' : 'light'">
    <router-view />
  </el-config-provider>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const isDark = ref(true) // 默认深色主题，或者根据系统设置

// 初始化暗色模式
if (isDark.value) {
  document.documentElement.classList.add('dark')
}
</script>

<style>
body {
  margin: 0;
  padding: 0;
  background-color: #000;
  color: #fff;
}
html.dark {
  color-scheme: dark;
}
</style>
