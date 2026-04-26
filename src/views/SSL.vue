<template>
  <div class="h-full flex flex-col space-y-4">
    <!-- 顶栏控制区 -->
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-4">
        <el-input
          v-model="searchDomain"
          placeholder="搜索域名..."
          class="w-64"
          :prefix-icon="Search"
          clearable
        />
      </div>
      <el-button type="primary" @click="openAddDialog" class="!bg-white !text-black hover:!bg-zinc-200 !border-none">
        <el-icon class="mr-2"><Plus /></el-icon> 添加证书
      </el-button>
    </div>

    <!-- 证书列表 -->
    <el-card shadow="never" class="flex-1 !bg-zinc-900 !border-zinc-800 flex flex-col">
      <el-table
        :data="filteredCerts"
        v-loading="loading"
        style="width: 100%"
        class="custom-dark-table"
      >
        <el-table-column prop="domain" label="绑定域名" min-width="180">
          <template #default="{ row }">
            <span class="text-zinc-200 font-medium">{{ row.domain }}</span>
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
                编辑
              </el-button>

              <el-button 
                link 
                type="danger" 
                class="!p-0 hover:opacity-80"
                @click="handleDelete(row)"
              >
                删除
              </el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 粘贴/上传/设置证书对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEditing ? '证书设置' : '添加证书'"
      width="800px"
      custom-class="dark-dialog"
      :destroy-on-close="true"
    >
      <el-form :model="certForm" label-position="top" class="mt-4 px-4">
        <el-form-item label="绑定域名" required>
          <el-input v-model="certForm.domain" placeholder="例如: api.godelion.com" :disabled="isEditing" />
        </el-form-item>

        <el-form-item label="输入方式" required>
          <el-radio-group v-model="inputMethod" class="!w-full flex">
            <el-radio label="paste">粘贴内容</el-radio>
            <el-radio label="upload">上传文件</el-radio>
          </el-radio-group>
        </el-form-item>

        <template v-if="inputMethod === 'paste'">
          <div class="flex gap-4 w-full">
            <el-form-item label="证书 (CRT/PEM)" required class="flex-1" label-position="top">
              <el-input 
                v-model="certForm.cert_content" 
                type="textarea" 
                :rows="12" 
                :resize="'none'"
                placeholder="-----BEGIN CERTIFICATE-----&#10;...&#10;-----END CERTIFICATE-----" 
                class="font-mono text-xs"
              />
            </el-form-item>

            <el-form-item label="私钥 (KEY)" required class="flex-1" label-position="top">
              <el-input 
                v-model="certForm.key_content" 
                type="textarea" 
                :rows="12" 
                :resize="'none'"
                placeholder="-----BEGIN PRIVATE KEY-----&#10;...&#10;-----END PRIVATE KEY-----" 
                class="font-mono text-xs"
              />
            </el-form-item>
          </div>
        </template>

        <template v-else-if="inputMethod === 'upload'">
          <div class="flex gap-4 w-full">
            <el-form-item label="证书文件" required class="flex-1" label-position="top">
              <input type="file" accept=".crt,.pem" @change="e => handleFileUpload(e, 'cert')" class="text-zinc-300 w-full" />
              <div class="text-xs text-zinc-500 mt-1" v-if="certForm.cert_content">已加载证书: {{ certForm.cert_content.length }} bytes</div>
            </el-form-item>
            <el-form-item label="私钥文件" required class="flex-1" label-position="top">
              <input type="file" accept=".key,.pem" @change="e => handleFileUpload(e, 'key')" class="text-zinc-300 w-full" />
              <div class="text-xs text-zinc-500 mt-1" v-if="certForm.key_content">已加载私钥: {{ certForm.key_content.length }} bytes</div>
            </el-form-item>
          </div>
        </template>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="dialogVisible = false" class="!bg-transparent !text-white !border-zinc-700">取消</el-button>
          <el-button type="primary" @click="submitCert" class="!bg-white !text-black !border-none hover:!bg-zinc-200">
            {{ isEditing ? '保存修改' : '确认添加' }}
          </el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getSSLCerts, createSSLCert, deleteSSLCert } from '../api'
import {
  Plus,
  Search
} from '@element-plus/icons-vue'

const loading = ref(false)
const certs = ref<any[]>([])
const searchDomain = ref('')

const dialogVisible = ref(false)
const isEditing = ref(false)
const inputMethod = ref<'paste' | 'upload'>('paste')

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

const filteredCerts = computed(() => {
  return certs.value.filter(c => c.domain.includes(searchDomain.value))
})

const resetForm = () => {
  certForm.domain = ''
  certForm.cert_content = ''
  certForm.key_content = ''
}

const openAddDialog = () => {
  isEditing.value = false
  inputMethod.value = 'paste'
  resetForm()
  dialogVisible.value = true
}

const handleSettings = (row: any) => {
  isEditing.value = true
  inputMethod.value = 'paste'
  resetForm()
  certForm.domain = row.domain
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
      ElMessage.success(isEditing.value ? '证书更新成功' : '证书保存成功')
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
:deep(.el-switch.is-checked .el-switch__core) {
  background-color: #3f3f46;
  border-color: #3f3f46;
}
:deep(.el-switch.is-checked .el-switch__action) {
  background-color: #fff;
}

/* Custom Scrollbar for textareas - hide by default, show on hover */
:deep(.el-textarea__inner::-webkit-scrollbar) {
  width: 6px;
  height: 6px;
  background-color: transparent;
}
:deep(.el-textarea__inner::-webkit-scrollbar-thumb) {
  background-color: transparent;
  border-radius: 4px;
}
:deep(.el-textarea__inner:hover::-webkit-scrollbar-thumb) {
  background-color: #3f3f46;
}
:deep(.el-textarea__inner::-webkit-scrollbar-track) {
  background-color: transparent;
}
:deep(.el-textarea__inner::-webkit-scrollbar-thumb:hover) {
  background-color: #52525b;
}
</style>
