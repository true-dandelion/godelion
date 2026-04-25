<template>
  <div class="h-full flex flex-col space-y-4">
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-4">
        <el-input
          v-model="searchQuery"
          placeholder="搜索容器名称或镜像..."
          class="w-64"
          :prefix-icon="Search"
          clearable
        />
        <el-select v-model="statusFilter" placeholder="状态过滤" class="w-32" clearable>
          <el-option label="运行中" value="running" />
          <el-option label="已停止" value="stopped" />
          <el-option label="异常" value="error" />
        </el-select>
      </div>
      <el-button type="primary" @click="deployDialogVisible = true" class="!bg-white !text-black hover:!bg-zinc-200 !border-none">
        <el-icon class="mr-2"><Plus /></el-icon> 部署容器
      </el-button>
    </div>

    <el-card shadow="never" class="flex-1 !bg-zinc-900 !border-zinc-800 flex flex-col">
      <el-table
        :data="filteredContainers"
        v-loading="loading"
        style="width: 100%"
        class="custom-dark-table"
        :row-class-name="tableRowClassName"
      >
        <el-table-column prop="name" label="容器名称" min-width="180">
          <template #default="{ row }">
            <div class="flex items-center gap-3">
              <div class="w-8 h-8 rounded-lg bg-zinc-800 flex items-center justify-center border border-zinc-700">
                <el-icon class="text-zinc-300"><Box /></el-icon>
              </div>
              <div class="flex flex-col">
                <span class="text-zinc-200 font-medium">{{ row.name }}</span>
                <span class="text-xs text-zinc-500 font-mono">{{ row.id.substring(0, 12) }}</span>
              </div>
            </div>
          </template>
        </el-table-column>
        
        <el-table-column prop="image" label="镜像" min-width="200">
          <template #default="{ row }">
            <span class="text-zinc-400 font-mono text-sm bg-zinc-800/50 px-2 py-1 rounded border border-zinc-700/50">{{ row.image }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="status" label="状态" width="120">
          <template #default="{ row }">
            <div class="flex items-center gap-2">
              <div class="w-2 h-2 rounded-full" :class="{
                'bg-green-500 shadow-[0_0_8px_rgba(34,197,94,0.5)]': row.status === 'running',
                'bg-zinc-500': row.status === 'stopped',
                'bg-red-500 shadow-[0_0_8px_rgba(239,68,68,0.5)]': row.status === 'error'
              }"></div>
              <span class="text-sm capitalize" :class="{
                'text-green-500': row.status === 'running',
                'text-zinc-400': row.status === 'stopped',
                'text-red-500': row.status === 'error'
              }">{{ row.status }}</span>
            </div>
          </template>
        </el-table-column>

        <el-table-column prop="ports" label="端口映射" min-width="150">
          <template #default="{ row }">
            <div v-if="row.parsedPorts && row.parsedPorts.length > 0" class="flex flex-wrap gap-1">
              <el-tag v-for="(port, idx) in row.parsedPorts" :key="idx" size="small" type="info" class="!bg-zinc-800 !border-zinc-700 !text-zinc-300 font-mono">
                {{ port.host }}:{{ port.container }}
              </el-tag>
            </div>
            <span v-else class="text-zinc-600 text-sm">-</span>
          </template>
        </el-table-column>

        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <div class="flex items-center gap-2">
              <el-button 
                v-if="row.status !== 'running' && row.status !== 'Up'" 
                link 
                type="success" 
                class="!p-0 hover:opacity-80"
                @click="handleAction(row, 'start')"
              >
                <el-icon :size="18"><VideoPlay /></el-icon>
              </el-button>
              
              <el-button 
                v-if="row.status === 'running' || row.status === 'Up'" 
                link 
                type="warning" 
                class="!p-0 hover:opacity-80"
                @click="handleAction(row, 'stop')"
              >
                <el-icon :size="18"><VideoPause /></el-icon>
              </el-button>

              <el-button 
                link 
                type="primary" 
                class="!p-0 hover:opacity-80"
                @click="handleAction(row, 'restart')"
              >
                <el-icon :size="18"><RefreshRight /></el-icon>
              </el-button>

              <el-divider direction="vertical" class="!border-zinc-700 mx-2" />

              <el-button 
                link 
                type="info" 
                class="!p-0 hover:opacity-80"
                @click="handleViewLogs(row)"
              >
                <el-icon :size="18"><Document /></el-icon>
              </el-button>

              <el-button 
                link 
                type="danger" 
                class="!p-0 hover:opacity-80"
                @click="handleDelete(row)"
              >
                <el-icon :size="18"><Delete /></el-icon>
              </el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 部署容器对话框 -->
    <el-dialog
      v-model="deployDialogVisible"
      title="部署前端/Node.js 项目"
      width="650px"
      custom-class="dark-dialog"
      :destroy-on-close="true"
    >
      <el-form :model="deployForm" label-width="120px" class="mt-4 pr-8">
        <el-form-item label="名称" required>
          <el-input v-model="deployForm.name" placeholder="例如: my-vue-app" />
        </el-form-item>

        <el-form-item label="Node.js 版本" required>
          <el-select v-model="deployForm.nodeVersion" class="w-full">
            <el-option label="Node 24 Alpine" value="node:24-alpine" />
            <el-option label="Node 25 Alpine" value="node:25-alpine" />
          </el-select>
        </el-form-item>

        <el-form-item label="项目目录" required>
          <el-cascader
            v-model="deployForm.projectDir"
            :props="folderProps"
            placeholder="选择 /godelion/user/ 下的项目文件夹"
            class="w-full"
            :show-all-levels="true"
            @change="handleProjectDirChange"
          />
        </el-form-item>

        <el-form-item label="启动命令" required>
          <el-select 
            v-model="deployForm.startCommand" 
            placeholder="选择解析的脚本，或手动输入 (如 node app.js)" 
            class="w-full"
            filterable
            allow-create
            default-first-option
          >
            <el-option 
              v-for="cmd in availableScripts" 
              :key="cmd" 
              :label="cmd" 
              :value="cmd" 
            />
          </el-select>
        </el-form-item>

        <el-form-item label="容器名称" required>
          <el-input v-model="deployForm.containerName" placeholder="例如: my-vue-app-container" />
        </el-form-item>

        <el-form-item label="包管理器" required v-if="hasPackageJson">
          <el-select v-model="deployForm.packageManager" class="w-full">
            <el-option label="npm" value="npm" />
            <el-option label="yarn" value="yarn" />
            <el-option label="pnpm" value="pnpm" />
          </el-select>
        </el-form-item>

        <el-form-item label="附加依赖" v-if="!hasPackageJson">
          <el-input 
            v-model="deployForm.dependencies" 
            placeholder="如需安装依赖，请用英文逗号隔开，例如: express, uuid" 
          />
        </el-form-item>

        <el-form-item label="端口映射">
          <div v-for="(port, index) in deployForm.ports" :key="index" class="flex gap-2 mb-2">
            <el-input v-model="port.host" placeholder="主机端口" class="flex-1" />
            <span class="text-zinc-500 pt-1">:</span>
            <el-input v-model="port.container" placeholder="容器端口 (例如 5173)" class="flex-1" />
            <el-button type="danger" link @click="removePort(index)">
              <el-icon><Remove /></el-icon>
            </el-button>
          </div>
          <el-button type="primary" link @click="addPort" class="!text-zinc-400 hover:!text-white mt-1">
            <el-icon class="mr-1"><Plus /></el-icon> 添加端口映射
          </el-button>
        </el-form-item>

        <el-form-item label="资源限制">
          <div class="flex gap-4 w-full">
            <el-input-number v-model="deployForm.cpu" :min="0.1" :max="8" :step="0.1" placeholder="CPU核数" class="!w-1/2" />
            <el-input v-model="deployForm.memory" placeholder="内存(例如: 512m)" class="!w-1/2" />
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="deployDialogVisible = false" class="!bg-transparent !text-white !border-zinc-700">取消</el-button>
          <el-button type="primary" @click="handleDeploy" class="!bg-white !text-black !border-none hover:!bg-zinc-200">
            确认部署
          </el-button>
        </span>
      </template>
    </el-dialog>

    <!-- 日志查看对话框 -->
    <el-dialog
      v-model="logDialogVisible"
      :title="`容器日志: ${currentLogContainerName}`"
      width="800px"
      custom-class="dark-dialog"
      :destroy-on-close="true"
    >
      <div 
        class="bg-black text-green-400 font-mono p-4 rounded-lg h-96 overflow-y-auto whitespace-pre-wrap text-sm border border-zinc-800"
      >
        <span v-if="containerLogs">{{ containerLogs }}</span>
        <span v-else class="text-zinc-500">暂无日志或正在加载中...</span>
      </div>
      <template #footer>
        <span class="dialog-footer flex justify-between items-center">
          <el-button link type="primary" @click="refreshLogs" class="!text-zinc-400 hover:!text-white">
            <el-icon class="mr-1"><RefreshRight /></el-icon> 刷新
          </el-button>
          <el-button @click="logDialogVisible = false" class="!bg-white !text-black !border-none hover:!bg-zinc-200">关闭</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getWorkloads, createWorkload, startWorkload, stopWorkload, getFiles, readFile, getWorkloadLogs } from '../api'
