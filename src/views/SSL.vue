<template>
  <div class="p-6">
    <div class="flex justify-between items-center mb-6">
      <h1 class="text-2xl font-bold text-white">SSL 证书管理</h1>
      <div class="flex gap-2">
        <el-button type="primary" @click="openPasteDialog" class="!bg-blue-600 !border-none hover:!bg-blue-500">
          <el-icon class="mr-1"><DocumentAdd /></el-icon>
          粘贴证书
        </el-button>
        <el-button type="success" @click="openUploadDialog" class="!bg-emerald-600 !border-none hover:!bg-emerald-500">
          <el-icon class="mr-1"><Upload /></el-icon>
          上传证书
        </el-button>
      </div>
    </div>

    <el-card class="!bg-zinc-900 !border-zinc-800 !text-zinc-300 shadow-xl rounded-xl">
      <el-table
        :data="certs"
        style="width: 100%"
        class="!bg-transparent custom-dark-table"
        :header-cell-style="{ background: '#18181b', color: '#a1a1aa', borderBottom: '1px solid #27272a' }"
        :row-style="{ background: 'transparent' }"
        :cell-style="{ borderBottom: '1px solid #27272a' }"
        v-loading="loading"
        element-loading-background="rgba(24, 24, 27, 0.8)"
      >
        <el-table-column prop="domain" label="绑定域名" min-width="180">
          <template #default="{ row }">
            <span class="text-white font-medium">{{ row.domain }}</span>
          </template>
        </el-table-column>
        
        <el-table-column prop="updated_at" label="更新时间" min-width="180">
          <template #default="{ row }">
            <span class="text-zinc-400">{{ new Date(row.updated_at).toLocaleString() }}</span>
          </template>
        </el-table-column>

        <el-table-column label="操作" width="150" align="right">
          <template #default="{ row }">
            <div class="flex items-center justify-end gap-3">
              <el-button 
                link 
                type="primary" 
                class="!p-0 hover:opacity-80"
                @click="handleSettings(row)"
              >
                <el-icon :size="18"><Setting /></el-icon>
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

    <!-- 粘贴/设置证书对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogMode === 'paste' ? '粘贴证书' : (dialogMode === 'upload' ? '上传证书' : '证书设置')"
      width="600px"
      custom-class="dark-dialog"
      :destroy-on-close="true"
    >
      <el-form :model="certForm" label-width="100px" class="mt-4 pr-8">
        <el-form-item label="绑定域名" required>
          <el-input v-model="certForm.domain" placeholder="例如: api.godelion.com" :disabled="dialogMode === 'settings'" />
        </el-form-item>

        <template v-if="dialogMode === 'paste' || dialogMode === 'settings'">
          <el-form-item label="证书 (CRT/PEM)" required>
            <el-input 
              v-model="certForm.cert_content" 
              type="textarea" 
              :rows="6" 
              placeholder="-----BEGIN CERTIFICATE-----&#10;...&#10;-----END CERTIFICATE-----" 
            />
          </el-form-item>

          <el-form-item label="私钥 (KEY)" required>
            <el-input 
              v-model="certForm.key_content" 
              type="textarea" 
              :rows="6" 
              placeholder="-----BEGIN PRIVATE KEY-----&#10;...&#10;-----END PRIVATE KEY-----" 
            />
          </el-form-item>
        </template>

        <template v-else-if="dialogMode === 'upload'">
          <el-form-item label="证书文件" required>
            <input type="file" accept=".crt,.pem" @change="e => handleFileUpload(e, 'cert')" class="text-zinc-300 w-full" />
            <div class="text-xs text-zinc-500 mt-1" v-if="certForm.cert_content">已加载证书: {{ certForm.cert_content.length }} bytes</div>
          </el-form-item>
          <el-form-item label="私钥文件" required>
            <input type="file" accept=".key,.pem" @change="e => handleFileUpload(e, 'key')" class="text-zinc-300 w-full" />
            <div class="text-xs text-zinc-500 mt-1" v-if="certForm.key_content">已加载私钥: {{ certForm.key_content.length }} bytes</div>
          </el-form-item>
        </template>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="dialogVisible = false" class="!bg-transparent !text-white !border-zinc-700">取消</el-button>
          <el-button type="primary" @click="submitCert" class="!bg-white !text-black !border-none hover:!bg-zinc-200">
            保存证书
          </el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getSSLCerts, createSSLCert, deleteSSLCert } from '../api'
