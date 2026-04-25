<template>
  <div class="h-full flex flex-col space-y-4">
    <!-- 工具栏 -->
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-4 bg-zinc-900 px-4 py-2 rounded-lg border border-zinc-800">
        <el-button link type="primary" class="!text-zinc-400 hover:!text-white !p-0" @click="navigateUp" :disabled="currentPath === '/'">
          <el-icon :size="20"><Top /></el-icon>
        </el-button>
        <el-divider direction="vertical" class="!border-zinc-700" />
        <el-breadcrumb separator="/" class="!text-sm flex items-center">
          <el-breadcrumb-item class="cursor-pointer font-mono" @click="navigatePath('/')">/godelion/user</el-breadcrumb-item>
          <el-breadcrumb-item v-for="(p, index) in pathSegments" :key="index" class="font-mono text-zinc-300">
            <span class="cursor-pointer hover:text-white transition-colors" @click="navigatePathSegment(index)">{{ p }}</span>
          </el-breadcrumb-item>
        </el-breadcrumb>
      </div>

      <div class="flex items-center gap-3">
        <el-button class="!bg-zinc-800 !text-white !border-zinc-700 hover:!bg-zinc-700" @click="handleCreateFolder">
          <el-icon class="mr-2"><FolderAdd /></el-icon> 新建文件夹
        </el-button>
        <el-upload
          action="#"
          :show-file-list="false"
          :before-upload="handleUpload"
          class="inline-block"
        >
          <el-button type="primary" class="!bg-white !text-black hover:!bg-zinc-200 !border-none">
            <el-icon class="mr-2"><UploadFilled /></el-icon> 上传文件
          </el-button>
        </el-upload>
      </div>
    </div>

    <!-- 文件列表 -->
    <el-card shadow="never" class="flex-1 !bg-zinc-900 !border-zinc-800 flex flex-col">
      <el-table
        :data="files"
        v-loading="loading"
        style="width: 100%"
        class="custom-dark-table"
        @row-click="handleRowClick"
      >
        <el-table-column prop="name" label="文件名" min-width="300">
          <template #default="{ row }">
            <div class="flex items-center gap-3 cursor-pointer group">
              <div class="w-10 h-10 rounded-lg flex items-center justify-center transition-colors" :class="{
                'bg-blue-500/10 text-blue-400 group-hover:bg-blue-500/20': row.type === 'dir',
                'bg-zinc-800 text-zinc-300 group-hover:bg-zinc-700': row.type === 'file'
              }">
                <el-icon :size="24">
                  <Folder v-if="row.type === 'dir'" />
                  <Document v-else />
                </el-icon>
              </div>
              <span class="text-zinc-200 font-medium group-hover:text-white transition-colors">{{ row.name }}</span>
            </div>
          </template>
        </el-table-column>
        
        <el-table-column prop="size" label="大小" width="150">
          <template #default="{ row }">
            <span class="text-zinc-400 text-sm">{{ formatSize(row.size) }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="updatedAt" label="修改时间" width="200">
          <template #default="{ row }">
            <span class="text-zinc-500 text-sm font-mono">{{ row.updatedAt }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="permissions" label="权限" width="150">
          <template #default="{ row }">
            <span class="text-zinc-500 font-mono text-xs bg-zinc-800/50 px-2 py-1 rounded border border-zinc-700/50">
              {{ row.permissions }}
            </span>
          </template>
        </el-table-column>

        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <div class="flex items-center gap-3" @click.stop>
              <el-button v-if="row.type === 'file'" link type="primary" class="!p-0 hover:opacity-80">
                <el-icon :size="18"><Download /></el-icon>
              </el-button>
              
              <el-button v-if="isArchive(row.name)" link type="success" class="!p-0 hover:opacity-80" @click.stop="handleExtract(row)">
                <el-icon :size="18"><Box /></el-icon>
              </el-button>

              <el-button link type="warning" class="!p-0 hover:opacity-80" @click.stop="openMoveDialog(row)">
                <el-icon :size="18"><Position /></el-icon>
              </el-button>

              <el-button link type="danger" class="!p-0 hover:opacity-80" @click.stop="handleDelete(row)">
                <el-icon :size="18"><Delete /></el-icon>
              </el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <div class="mt-4 flex items-center justify-between text-xs text-zinc-500 px-2">
        <span>当前存储限制: 10GB / 已使用 2.4GB</span>
        <span class="flex items-center gap-1"><el-icon><Lock /></el-icon> 目录权限隔离开启中</span>
      </div>
    </el-card>

    <!-- 移动文件对话框 -->
    <el-dialog
      v-model="moveDialogVisible"
      title="移动到..."
      width="500px"
      custom-class="dark-dialog"
      :destroy-on-close="true"
    >
      <el-form label-width="80px" class="mt-4">
        <el-form-item label="目标目录">
          <el-cascader
            v-model="moveTargetPath"
            :props="folderProps"
            placeholder="选择目标目录"
            class="w-full"
            :show-all-levels="true"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="moveDialogVisible = false" class="!bg-transparent !text-white !border-zinc-700">取消</el-button>
          <el-button type="primary" @click="confirmMove" class="!bg-white !text-black !border-none hover:!bg-zinc-200">
            确认移动
          </el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getFiles, uploadFile, deleteFile, createFolder, moveFile, extractArchive } from '../api'
import {
  Top,
  Folder,
  FolderAdd,
  Document,
  UploadFilled,
  Download,
  Delete,
  Lock,
  Box,
  Position
} from '@element-plus/icons-vue'

const loading = ref(false)
const currentPath = ref('/')
const moveDialogVisible = ref(false)
const moveTargetRow = ref<any>(null)
const moveTargetPath = ref<string[]>([])

const folderProps = {
  lazy: true,
  checkStrictly: true,
  async lazyLoad(node: any, resolve: any) {
    const { level, path } = node
    
    try {
      // 动态解析真实路径以查询后端
      const currentQueryPath = level === 0 
        ? '/' 
        : '/' + path.join('/')
        
      const res = await getFiles(currentQueryPath)
      let items: any[] = []

      if (level === 0) {
        items.push({
          value: '/',
          label: '根目录 ( / )',
          leaf: true
        })
      }

      if (res.code === 200 && res.data) {
        const folderItems = res.data
          .filter((f: any) => f.is_dir)
          .map((f: any) => ({
            value: f.name,
            label: f.name,
            leaf: false 
          }))
        items = items.concat(folderItems)
      }
      resolve(items)
    } catch {
      resolve(level === 0 ? [{ value: '/', label: '根目录 ( / )', leaf: true }] : [])
    }
  }
}

const isArchive = (name: string) => {
  const lower = name.toLowerCase()
  return lower.endsWith('.zip') || lower.endsWith('.tar.gz') || lower.endsWith('.tgz')
}

const handleExtract = async (row: any) => {
  const filePath = currentPath.value === '/' ? `/${row.name}` : `${currentPath.value}/${row.name}`
  ElMessage.info(`正在解压 ${row.name}...`)
  try {
    const res = await extractArchive({ path: filePath })
    if (res.code === 200) {
      ElMessage.success('解压成功')
      fetchFiles()
    } else {
      ElMessage.error(res.message || '解压失败')
    }
  } catch (error) {
    ElMessage.error('解压异常')
  }
}

const openMoveDialog = (row: any) => {
  moveTargetRow.value = row
  moveTargetPath.value = []
  moveDialogVisible.value = true
}

const confirmMove = async () => {
  if (!moveTargetPath.value || moveTargetPath.value.length === 0) {
    ElMessage.warning('请选择目标目录')
    return
  }

  const sourcePath = currentPath.value === '/' ? `/${moveTargetRow.value.name}` : `${currentPath.value}/${moveTargetRow.value.name}`
  
  const targetDirSegments = moveTargetPath.value.filter(p => p !== '/')
  const targetDir = targetDirSegments.length === 0 ? '/' : '/' + targetDirSegments.join('/')
  const targetPath = targetDir === '/' ? `/${moveTargetRow.value.name}` : `${targetDir}/${moveTargetRow.value.name}`
  
  try {
    const res = await moveFile({ source_path: sourcePath, target_path: targetPath })
    if (res.code === 200) {
      ElMessage.success('移动成功')
      moveDialogVisible.value = false
      fetchFiles()
    } else {
      ElMessage.error(res.message || '移动失败')
    }
  } catch (error) {
    ElMessage.error('移动异常')
  }
}

const pathSegments = computed(() => {
  return currentPath.value.split('/').filter(p => p)
})

const files = ref<any[]>([])

const formatSize = (bytes: number) => {
  if (bytes === 0) return '-'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const navigateUp = () => {
  if (currentPath.value === '/') return
  const segments = currentPath.value.split('/').filter(p => p)
  segments.pop()
  currentPath.value = segments.length ? '/' + segments.join('/') : '/'
  fetchFiles()
}

const navigatePath = (path: string) => {
  if (currentPath.value === path) return
  currentPath.value = path
  fetchFiles()
}

const navigatePathSegment = (index: number) => {
  const newPath = '/' + pathSegments.value.slice(0, index + 1).join('/')
  navigatePath(newPath)
}

const handleRowClick = (row: any) => {
  if (row.type === 'dir') {
    currentPath.value = currentPath.value === '/' 
      ? '/' + row.name 
      : currentPath.value + '/' + row.name
    fetchFiles()
  }
}

const fetchFiles = async () => {
  loading.value = true
  try {
    const res = await getFiles(currentPath.value)
    if (res.code === 200) {
      files.value = (res.data || []).map((f: any) => ({
        id: f.name,
        name: f.name,
        type: f.is_dir ? 'dir' : 'file',
        size: f.size,
        updatedAt: 'N/A', // Backend API needs to provide this if necessary
        permissions: f.is_dir ? 'drwxr-xr-x' : '-rw-r--r--'
      }))
    } else {
      ElMessage.error(res.message || '获取文件列表失败')
      files.value = []
    }
  } catch (error) {
    ElMessage.error('获取文件列表异常')
    files.value = []
  } finally {
    loading.value = false
  }
}

const handleUpload = (file: File) => {
  ElMessage.info(`开始上传文件 ${file.name}...`)
  loading.value = true
  uploadFile(currentPath.value, file).then((res: any) => {
    if (res.code === 200) {
      ElMessage.success('上传成功')
      fetchFiles()
    } else {
      ElMessage.error(res.message || '上传失败')
      loading.value = false
    }
  }).catch((error: any) => {
    ElMessage.error('上传异常')
    loading.value = false
  })
  
  return false // Prevent default upload behavior
}

const handleCreateFolder = () => {
  ElMessageBox.prompt('请输入文件夹名称', '新建文件夹', {
    confirmButtonText: '确认',
    cancelButtonText: '取消',
    inputPattern: /^[\w.-]+$/,
    inputErrorMessage: '文件夹名称格式不正确',
    customClass: 'dark-message-box'
  }).then(async ({ value }) => {
    try {
      const res = await createFolder(currentPath.value, value)
      if (res.code === 200) {
        ElMessage.success('创建成功')
        fetchFiles()
      } else {
        ElMessage.error(res.message || '创建失败')
      }
    } catch (error) {
      ElMessage.error('创建异常')
    }
  }).catch(() => {})
}

const handleDelete = (row: any) => {
  ElMessageBox.confirm(`确定要永久删除 ${row.name} 吗？此操作无法恢复。`, '危险操作', {
    confirmButtonText: '删除',
    cancelButtonText: '取消',
    type: 'error',
    customClass: 'dark-message-box'
  }).then(async () => {
    const filePath = currentPath.value === '/' ? `/${row.name}` : `${currentPath.value}/${row.name}`
    try {
      const res = await deleteFile(filePath)
      if (res.code === 200) {
        ElMessage.success('已删除')
        fetchFiles()
      } else {
        ElMessage.error(res.message || '删除失败')
      }
    } catch (error) {
      ElMessage.error('删除异常')
    }
  }).catch(() => {})
}

onMounted(() => {
  fetchFiles()
})
</script>

<style scoped>
:deep(.el-breadcrumb__inner) {
  color: #a1a1aa !important;
}
:deep(.el-breadcrumb__item:last-child .el-breadcrumb__inner) {
  color: #fff !important;
  font-weight: 600;
}
:deep(.el-breadcrumb__separator) {
  color: #52525b !important;
}
</style>