import {
  Search,
  Plus,
  Box,
  VideoPlay,
  VideoPause,
  RefreshRight,
  Delete,
  Document,
  Remove
} from '@element-plus/icons-vue'

const loading = ref(false)
const searchQuery = ref('')
const statusFilter = ref('')
const deployDialogVisible = ref(false)

const logDialogVisible = ref(false)
const containerLogs = ref('')
const currentLogContainerId = ref('')
const currentLogContainerName = ref('')

const handleViewLogs = async (row: any) => {
  currentLogContainerId.value = row.id
  currentLogContainerName.value = row.name
  containerLogs.value = ''
  logDialogVisible.value = true
  await fetchLogs()
}

const refreshLogs = async () => {
  ElMessage.info('刷新日志...')
  await fetchLogs()
}

const fetchLogs = async () => {
  if (!currentLogContainerId.value) return
  try {
    const res = await getWorkloadLogs(currentLogContainerId.value)
    if (res.code === 200) {
      containerLogs.value = res.data || '容器暂无输出日志'
    } else {
      containerLogs.value = `获取日志失败: ${res.message}`
    }
  } catch (error) {
    containerLogs.value = '获取日志异常，请检查后端连接'
  }
}

const deployForm = reactive({
  name: '',
  nodeVersion: 'node:24-alpine',
  projectDir: [] as string[],
  startCommand: '',
  containerName: '',
  packageManager: 'npm',
  dependencies: '', // for comma-separated deps
  ports: [{ host: '', container: '' }],
  cpu: 1,
  memory: '512m'
})

