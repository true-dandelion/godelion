import { defineStore } from 'pinia'

// Cookie helper functions
const setCookie = (name: string, value: string, days: number = 7) => {
  const expires = new Date(Date.now() + days * 864e5).toUTCString()
  document.cookie = `${name}=${encodeURIComponent(value)}; expires=${expires}; path=/; SameSite=Strict`
}

const getCookie = (name: string): string => {
  const match = document.cookie.match(new RegExp('(^| )' + name + '=([^;]+)'))
  return match ? decodeURIComponent(match[2]) : ''
}

const removeCookie = (name: string) => {
  document.cookie = `${name}=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;`
}



export const useUserStore = defineStore('user', {
  state: () => ({
    token: localStorage.getItem('token') || '',
    delionId: getCookie('d_delion_id') || '',
    userInfo: null as any
  }),
  actions: {
    setToken(token: string) {
      this.token = token
      localStorage.setItem('token', token)
    },
    setDelionId(delionId: string) {
      this.delionId = delionId
      setCookie('d_delion_id', delionId)
    },
    setUserInfo(info: any) {
      this.userInfo = info
    },
    logout() {
      this.token = ''
      this.delionId = ''
      this.userInfo = null
      localStorage.removeItem('token')
      removeCookie('d_delion_id')
    }
  }
})
