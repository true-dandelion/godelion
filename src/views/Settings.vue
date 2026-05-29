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
          @click="activeTab = tab.key"
          class="px-5 py-2.5 rounded-lg text-sm font-medium transition-all duration-200 flex items-center gap-2"
          :class="activeTab === tab.key ? 'bg-zinc-800 text-white shadow-sm' : 'text-zinc-500 hover:text-zinc-300'"
        >
          <el-icon :size="16"><User v-if="tab.icon === 'User'" /><Monitor v-else-if="tab.icon === 'Monitor'" /><Lock v-else /></el-icon>
          {{ tab.label }}
        </button>
      </div>

      <!-- 用户设置 -->
      <div v-show="activeTab === 'user'" class="space-y-6">
        <!-- 账户信息卡片 -->
        <div class="bg-zinc-900/50 border border-zinc-800 rounded-2xl p-6">
          <div class="flex items-center gap-4 mb-6">
            <div class="w-16 h-16 rounded-2xl bg-gradient-to-br from-blue-500/20 to-purple-500/20 flex items-center justify-center">
              <el-icon :size="28" class="text-blue-400"><User /></el-icon>
            </div>
            <div>
              <h3 class="text-lg font-semibold text-white">账户信息</h3>
              <p class="text-zinc-500 text-sm">管理您的登录凭据</p>
            </div>
          </div>

          <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <div class="space-y-4">
              <div>
                <label class="block text-sm font-medium text-zinc-400 mb-2">用户名</label>
                <el-input v-model="userForm.new_username" placeholder="输入新用户名" size="large" class="!bg-zinc-800/50" />
              </div>
              <div>
                <label class="block text-sm font-medium text-zinc-400 mb-2">当前密码</label>
                <el-input v-model="userForm.current_password" type="password" placeholder="输入当前密码" show-password size="large" class="!bg-zinc-800/50" />
              </div>
              <div>
                <label class="block text-sm font-medium text-zinc-400 mb-2">新密码</label>
                <el-input v-model="userForm.new_password" type="password" placeholder="输入新密码" show-password size="large" class="!bg-zinc-800/50" />
              </div>
            </div>
            <div class="bg-zinc-800/30 rounded-xl p-5 border border-zinc-800">
              <h4 class="text-sm font-medium text-zinc-300 mb-3 flex items-center gap-2">
                <el-icon :size="14" class="text-amber-500"><WarningFilled /></el-icon>
                安全提示
              </h4>
              <ul class="space-y-2 text-sm text-zinc-500">
                <li class="flex items-start gap-2">
                  <span class="text-zinc-600">•</span>
                  修改用户名后需要重新登录
                </li>
                <li class="flex items-start gap-2">
                  <span class="text-zinc-600">•</span>
                  建议使用包含大小写字母、数字的强密码
                </li>
                <li class="flex items-start gap-2">
                  <span class="text-zinc-600">•</span>
                  定期更换密码可提高账户安全性
                </li>
              </ul>
            </div>
          </div>

          <div class="mt-6 pt-6 border-t border-zinc-800 flex justify-end">
            <el-button type="primary" :loading="userSaving" @click="handleUserSave" size="large" class="!bg-white !text-black !border-none hover:!bg-zinc-200 px-8">
              保存更改
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
                <el-switch v-model="configForm.password_complexity" size="large" />
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

              <!-- 通行密钥 -->
              <div>
                <div class="flex items-center justify-between mb-3">
                  <div class="flex items-center gap-2">
                    <label class="text-sm font-medium text-zinc-400">通行密钥</label>
                    <el-tooltip content="用于快速登录，最多绑定 5 个" placement="top">
                      <el-icon :size="14" class="text-zinc-600 cursor-help"><QuestionFilled /></el-icon>
                    </el-tooltip>
                  </div>
                  <span class="text-xs text-zinc-500">{{ passkeys.length }}/5</span>
                </div>
                <div class="flex items-center gap-3">
                  <el-button @click="openPasskeyDialog" class="!bg-zinc-800 !text-white !border-zinc-700 hover:!bg-zinc-700">
                    <el-icon class="mr-2"><Key /></el-icon>
                    管理密钥
                  </el-button>
                  <span v-if="passkeys.length > 0" class="text-sm text-zinc-500">已绑定 {{ passkeys.length }} 个密钥</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="flex justify-end">
          <el-button type="primary" :loading="configSaving" @click="handleConfigSave" size="large" class="!bg-white !text-black !border-none hover:!bg-zinc-200 px-8">
            保存安全设置
          </el-button>
        </div>
      </div>
    </div>

    <!-- 2FA 开启对话框 -->
    <el-dialog v-model="twoFADialogVisible" title="开启两步验证" width="460px" custom-class="dark-dialog" :destroy-on-close="true">
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
          <p class="text-sm text-zinc-400 mb-1">手动输入密钥</p>
          <code class="text-white text-sm font-mono break-all">{{ twoFASecret }}</code>
        </div>
        <div>
          <label class="block text-sm font-medium text-zinc-400 mb-2">输入验证码以确认开启</label>
          <el-input v-model="twoFAVerifyCode" placeholder="请输入 6 位验证码" maxlength="6" size="large" class="!bg-zinc-800/50" />
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
    <el-dialog v-model="twoFADisableDialogVisible" title="关闭两步验证" width="400px" custom-class="dark-dialog" :destroy-on-close="true">
      <div class="text-center py-4">
        <div class="w-16 h-16 rounded-2xl bg-red-500/10 flex items-center justify-center mx-auto mb-4">
          <el-icon :size="28" class="text-red-400"><WarningFilled /></el-icon>
        </div>
        <p class="text-white font-medium mb-2">确定要关闭两步验证吗？</p>
        <p class="text-sm text-zinc-500 mb-4">关闭后账户安全性将降低</p>
        <div>
          <label class="block text-sm font-medium text-zinc-400 mb-2">输入当前验证码以确认关闭</label>
          <el-input v-model="twoFADisableCode" placeholder="请输入 6 位验证码" maxlength="6" size="large" class="!bg-zinc-800/50" />
        </div>
      </div>
      <template #footer>
        <el-button @click="twoFADisableDialogVisible = false" class="!bg-transparent !text-white !border-zinc-700">取消</el-button>
        <el-button type="danger" :loading="twoFADisabling" @click="handleDisable2FA" class="!bg-red-500 !text-white !border-none hover:!bg-red-600">
          确认关闭
        </el-button>
      </template>
    </el-dialog>

    <!-- 通行密钥管理对话框 -->
    <el-dialog v-model="passkeyDialogVisible" title="通行密钥管理" width="520px" custom-class="dark-dialog" :destroy-on-close="true">
      <div class="mb-4 flex items-center justify-between">
        <div>
          <p class="text-white font-medium">我的通行密钥</p>
          <p class="text-sm text-zinc-500">使用通行密钥可以快速安全地登录</p>
        </div>
        <el-button type="primary" @click="handleAddPasskey" :disabled="passkeys.length >= 5" class="!bg-white !text-black !border-none hover:!bg-zinc-200">
          <el-icon class="mr-2"><Plus /></el-icon>
          添加密钥
        </el-button>
      </div>
      <div v-if="passkeys.length === 0" class="text-center py-8 text-zinc-500">
        <el-icon :size="48" class="mb-3 opacity-30"><Key /></el-icon>
        <p>暂无通行密钥</p>
      </div>
      <div v-else class="space-y-2">
        <div v-for="key in passkeys" :key="key.id" class="flex items-center justify-between p-4 bg-zinc-800/50 rounded-xl border border-zinc-800">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-lg bg-zinc-800 flex items-center justify-center">
              <el-icon :size="20" class="text-zinc-400"><Key /></el-icon>
            </div>
            <div>
              <p class="text-white font-medium">{{ key.name }}</p>
              <p class="text-xs text-zinc-500">创建于 {{ formatDate(key.created_at) }}</p>
            </div>
          </div>
          <el-button link type="danger" @click="handleDeletePasskey(key.id)">
            <el-icon><Delete /></el-icon>
          </el-button>
        </div>
      </div>
    </el-dialog>

    <!-- 添加通行密钥对话框 -->
    <el-dialog v-model="addPasskeyDialogVisible" title="添加通行密钥" width="400px" custom-class="dark-dialog" :destroy-on-close="true">
      <div class="text-center py-4">
        <div class="w-16 h-16 rounded-2xl bg-gradient-to-br from-blue-500/20 to-purple-500/20 flex items-center justify-center mx-auto mb-4">
          <el-icon :size="28" class="text-blue-400"><Key /></el-icon>
        </div>
        <h4 class="text-white font-medium mb-2">创建新的通行密钥</h4>
        <p class="text-sm text-zinc-500 mb-6">为您的账户添加一个安全的登录方式</p>
      </div>
      <el-form :model="newPasskeyForm" label-position="top">
        <el-form-item label="密钥名称">
          <el-input v-model="newPasskeyForm.name" placeholder="例如：我的 MacBook" size="large" class="!bg-zinc-800/50" />
        </el-form-item>
      </el-form>
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
  Monitor,
  Lock,
  Key,
  Plus,
  Delete,
  QuestionFilled,
  Unlock,
  CircleCheck,
  InfoFilled,
  WarningFilled
} from '@element-plus/icons-vue'
import {
  getProfile,
  getSystemConfig,
  updateSystemConfig,
  changeUsername,
  changePassword,
  getPasskeys,
  createPasskey,
  deletePasskey,
  get2FAStatus,
  generate2FA,
  verify2FA,
  disable2FA
} from '../api'
import { useRouter } from 'vue-router'
import { useUserStore } from '../store/user'

