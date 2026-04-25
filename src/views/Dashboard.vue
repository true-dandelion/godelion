<template>
  <div class="space-y-6">
    <!-- Header Summary -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
      <el-card shadow="never" class="!bg-zinc-900 !border-zinc-800">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-zinc-400 text-sm font-medium mb-1">运行中容器</p>
            <h3 class="text-3xl font-bold text-white tracking-tight">{{ runningCount }}</h3>
          </div>
          <div class="w-12 h-12 rounded-full bg-green-500/10 flex items-center justify-center">
            <el-icon :size="24" class="text-green-500"><VideoPlay /></el-icon>
          </div>
        </div>
        <div class="mt-4 flex items-center text-xs text-zinc-500">
          <span class="text-green-500 flex items-center mr-2">
            <el-icon><TopRight /></el-icon> +2
          </span>
          较上周增长
        </div>
      </el-card>

      <el-card shadow="never" class="!bg-zinc-900 !border-zinc-800">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-zinc-400 text-sm font-medium mb-1">中继规则</p>
            <h3 class="text-3xl font-bold text-white tracking-tight">8</h3>
          </div>
          <div class="w-12 h-12 rounded-full bg-blue-500/10 flex items-center justify-center">
            <el-icon :size="24" class="text-blue-500"><Connection /></el-icon>
          </div>
        </div>
        <div class="mt-4 flex items-center text-xs text-zinc-500">
          正常运行中
        </div>
      </el-card>

      <el-card shadow="never" class="!bg-zinc-900 !border-zinc-800">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-zinc-400 text-sm font-medium mb-1">存储占用</p>
            <h3 class="text-3xl font-bold text-white tracking-tight">45<span class="text-lg text-zinc-500 font-normal">GB</span></h3>
          </div>
          <div class="w-12 h-12 rounded-full bg-orange-500/10 flex items-center justify-center">
            <el-icon :size="24" class="text-orange-500"><Coin /></el-icon>
          </div>
        </div>
        <div class="mt-4 w-full">
          <el-progress :percentage="45" :stroke-width="4" :show-text="false" color="#f97316" />
        </div>
      </el-card>

      <el-card shadow="never" class="!bg-zinc-900 !border-zinc-800">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-zinc-400 text-sm font-medium mb-1">系统负载</p>
            <h3 class="text-3xl font-bold text-white tracking-tight">2.4<span class="text-lg text-zinc-500 font-normal">%</span></h3>
          </div>
          <div class="w-12 h-12 rounded-full bg-purple-500/10 flex items-center justify-center">
            <el-icon :size="24" class="text-purple-500"><Cpu /></el-icon>
          </div>
        </div>
        <div class="mt-4 w-full">
          <el-progress :percentage="2.4" :stroke-width="4" :show-text="false" color="#a855f7" />
        </div>
      </el-card>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <!-- Recent Activity -->
      <el-card shadow="never" class="lg:col-span-2 !bg-zinc-900 !border-zinc-800">
        <template #header>
          <div class="flex items-center justify-between">
            <span class="text-lg font-semibold text-white">近期系统活动</span>
            <el-button link type="primary" class="!text-zinc-400 hover:!text-white">查看全部</el-button>
          </div>
        </template>
        <el-timeline class="mt-4">
          <el-timeline-item
            v-for="(activity, index) in activities"
            :key="index"
            :type="activity.type"
            :color="activity.color"
            :timestamp="activity.timestamp"
            placement="top"
          >
            <div class="text-zinc-300 font-medium">{{ activity.content }}</div>
            <div class="text-xs text-zinc-500 mt-1">{{ activity.description }}</div>
          </el-timeline-item>
        </el-timeline>
      </el-card>

      <!-- Quick Actions & System Status -->
      <div class="space-y-6">
        <el-card shadow="never" class="!bg-zinc-900 !border-zinc-800">
          <template #header>
            <span class="text-lg font-semibold text-white">快捷操作</span>
          </template>
          <div class="grid grid-cols-2 gap-4">
            <el-button class="!h-24 !bg-zinc-800 hover:!bg-zinc-700 !border-none !text-zinc-300 flex flex-col items-center justify-center gap-2 !m-0 transition-colors">
              <el-icon :size="24"><Box /></el-icon>
              <span>部署容器</span>
            </el-button>
            <el-button class="!h-24 !bg-zinc-800 hover:!bg-zinc-700 !border-none !text-zinc-300 flex flex-col items-center justify-center gap-2 !m-0 transition-colors">
              <el-icon :size="24"><Link /></el-icon>
              <span>添加路由</span>
            </el-button>
            <el-button class="!h-24 !bg-zinc-800 hover:!bg-zinc-700 !border-none !text-zinc-300 flex flex-col items-center justify-center gap-2 !m-0 transition-colors">
              <el-icon :size="24"><Upload /></el-icon>
              <span>上传文件</span>
            </el-button>
            <el-button class="!h-24 !bg-zinc-800 hover:!bg-zinc-700 !border-none !text-zinc-300 flex flex-col items-center justify-center gap-2 !m-0 transition-colors">
              <el-icon :size="24"><Setting /></el-icon>
              <span>系统设置</span>
            </el-button>
          </div>
        </el-card>

        <el-card shadow="never" class="!bg-zinc-900 !border-zinc-800">
          <template #header>
            <span class="text-lg font-semibold text-white">节点健康状态</span>
          </template>
          <div class="space-y-4">
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-2">
                <div class="w-2 h-2 rounded-full bg-green-500"></div>
                <span class="text-zinc-300 text-sm">Docker Daemon</span>
              </div>
              <span class="text-xs text-green-500 bg-green-500/10 px-2 py-1 rounded">在线</span>
            </div>
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-2">
                <div class="w-2 h-2 rounded-full bg-green-500"></div>
                <span class="text-zinc-300 text-sm">SQLite Database</span>
              </div>
              <span class="text-xs text-green-500 bg-green-500/10 px-2 py-1 rounded">正常</span>
            </div>
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-2">
                <div class="w-2 h-2 rounded-full bg-green-500"></div>
                <span class="text-zinc-300 text-sm">Godelion Proxy</span>
              </div>
              <span class="text-xs text-green-500 bg-green-500/10 px-2 py-1 rounded">在线</span>
            </div>
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-2">
                <div class="w-2 h-2 rounded-full bg-zinc-500"></div>
                <span class="text-zinc-300 text-sm">系统更新</span>
              </div>
              <span class="text-xs text-zinc-400">已是最新</span>
            </div>
          </div>
        </el-card>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getWorkloads } from '../api'
import {
  VideoPlay,
  Connection,
  Coin,
  Cpu,
  TopRight,
  Box,
  Link,
  Upload,
  Setting
} from '@element-plus/icons-vue'

const runningCount = ref(0)

const fetchDashboardData = async () => {
  try {
    const res = await getWorkloads()
    if (res.code === 200) {
      const workloads = res.data || []
      runningCount.value = workloads.filter((w: any) => w.status === 'running' || w.status === 'Up').length
    }
  } catch (error) {
    console.error('Failed to fetch dashboard data', error)
  }
}

onMounted(() => {
  fetchDashboardData()
})

const activities = ref([
  {
    content: '系统初始化完成',
    description: 'Godelion 服务冷启动自检通过',
    timestamp: '刚刚',
    color: '#a1a1aa'
  }
])
</script>
