import request from '../utils/request'

export interface ApiResponse<T = any> {
  code: number
  message: string
  data: T
}

export const login = (data: any) => request.post<any, ApiResponse>('/auth/login', data)

export const getWorkloads = () => request.get<any, ApiResponse>('/workloads')
export const createWorkload = (data: any) => request.post<any, ApiResponse>('/workloads', data, { timeout: 300000 }) // Increase timeout to 5 mins for Docker image pull
export const startWorkload = (id: string) => request.post<any, ApiResponse>(`/workloads/${id}/start`, {}, { timeout: 60000 })
export const stopWorkload = (id: string) => request.post<any, ApiResponse>(`/workloads/${id}/stop`, {}, { timeout: 60000 })
export const deleteWorkload = (id: string) => request.delete<any, ApiResponse>(`/workloads/${id}`, { timeout: 60000 })
export const updateWorkload = (id: string, data: any) => request.put<any, ApiResponse>(`/workloads/${id}`, data)
export const getWorkloadLogs = (id: string) => request.get<any, ApiResponse>(`/workloads/${id}/logs`)

// Gateway & Network
export const getGateways = () => request.get<any, ApiResponse>('/gateways')
export const createGateway = (data: any) => request.post<any, ApiResponse>('/gateways', data)
export const updateGateway = (id: string, data: any) => request.put<any, ApiResponse>(`/gateways/${id}`, data)
export const deleteGateway = (id: string) => request.delete<any, ApiResponse>(`/gateways/${id}`)

export const getSSLCerts = () => request.get<any, ApiResponse>('/ssl')
export const createSSLCert = (data: any) => request.post<any, ApiResponse>('/ssl', data)
export const deleteSSLCert = (id: string) => request.delete<any, ApiResponse>(`/ssl/${id}`)

// Storage
export const getFiles = (path: string = '/') => request.get<any, ApiResponse>(`/storage/list?path=${encodeURIComponent(path)}`)
export const uploadFile = (path: string, file: File) => {
  const formData = new FormData()
  formData.append('file', file)
  return request.post(`/storage/upload?path=${encodeURIComponent(path)}`, formData, {
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}

export const createFolder = (path: string, name: string) => {
  return request.post(`/storage/folder?path=${encodeURIComponent(path)}`, { name })
}

export const moveFile = (data: { source_path: string, target_path: string }) => {
  return request.post('/storage/move', data)
}

export const extractArchive = (data: { path: string }) => {
  return request.post('/storage/extract', data)
}

export const readFile = (path: string) => request.get<any, ApiResponse>(`/storage/read?path=${encodeURIComponent(path)}`)

export const deleteFile = (path: string) => request.delete<any, ApiResponse>(`/storage/delete?path=${encodeURIComponent(path)}`)
export const getDockerStatus = () => {
  return request.get('/system/docker/status')
}

export const installDocker = () => {
  return request.post('/system/docker/install', {}, { timeout: 300000 }) // 5 min timeout for installation
}

export const startDocker = () => request.post('/system/docker/start')
export const stopDocker = () => request.post('/system/docker/stop')
export const restartDocker = () => request.post<any, ApiResponse>('/system/docker/restart')
export const getDockerConfig = () => request.get<any, ApiResponse>('/system/docker/config')
export const updateDockerConfig = (config: string) => request.post<any, ApiResponse>('/system/docker/config', { config })

export const getMetrics = () => request.get<any, ApiResponse>('/metrics')
export const getAuditLogs = () => request.get<any, ApiResponse>('/audit')
