<template>
  <div class="min-h-full">
    <!-- 页面标题 -->
    <div class="px-8 py-6 border-b border-zinc-800">
      <h1 class="text-2xl font-bold text-white">设置</h1>
      <p class="text-zinc-500 text-sm mt-1">管理您的账户和系统配置</p>
    </div>

    <div class="p-8 max-w-6xl pb-16">
      <!-- 标签页导航 -->
      <div class="flex gap-1 mb-8 bg-zinc-900/50 p-1 rounded-xl w-fit">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          @click="switchTab(tab.key)"
          class="px-5 py-2.5 rounded-lg text-sm font-medium transition-all duration-200 flex items-center gap-2"
          :class="activeTab === tab.key ? 'bg-zinc-800 text-white shadow-sm' : 'text-zinc-500 hover:text-zinc-300'"
        >
          <el-icon :size="16"><User v-if="tab.icon === 'User'" /><Lock v-else-if="tab.icon === 'Lock'" /><Monitor v-else-if="tab.icon === 'Monitor'" /><Unlock v-else /></el-icon>
          {{ tab.label }}
        </button>
      </div>

      <!-- 用户信息 -->
      <div v-show="activeTab === 'user'" class="space-y-6">
        <div class="bg-zinc-900/50 border border-zinc-800 rounded-2xl p-6">
          <div class="flex items-center gap-4 mb-6">
            <div class="w-16 h-16 rounded-2xl bg-gradient-to-br from-blue-500/20 to-purple-500/20 flex items-center justify-center">
              <el-icon :size="28" class="text-blue-400"><User /></el-icon>
            </div>
            <div>
              <h3 class="text-lg font-semibold text-white">用户信息</h3>
              <p class="text-zinc-500 text-sm">修改您的用户名</p>
            </div>
          </div>

          <div class="max-w-md">
            <div>
              <label class="block text-sm font-medium text-zinc-400 mb-2">用户名</label>
              <el-input v-model="userForm.new_username" placeholder="输入新用户名" size="large" class="!bg-zinc-800/50" />
            </div>
            <div class="mt-4 p-4 bg-zinc-800/30 rounded-xl border border-zinc-800">
              <p class="text-sm text-zinc-500">
                <el-icon :size="14" class="text-amber-500 mr-1"><WarningFilled /></el-icon>
                修改用户名后需要重新登录
              </p>
            </div>
          </div>

          <div class="mt-6 pt-6 border-t border-zinc-800 flex justify-end">
            <el-button type="primary" :loading="userSaving" @click="handleUsernameSave" size="large" class="!bg-white !text-black !border-none hover:!bg-zinc-200 px-8">
              保存用户名
            </el-button>
          </div>
        </div>
      </div>

      <!-- 修改密码 -->
      <div v-show="activeTab === 'password'" class="space-y-6">
        <div class="bg-zinc-900/50 border border-zinc-800 rounded-2xl p-6">
          <div class="flex items-center gap-4 mb-6">
            <div class="w-16 h-16 rounded-2xl bg-gradient-to-br from-amber-500/20 to-red-500/20 flex items-center justify-center">
              <el-icon :size="28" class="text-amber-400"><Lock /></el-icon>
            </div>
            <div>
              <h3 class="text-lg font-semibold text-white">修改密码</h3>
              <p class="text-zinc-500 text-sm">更改您的登录密码</p>
            </div>
          </div>

          <div class="max-w-md space-y-4">
            <div>
              <label class="block text-sm font-medium text-zinc-400 mb-2">当前密码</label>
              <el-input v-model="userForm.current_password" type="password" placeholder="输入当前密码" show-password size="large" class="!bg-zinc-800/50" />
            </div>
            <div>
              <label class="block text-sm font-medium text-zinc-400 mb-2">新密码</label>
              <el-input v-model="userForm.new_password" type="password" placeholder="输入新密码" show-password size="large" class="!bg-zinc-800/50" />
            </div>
            <div class="p-4 bg-zinc-800/30 rounded-xl border border-zinc-800">
              <p class="text-sm text-zinc-500">
                <el-icon :size="14" class="text-amber-500 mr-1"><WarningFilled /></el-icon>
                修改密码后需要重新登录
              </p>
            </div>
          </div>

          <div class="mt-6 pt-6 border-t border-zinc-800 flex justify-end">
            <el-button type="primary" :loading="userSaving" @click="handlePasswordSave" size="large" class="!bg-white !text-black !border-none hover:!bg-zinc-200 px-8">
              保存密码
            </el-button>
          </div>
        </div>
      </div>

      <!-- 面板设置 -->
      <div v-show="activeTab === 'panel'" class="space-y-6">
        <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
          <!-- 基本信息 -->
          <div class="lg:col-span-2 bg-zinc-900/50 border border-zinc-800 rounded-2xl p-6">
            <h3 class="text-lg font-semibold text-white mb-6 flex items-center gap-2">
              <el-icon :size="18" class="text-purple-400"><Monitor /></el-icon>
              基本信息
            </h3>
            <div class="space-y-5">
              <div>
                <label class="block text-sm font-medium text-zinc-400 mb-2">面板名称</label>
                <el-input v-model="configForm.panel_name" placeholder="Godelion" size="large" class="!bg-zinc-800/50" />
              </div>
              <div class="grid grid-cols-2 gap-4">
                <div>
                  <label class="block text-sm font-medium text-zinc-400 mb-2">HTTP 端口</label>
                  <el-input-number v-model="configForm.http_port" :min="1" :max="65535" size="large" class="w-full !bg-zinc-800/50" controls-position="right" />
                </div>
                <div>
                  <label class="block text-sm font-medium text-zinc-400 mb-2">HTTPS 端口</label>
                  <el-input-number v-model="configForm.https_port" :min="1" :max="65535" size="large" class="w-full !bg-zinc-800/50" controls-position="right" />
                </div>
              </div>
            </div>
          </div>

          <!-- HTTPS 开关 -->
          <div class="bg-zinc-900/50 border border-zinc-800 rounded-2xl p-6">
            <div class="flex items-center justify-between mb-4">
              <h3 class="text-lg font-semibold text-white">HTTPS</h3>
              <el-switch v-model="configForm.enable_https" size="large" />
            </div>
            <p class="text-sm text-zinc-500 mb-4">启用 HTTPS 加密传输，提升访问安全性</p>
            <div class="flex items-center gap-2 text-xs text-zinc-600 bg-zinc-800/50 rounded-lg p-3">
              <el-icon :size="14"><InfoFilled /></el-icon>
              需要配置 SSL 证书才能正常使用
            </div>
          </div>
        </div>

        <!-- 访问控制 -->
        <div class="bg-zinc-900/50 border border-zinc-800 rounded-2xl p-6">
          <h3 class="text-lg font-semibold text-white mb-6 flex items-center gap-2">
            <el-icon :size="18" class="text-green-400"><Unlock /></el-icon>
            访问控制
          </h3>
          <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
            <div>
              <div class="flex items-center gap-2 mb-2">
                <label class="text-sm font-medium text-zinc-400">超时时间</label>
                <el-tooltip content="超过设定时间未操作将自动退出登录" placement="top">
                  <el-icon :size="14" class="text-zinc-600 cursor-help"><QuestionFilled /></el-icon>
                </el-tooltip>
              </div>
              <el-input-number v-model="configForm.session_timeout" :min="60" :max="864000" size="large" class="w-full !bg-zinc-800/50" controls-position="right">
                <template #suffix>秒</template>
              </el-input-number>
              <p class="text-xs text-zinc-600 mt-1">默认 86400 秒（24小时）</p>
            </div>
            <div>
              <div class="flex items-center gap-2 mb-2">
                <label class="text-sm font-medium text-zinc-400">安全入口</label>
                <el-tooltip content="设置后只能通过指定路径访问面板" placement="top">
                  <el-icon :size="14" class="text-zinc-600 cursor-help"><QuestionFilled /></el-icon>
                </el-tooltip>
              </div>
              <el-input v-model="configForm.secure_entrypoint" placeholder="例如: /admin" size="large" class="!bg-zinc-800/50" />
            </div>
            <div>
              <div class="flex items-center gap-2 mb-2">
                <label class="text-sm font-medium text-zinc-400">域名绑定</label>
                <el-tooltip content="设置后只能通过指定域名访问" placement="top">
                  <el-icon :size="14" class="text-zinc-600 cursor-help"><QuestionFilled /></el-icon>
                </el-tooltip>
              </div>
              <el-input v-model="configForm.domain_binding" placeholder="例如: panel.example.com" size="large" class="!bg-zinc-800/50" />
            </div>
          </div>
          <div class="mt-5">
            <div class="flex items-center gap-2 mb-2">
              <label class="text-sm font-medium text-zinc-400">授权 IP 白名单</label>
              <el-tooltip content="设置后只有指定 IP 可以访问面板，多个 IP 用逗号分隔" placement="top">
                <el-icon :size="14" class="text-zinc-600 cursor-help"><QuestionFilled /></el-icon>
              </el-tooltip>
            </div>
            <el-input v-model="configForm.authorized_ips" placeholder="192.168.1.100, 10.0.0.0/8, 172.16.0.1" size="large" class="!bg-zinc-800/50" />
            <p class="text-xs text-zinc-600 mt-1">留空表示允许所有 IP 访问</p>
          </div>
        </div>

        <div class="flex justify-end">
          <el-button type="primary" :loading="configSaving" @click="handleConfigSave" size="large" class="!bg-white !text-black !border-none hover:!bg-zinc-200 px-8">
            保存面板设置
          </el-button>
        </div>
      </div>

      <!-- 安全设置 -->
      <div v-show="activeTab === 'security'" class="space-y-6">
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <!-- 密码策略 -->
          <div class="bg-zinc-900/50 border border-zinc-800 rounded-2xl p-6">
            <h3 class="text-lg font-semibold text-white mb-6 flex items-center gap-2">
              <el-icon :size="18" class="text-amber-400"><Lock /></el-icon>
              密码策略
            </h3>
            <div class="space-y-5">
              <div class="flex items-center justify-between py-3 border-b border-zinc-800">
                <div>
                  <p class="text-white font-medium">密码复杂度验证</p>
                  <p class="text-sm text-zinc-500">密码必须包含字母、数字、特殊字符至少两项</p>
                </div>
                <el-switch v-model="configForm.password_complexity" size="large" @change="handlePasswordComplexityChange" />
              </div>
              <div>
                <div class="flex items-center gap-2 mb-3">
                  <label class="text-sm font-medium text-zinc-400">密码过期时间</label>
                  <el-tooltip content="设置密码有效期，过期后需重新设置" placement="top">
                    <el-icon :size="14" class="text-zinc-600 cursor-help"><QuestionFilled /></el-icon>
                  </el-tooltip>
                </div>
                <div class="flex items-center gap-4">
                  <el-slider v-model="configForm.password_expiry_days" :max="90" :step="1" class="flex-1" />
                  <span class="text-white font-mono w-16 text-right">{{ configForm.password_expiry_days }} 天</span>
                </div>
                <p class="text-xs text-zinc-600 mt-1">0 表示永不过期</p>
              </div>
              <div class="pt-2 flex justify-end">
                <el-button type="primary" :loading="configSaving" @click="handleConfigSave" size="large" class="!bg-white !text-black !border-none hover:!bg-zinc-200 px-6">
                  保存密码策略
                </el-button>
              </div>
            </div>
          </div>

          <!-- 双重认证 -->
          <div class="bg-zinc-900/50 border border-zinc-800 rounded-2xl p-6">
            <h3 class="text-lg font-semibold text-white mb-6 flex items-center gap-2">
              <el-icon :size="18" class="text-cyan-400"><CircleCheck /></el-icon>
              双重认证
            </h3>
            <div class="space-y-5">
              <!-- 2FA 状态 -->
              <div class="flex items-center justify-between py-3 border-b border-zinc-800">
                <div>
                  <p class="text-white font-medium">两步验证 (2FA)</p>
                  <p class="text-sm text-zinc-500">
                    <span v-if="twoFAEnabled" class="text-green-400">已开启</span>
                    <span v-else>未开启</span>
                    — 登录时需要输入动态验证码
                  </p>
                </div>
                <div class="flex items-center gap-3">
                  <el-button v-if="!twoFAEnabled" @click="open2FADialog" type="primary" class="!bg-cyan-500 !text-white !border-none hover:!bg-cyan-600">
                    开启
                  </el-button>
                  <el-button v-else @click="open2FADisableDialog" class="!bg-red-500/10 !text-red-400 !border-red-500/30 hover:!bg-red-500/20">
                    关闭
                  </el-button>
                </div>
              </div>


            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 2FA 开启对话框 -->
    <el-dialog v-model="twoFADialogVisible" title="开启两步验证" width="460px" custom-class="dark-dialog" :destroy-on-close="true" :close-on-click-modal="false">
      <div v-if="!twoFAQRCode" class="text-center py-4">
        <el-button type="primary" :loading="twoFALoading" @click="handleGenerate2FA" class="!bg-cyan-500 !text-white !border-none hover:!bg-cyan-600">
          生成二维码
        </el-button>
      </div>
      <div v-else>
        <div class="text-center mb-6">
          <p class="text-white font-medium mb-4">使用 Google Authenticator 扫描二维码</p>
          <img :src="twoFAQRCode" alt="2FA QR Code" class="mx-auto rounded-xl border border-zinc-700" style="width: 200px; height: 200px;" />
        </div>
        <div class="bg-zinc-800/50 rounded-xl p-4 mb-4">
          <div class="flex items-center justify-between mb-1">
            <p class="text-sm text-zinc-400">手动输入密钥</p>
            <el-button link size="small" @click="copySecret" class="!text-zinc-400 hover:!text-white">
              <el-icon class="mr-1"><CopyDocument /></el-icon>复制
            </el-button>
          </div>
          <code class="text-white text-sm font-mono break-all">{{ twoFASecret }}</code>
        </div>
        <div>
          <label class="block text-sm font-medium text-zinc-400 mb-2">输入验证码以确认开启</label>
          <OTPInput v-model="twoFAVerifyCode" />
        </div>
      </div>
      <template #footer>
        <el-button @click="twoFADialogVisible = false" class="!bg-transparent !text-white !border-zinc-700">取消</el-button>
        <el-button v-if="twoFAQRCode" type="primary" :loading="twoFAVerifying" @click="handleVerify2FA" class="!bg-cyan-500 !text-white !border-none hover:!bg-cyan-600">
          确认开启
        </el-button>
      </template>
    </el-dialog>

    <!-- 2FA 关闭对话框 -->
    <el-dialog v-model="twoFADisableDialogVisible" title="关闭两步验证" width="400px" custom-class="dark-dialog" :destroy-on-close="true" :close-on-click-modal="false">
      <div class="text-center py-4">
        <div class="w-16 h-16 rounded-2xl bg-red-500/10 flex items-center justify-center mx-auto mb-4">
          <el-icon :size="28" class="text-red-400"><WarningFilled /></el-icon>
        </div>
        <p class="text-white font-medium mb-2">确定要关闭两步验证吗？</p>
        <p class="text-sm text-zinc-500 mb-4">关闭后账户安全性将降低</p>

        <!-- 验证方式切换 -->
        <div class="flex gap-2 mb-4 justify-center">
          <el-button
            :type="disable2FAMethod === 'code' ? 'primary' : 'default'"
            size="small"
            @click="disable2FAMethod = 'code'"
            class="!rounded-lg"
          >
            验证码验证
          </el-button>
          <el-button
            :type="disable2FAMethod === 'password' ? 'primary' : 'default'"
            size="small"
            @click="disable2FAMethod = 'password'"
            class="!rounded-lg"
          >
            密码验证
          </el-button>
        </div>

        <div v-if="disable2FAMethod === 'code'">
          <label class="block text-sm font-medium text-zinc-400 mb-2">输入当前验证码以确认关闭</label>
          <OTPInput v-model="twoFADisableCode" />
        </div>
        <div v-else>
          <label class="block text-sm font-medium text-zinc-400 mb-2">输入登录密码以确认关闭</label>
          <el-input v-model="twoFADisablePassword" type="password" placeholder="请输入登录密码" size="large" show-password class="!bg-zinc-800/50" />
        </div>
      </div>
      <template #footer>
        <el-button @click="twoFADisableDialogVisible = false" class="!bg-transparent !text-white !border-zinc-700">取消</el-button>
        <el-button type="danger" :loading="twoFADisabling" @click="handleDisable2FA" class="!bg-red-500 !text-white !border-none hover:!bg-red-600">
          确认关闭
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
  Monitor,
  Lock,
  QuestionFilled,
  Unlock,
  CircleCheck,
  InfoFilled,
  WarningFilled,
  CopyDocument
} from '@element-plus/icons-vue'
import {
  getProfile,
  getSystemConfig,
  updateSystemConfig,
  changeUsername,
  changePassword,
  get2FAStatus,
  generate2FA,
  verify2FA,
  disable2FA
} from '../api'
import { useRouter, useRoute } from 'vue-router'
import { useUserStore } from '../store/user'
import OTPInput from '../components/OTPInput.vue'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()

