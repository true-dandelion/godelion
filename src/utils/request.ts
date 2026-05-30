import axios from 'axios'
import { ElMessage } from 'element-plus'
import { useUserStore } from '../store/user'
import router from '../router'

const service = axios.create({
  baseURL: 'http://localhost:8080/sys/v1',
  timeout: 60000 // 增加默认超时时间到 60 秒
})

// Deduplicate error messages: same message within 3s only shows once
let lastErrorMessage = ''
let lastErrorTime = 0
const ERROR_DEBOUNCE_MS = 3000

const showError = (message: string) => {
  const now = Date.now()
  if (message === lastErrorMessage && now - lastErrorTime < ERROR_DEBOUNCE_MS) {
    return // Skip duplicate within debounce window
  }
  lastErrorMessage = message
  lastErrorTime = now
  ElMessage.error(message)
}

service.interceptors.request.use(
  config => {
    const userStore = useUserStore()
    // Send JWT in Authorization header
    if (userStore.token) {
      config.headers.Authorization = `Bearer ${userStore.token}`
    }
    // Send d_delion_id in custom header
    if (userStore.delionId) {
      config.headers['X-Delion-Id'] = userStore.delionId
    }
    return config
  },
  error => {
    return Promise.reject(error)
  }
)

service.interceptors.response.use(
  response => {
    const { data } = response
    // If the custom code is not 0 or 200, it is judged as an error.
    if (data.code !== 0 && data.code !== 200) {
      showError(data.message || '请求失败')
      if (data.code === 401 || data.code === 403) {
        const userStore = useUserStore()
        userStore.logout()
        router.push('/login')
      }
      return Promise.reject(new Error(data.message || 'Error'))
    }
    return data
  },
  error => {
    if (error.response) {
      // Show backend error message if available (support both "message" and "error" formats)
      const message = error.response.data?.message || error.response.data?.error || '请求失败'
      showError(message)
      if (error.response.status === 401 || error.response.status === 403) {
        const userStore = useUserStore()
        userStore.logout()
        router.push('/login')
      }
    } else {
      showError('请求失败')
    }
    return Promise.reject(error)
  }
)

export default service