const availableScripts = ref<string[]>([])
const hasPackageJson = ref(true)

const handleProjectDirChange = async (val: string[]) => {
  if (!val || val.length === 0) {
    availableScripts.value = []
    hasPackageJson.value = true
    return
  }
  const path = '/' + val.join('/')
  const pkgJsonPath = `${path === '/' ? '' : path}/package.json`

  try {
    const res = await readFile(pkgJsonPath)
    if (res.code === 200 && res.data) {
      hasPackageJson.value = true
      const pkg = typeof res.data === 'string' ? JSON.parse(res.data) : res.data
      if (pkg && pkg.scripts) {
        availableScripts.value = Object.keys(pkg.scripts)
        if (availableScripts.value.length > 0 && !deployForm.startCommand) {
          deployForm.startCommand = availableScripts.value[0]
        }
        ElMessage.success('已成功解析 package.json 脚本')
      } else {
        availableScripts.value = []
        ElMessage.info('未发现 scripts 配置，请手动输入启动命令')
      }
    } else {
      hasPackageJson.value = false
      availableScripts.value = []
      ElMessage.info('未找到 package.json，请手动输入启动命令。若需要安装依赖，请在下方填写')
    }
  } catch (err) {
    hasPackageJson.value = false
    availableScripts.value = []
    ElMessage.info('未找到 package.json，请手动输入启动命令。若需要安装依赖，请在下方填写')
  }
}

