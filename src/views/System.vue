<template>
  <div class="space-y-6 max-w-4xl">
    <el-card shadow="never" class="!bg-zinc-900 !border-zinc-800">
      <template #header>
        <div class="flex items-center justify-between">
          <span class="text-lg font-semibold text-white">Docker 运行环境状态</span>
        </div>
      </template>

      <div class="py-4 space-y-6">
        <!-- 状态展示 -->
        <div class="flex items-center gap-6">
          <div class="w-16 h-16 rounded-xl flex items-center justify-center border transition-colors"
               :class="status.installed ? 'bg-blue-500/10 border-blue-500/20' : 'bg-zinc-800 border-zinc-700'">
            <el-icon :size="32" :class="status.installed ? 'text-blue-500' : 'text-zinc-500'">
              <Monitor />
            </el-icon>
          </div>
          <div class="flex-1">
            <h3 class="text-xl font-medium text-white mb-1">
              Docker Engine
              <el-tag v-if="loading" size="small" type="info" class="ml-2 !bg-zinc-800 !border-zinc-700 !text-zinc-400">检查中...</el-tag>
              <el-tag v-else-if="status.installed && status.running" size="small" type="success" class="ml-2 !bg-green-500/20 !border-green-500/30 !text-green-500">运行中</el-tag>
              <el-tag v-else-if="status.installed && !status.running" size="small" type="warning" class="ml-2 !bg-yellow-500/20 !border-yellow-500/30 !text-yellow-500">已停止</el-tag>
              <el-tag v-else size="small" type="danger" class="ml-2 !bg-red-500/20 !border-red-500/30 !text-red-500">未安装</el-tag>
            </h3>
            <p class="text-zinc-400 text-sm">
              Godelion 核心管控和中继服务依赖底层的 Docker Engine 运行。
              <span v-if="!status.installed" class="text-red-400">目前系统未检测到 Docker 环境，部分核心功能将被禁用。</span>
            </p>
          </div>
        </div>

        <el-divider class="!border-zinc-800" />

        <!-- 操作区域 -->
        <div class="flex items-center gap-4">
          <el-button 
            v-if="!status.installed"
            type="primary" 
            size="large"
            :loading="installing"
            :disabled="status.installed || loading"
            @click="handleInstallDocker"
            class="!bg-white !text-black hover:!bg-zinc-200 !border-none"
          >
            <el-icon class="mr-2"><Download /></el-icon>
            {{ installing ? 'Docker 安装中 (预计 1-3 分钟)...' : '一键安装 Docker' }}
          </el-button>

          <el-button
            v-if="status.installed && !status.running"
            type="success"
            size="large"
            :loading="operating"
            :disabled="operating || loading"
            @click="handleStartDocker"
            class="!bg-green-600 !text-white hover:!bg-green-500 !border-none"
          >
            <el-icon class="mr-2"><VideoPlay /></el-icon>
            启动 Docker
          </el-button>

          <el-button
            v-if="status.installed && status.running"
            type="warning"
            size="large"
            :loading="operating"
            :disabled="operating || loading"
            @click="handleStopDocker"
            class="!bg-yellow-600 !text-white hover:!bg-yellow-500 !border-none"
          >
            <el-icon class="mr-2"><VideoPause /></el-icon>
            停止 Docker
          </el-button>

          <el-button 
            size="large"
            :loading="loading"
            @click="fetchDockerStatus"
            class="!bg-zinc-800 !text-zinc-300 hover:!text-white hover:!bg-zinc-700 !border-zinc-700"
          >
            <el-icon class="mr-2"><Refresh /></el-icon>
            刷新状态
          </el-button>
        </div>

        <!-- 安装进度提示 -->
        <el-alert
          v-if="installing"
          title="正在通过国内镜像源下载并安装 Docker，此过程可能需要几分钟，请耐心等待，切勿关闭页面..."
          type="info"
          :closable="false"
          show-icon
          class="!bg-blue-500/10 !text-blue-400 !border-blue-500/20 border"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getDockerStatus, installDocker, startDocker, stopDocker } from '../api'
import { Monitor, Download, Refresh, VideoPlay, VideoPause } from '@element-plus/icons-vue'

const loading = ref(true)
const installing = ref(false)
const operating = ref(false)

const status = ref({
  installed: false,
  running: false
})

const fetchDockerStatus = async () => {
  loading.value = true
  try {
    const res: any = await getDockerStatus()
    if (res.code === 200) {
      status.value = res.data
    }
  } catch (error) {
    console.error('Failed to get docker status:', error)
  } finally {
    loading.value = false
  }
}

const handleInstallDocker = () => {
  ElMessageBox.confirm(
    '此操作将使用国内阿里云镜像源自动安装 Docker（以避免官方源被墙导致下载失败）。确认安装吗？',
    '安装 Docker',
    {
      confirmButtonText: '确认安装',
      cancelButtonText: '取消',
      type: 'warning',
      customClass: 'dark-message-box'
    }
  ).then(async () => {
    installing.value = true
    try {
      const res: any = await installDocker()
      if (res.code === 200) {
        ElMessage.success('Docker 安装并启动成功！')
        await fetchDockerStatus()
      } else {
        ElMessage.error(res.message || '安装失败')
      }
    } catch (error: any) {
      ElMessage.error(error.response?.data?.error || '安装过程发生异常')
    } finally {
      installing.value = false
    }
  }).catch(() => {})
}

const handleStartDocker = async () => {
  operating.value = true
  ElMessage.info('正在启动 Docker 服务...')
  try {
    const res: any = await startDocker()
    if (res.code === 200) {
      ElMessage.success('Docker 启动成功')
      await fetchDockerStatus()
    } else {
      ElMessage.error(res.message || '启动失败')
    }
  } catch (error: any) {
    ElMessage.error(error.response?.data?.error || '启动异常')
  } finally {
    operating.value = false
  }
}

const handleStopDocker = async () => {
  operating.value = true
  ElMessage.warning('正在停止 Docker 服务...')
  try {
    const res: any = await stopDocker()
    if (res.code === 200) {
      ElMessage.success('Docker 停止成功')
      await fetchDockerStatus()
    } else {
      ElMessage.error(res.message || '停止失败')
    }
  } catch (error: any) {
    ElMessage.error(error.response?.data?.error || '停止异常')
  } finally {
    operating.value = false
  }
}

onMounted(() => {
  fetchDockerStatus()
})
</script>