import { defineStore } from 'pinia'

export const useUserStore = defineStore('user', {
  state: () => ({
    token: localStorage.getItem('token') || '',
    userInfo: null as any
  }),
  actions: {
    setToken(token: string) {
      this.token = token
      localStorage.setItem('token', token)
    },
    setUserInfo(info: any) {
      this.userInfo = info
    },
    logout() {
      this.token = ''
      this.userInfo = null
      localStorage.removeItem('token')
    }
  }
})
