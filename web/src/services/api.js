import axios from 'axios'
import { useAuth } from './auth.jsx'

// 创建axios实例
const api = axios.create({
  baseURL: '/api',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 获取当前认证令牌的方法
const getAuthToken = () => {
  // 优先从auth上下文获取令牌
  try {
    // 直接从存储中获取令牌作为备用方案
    const tempToken = sessionStorage.getItem('tempToken')
    if (tempToken) {
      return tempToken
    }
    return localStorage.getItem('token') || sessionStorage.getItem('token')
  } catch (error) {
    console.error('Error getting auth token:', error)
    return null
  }
}

// 请求拦截器
api.interceptors.request.use(
  config => {
    const token = getAuthToken()
    if (token) {
      config.headers['Authorization'] = `Bearer ${token}`
    }
    return config
  },
  error => {
    return Promise.reject(error)
  }
)

// 响应拦截器
api.interceptors.response.use(
  response => {
    return response
  },
  error => {
    // 处理401错误
    if (error.response && error.response.status === 401) {
      // 清除所有认证信息
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      sessionStorage.removeItem('token')
      sessionStorage.removeItem('user')
      sessionStorage.removeItem('tempToken')
      sessionStorage.removeItem('isSetupWizard')
      
      // 根据当前路径决定重定向到哪里
      const currentPath = window.location.pathname
      if (currentPath.startsWith('/setup')) {
        // 设置向导路径，可能需要重新获取临时令牌
        window.location.href = '/setup'
      } else {
        // 其他路径重定向到登录页
        window.location.href = '/login'
      }
    }
    return Promise.reject(error)
  }
)

// 认证相关API
export const authAPI = {
  // 用户注册
  register: (data) => api.post('/auth/register', data),
  // 用户登录
  login: (data) => api.post('/auth/login', data),
  // 获取用户信息
  getProfile: () => api.get('/profile')
}

// ZeroTier网络相关API
export const networkAPI = {
  // 获取所有网络
  getAllNetworks: () => api.get('/networks'),
  // 获取单个网络
  getNetwork: (networkId) => api.get(`/networks/${networkId}`),
  // 创建网络
  createNetwork: (data) => api.post('/networks', data),
  // 更新网络
  updateNetwork: (networkId, data) => api.put(`/networks/${networkId}`, data),
  // 删除网络
  deleteNetwork: (networkId) => api.delete(`/networks/${networkId}`)
}

// 成员相关API
export const memberAPI = {
  // 获取网络成员
  getMembers: (networkId) => api.get(`/networks/${networkId}/members`),
  // 获取单个成员
  getMember: (networkId, memberId) => api.get(`/networks/${networkId}/members/${memberId}`),
  // 更新成员
  updateMember: (networkId, memberId, data) => api.put(`/networks/${networkId}/members/${memberId}`, data),
  // 删除成员
  deleteMember: (networkId, memberId) => api.delete(`/networks/${networkId}/members/${memberId}`)
}

// 系统状态API
export const statusAPI = {
  getStatus: () => api.get('/status')
}

// 系统相关API
export const systemAPI = {
  // 获取系统状态
  getStatus: () => api.get('/status'),
  // 获取系统设置状态（用于检测是否是首次运行）
  getSetupStatus: () => api.get('/system/status'),
  // 配置数据库
  configureDatabase: (config) => api.post('/system/database', config),
  // 初始化ZeroTier客户端
  initZeroTierClient: () => api.post('/system/zerotier/init'),
  // 测试ZeroTier连接
  testZtConnection: () => api.post('/system/zerotier/test'),
  // 保存ZeroTier配置
  saveZtConfig: (config) => api.post('/system/zerotier/config', config),
  // 更新系统设置
  updateSettings: (settings) => api.put('/settings', settings),
  // 获取系统信息
  getSystemInfo: () => api.get('/system/info'),
  // 设置系统初始化状态
  setInitialized: (initialized) => api.post('/system/initialized', { initialized }),
  // 初始化管理员账户创建步骤
  initializeAdminCreation: () => api.post('/system/admin/init'),
  // 生成设置向导临时令牌
  generateSetupWizardToken: () => api.post('/system/setup/token'),
  // 完成设置向导
  completeSetupWizard: () => api.post('/system/setup/complete')
}

export default api