<template>
  <div class="space-y-6 max-w-2xl mx-auto mt-8">
    <el-card shadow="never" class="!bg-zinc-900 !border-zinc-800">
      <template #header>
        <div class="flex items-center gap-3">
          <el-icon class="text-white text-xl"><User /></el-icon>
          <span class="text-lg font-semibold text-white">个人中心设置</span>
        </div>
      </template>

      <div class="py-4">
        <el-form 
          :model="form" 
          label-position="top" 
          v-loading="loading"
          element-loading-background="rgba(24, 24, 27, 0.8)"
        >
          <el-form-item label="登录账号">
            <el-input 
              v-model="form.username" 
              placeholder="请输入新的账号名称" 
              class="w-full"
            />
          </el-form-item>

          <el-divider class="!border-zinc-800 my-8" />
          
          <h3 class="text-white font-medium mb-4">修改密码</h3>

          <el-form-item label="当前密码">
            <el-input 
              v-model="form.old_password" 
              type="password" 
              placeholder="需要修改密码时必填" 
              show-password
            />
          </el-form-item>

          <el-form-item label="新密码">
            <el-input 
              v-model="form.new_password" 
              type="password" 
              placeholder="请输入新密码" 
              show-password
            />
          </el-form-item>

          <div class="mt-8 flex justify-end">
            <el-button 
              type="primary" 
              :loading="saving"
              @click="handleSave"
              class="!bg-white !text-black !border-none hover:!bg-zinc-200 px-8"
            >
              保存修改
            </el-button>
          </div>
        </el-form>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getProfile, updateProfile } from '../api'
import { User } from '@element-plus/icons-vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '../store/user'

const router = useRouter()
const userStore = useUserStore()

const loading = ref(true)
const saving = ref(false)

const form = reactive({
  username: '',
  old_password: '',
  new_password: ''
})

const fetchUserProfile = async () => {
  loading.value = true
  try {
    const res: any = await getProfile()
    if (res.code === 200) {
      form.username = res.data.username
    } else {
      ElMessage.error(res.message || '获取个人信息失败')
    }
  } catch (error) {
    ElMessage.error('获取信息异常')
  } finally {
    loading.value = false
  }
}

const handleSave = async () => {
  if (!form.username) {
    ElMessage.warning('账号名称不能为空')
    return
  }

  if (form.new_password && !form.old_password) {
    ElMessage.warning('修改密码时必须输入当前密码')
    return
  }

  saving.value = true
  try {
    const res: any = await updateProfile({
      username: form.username,
      old_password: form.old_password,
      new_password: form.new_password
    })

    if (res.code === 200) {
      ElMessage.success('保存成功')
      
      // 如果修改了密码或用户名，要求重新登录
      if (form.new_password || form.username !== userStore.user?.username) {
        ElMessage.info('账号信息已更新，请重新登录')
        setTimeout(() => {
          userStore.logout()
          router.push('/login')
        }, 1500)
      }
      
      form.old_password = ''
      form.new_password = ''
    } else {
      ElMessage.error(res.message || '保存失败')
    }
  } catch (error: any) {
    ElMessage.error(error.response?.data?.error || '保存异常')
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  fetchUserProfile()
})
</script>
