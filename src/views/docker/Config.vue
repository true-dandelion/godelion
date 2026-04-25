<template>
  <div class="space-y-6 max-w-4xl">
    <el-card shadow="never" class="!bg-zinc-900 !border-zinc-800">
      <template #header>
        <div class="flex items-center justify-between">
          <span class="text-lg font-semibold text-white">Docker 配置文件 (daemon.json)</span>
          <el-button
            type="warning"
            :loading="saving"
            @click="handleSaveAndRestart"
            class="!bg-orange-600 !text-white hover:!bg-orange-500 !border-none"
          >
            <el-icon class="mr-2"><Check /></el-icon>
            保存并重启 Docker
          </el-button>
        </div>
      </template>

      <div class="py-4 space-y-4" v-loading="loading">
        <el-alert
          title="警告: 修改 Docker daemon.json 可能会导致 Docker 无法启动。修改后将自动重启 Docker 生效。"
          type="warning"
          :closable="false"
          show-icon
          class="!bg-orange-500/10 !text-orange-400 !border-orange-500/20 border mb-4"
        />

        <el-input
          v-model="configContent"
          type="textarea"
          :rows="15"
          placeholder="请输入 JSON 格式的配置..."
          class="font-mono text-sm custom-textarea"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getDockerConfig, updateDockerConfig, restartDocker } from '../../api'
import { Check } from '@element-plus/icons-vue'

const loading = ref(true)
const saving = ref(false)
const configContent = ref('')

const fetchConfig = async () => {
  loading.value = true
  try {
    const res: any = await getDockerConfig()
    if (res.code === 200) {
      configContent.value = res.data || '{\n}'
    } else {
      ElMessage.error(res.message || '获取配置失败')
    }
  } catch (error) {
    ElMessage.error('获取配置异常')
  } finally {
    loading.value = false
  }
}

const handleSaveAndRestart = async () => {
  if (!configContent.value) {
    ElMessage.warning('配置内容不能为空')
    return
  }

  // Basic JSON validation before sending to backend
  try {
    JSON.parse(configContent.value)
  } catch (e) {
    ElMessage.error('JSON 格式无效，请检查语法')
    return
  }

  ElMessageBox.confirm(
    '确认保存修改并立即重启 Docker 服务吗？这会导致所有正在运行的容器短暂中断。',
    '保存并重启',
    {
      confirmButtonText: '确认',
      cancelButtonText: '取消',
      type: 'warning',
      customClass: 'dark-message-box'
    }
  ).then(async () => {
    saving.value = true
    try {
      // Save config
      const saveRes: any = await updateDockerConfig(configContent.value)
      if (saveRes.code !== 200) {
        ElMessage.error(saveRes.message || '保存配置失败')
        saving.value = false
        return
      }

      ElMessage.success('配置已保存，正在重启 Docker...')
      
      // Restart Docker
      const restartRes: any = await restartDocker()
      if (restartRes.code === 200) {
        ElMessage.success('Docker 重启成功，配置已生效')
      } else {
        ElMessage.error(restartRes.message || 'Docker 重启失败')
      }
    } catch (error: any) {
      ElMessage.error(error.response?.data?.error || '操作异常')
    } finally {
      saving.value = false
    }
  }).catch(() => {})
}

onMounted(() => {
  fetchConfig()
})
</script>

<style>
.custom-textarea .el-textarea__inner {
  background-color: #09090b !important;
  color: #e4e4e7 !important;
  border-color: #27272a !important;
  box-shadow: none !important;
  padding: 16px;
  line-height: 1.6;
}
.custom-textarea .el-textarea__inner:focus {
  border-color: #52525b !important;
}
</style>
