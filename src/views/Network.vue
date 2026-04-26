<template>
  <div class="h-full flex flex-col space-y-4">
    <!-- 顶栏控制区 -->
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-4">
        <el-input
          v-model="searchDomain"
          placeholder="搜索域名或容器 ID..."
          class="w-64"
          :prefix-icon="Search"
          clearable
        />
        <el-select v-model="protocolFilter" placeholder="协议" class="w-32" clearable>
          <el-option label="HTTP" value="http" />
          <el-option label="HTTPS" value="https" />
          <el-option label="TCP" value="tcp" />
        </el-select>
      </div>
      <el-button type="primary" @click="ruleDialogVisible = true" class="!bg-white !text-black hover:!bg-zinc-200 !border-none">
        <el-icon class="mr-2"><Plus /></el-icon> 添加路由规则
      </el-button>
    </div>

    <!-- 路由规则列表 -->
    <el-card shadow="never" class="flex-1 !bg-zinc-900 !border-zinc-800 flex flex-col">
      <el-table
        :data="filteredRules"
        v-loading="loading"
        style="width: 100%"
        class="custom-dark-table"
      >
        <el-table-column prop="domain" label="访问域名 / 主机端口" min-width="220">
          <template #default="{ row }">
            <div class="flex items-center gap-2">
              <el-icon v-if="row.tls_enabled" class="text-green-500"><Lock /></el-icon>
              <el-icon v-else class="text-zinc-500"><Unlock /></el-icon>
              <span v-if="row.is_port_mapping" class="text-zinc-200 font-mono">{{ row.domain }}</span>
              <a v-else :href="(row.tls_enabled ? 'https://' : 'http://') + row.domain" target="_blank" class="text-blue-400 hover:text-blue-300 font-medium underline decoration-blue-500/30 underline-offset-4">
                {{ row.domain }}
              </a>
            </div>
          </template>
        </el-table-column>
        
        <el-table-column prop="listen_ports" label="监听端口" width="150">
          <template #default="{ row }">
            <el-tag size="small" type="success" class="!bg-green-500/10 !border-green-500/20 !text-green-400">
              {{ row.listen_ports || '80, 443' }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="target_urls" label="目标地址/端口" min-width="250">
          <template #default="{ row }">
            <span class="text-zinc-400 text-sm font-mono break-all">{{ row.target_urls }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="protocol" label="协议支持" width="150">
          <template #default="{ row }">
            <div class="flex gap-1">
              <el-tag size="small" :type="row.tls_enabled ? 'success' : 'info'" class="!bg-zinc-800 !border-zinc-700 !text-zinc-300">
                {{ row.tls_enabled ? 'HTTPS' : 'HTTP' }}
              </el-tag>
              <el-tag v-if="row.ws_enabled" size="small" type="warning" class="!bg-zinc-800 !border-zinc-700 !text-zinc-300">
                WS
              </el-tag>
            </div>
          </template>
        </el-table-column>

        <el-table-column prop="certStatus" label="证书状态" width="150">
          <template #default="{ row }">
            <span v-if="!row.tls_enabled" class="text-zinc-500 text-sm">-</span>
            <div v-else class="flex items-center gap-2">
              <div class="w-2 h-2 rounded-full bg-green-500 shadow-[0_0_8px_rgba(34,197,94,0.5)]"></div>
              <span class="text-sm text-green-500">正常 (28天后过期)</span>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <div class="flex items-center gap-3">
              <template v-if="!row.is_port_mapping">
                <el-button link type="primary" class="!p-0 hover:opacity-80">
                  编辑
                </el-button>
                <el-button link type="warning" class="!p-0 hover:opacity-80" v-if="row.tls_enabled">
                  续签证书
                </el-button>
                <el-button link type="danger" class="!p-0 hover:opacity-80" @click="handleDelete(row)">
                  删除
                </el-button>
              </template>
              <template v-else>
                <el-tag size="small" type="info" class="!bg-zinc-800 !border-zinc-700 !text-zinc-500">
                  容器随附映射，不可单删
                </el-tag>
              </template>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 添加路由规则对话框 -->
    <el-dialog
      v-model="ruleDialogVisible"
      title="配置中继规则"
      width="500px"
      custom-class="dark-dialog"
      :destroy-on-close="true"
    >
      <el-form :model="ruleForm" label-width="100px" class="mt-4 pr-8">
        <el-form-item label="访问域名" required>
          <el-input v-model="ruleForm.domain" placeholder="例如: api.godelion.com" />
        </el-form-item>
        <el-form-item label="监听端口" prop="listen_ports" required>
          <el-input v-model="ruleForm.listen_ports" placeholder="如: 80, 443, 8080" />
        </el-form-item>

        <el-form-item label="目标类型" required>
          <el-radio-group v-model="targetType" class="!w-full flex">
            <el-radio label="custom">自定义地址</el-radio>
            <el-radio label="container">选择容器</el-radio>
          </el-radio-group>
        </el-form-item>

        <template v-if="targetType === 'custom'">
          <el-form-item label="目标地址" prop="target_urls" required>
            <el-input v-model="ruleForm.target_urls" placeholder="如: 127.0.0.1:3000 (支持多个，逗号隔开)" />
          </el-form-item>
        </template>

        <template v-else>
          <el-form-item label="目标容器" required>
            <el-select v-model="selectedContainer" placeholder="请选择目标容器" class="w-full" @change="handleContainerSelect">
              <el-option
                v-for="c in containersList"
                :key="c.id"
                :label="c.name"
                :value="c.id"
              >
                <div class="flex justify-between items-center">
                  <span>{{ c.name }}</span>
                  <span class="text-zinc-500 text-xs ml-2">{{ c.parsedPorts?.map((p: any) => p.container).join(', ') || '无映射' }}</span>
                </div>
              </el-option>
            </el-select>
          </el-form-item>
          
          <el-form-item label="容器端口" required v-if="selectedContainer">
            <el-select v-model="selectedContainerPort" placeholder="选择要转发到的容器内端口" class="w-full">
              <el-option
                v-for="port in availableContainerPorts"
                :key="port"
                :label="port"
                :value="port"
              />
            </el-select>
          </el-form-item>
        </template>

        <el-form-item label="HTTPS 开关" prop="tls_enabled">
          <div class="flex flex-col gap-1 w-full">
            <el-switch v-model="ruleForm.tls_enabled" />
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="ruleDialogVisible = false" class="!bg-transparent !text-white !border-zinc-700">取消</el-button>
          <el-button type="primary" @click="handleAddRule" class="!bg-white !text-black !border-none hover:!bg-zinc-200">
            保存规则
          </el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getGateways, createGateway, deleteGateway, getWorkloads } from '../api'
import {
  Search,
  Plus,
  Lock,
  Unlock,
  Right
} from '@element-plus/icons-vue'

const loading = ref(false)
const searchDomain = ref('')
const protocolFilter = ref('')
const ruleDialogVisible = ref(false)

const targetType = ref('custom')
const containersList = ref<any[]>([])
const selectedContainer = ref('')
const selectedContainerPort = ref('')
const availableContainerPorts = ref<string[]>([])

const ruleForm = reactive({
  domain: '',
  listen_ports: '80, 443',
  target_urls: '',
  tls_enabled: false
})

const rules = ref<any[]>([])

const fetchRules = async () => {
  loading.value = true
  try {
    const res = await getGateways()
    if (res.code === 200) {
      rules.value = res.data || []
    } else {
      ElMessage.error(res.message || '获取规则失败')
    }
  } catch (error) {
    ElMessage.error('获取规则异常')
  } finally {
    loading.value = false
  }
}

const fetchContainers = async () => {
  try {
    const res = await getWorkloads()
    if (res.code === 200) {
      containersList.value = (res.data || []).map((c: any) => {
        try {
          const parsed = typeof c.ports === 'string' && c.ports ? JSON.parse(c.ports) : []
          c.parsedPorts = Array.isArray(parsed) ? parsed.filter((p: any) => p.container) : []
        } catch {
          c.parsedPorts = []
        }
        return c
      })
    }
  } catch (error) {
    console.error('获取容器列表异常', error)
  }
}

const handleContainerSelect = (val: string) => {
  const container = containersList.value.find(c => c.id === val)
  if (container) {
    availableContainerPorts.value = container.parsedPorts.map((p: any) => p.container)
    if (availableContainerPorts.value.length > 0) {
      selectedContainerPort.value = availableContainerPorts.value[0]
    } else {
      selectedContainerPort.value = ''
    }
  }
}

onMounted(() => {
  fetchRules()
  fetchContainers()
})

const filteredRules = computed(() => {
  return rules.value.filter(r => {
    const matchQuery = r.domain.includes(searchDomain.value) || r.container_id.includes(searchDomain.value)
    const matchProtocol = protocolFilter.value 
      ? (protocolFilter.value === 'https' ? r.tls_enabled : (protocolFilter.value === 'http' ? !r.tls_enabled : true))
      : true
    return matchQuery && matchProtocol
  })
})

const handleAddRule = async () => {
  if (targetType.value === 'container') {
    if (!selectedContainer.value) {
      ElMessage.warning('请选择目标容器')
      return
    }
    if (!selectedContainerPort.value) {
      ElMessage.warning('请选择目标容器的内部端口')
      return
    }
  }

  if (targetType.value === 'custom' && (!ruleForm.domain || !ruleForm.listen_ports || !ruleForm.target_urls)) {
    ElMessage.warning('请填写完整的路由规则')
    return
  }
  
  loading.value = true
  try {
    const payload = {
      domain: ruleForm.domain,
      listen_ports: ruleForm.listen_ports,
      target_urls: targetType.value === 'custom' ? ruleForm.target_urls : '',
      tls_enabled: ruleForm.tls_enabled,
      container_id: targetType.value === 'container' ? selectedContainer.value : '',
      target_port: targetType.value === 'container' ? parseInt(selectedContainerPort.value) : 0
    }
    const res = await createGateway(payload)
    if (res.code === 200) {
      ElMessage.success('规则添加成功')
      ruleDialogVisible.value = false
      ruleForm.domain = ''
      ruleForm.listen_ports = '80, 443'
      ruleForm.target_urls = ''
      ruleForm.tls_enabled = false
      targetType.value = 'custom'
      selectedContainer.value = ''
      selectedContainerPort.value = ''
      fetchRules()
    } else {
      ElMessage.error(res.message || '添加失败')
    }
  } catch (error: any) {
    ElMessage.error(error.response?.data?.error || '添加异常')
  } finally {
    loading.value = false
  }
}

const handleDelete = (row: any) => {
  ElMessageBox.confirm(`确定要删除域名 ${row.domain} 的路由规则吗？`, '危险操作', {
    confirmButtonText: '删除',
    cancelButtonText: '取消',
    type: 'error',
    customClass: 'dark-message-box'
  }).then(async () => {
    try {
      const res = await deleteGateway(row.id)
      if (res.code === 200) {
        ElMessage.success('删除成功')
        fetchRules()
      } else {
        ElMessage.error(res.message || '删除失败')
      }
    } catch (error) {
      ElMessage.error('删除异常')
    }
  }).catch(() => {})
}

const tableRowClassName = ({ rowIndex }: { rowIndex: number }) => {
  return 'dark-row transition-colors hover:bg-zinc-800/50'
}
</script>

<style scoped>
:deep(.el-switch.is-checked .el-switch__core) {
  background-color: #3f3f46;
  border-color: #3f3f46;
}
:deep(.el-switch.is-checked .el-switch__action) {
  background-color: #fff;
}
</style>