import {
  Upload,
  DocumentAdd,
  Delete,
  Setting
} from '@element-plus/icons-vue'

const loading = ref(false)
const certs = ref<any[]>([])

const dialogVisible = ref(false)
const dialogMode = ref<'paste' | 'upload' | 'settings'>('paste')

const certForm = reactive({
  domain: '',
  cert_content: '',
  key_content: ''
})

const fetchCerts = async () => {
  loading.value = true
  try {
    const res = await getSSLCerts()
    if (res.code === 200) {
      certs.value = res.data || []
    } else {
      ElMessage.error(res.message || '获取证书列表失败')
    }
  } catch (error) {
    ElMessage.error('获取证书异常')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchCerts()
})

const resetForm = () => {
  certForm.domain = ''
  certForm.cert_content = ''
  certForm.key_content = ''
}

const openPasteDialog = () => {
  dialogMode.value = 'paste'
  resetForm()
  dialogVisible.value = true
}

const openUploadDialog = () => {
  dialogMode.value = 'upload'
  resetForm()
  dialogVisible.value = true
}

const handleSettings = (row: any) => {
  dialogMode.value = 'settings'
  resetForm()
  certForm.domain = row.domain
  // We don't fetch private keys securely in list API, so they are blank here unless we add a get endpoint
  // But for now we just allow them to overwrite
  certForm.cert_content = ''
  certForm.key_content = ''
  dialogVisible.value = true
}

const handleFileUpload = (event: Event, type: 'cert' | 'key') => {
  const target = event.target as HTMLInputElement
  if (target.files && target.files.length > 0) {
    const file = target.files[0]
    const reader = new FileReader()
    reader.onload = (e) => {
      if (e.target && typeof e.target.result === 'string') {
        if (type === 'cert') {
          certForm.cert_content = e.target.result
        } else {
          certForm.key_content = e.target.result
        }
      }
    }
    reader.readAsText(file)
  }
}

const submitCert = async () => {
  if (!certForm.domain) {
    ElMessage.warning('请输入绑定域名')
    return
  }
  if (!certForm.cert_content || !certForm.key_content) {
    ElMessage.warning('证书和私钥内容不能为空')
    return
  }

  loading.value = true
  try {
    const payload = {
      domain: certForm.domain,
      cert_content: certForm.cert_content,
      key_content: certForm.key_content
    }
    
    const res = await createSSLCert(payload)

    if (res.code === 200) {
      ElMessage.success('证书保存成功')
      dialogVisible.value = false
      fetchCerts()
    } else {
      ElMessage.error(res.message || '保存失败')
    }
  } catch (error: any) {
    ElMessage.error(error.response?.data?.error || '保存异常')
  } finally {
    loading.value = false
  }
}

const handleDelete = (row: any) => {
  ElMessageBox.confirm(`确定要删除域名 ${row.domain} 的证书吗？会导致对应的 HTTPS 访问失效。`, '危险操作', {
    confirmButtonText: '确认删除',
    cancelButtonText: '取消',
    type: 'error',
    customClass: 'dark-message-box'
  }).then(async () => {
    loading.value = true
    try {
      const res = await deleteSSLCert(row.id)
      if (res.code === 200) {
        ElMessage.success('证书已删除')
        fetchCerts()
      } else {
        ElMessage.error(res.message || '删除失败')
      }
    } catch (error) {
      ElMessage.error('请求异常')
    } finally {
      loading.value = false
    }
  }).catch(() => {})
}
</script>

<style scoped>
.custom-dark-table {
  --el-table-border-color: #27272a;
  --el-table-header-bg-color: #18181b;
  --el-table-bg-color: transparent;
  --el-table-tr-bg-color: transparent;
  --el-table-row-hover-bg-color: rgba(255, 255, 255, 0.03);
}

:deep(.el-table tbody tr:hover > td) {
  background-color: rgba(255, 255, 255, 0.03) !important;
}

:deep(.el-table::before) {
  display: none;
}
</style>
