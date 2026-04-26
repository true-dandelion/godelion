<template>
  <div class="space-y-6">
    <!-- Header Summary -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
      <el-card shadow="never" class="!bg-zinc-900 !border-zinc-800">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-zinc-400 text-sm font-medium mb-1">容器数量</p>
            <h3 class="text-3xl font-bold text-white tracking-tight">{{ workloads.length }}</h3>
          </div>
          <div class="w-12 h-12 rounded-full bg-green-500/10 flex items-center justify-center">
            <el-icon :size="24" class="text-green-500"><Box /></el-icon>
          </div>
        </div>
        <div class="mt-4 flex items-center text-xs text-zinc-500">
          <span class="text-green-500 flex items-center mr-2">
            <div class="w-1.5 h-1.5 rounded-full bg-green-500 mr-1"></div>
            {{ runningCount }} 运行
          </span>
          <span class="text-zinc-500 flex items-center">
            <div class="w-1.5 h-1.5 rounded-full bg-zinc-500 mr-1"></div>
            {{ workloads.length - runningCount }} 未运行
          </span>
        </div>
      </el-card>

      <el-card shadow="never" class="!bg-zinc-900 !border-zinc-800">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-zinc-400 text-sm font-medium mb-1">中继数量</p>
            <h3 class="text-3xl font-bold text-white tracking-tight">{{ gateways.length }}</h3>
          </div>
          <div class="w-12 h-12 rounded-full bg-blue-500/10 flex items-center justify-center">
            <el-icon :size="24" class="text-blue-500"><Connection /></el-icon>
          </div>
        </div>
        <div class="mt-4 flex items-center text-xs text-zinc-500">
          <span class="text-blue-400 flex items-center mr-2">
            <div class="w-1.5 h-1.5 rounded-full bg-blue-400 mr-1"></div>
            {{ normalGatewayCount }} 正常使用
          </span>
          <span class="text-zinc-500 flex items-center">
            <div class="w-1.5 h-1.5 rounded-full bg-zinc-500 mr-1"></div>
            {{ gateways.length - normalGatewayCount }} 停止使用
          </span>
        </div>
      </el-card>

      <el-card shadow="never" class="!bg-zinc-900 !border-zinc-800">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-zinc-400 text-sm font-medium mb-1">CPU 占用</p>
            <h3 class="text-3xl font-bold text-white tracking-tight">{{ cpuUsage }}<span class="text-lg text-zinc-500 font-normal">%</span></h3>
          </div>
          <div class="w-12 h-12 rounded-full bg-orange-500/10 flex items-center justify-center">
            <el-icon :size="24" class="text-orange-500"><Cpu /></el-icon>
          </div>
        </div>
        <div class="mt-4 w-full">
          <el-progress :percentage="cpuUsage" :stroke-width="4" :show-text="false" color="#f97316" />
        </div>
      </el-card>

      <el-card shadow="never" class="!bg-zinc-900 !border-zinc-800">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-zinc-400 text-sm font-medium mb-1">内存占用</p>
            <h3 class="text-3xl font-bold text-white tracking-tight">{{ memUsage }}<span class="text-lg text-zinc-500 font-normal">%</span></h3>
          </div>
          <div class="w-12 h-12 rounded-full bg-purple-500/10 flex items-center justify-center">
            <el-icon :size="24" class="text-purple-500"><Odometer /></el-icon>
          </div>
        </div>
        <div class="mt-4 w-full">
          <el-progress :percentage="memUsage" :stroke-width="4" :show-text="false" color="#a855f7" />
        </div>
      </el-card>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <!-- Recent Activity -->
      <el-card shadow="never" class="lg:col-span-2 !bg-zinc-900 !border-zinc-800">
        <template #header>
          <div class="flex items-center justify-between">
            <span class="text-lg font-semibold text-white">近期系统活动</span>
            <el-button link type="primary" class="!text-zinc-400 hover:!text-white" @click="fetchAuditLogs">刷新记录</el-button>
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
          <div v-if="activities.length === 0" class="text-center text-zinc-500 py-4">暂无活动记录</div>
        </el-timeline>
      </el-card>

      <!-- Quick Actions & System Status -->
      <div class="space-y-6">
        <el-card shadow="never" class="!bg-zinc-900 !border-zinc-800">
          <template #header>
            <span class="text-lg font-semibold text-white">快捷操作</span>
          </template>
          <div class="grid grid-cols-2 gap-4">
            <el-button @click="$router.push('/container')" class="!h-24 !bg-zinc-800 hover:!bg-zinc-700 !border-none !text-zinc-300 flex flex-col items-center justify-center gap-2 !m-0 transition-colors">
              <el-icon :size="24"><Box /></el-icon>
              <span>容器管控</span>
            </el-button>
            <el-button @click="$router.push('/network')" class="!h-24 !bg-zinc-800 hover:!bg-zinc-700 !border-none !text-zinc-300 flex flex-col items-center justify-center gap-2 !m-0 transition-colors">
              <el-icon :size="24"><Connection /></el-icon>
              <span>网络中继</span>
            </el-button>
            <el-button @click="$router.push('/ssl')" class="!h-24 !bg-zinc-800 hover:!bg-zinc-700 !border-none !text-zinc-300 flex flex-col items-center justify-center gap-2 !m-0 transition-colors">
              <el-icon :size="24"><Lock /></el-icon>
              <span>SSL 证书</span>
            </el-button>
            <el-button @click="$router.push('/file')" class="!h-24 !bg-zinc-800 hover:!bg-zinc-700 !border-none !text-zinc-300 flex flex-col items-center justify-center gap-2 !m-0 transition-colors">
              <el-icon :size="24"><Folder /></el-icon>
              <span>文件隔离</span>
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
          </div>
        </el-card>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { getWorkloads, getGateways, getAuditLogs } from '../api'