const activeTab = ref('user')
const tabs = [
  { key: 'user', label: '用户信息', icon: 'User', path: '/settings' },
  { key: 'password', label: '修改密码', icon: 'Lock', path: '/settings/password' },
  { key: 'panel', label: '面板设置', icon: 'Monitor', path: '/settings/panel' },
  { key: 'security', label: '安全设置', icon: 'Unlock', path: '/settings/security' }
]

// Sync tab with route
const switchTab = (key: string) => {
  activeTab.value = key
  const tab = tabs.find(t => t.key === key)
  if (tab && route.path !== tab.path) {
    router.replace(tab.path)
  }
}

// Initialize tab from route
const initTabFromRoute = () => {
  const path = route.path
  if (path === '/settings/password') activeTab.value = 'password'
  else if (path === '/settings/panel') activeTab.value = 'panel'
  else if (path === '/settings/security') activeTab.value = 'security'
  else activeTab.value = 'user'
}

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

const fetchUserProfile = async () => {
  userLoading.value = true
  try {
    const res: any = await getProfile()
    if (res.code === 200) {
      userForm.new_username = res.data.username
    }
  } catch {
  } finally {
    userLoading.value = false
  }
}

const fetchSystemConfig = async () => {
  configLoading.value = true
  try {
    const res: any = await getSystemConfig()
    if (res.code === 200 && res.data) {
      Object.assign(configForm, res.data)
    }
  } catch {
  } finally {
    configLoading.value = false
  }
}

