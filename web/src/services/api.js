import axios from 'axios'

// 创建axios实例
const api = axios.create({
  baseURL: '/api',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器
api.interceptors.request.use(
  config => {
    const token = localStorage.getItem('token')
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
      localStorage.removeItem('token')
      window.location.href = '/login'
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
  // 更新系统设置
  updateSettings: (settings) => api.put('/settings', settings),
  // 获取系统信息
  getSystemInfo: () => api.get('/system/info')
}

export default api