const folderProps = {
  lazy: true,
  checkStrictly: true,
  async lazyLoad(node: any, resolve: any) {
    const { level, path } = node
    
    try {
      // 动态解析真实路径以查询后端
      const currentPath = level === 0 
        ? '/' 
        : '/' + path.join('/')
        
      const res = await getFiles(currentPath)
      let items: any[] = []

      if (res.code === 200 && res.data) {
        // 只筛选文件夹，项目目录通常是文件夹
        const folderItems = res.data
          .filter((f: any) => f.is_dir)
          .map((f: any) => ({
            value: f.name,
            label: f.name,
            leaf: false // 文件夹不是叶子节点，可以继续点开
          }))
        items = items.concat(folderItems)
      }
      resolve(items)
    } catch {
      resolve([])
    }
  }
}

const addVolume = () => {
  deployForm.volumes.push({ host: [], container: '' })
}

const removeVolume = (index: number) => {
  deployForm.volumes.splice(index, 1)
}

const containers = ref<any[]>([])

const fetchContainers = async () => {
  loading.value = true
  try {
    const res = await getWorkloads()
    if (res.code === 200) {
      containers.value = (res.data || []).map((c: any) => {
        try {
          const parsed = typeof c.ports === 'string' && c.ports ? JSON.parse(c.ports) : []
          c.parsedPorts = Array.isArray(parsed) ? parsed.filter((p: any) => p.host || p.container) : []
        } catch {
          c.parsedPorts = []
        }
        return c
      })
    } else {
      ElMessage.error(res.message || '获取容器列表失败')
    }
  } catch (error) {
    ElMessage.error('获取容器列表异常')
  } finally {
    loading.value = false
  }
}

const filteredContainers = computed(() => {
  return containers.value.filter(c => {
    const matchQuery = c.name.includes(searchQuery.value) || c.image.includes(searchQuery.value)
    const matchStatus = statusFilter.value ? c.status === statusFilter.value : true
    return matchQuery && matchStatus
  })
})

const tableRowClassName = ({ rowIndex }: { rowIndex: number }) => {
  return 'dark-row transition-colors hover:bg-zinc-800/50'
}

const addPort = () => {
  deployForm.ports.push({ host: '', container: '' })
}

const removePort = (index: number) => {
  deployForm.ports.splice(index, 1)
}

const handleAction = async (row: any, action: string) => {
  ElMessage.info(`正在${action === 'start' ? '启动' : '停止'}容器 ${row.name}...`)
  try {
    let res: any
    if (action === 'start') {
      res = await startWorkload(row.id)
    } else if (action === 'stop') {
      res = await stopWorkload(row.id)
    }
    
    if (res && res.code === 200) {
      ElMessage.success(`${action === 'start' ? '启动' : '停止'}成功`)
      fetchContainers()
    } else {
      ElMessage.error(res?.message || '操作失败')
    }
  } catch (error) {
    ElMessage.error('操作异常')
  }
}

const handleDelete = (row: any) => {
  ElMessageBox.confirm(`确定要删除容器 ${row.name} 吗？此操作不可恢复。`, '危险操作', {
    confirmButtonText: '强制删除',
    cancelButtonText: '取消',
    type: 'error',
    customClass: 'dark-message-box'
  }).then(() => {
    ElMessage.warning('后端暂未实现物理删除接口')
  }).catch(() => {})
}