const handleUsernameSave = async () => {
  if (!userForm.new_username) {
    ElMessage.warning('请输入用户名')
    return
  }
  if (userForm.new_username === userStore.user?.username) {
    ElMessage.info('用户名未修改')
    return
  }

  userSaving.value = true
  try {
    const usernameRes: any = await changeUsername({ new_username: userForm.new_username })
    if (usernameRes.code === 200) {
      ElMessage.success('用户名已修改')
      ElMessage.info('请重新登录')
      setTimeout(() => {
        userStore.logout()
        router.push('/login')
      }, 1500)
    }
  } catch {
  } finally {
    userSaving.value = false
  }
}

const handlePasswordSave = async () => {
  if (!userForm.current_password) {
    ElMessage.warning('请输入当前密码')
    return
  }
  if (!userForm.new_password) {
    ElMessage.warning('请输入新密码')
    return
  }

  // Frontend password complexity check (if enabled)
  if (configForm.password_complexity) {
    if (userForm.new_password.length < 8) {
      ElMessage.warning('密码长度至少为 8 位')
      return
    }
    const hasLetter = /[a-zA-Z]/.test(userForm.new_password)
    const hasDigit = /[0-9]/.test(userForm.new_password)
    const hasSpecial = /[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]/.test(userForm.new_password)
    const typeCount = [hasLetter, hasDigit, hasSpecial].filter(Boolean).length
    if (typeCount < 2) {
      ElMessage.warning('密码必须包含字母、数字、特殊字符至少两项')
      return
    }
  }

  userSaving.value = true
  try {
    const passwordRes: any = await changePassword({
      current_password: userForm.current_password,
      new_password: userForm.new_password
    })
    if (passwordRes.code === 200) {
      ElMessage.success('密码已修改')
      ElMessage.info('请重新登录')
      setTimeout(() => {
        userStore.logout()
        router.push('/login')
      }, 1500)
      userForm.current_password = ''
      userForm.new_password = ''
    }
  } catch {
  } finally {
    userSaving.value = false
  }
}

