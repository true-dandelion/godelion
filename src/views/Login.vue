<template>
  <div class="min-h-screen bg-black flex items-center justify-center relative overflow-hidden">
    <!-- 背景装饰 -->
    <div class="absolute inset-0 pointer-events-none">
      <div class="absolute top-1/4 left-1/4 w-96 h-96 bg-blue-900/20 rounded-full blur-[100px]"></div>
      <div class="absolute bottom-1/4 right-1/4 w-96 h-96 bg-indigo-900/20 rounded-full blur-[100px]"></div>
      <div class="absolute inset-0 bg-[url('data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSI0IiBoZWlnaHQ9IjQiPgo8cmVjdCB3aWR0aD0iNCIgaGVpZ2h0PSI0IiBmaWxsPSIjZmZmIiBmaWxsLW9wYWNpdHk9IjAuMDUiLz4KPC9zdmc+')] opacity-20"></div>
    </div>

    <!-- 登录面板 -->
    <div class="w-full max-w-md p-8 relative z-10">
      <div class="backdrop-blur-xl bg-zinc-900/80 border border-zinc-800 rounded-2xl shadow-2xl p-8">
        <div class="text-center mb-8">
          <div class="inline-flex items-center justify-center w-16 h-16 rounded-xl bg-gradient-to-br from-zinc-800 to-zinc-900 border border-zinc-700 shadow-inner mb-4">
            <el-icon :size="32" class="text-white"><Platform /></el-icon>
          </div>
          <h1 class="text-3xl font-bold text-white tracking-tight">Godelion</h1>
          <p class="text-zinc-400 mt-2 text-sm">企业级安全管控与中继网关</p>
        </div>

        <!-- 登录表单 -->
        <el-form 
          v-if="!require2FA"
          ref="loginFormRef"
          :model="loginForm"
          :rules="loginRules"
          @keyup.enter="handleLogin"
          class="space-y-6"
        >
          <el-form-item prop="username" class="mb-4">
            <el-input 
              v-model="loginForm.username" 
              placeholder="管理员账号"
              :prefix-icon="User"
              size="large"
              class="!bg-transparent"
            />
          </el-form-item>
          
          <el-form-item prop="password" class="mb-6">
            <el-input 
              v-model="loginForm.password" 
              type="password" 
              placeholder="登录密码"
              :prefix-icon="Lock"
              show-password
              size="large"
            />
          </el-form-item>

          <el-button 
            type="primary" 
            class="w-full h-12 text-base font-medium !bg-white !text-black hover:!bg-zinc-200 border-none rounded-lg transition-colors"
            :loading="loading"
            @click="handleLogin"
          >
            进入控制台
          </el-button>
        </el-form>

        <!-- 2FA 验证码输入 -->
        <div v-else class="space-y-6">
          <div class="text-center">
            <div class="inline-flex items-center justify-center w-14 h-14 rounded-xl bg-cyan-500/10 border border-cyan-500/20 mb-4">
              <el-icon :size="28" class="text-cyan-400"><Key /></el-icon>
            </div>
            <h2 class="text-xl font-semibold text-white">两步验证</h2>
            <p class="text-zinc-500 text-sm mt-2">请输入 Google Authenticator 中的验证码</p>
          </div>

          <el-input
            v-model="twoFACode"
            placeholder="请输入 6 位验证码"
            size="large"
            maxlength="6"
            class="!bg-transparent text-center text-2xl tracking-[0.5em] font-mono"
            @keyup.enter="handleVerify2FA"
          />

          <el-button 
            type="primary" 
            class="w-full h-12 text-base font-medium !bg-cyan-500 !text-white hover:!bg-cyan-600 border-none rounded-lg transition-colors"
            :loading="verifying2FA"
            @click="handleVerify2FA"
          >
            验证
          </el-button>

          <div class="text-center">
            <el-button link class="!text-zinc-500 hover:!text-zinc-300 text-sm" @click="cancel2FA">
              返回登录
            </el-button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { User, Lock, Platform, Key } from '@element-plus/icons-vue'
import { useUserStore } from '../store/user'
import { login, verifyLogin2FA } from '../api'

const router = useRouter()
const userStore = useUserStore()

const loginFormRef = ref()
const loading = ref(false)
const require2FA = ref(false)
const tempToken = ref('')
const twoFACode = ref('')
const verifying2FA = ref(false)

const loginForm = reactive({
  username: '',
  password: ''
})

const loginRules = {
  username: [{ required: true, message: '请输入账号', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
}

const handleLogin = async () => {
  if (!loginFormRef.value) return
  
  await loginFormRef.value.validate(async (valid: boolean) => {
    if (valid) {
      loading.value = true
      try {
        const res: any = await login({ username: loginForm.username, password: loginForm.password })
        if (res.code === 200) {
          if (res.data.require_2fa) {
            // 2FA enabled: show verification code input
            tempToken.value = res.data.temp_token
            require2FA.value = true
          } else {
            // Normal login
            userStore.setToken(res.data.token)
            userStore.setDelionId(res.data.d_delion_id)
            userStore.setUserInfo(res.data.user)
            ElMessage.success('登录成功')
            router.push('/')
          }
        }
      } catch {
      } finally {
        loading.value = false
      }
    }
  })
}

const handleVerify2FA = async () => {
  if (!twoFACode.value || twoFACode.value.length !== 6) {
    ElMessage.warning('请输入 6 位验证码')
    return
  }

  verifying2FA.value = true
  try {
    const res: any = await verifyLogin2FA({
      temp_token: tempToken.value,
      code: twoFACode.value
    })
    if (res.code === 200) {
      userStore.setToken(res.data.token)
      userStore.setDelionId(res.data.d_delion_id)
      userStore.setUserInfo(res.data.user)
      ElMessage.success('登录成功')
      router.push('/')
    }
  } catch {
    // Verification failed: clear state
    tempToken.value = ''
    twoFACode.value = ''
    require2FA.value = false
    userStore.logout()
  } finally {
    verifying2FA.value = false
  }
}

const cancel2FA = () => {
  require2FA.value = false
  tempToken.value = ''
  twoFACode.value = ''
}
</script>

<style scoped>
:deep(.el-input__wrapper) {
  background-color: rgba(39, 39, 42, 0.5) !important;
  box-shadow: 0 0 0 1px rgba(63, 63, 70, 0.5) inset !important;
}
:deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 1px #fff inset !important;
}
:deep(.el-input__inner) {
  color: #fff !important;
}
</style>