const router = useRouter()
const userStore = useUserStore()

const activeTab = ref('user')
const tabs = [
  { key: 'user', label: '用户设置', icon: 'User' },
  { key: 'panel', label: '面板设置', icon: 'Monitor' },
  { key: 'security', label: '安全设置', icon: 'Lock' }
]

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

const handleUserSave = async () => {
  if (!userForm.new_password) {
    ElMessage.warning('请输入新密码')
    return
  }
  if (!userForm.current_password) {
    ElMessage.warning('请输入当前密码')
    return
  }

  userSaving.value = true
  try {
    if (userForm.new_username !== userStore.user?.username) {
      const usernameRes: any = await changeUsername({ new_username: userForm.new_username })
      if (usernameRes.code !== 200) {
        ElMessage.error(usernameRes.message || '修改用户名失败')
        userSaving.value = false
        return
      }
    }

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
  return date.toLocaleDateString()
}

// 2FA
const twoFAEnabled = ref(false)
const twoFADialogVisible = ref(false)
const twoFADisableDialogVisible = ref(false)
const twoFALoading = ref(false)
const twoFAVerifying = ref(false)
const twoFADisabling = ref(false)
const twoFAQRCode = ref('')
const twoFASecret = ref('')
const twoFAVerifyCode = ref('')
const twoFADisableCode = ref('')

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
    } else {
      ElMessage.error(res.message || '生成失败')
    }
  } catch {
    ElMessage.error('生成异常')
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
    } else {
      ElMessage.error(res.message || res.error || '验证失败')
    }
  } catch (error: any) {
    ElMessage.error(error.response?.data?.error || '验证异常')
  } finally {
    twoFAVerifying.value = false
  }
}

const handleDisable2FA = async () => {
  if (!twoFADisableCode.value || twoFADisableCode.value.length !== 6) {
    ElMessage.warning('请输入 6 位验证码')
    return
  }
  twoFADisabling.value = true
  try {
    const res: any = await disable2FA({ code: twoFADisableCode.value })
    if (res.code === 200) {
      ElMessage.success('两步验证已关闭')
      twoFADisableDialogVisible.value = false
      twoFAEnabled.value = false
    } else {
      ElMessage.error(res.message || res.error || '关闭失败')
    }
  } catch (error: any) {
    ElMessage.error(error.response?.data?.error || '关闭异常')
  } finally {
    twoFADisabling.value = false
  }
}

onMounted(() => {
  fetchUserProfile()
  fetchSystemConfig()
  fetchPasskeys()
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