const handleConfigSave = async () => {
  configSaving.value = true
  try {
    const res: any = await updateSystemConfig(configForm)
    if (res.code === 200) {
      ElMessage.success('设置已保存')
      fetchSystemConfig()
      fetch2FAStatus()
    }
  } catch {
  } finally {
    configSaving.value = false
  }
}

const handlePasswordComplexityChange = async (val: boolean) => {
  try {
    const res: any = await updateSystemConfig({ password_complexity: val })
    if (res.code === 200) {
      ElMessage.success(val ? '密码复杂度验证已开启' : '密码复杂度验证已关闭')
    } else {
      fetchSystemConfig()
    }
  } catch {
    fetchSystemConfig()
  }
}

const formatDate = (dateStr: string) => {
  if (!dateStr) return 'N/A'
  const date = new Date(dateStr)
  return date.toLocaleDateString()
}

// 2FA
const twoFAEnabled = ref(false)
const twoFADialogVisible = ref(false)
const twoFADisableDialogVisible = ref(false)
const twoFALoading = ref(false)
const twoFAVerifying = ref(false)
const twoFADisabling = ref(false)
const twoFADisableCode = ref('')
const twoFADisablePassword = ref('')
const disable2FAMethod = ref<'code' | 'password'>('code')
const twoFASecret = ref('')
const twoFAVerifyCode = ref('')

