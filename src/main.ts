import { createApp } from 'vue'
import './style.css'
import 'element-plus/dist/index.css'
import 'element-plus/theme-chalk/dark/css-vars.css'
import ElementPlus from 'element-plus'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'

// 创建Vue应用实例
const app = createApp(App)

// 注册所有图标
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

// 使用Pinia
const pinia = createPinia()
app.use(pinia)

// 使用Element Plus
app.use(ElementPlus)

// 使用路由
app.use(router)

// 挂载应用
app.mount('#app')