const handleDeploy = async () => {
  if (!deployForm.name || !deployForm.projectDir.length || !deployForm.containerName || !deployForm.startCommand) {
    ElMessage.warning('请填写完整的容器和项目信息')
    return
  }
  loading.value = true
  
  try {
    const validPorts = deployForm.ports.filter(p => p.host && p.container)
    
    // Resolve project directory path
    const projectDirPath = '/' + deployForm.projectDir.join('/')

    const payload = {
      name: deployForm.name,
      node_version: deployForm.nodeVersion,
      project_dir: projectDirPath,
      start_command: deployForm.startCommand,
      container_name: deployForm.containerName,
      package_manager: deployForm.packageManager,
      dependencies: deployForm.dependencies,
      ports: validPorts,
      resource_limits: `cpu=${deployForm.cpu},mem=${deployForm.memory}`
    }
    
    const res = await createWorkload(payload)
    if (res.code === 200) {
      ElMessage.success('前端/Node项目部署成功')
      deployDialogVisible.value = false
      deployForm.name = ''
      deployForm.nodeVersion = 'node:24-alpine'
      deployForm.projectDir = []
      deployForm.startCommand = ''
      deployForm.containerName = ''
      deployForm.packageManager = 'npm'
      deployForm.dependencies = ''
      deployForm.ports = [{ host: '', container: '' }]
      availableScripts.value = []
      fetchContainers()
    } else {
      ElMessage.error(res.message || '创建失败')
    }
  } catch (error: any) {
    ElMessage.error(error.response?.data?.error || '创建异常')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchContainers()
})
</script>

<style>
.custom-dark-table {
  background-color: transparent !important;
  --el-table-border-color: #27272a;
  --el-table-header-bg-color: #18181b;
  --el-table-header-text-color: #a1a1aa;
  --el-table-text-color: #e4e4e7;
  --el-table-row-hover-bg-color: #27272a;
}
.custom-dark-table th.el-table__cell {
  background-color: #18181b !important;
  border-bottom: 1px solid #27272a !important;
  font-weight: 600;
}
.custom-dark-table td.el-table__cell {
  border-bottom: 1px solid #27272a !important;
  background-color: transparent !important;
}
.custom-dark-table::before {
  display: none;
}
.dark-row {
  background-color: transparent !important;
}

.dark-dialog {
  background-color: #18181b !important;
  border: 1px solid #27272a;
  border-radius: 12px;
}
.dark-dialog .el-dialog__title {
  color: #fff;
  font-weight: 600;
}
.dark-dialog .el-dialog__headerbtn .el-dialog__close {
  color: #a1a1aa;
}
.dark-dialog .el-dialog__headerbtn:hover .el-dialog__close {
  color: #fff;
}
.dark-dialog .el-form-item__label {
  color: #a1a1aa;
}
.dark-dialog .el-input__wrapper {
  background-color: #09090b !important;
  box-shadow: 0 0 0 1px #27272a inset !important;
}
.dark-dialog .el-input__wrapper.is-focus {
  box-shadow: 0 0 0 1px #fff inset !important;
}
.dark-dialog .el-input__inner {
  color: #fff !important;
}

/* Fix Cascader Dark Theme */
.el-cascader-panel {
  background-color: #18181b !important;
  border-color: #27272a !important;
}
.el-cascader-menu {
  border-right: 1px solid #27272a !important;
  color: #e4e4e7 !important;
}
.el-cascader-node:hover, .el-cascader-node.is-selectable.in-active-path {
  background-color: #27272a !important;
}
.el-cascader-node.is-active {
  color: #fff !important;
  font-weight: bold;
}
.el-popper.is-light {
  background: #18181b !important;
  border: 1px solid #27272a !important;
}
.el-popper.is-light .el-popper__arrow::before {
  background: #18181b !important;
  border: 1px solid #27272a !important;
}
</style>