const fetch2FAStatus = async () => {
  try {
    const res: any = await get2FAStatus()
    if (res.code === 200) {
      twoFAEnabled.value = res.data.enabled
    }
  } catch {}
}

const open2FADialog = () => {
  twoFADialogVisible.value = true
  twoFAQRCode.value = ''
  twoFASecret.value = ''
  twoFAVerifyCode.value = ''
  // Auto generate QR code
  handleGenerate2FA()
}

const copySecret = () => {
  navigator.clipboard.writeText(twoFASecret.value).then(() => {
    ElMessage.success('密钥已复制')
  }).catch(() => {
    ElMessage.error('复制失败')
  })
}

const open2FADisableDialog = () => {
  twoFADisableDialogVisible.value = true
  twoFADisableCode.value = ''
}

const handleGenerate2FA = async () => {
  twoFALoading.value = true
  try {
    const res: any = await generate2FA()
    if (res.code === 200) {
      twoFAQRCode.value = res.data.qr_code
      twoFASecret.value = res.data.secret
    }
  } catch {
  } finally {
    twoFALoading.value = false
  }
}

const handleVerify2FA = async () => {
  if (!twoFAVerifyCode.value || twoFAVerifyCode.value.length !== 6) {
    ElMessage.warning('请输入 6 位验证码')
    return
  }
  twoFAVerifying.value = true
  try {
    const res: any = await verify2FA({ code: twoFAVerifyCode.value })
    if (res.code === 200) {
      ElMessage.success('两步验证已开启')
      twoFADialogVisible.value = false
      twoFAEnabled.value = true
    }
  } catch {
  } finally {
    twoFAVerifying.value = false
  }
}

