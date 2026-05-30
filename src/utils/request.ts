import axios from 'axios'
import { ElMessage } from 'element-plus'
import { useUserStore } from '../store/user'
import router from '../router'

const service = axios.create({
  baseURL: 'http://localhost:8080/sys/v1',
  timeout: 60000 // 增加默认超时时间到 60 秒
})

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
      ElMessage.error(data.message || 'Error')
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
    let message = error.message
    if (error.response) {
      message = error.response.data?.message || 'Server Error'
      if (error.response.status === 401 || error.response.status === 403) {
        const userStore = useUserStore()
        userStore.logout()
        router.push('/login')
      }
    }
    ElMessage.error(message)
    return Promise.reject(error)
  }
)

export default service
