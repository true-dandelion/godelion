<template>
  <div class="space-y-6 max-w-4xl mx-auto mt-8">
    <!-- 用户设置 -->
    <el-card shadow="never" class="!bg-zinc-900 !border-zinc-800">
      <template #header>
        <div class="flex items-center gap-3">
          <el-icon class="text-white text-xl"><User /></el-icon>
          <span class="text-lg font-semibold text-white">用户设置</span>
        </div>
      </template>
      <div class="py-4">
        <el-form :model="userForm" label-position="top" v-loading="userLoading">
          <el-form-item label="用户">
            <el-input v-model="userForm.new_username" placeholder="更改当前用户名" class="w-full" />
          </el-form-item>
          <el-form-item label="密码">
            <el-input v-model="userForm.new_password" type="password" placeholder="修改当前用户的密码" show-password class="w-full" />
          </el-form-item>
          <el-form-item label="当前密码（修改密码时必填）">
            <el-input v-model="userForm.current_password" type="password" placeholder="输入当前密码以确认修改" show-password class="w-full" />
          </el-form-item>
          <div class="mt-4 flex justify-end">
            <el-button type="primary" :loading="userSaving" @click="handleUserSave" class="!bg-white !text-black !border-none hover:!bg-zinc-200 px-8">
              保存用户设置
            </el-button>
          </div>
        </el-form>
      </div>
    </el-card>

    <!-- 面板设置 -->
    <el-card shadow="never" class="!bg-zinc-900 !border-zinc-800">
      <template #header>
        <div class="flex items-center gap-3">
          <el-icon class="text-white text-xl"><Setting /></el-icon>
          <span class="text-lg font-semibold text-white">面板设置</span>
        </div>
      </template>
      <div class="py-4">
        <el-form :model="configForm" label-position="top" v-loading="configLoading">
          <el-form-item label="面板名称">
            <el-input v-model="configForm.panel_name" placeholder="设置面板名称" class="w-full" />
          </el-form-item>
          
          <el-form-item>
            <div class="flex items-center gap-2">
              <span class="text-zinc-300">超时时间</span>
              <el-tooltip content="如果用户超过 86400 秒未操作面板，面板将自动退出登录" placement="top">
                <el-icon class="text-zinc-500 cursor-help"><QuestionFilled /></el-icon>
              </el-tooltip>
            </div>
            <el-input-number v-model="configForm.session_timeout" :min="60" :max="864000" class="w-full" />
          </el-form-item>

          <el-form-item>
            <div class="flex items-center gap-2">
              <span class="text-zinc-300">面板 HTTPS</span>
              <el-tooltip content="为面板设置HTTPS可提升访问安全性" placement="top">
                <el-icon class="text-zinc-500 cursor-help"><QuestionFilled /></el-icon>
              </el-tooltip>
            </div>
            <el-switch v-model="configForm.enable_https" />
          </el-form-item>

          <el-form-item label="端口">
            <div class="flex gap-4">
              <div class="flex-1">
                <span class="text-zinc-500 text-sm mb-1 block">HTTP 端口</span>
                <el-input-number v-model="configForm.http_port" :min="1" :max="65535" class="w-full" />
              </div>
              <div class="flex-1">
                <span class="text-zinc-500 text-sm mb-1 block">HTTPS 端口</span>
                <el-input-number v-model="configForm.https_port" :min="1" :max="65535" class="w-full" />
              </div>
            </div>
          </el-form-item>

          <el-form-item>
            <div class="flex items-center gap-2">
              <span class="text-zinc-300">安全入口</span>
              <el-tooltip content="开启安全访问接口时，只有访问正确的接口时，才能够访问到" placement="top">
                <el-icon class="text-zinc-500 cursor-help"><QuestionFilled /></el-icon>
              </el-tooltip>
            </div>
            <el-input v-model="configForm.secure_entrypoint" placeholder="例如: /admin" class="w-full" />
          </el-form-item>

          <el-form-item>
            <div class="flex items-center gap-2">
              <span class="text-zinc-300">授权 IP</span>
              <el-tooltip content="设置授权 IP 后，仅有设置中的 IP 可以访问本服务" placement="top">
                <el-icon class="text-zinc-500 cursor-help"><QuestionFilled /></el-icon>
              </el-tooltip>
            </div>
            <el-input v-model="configForm.authorized_ips" placeholder="多个 IP 用逗号分隔，例如: 192.168.1.1,10.0.0.1" class="w-full" />
          </el-form-item>

          <el-form-item>
            <div class="flex items-center gap-2">
              <span class="text-zinc-300">域名绑定</span>
              <el-tooltip content="设置域名绑定后，仅能通过设置中域名访问本服务" placement="top">
                <el-icon class="text-zinc-500 cursor-help"><QuestionFilled /></el-icon>
              </el-tooltip>
            </div>
            <el-input v-model="configForm.domain_binding" placeholder="例如: example.com" class="w-full" />
          </el-form-item>

          <div class="mt-4 flex justify-end">
            <el-button type="primary" :loading="configSaving" @click="handleConfigSave" class="!bg-white !text-black !border-none hover:!bg-zinc-200 px-8">
              保存面板设置
            </el-button>
          </div>
        </el-form>
      </div>
    </el-card>

    <!-- 安全设置 -->
    <el-card shadow="never" class="!bg-zinc-900 !border-zinc-800">
      <template #header>
        <div class="flex items-center gap-3">
          <el-icon class="text-white text-xl"><Lock /></el-icon>
          <span class="text-lg font-semibold text-white">安全设置</span>
        </div>
      </template>
      <div class="py-4">
        <el-form :model="configForm" label-position="top" v-loading="configLoading">
          <el-form-item>
            <div class="flex items-center gap-2">
              <span class="text-zinc-300">密码过期时间（天）</span>
              <el-tooltip content="为面板密码设置过期时间，过期后需要重新设置密码" placement="top">
                <el-icon class="text-zinc-500 cursor-help"><QuestionFilled /></el-icon>
              </el-tooltip>
            </div>
            <el-input-number v-model="configForm.password_expiry_days" :min="0" :max="365" class="w-full" />
            <span class="text-zinc-500 text-xs mt-1">0 表示永不过期</span>
          </el-form-item>

          <el-form-item>
            <div class="flex items-center gap-2">
              <span class="text-zinc-300">密码复杂度验证</span>
              <el-tooltip content="开启后密码必须满足长度为 8-30 位且包含字母、数字、特殊字符至少两项" placement="top">
                <el-icon class="text-zinc-500 cursor-help"><QuestionFilled /></el-icon>
              </el-tooltip>
            </div>
            <el-switch v-model="configForm.password_complexity" />
          </el-form-item>

          <el-form-item>
            <div class="flex items-center gap-2">
              <span class="text-zinc-300">两步验证</span>
              <el-tooltip content="开启后会验证 2FA" placement="top">
                <el-icon class="text-zinc-500 cursor-help"><QuestionFilled /></el-icon>
              </el-tooltip>
            </div>
            <el-switch v-model="configForm.two_factor_enabled" />
          </el-form-item>

          <el-form-item>
            <div class="flex items-center gap-2">
              <span class="text-zinc-300">通行密钥</span>
              <el-tooltip content="用于快速登录，最多可绑定 5 个通行密钥" placement="top">
                <el-icon class="text-zinc-500 cursor-help"><QuestionFilled /></el-icon>
              </el-tooltip>
            </div>
            <el-button @click="openPasskeyDialog" class="!bg-zinc-800 !text-white !border-zinc-700 hover:!bg-zinc-700">
              <el-icon class="mr-2"><Key /></el-icon> 管理通行密钥
            </el-button>
            <span class="text-zinc-500 text-xs ml-2">已绑定 {{ passkeys.length }} 个密钥</span>
          </el-form-item>

          <div class="mt-4 flex justify-end">
            <el-button type="primary" :loading="configSaving" @click="handleConfigSave" class="!bg-white !text-black !border-none hover:!bg-zinc-200 px-8">
              保存安全设置
            </el-button>
          </div>
        </el-form>
      </div>
    </el-card>

    <!-- 通行密钥管理对话框 -->
    <el-dialog v-model="passkeyDialogVisible" title="通行密钥管理" width="500px" custom-class="dark-dialog" :destroy-on-close="true">
      <div class="mb-4">
        <el-button type="primary" @click="handleAddPasskey" class="!bg-white !text-black !border-none hover:!bg-zinc-200">
          <el-icon class="mr-2"><Plus /></el-icon> 添加通行密钥
        </el-button>
        <span class="text-zinc-500 text-xs ml-2">最多 5 个</span>
      </div>
      <el-table :data="passkeys" v-loading="passkeyLoading" class="custom-dark-table">
        <el-table-column prop="name" label="名称" />
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">
            <span class="text-zinc-500 text-sm">{{ formatDate(row.created_at) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100">
          <template #default="{ row }">
            <el-button link type="danger" @click="handleDeletePasskey(row.id)">
              <el-icon><Delete /></el-icon>
            </el-button>
          </template>
        </el-table-column>
      </el-table>
      <template #footer>
        <el-button @click="passkeyDialogVisible = false" class="!bg-transparent !text-white !border-zinc-700">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 添加通行密钥对话框 -->
    <el-dialog v-model="addPasskeyDialogVisible" title="添加通行密钥" width="400px" custom-class="dark-dialog" :destroy-on-close="true">
      <el-form :model="newPasskeyForm" label-position="top">
        <el-form-item label="密钥名称">
          <el-input v-model="newPasskeyForm.name" placeholder="为这个密钥命名" />
        </el-form-item>
      </el-form>
      <div class="text-zinc-500 text-sm mb-4">
        <p>点击"生成密钥"后，系统将为您生成一个通行密钥。请妥善保存。</p>
      </div>
      <template #footer>
        <el-button @click="addPasskeyDialogVisible = false" class="!bg-transparent !text-white !border-zinc-700">取消</el-button>
        <el-button type="primary" :loading="addPasskeyLoading" @click="handleGeneratePasskey" class="!bg-white !text-black !border-none hover:!bg-zinc-200">
          生成密钥
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  User,
  Setting,
  Lock,
  Key,
  Plus,
  Delete,
  QuestionFilled
} from '@element-plus/icons-vue'
import {
  getProfile,
  getSystemConfig,
  updateSystemConfig,
  changeUsername,
  changePassword,
  getPasskeys,
  createPasskey,
  deletePasskey
} from '../api'
import { useRouter } from 'vue-router'
import { useUserStore } from '../store/user'

const router = useRouter()
const userStore = useUserStore()

// User form
const userLoading = ref(false)
const userSaving = ref(false)
const userForm = reactive({
  new_username: '',
  current_password: '',
  new_password: ''
})

// Config form
const configLoading = ref(false)
const configSaving = ref(false)
const configForm = reactive({
  panel_name: 'Godelion',
  session_timeout: 86400,
  enable_https: false,
  http_port: 8080,
  https_port: 443,
  secure_entrypoint: '',
  authorized_ips: '',
  domain_binding: '',
  password_expiry_days: 0,
  password_complexity: false,
  two_factor_enabled: false
})

// Passkeys
const passkeyDialogVisible = ref(false)
const addPasskeyDialogVisible = ref(false)
const passkeyLoading = ref(false)
const addPasskeyLoading = ref(false)
const passkeys = ref<any[]>([])
const newPasskeyForm = reactive({
  name: ''
})

// Fetch user profile
const fetchUserProfile = async () => {
  userLoading.value = true
  try {
    const res: any = await getProfile()
    if (res.code === 200) {
      userForm.new_username = res.data.username
    }
  } catch {
    ElMessage.error('获取用户信息失败')
  } finally {
    userLoading.value = false
  }
}

// Fetch system config
const fetchSystemConfig = async () => {
  configLoading.value = true
  try {
    const res: any = await getSystemConfig()
    if (res.code === 200 && res.data) {
      Object.assign(configForm, res.data)
    }
  } catch {
    ElMessage.error('获取系统配置失败')
  } finally {
    configLoading.value = false
  }
}

// Fetch passkeys
const fetchPasskeys = async () => {
  passkeyLoading.value = true
  try {
    const res: any = await getPasskeys()
    if (res.code === 200) {
      passkeys.value = res.data || []
    }
  } catch {
    ElMessage.error('获取通行密钥失败')
  } finally {
    passkeyLoading.value = false
  }
}

// Save user settings
const handleUserSave = async () => {
  if (userForm.new_password && !userForm.current_password) {
    ElMessage.warning('修改密码时必须输入当前密码')
    return
  }

  userSaving.value = true
  try {
    // Change username if different
    if (userForm.new_username !== userStore.user?.username) {
      const usernameRes: any = await changeUsername({ new_username: userForm.new_username })
      if (usernameRes.code !== 200) {
        ElMessage.error(usernameRes.message || '修改用户名失败')
        userSaving.value = false
        return
      }
    }

    // Change password if provided
    if (userForm.new_password) {
      const passwordRes: any = await changePassword({
        current_password: userForm.current_password,
        new_password: userForm.new_password
      })
      if (passwordRes.code !== 200) {
        ElMessage.error(passwordRes.message || '修改密码失败')
        userSaving.value = false
        return
      }
    }

    ElMessage.success('用户设置已保存')
    
    // If password changed, require re-login
    if (userForm.new_password) {
      ElMessage.info('密码已修改，请重新登录')
      setTimeout(() => {
        userStore.logout()
        router.push('/login')
      }, 1500)
    }

    userForm.current_password = ''
    userForm.new_password = ''
  } catch (error: any) {
    ElMessage.error(error.response?.data?.error || '保存失败')
  } finally {
    userSaving.value = false
  }
}

// Save config settings
const handleConfigSave = async () => {
  configSaving.value = true
  try {
    const res: any = await updateSystemConfig(configForm)
    if (res.code === 200) {
      ElMessage.success('设置已保存')
    } else {
      ElMessage.error(res.message || '保存失败')
    }
  } catch {
    ElMessage.error('保存异常')
  } finally {
    configSaving.value = false
  }
}

// Passkey management
const openPasskeyDialog = () => {
  passkeyDialogVisible.value = true
  fetchPasskeys()
}

const handleAddPasskey = () => {
  if (passkeys.value.length >= 5) {
    ElMessage.warning('最多只能绑定 5 个通行密钥')
    return
  }
  addPasskeyDialogVisible.value = true
  newPasskeyForm.name = ''
}

const handleGeneratePasskey = async () => {
  if (!newPasskeyForm.name) {
    ElMessage.warning('请输入密钥名称')
    return
  }

  addPasskeyLoading.value = true
  try {
    // Generate random credential (simplified for demo)
    const credentialId = Math.random().toString(36).substring(2, 15)
    const publicKey = Math.random().toString(36).substring(2, 30)
    
    const res: any = await createPasskey({
      name: newPasskeyForm.name,
      credential_id: credentialId,
      public_key: publicKey,
      counter: 0
    })
    if (res.code === 200) {
      ElMessage.success('通行密钥已创建')
      addPasskeyDialogVisible.value = false
      fetchPasskeys()
    } else {
      ElMessage.error(res.message || '创建失败')
    }
  } catch {
    ElMessage.error('创建异常')
  } finally {
    addPasskeyLoading.value = false
  }
}

const handleDeletePasskey = async (id: number) => {
  try {
    await ElMessageBox.confirm('确定要删除这个通行密钥吗？', '删除确认', {
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      type: 'warning'
    })
    const res: any = await deletePasskey(id)
    if (res.code === 200) {
      ElMessage.success('已删除')
      fetchPasskeys()
    } else {
      ElMessage.error(res.message || '删除失败')
    }
  } catch {}
}

const formatDate = (dateStr: string) => {
  if (!dateStr) return 'N/A'
  const date = new Date(dateStr)
  return date.toLocaleString()
}

onMounted(() => {
  fetchUserProfile()
  fetchSystemConfig()
  fetchPasskeys()
})
</script>

<style scoped>
:deep(.el-input-number) {
  width: 100%;
}
</style>