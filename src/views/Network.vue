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
      <el-button type="primary" @click="openAddDialog" class="!bg-white !text-black hover:!bg-zinc-200 !border-none">
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
        
        <el-table-column label="监听端口" width="180">
          <template #default="{ row }">
            <div class="flex gap-1">
              <el-tag v-if="row.http_port" size="small" type="info" class="!bg-blue-500/10 !border-blue-500/20 !text-blue-400">
                HTTP:{{ row.http_port }}
              </el-tag>
              <el-tag v-if="row.tls_enabled && row.https_port" size="small" type="success" class="!bg-green-500/10 !border-green-500/20 !text-green-400">
                HTTPS:{{ row.https_port }}
              </el-tag>
            </div>
          </template>
        </el-table-column>

        <el-table-column prop="target_urls" label="目标地址/端口" min-width="250">
          <template #default="{ row }">
            <span class="text-zinc-400 text-sm font-mono break-all">{{ row.target_urls }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="certStatus" label="证书状态" width="150">
          <template #default="{ row }">
            <span v-if="!row.tls_enabled" class="text-zinc-500 text-sm">-</span>
            <div v-else class="flex items-center gap-2">
              <template v-if="row.ssl_cert_id">
                <div class="w-2 h-2 rounded-full bg-green-500 shadow-[0_0_8px_rgba(34,197,94,0.5)]"></div>
                <span class="text-sm text-green-500">已绑定</span>
              </template>
              <template v-else>
                <div class="w-2 h-2 rounded-full bg-yellow-500 shadow-[0_0_8px_rgba(234,179,8,0.5)]"></div>
                <span class="text-sm text-yellow-500">自动匹配中</span>
              </template>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <div class="flex items-center gap-3">
              <template v-if="!row.is_port_mapping">
                <el-button link type="primary" class="!p-0 hover:opacity-80" @click="handleEdit(row)">
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

    <!-- 添加/编辑路由规则对话框 -->
    <el-dialog
      v-model="ruleDialogVisible"
      :title="isEditing ? '编辑中继规则' : '配置中继规则'"
      width="520px"
      custom-class="dark-dialog"
      :destroy-on-close="true"
    >
      <el-form :model="ruleForm" label-width="100px" class="mt-4 pr-8">
        <el-form-item label="访问域名" required>
          <el-input v-model="ruleForm.domain" placeholder="例如: api.godelion.com" />
        </el-form-item>

        <!-- 规则类型 -->
        <el-form-item label="规则类型" required>
          <el-radio-group v-model="ruleForm.rule_type" class="!w-full flex">
            <el-radio label="proxy">反向代理</el-radio>
            <el-radio label="redirect">跳转重定向</el-radio>
          </el-radio-group>
        </el-form-item>

        <!-- 跳转配置（仅跳转类型显示） -->
        <template v-if="ruleForm.rule_type === 'redirect'">
          <el-form-item label="跳转地址" required>
            <el-input v-model="ruleForm.redirect_url" placeholder="例如: https://new.example.com" />
          </el-form-item>
          <el-form-item label="跳转类型" required>
            <el-radio-group v-model="ruleForm.redirect_code">
              <el-radio :label="301">301 永久重定向</el-radio>
              <el-radio :label="302">302 临时重定向</el-radio>
            </el-radio-group>
          </el-form-item>
        </template>

        <!-- HTTP 端口 -->
        <el-form-item label="HTTP 端口" :required="!ruleForm.tls_enabled">
          <el-input v-model="ruleForm.http_port" placeholder="例如: 80" :disabled="false" />
          <p v-if="ruleForm.tls_enabled" class="text-xs text-zinc-500 mt-1">留空则不监听 HTTP，填写后自动跳转到 HTTPS</p>
        </el-form-item>

        <!-- HTTPS 开关 -->
        <el-form-item label="HTTPS">
          <el-switch v-model="ruleForm.tls_enabled" />
        </el-form-item>

        <!-- HTTPS 端口（仅开启 HTTPS 时显示） -->
        <el-form-item v-if="ruleForm.tls_enabled" label="HTTPS 端口" required>
          <el-input v-model="ruleForm.https_port" placeholder="例如: 443" />
        </el-form-item>

        <!-- SSL 证书（仅开启 HTTPS 时显示） -->
        <el-form-item v-if="ruleForm.tls_enabled" label="SSL 证书" required>
          <el-select v-model="ruleForm.ssl_cert_id" placeholder="请选择绑定的 SSL 证书" class="w-full">
            <el-option
              v-for="cert in sslCerts"
              :key="cert.id"
              :label="cert.domain"
              :value="cert.id"
            >
              <div class="flex justify-between items-center">
                <span>{{ cert.domain }}</span>
                <span class="text-zinc-500 text-xs ml-2">{{ new Date(cert.expires_at).toLocaleDateString() }} 到期</span>
              </div>
            </el-option>
          </el-select>
        </el-form-item>

        <!-- 目标类型（仅代理类型显示） -->
        <el-form-item v-if="ruleForm.rule_type === 'proxy'" label="目标类型" required>
          <el-radio-group v-model="targetType" class="!w-full flex">
            <el-radio label="custom">自定义地址</el-radio>
            <el-radio label="container">选择容器</el-radio>
          </el-radio-group>
        </el-form-item>

        <template v-if="ruleForm.rule_type === 'proxy' && targetType === 'custom'">
          <el-form-item label="目标地址" prop="target_urls" required>
            <el-input v-model="ruleForm.target_urls" placeholder="如: 127.0.0.1:3000 (支持多个，逗号隔开)" />
          </el-form-item>
        </template>

        <template v-if="ruleForm.rule_type === 'proxy' && targetType !== 'custom'">
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
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="ruleDialogVisible = false" class="!bg-transparent !text-white !border-zinc-700">取消</el-button>
          <el-button type="primary" @click="handleAddRule" class="!bg-white !text-black !border-none hover:!bg-zinc-200">
            {{ isEditing ? '保存修改' : '确认添加' }}
          </el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getGateways, createGateway, deleteGateway, getWorkloads, updateGateway, getSSLCerts } from '../api'