const handleDisable2FA = async () => {
  if (disable2FAMethod.value === 'code') {
    if (!twoFADisableCode.value || twoFADisableCode.value.length !== 6) {
      ElMessage.warning('请输入 6 位验证码')
      return
    }
  } else {
    if (!twoFADisablePassword.value) {
      ElMessage.warning('请输入登录密码')
      return
    }
  }

  twoFADisabling.value = true
  try {
    const payload: any = {
      method: disable2FAMethod.value,
    }
    if (disable2FAMethod.value === 'code') {
      payload.code = twoFADisableCode.value
    } else {
      payload.password = twoFADisablePassword.value
    }
    const res: any = await disable2FA(payload)
    if (res.code === 200) {
      ElMessage.success('两步验证已关闭')
      twoFADisableDialogVisible.value = false
      twoFAEnabled.value = false
    }
  } catch {
  } finally {
    twoFADisabling.value = false
  }
}

onMounted(() => {
  initTabFromRoute()
  fetchUserProfile()
  fetchSystemConfig()
  fetch2FAStatus()
})
</script>

<style scoped>
:deep(.el-input__wrapper) {
  background-color: rgba(39, 39, 42, 0.5);
  box-shadow: none;
  border: 1px solid rgba(63, 63, 70, 0.5);
}
:deep(.el-input__wrapper:hover) {
  border-color: rgba(82, 82, 91, 0.8);
}
:deep(.el-input__wrapper.is-focus) {
  border-color: #3b82f6;
}
:deep(.el-input__inner) {
  color: white;
}
:deep(.el-slider__runway) {
  background-color: rgba(63, 63, 70, 0.5);
}
:deep(.el-slider__bar) {
  background-color: #3b82f6;
}
:deep(.el-slider__button) {
  border-color: #3b82f6;
  background-color: white;
}
</style>