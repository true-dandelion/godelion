<template>
  <el-container class="h-screen bg-black text-zinc-300">
    <el-aside :width="isCollapse ? '64px' : '240px'" class="border-r border-zinc-800 transition-all duration-300 ease-in-out flex flex-col bg-zinc-950">
      <div class="h-16 flex items-center justify-center border-b border-zinc-800">
        <el-icon v-if="isCollapse" :size="24" class="text-white"><Platform /></el-icon>
        <div v-else class="flex items-center gap-3">
          <el-icon :size="24" class="text-white"><Platform /></el-icon>
          <span class="text-lg font-bold tracking-tight text-white">Godelion</span>
        </div>
      </div>
      
      <el-menu
        :default-active="activeMenu"
        class="border-none bg-transparent !w-full flex-1"
        :collapse="isCollapse"
        background-color="transparent"
        text-color="#a1a1aa"
        active-text-color="#fff"
        router
      >
        <el-menu-item index="/dashboard">
          <el-icon><Odometer /></el-icon>
          <template #title>仪表盘</template>
        </el-menu-item>
        
        <el-menu-item index="/container">
          <el-icon><Box /></el-icon>
          <template #title>容器管控</template>
        </el-menu-item>
        
        <el-menu-item index="/network">
          <el-icon><Connection /></el-icon>
          <template #title>网络中继</template>
        </el-menu-item>

        <el-menu-item index="/file">
          <el-icon><Folder /></el-icon>
          <template #title>文件隔离</template>
        </el-menu-item>

        <el-menu-item index="/ssl">
          <el-icon><Lock /></el-icon>
          <template #title>SSL证书</template>
        </el-menu-item>

        <el-sub-menu index="/docker">
          <template #title>
            <el-icon><Monitor /></el-icon>
            <span>Docker 管理</span>
          </template>
          <el-menu-item index="/docker/status">运行环境状态</el-menu-item>
          <el-menu-item index="/docker/config">配置</el-menu-item>
        </el-sub-menu>

        <el-menu-item index="/settings">
          <el-icon><Setting /></el-icon>
          <template #title>设置</template>
        </el-menu-item>
      </el-menu>

      <div class="p-4 border-t border-zinc-800 text-center text-xs text-zinc-500">
        <span v-if="!isCollapse">v1.0.0-beta</span>
      </div>
    </el-aside>

    <el-container class="flex-1 bg-[#0a0a0a]">
      <el-header class="h-16 border-b border-zinc-800 flex items-center justify-between px-6 bg-zinc-950/80 backdrop-blur">
        <div class="flex items-center gap-4">
          <el-button link @click="isCollapse = !isCollapse" class="text-zinc-400 hover:text-white">
            <el-icon :size="20">
              <Expand v-if="isCollapse" />
              <Fold v-else />
            </el-icon>
          </el-button>
          <h2 class="text-lg font-semibold text-white">{{ currentTitle }}</h2>
        </div>
        
        <div class="flex items-center gap-6">
          <el-badge is-dot class="item">
            <el-icon :size="20" class="text-zinc-400 hover:text-white cursor-pointer transition-colors"><Bell /></el-icon>
          </el-badge>
          <el-dropdown trigger="click" @command="handleCommand">
            <span class="el-dropdown-link flex items-center gap-2 cursor-pointer outline-none">
              <el-avatar :size="32" class="bg-zinc-800 text-zinc-300">
                <el-icon><UserFilled /></el-icon>
              </el-avatar>
              <span class="text-sm text-zinc-300 hover:text-white transition-colors">Admin</span>
              <el-icon class="text-zinc-500"><CaretBottom /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu class="bg-zinc-900 border-zinc-800">
                <el-dropdown-item command="profile">个人中心</el-dropdown-item>
                <el-dropdown-item command="logout" divided class="text-red-400">退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>

      <el-main class="p-6 overflow-x-hidden relative">
        <router-view v-slot="{ Component }">
          <transition name="fade-transform" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '../../store/user'
import { ElMessageBox, ElMessage } from 'element-plus'
import {
  Platform,
  Odometer,
  Box,
  Connection,
  Folder,
  Lock,
  Expand,
  Fold,
  Bell,
  UserFilled,
  CaretBottom,
  Setting,
  Monitor
} from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const isCollapse = ref(false)

const activeMenu = computed(() => route.path)
const currentTitle = computed(() => route.meta.title as string || 'Dashboard')

const handleCommand = (command: string) => {
  if (command === 'logout') {
    ElMessageBox.confirm('确定要退出登录吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
      customClass: 'dark-message-box'
    }).then(() => {
      userStore.logout()
      ElMessage.success('已退出登录')
      router.push('/login')
    }).catch(() => {})
  } else if (command === 'profile') {
    router.push('/settings')
  }
}
</script>

<style>
/* Custom styling for Element Plus Dark Theme */
.dark-message-box {
  background-color: #18181b !important;
  border-color: #27272a !important;
}
.dark-message-box .el-message-box__title,
.dark-message-box .el-message-box__content {
  color: #fff !important;
}
.dark-message-box .el-button--default {
  background-color: transparent !important;
  border-color: #3f3f46 !important;
  color: #fff !important;
}
.dark-message-box .el-button--primary {
  background-color: #fff !important;
  border-color: #fff !important;
  color: #000 !important;
}

.el-menu-item.is-active {
  background-color: rgba(255, 255, 255, 0.05) !important;
  border-right: 2px solid #fff;
}
.el-menu-item:hover {
  background-color: rgba(255, 255, 255, 0.02) !important;
}

/* Page Transitions */
.fade-transform-leave-active,
.fade-transform-enter-active {
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}
.fade-transform-enter-from {
  opacity: 0;
  transform: translateX(-10px);
}
.fade-transform-leave-to {
  opacity: 0;
  transform: translateX(10px);
}
</style>