import {
  Search,
  Plus,
  Lock,
  Unlock,
  Right,
  Edit
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
const isEditing = ref(false)
const currentEditId = ref('')

const ruleForm = reactive({
  domain: '',
  http_port: '80',
  https_port: '443',
  tls_enabled: false,
  ssl_cert_id: '',
  target_urls: '',
  rule_type: 'proxy',
  redirect_url: '',
  redirect_code: 301
})

const rules = ref<any[]>([])
const sslCerts = ref<any[]>([])

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

const fetchSSLCerts = async () => {
  try {
    const res = await getSSLCerts()
    if (res.code === 200) {
      sslCerts.value = res.data || []
    }
  } catch (error) {
    console.error('获取证书列表异常', error)
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
  fetchSSLCerts()
})

const filteredRules = computed(() => {
  return rules.value.filter(r => {
    const matchQuery = r.domain.includes(searchDomain.value) || (r.container_id && r.container_id.includes(searchDomain.value))
    const matchProtocol = protocolFilter.value 
      ? (protocolFilter.value === 'https' ? r.tls_enabled : (protocolFilter.value === 'http' ? !r.tls_enabled : true))
      : true
    return matchQuery && matchProtocol
  })
})

const handleEdit = (row: any) => {
  isEditing.value = true
  currentEditId.value = row.id
  ruleForm.domain = row.domain
  ruleForm.http_port = row.http_port || ''
  ruleForm.https_port = row.https_port || '443'
  ruleForm.tls_enabled = row.tls_enabled
  ruleForm.ssl_cert_id = row.ssl_cert_id || ''
  ruleForm.rule_type = row.rule_type || 'proxy'
  ruleForm.redirect_url = row.redirect_url || ''
  ruleForm.redirect_code = row.redirect_code || 301

  if (row.target_urls && row.target_urls.startsWith('Container: ')) {
    targetType.value = 'container'
    const match = row.target_urls.match(/Container: (.+) \((\d+)\)/)
    if (match) {
      const cName = match[1]
      const cPort = match[2]
      const container = containersList.value.find(c => c.name === cName)
      if (container) {
        selectedContainer.value = container.id
        handleContainerSelect(container.id)
        selectedContainerPort.value = cPort
      }
    }
  } else {
    targetType.value = 'custom'
    ruleForm.target_urls = row.target_urls || ''
  }

  ruleDialogVisible.value = true
}

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

  if (targetType.value === 'custom' && ruleForm.rule_type === 'proxy' && (!ruleForm.domain || !ruleForm.target_urls)) {
    ElMessage.warning('请填写完整的路由规则')
    return
  }

  // Validate redirect
  if (ruleForm.rule_type === 'redirect') {
    if (!ruleForm.redirect_url) {
      ElMessage.warning('请填写跳转地址')
      return
    }
  }

  // Validate port requirements
  if (ruleForm.tls_enabled) {
    if (!ruleForm.https_port) {
      ElMessage.warning('开启 HTTPS 时必须填写 HTTPS 端口')
      return
    }
    if (!ruleForm.ssl_cert_id) {
      ElMessage.warning('开启 HTTPS 时必须选择 SSL 证书')
      return
    }
  } else {
    if (!ruleForm.http_port) {
      ElMessage.warning('未开启 HTTPS 时必须填写 HTTP 端口')
      return
    }
  }
  
  loading.value = true
  try {
    const payload = {
      domain: ruleForm.domain,
      http_port: ruleForm.http_port,
      https_port: ruleForm.tls_enabled ? ruleForm.https_port : '',
      target_urls: ruleForm.rule_type === 'proxy' && targetType.value === 'custom' ? ruleForm.target_urls : '',
      tls_enabled: ruleForm.tls_enabled,
      ssl_cert_id: ruleForm.tls_enabled ? ruleForm.ssl_cert_id : '',
      container_id: ruleForm.rule_type === 'proxy' && targetType.value === 'container' ? selectedContainer.value : '',
      target_port: ruleForm.rule_type === 'proxy' && targetType.value === 'container' ? parseInt(selectedContainerPort.value) : 0,
      rule_type: ruleForm.rule_type,
      redirect_url: ruleForm.rule_type === 'redirect' ? ruleForm.redirect_url : '',
      redirect_code: ruleForm.rule_type === 'redirect' ? ruleForm.redirect_code : 301
    }
    
    let res
    if (isEditing.value) {
      res = await updateGateway(currentEditId.value, payload)
    } else {
      res = await createGateway(payload)
    }

    if (res.code === 200) {
      ElMessage.success(isEditing.value ? '规则更新成功' : '规则添加成功')
      ruleDialogVisible.value = false
      resetForm()
      fetchRules()
    } else {
      ElMessage.error(res.message || (isEditing.value ? '更新失败' : '添加失败'))
    }
  } catch (error: any) {
    ElMessage.error(error.response?.data?.error || (isEditing.value ? '更新异常' : '添加异常'))
  } finally {
    loading.value = false
  }
}

const resetForm = () => {
  ruleForm.domain = ''
  ruleForm.http_port = '80'
  ruleForm.https_port = '443'
  ruleForm.tls_enabled = false
  ruleForm.ssl_cert_id = ''
  ruleForm.target_urls = ''
  ruleForm.rule_type = 'proxy'
  ruleForm.redirect_url = ''
  ruleForm.redirect_code = 301
  targetType.value = 'custom'
  selectedContainer.value = ''
  selectedContainerPort.value = ''
  isEditing.value = false
  currentEditId.value = ''
}

const openAddDialog = () => {
  resetForm()
  ruleDialogVisible.value = true
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