import {
  VideoPlay,
  Connection,
  Coin,
  Cpu,
  TopRight,
  Box,
  Link,
  Upload,
  Setting,
  Lock,
  Folder,
  Odometer
} from '@element-plus/icons-vue'

const router = useRouter()

const workloads = ref<any[]>([])
const runningCount = ref(0)

const gateways = ref<any[]>([])
const normalGatewayCount = ref(0)

const cpuUsage = ref(Math.floor(Math.random() * 20) + 1)
const memUsage = ref(Math.floor(Math.random() * 40) + 20)

const activities = ref<any[]>([])

const fetchDashboardData = async () => {
  try {
    const [resW, resG] = await Promise.all([
      getWorkloads(),
      getGateways()
    ])

    if (resW.code === 200) {
      workloads.value = resW.data || []
      runningCount.value = workloads.value.filter((w: any) => w.status === 'running' || w.status === 'Up').length
    }

    if (resG.code === 200) {
      gateways.value = resG.data || []
      
      // Calculate normal vs stopped gateways. 
      // Custom gateways (no container id) are considered normal.
      // Gateways bound to containers check the container's status.
      normalGatewayCount.value = gateways.value.filter((g: any) => {
        if (!g.container_id) return true // Custom targets are always "normal"
        const targetContainer = workloads.value.find((w: any) => w.id === g.container_id)
        if (targetContainer && (targetContainer.status === 'running' || targetContainer.status === 'Up')) {
          return true
        }
        return false
      }).length
    }
  } catch (error) {
    console.error('Failed to fetch dashboard data', error)
  }
}

const fetchAuditLogs = async () => {
  try {
    const res = await getAuditLogs()
    if (res.code === 200 && res.data) {
      activities.value = res.data.map((log: any) => {
        let color = '#a1a1aa'
        let type = 'info'
        
        switch (log.action) {
          case 'Login': color = '#3b82f6'; type = 'primary'; break
          case 'Create': 
          case 'Deploy':
          case 'Start': color = '#22c55e'; type = 'success'; break
          case 'Update': color = '#eab308'; type = 'warning'; break
          case 'Delete': 
          case 'Stop': color = '#ef4444'; type = 'danger'; break
        }

        return {
          content: `${log.action} ${log.resource}`,
          description: log.details,
          timestamp: new Date(log.created_at).toLocaleString(),
          color: color,
          type: type
        }
      })
    }
  } catch (error) {
    console.error('Failed to fetch audit logs', error)
  }
}

onMounted(() => {
  fetchDashboardData()
  fetchAuditLogs()
  
  // Mock CPU/Mem usage updates
  setInterval(() => {
    cpuUsage.value = Math.max(1, Math.min(100, cpuUsage.value + (Math.random() * 10 - 5)))
    cpuUsage.value = Math.round(cpuUsage.value * 10) / 10
    memUsage.value = Math.max(10, Math.min(90, memUsage.value + (Math.random() * 4 - 2)))
    memUsage.value = Math.round(memUsage.value * 10) / 10
  }, 5000)
})
</script>
