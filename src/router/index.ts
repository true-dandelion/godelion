import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { useUserStore } from '../store/user'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/Login.vue'),
    meta: { public: true }
  },
  {
    path: '/',
    name: 'Layout',
    component: () => import('../views/layout/index.vue'),
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('../views/Dashboard.vue'),
        meta: { title: '仪表盘' }
      },
      {
        path: 'container',
        name: 'Container',
        component: () => import('../views/Container.vue'),
        meta: { title: '容器管控' }
      },
      {
        path: 'network',
        name: 'Network',
        component: () => import('../views/Network.vue'),
        meta: { title: '网络中继' }
      },
      {
        path: 'file',
        name: 'File',
        component: () => import('../views/File.vue'),
        meta: { title: '文件隔离' }
      },
      {
        path: 'ssl',
        name: 'SSL',
        component: () => import('../views/SSL.vue'),
        meta: { title: 'SSL证书' }
      },
      {
        path: 'docker/status',
        name: 'DockerStatus',
        component: () => import('../views/docker/Status.vue'),
        meta: { title: '运行环境状态' }
      },
      {
        path: 'docker/config',
        name: 'DockerConfig',
        component: () => import('../views/docker/Config.vue'),
        meta: { title: 'Docker 配置' }
      },
      {
        path: 'settings',
        name: 'Settings',
        component: () => import('../views/Settings.vue'),
        meta: { title: '设置' }
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, from, next) => {
  const userStore = useUserStore()
  if (!to.meta.public && !userStore.token) {
    next('/login')
  } else {
    next()
  }
})

export default router
