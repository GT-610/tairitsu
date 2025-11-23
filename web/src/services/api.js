/**
 * API服务模块
 * 提供与后端交互的HTTP请求功能，包含请求/响应拦截器和各功能模块的API封装
 */
import axios from 'axios'

/**
 * 创建axios实例，配置基础URL、超时时间和默认请求头
 * @type {axios.AxiosInstance}
 */
const api = axios.create({
  baseURL: '/api',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
})

/**
 * 请求拦截器
 * 负责在请求发送前添加认证令牌
 */
api.interceptors.request.use(
  (config) => {
    // 从localStorage获取认证令牌并添加到请求头
    const token = localStorage.getItem('token')
    if (token) {
      config.headers['Authorization'] = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    // 请求错误处理
    return Promise.reject(error)
  }
)

/**
 * 响应拦截器
 * 负责统一处理响应和错误
 */
api.interceptors.response.use(
  (response) => {
    // 成功响应直接返回
    return response
  },
  (error) => {
    // 处理401认证错误 - 清除token并重定向到登录页
    if (error.response && error.response.status === 401) {
      localStorage.removeItem('token')
      window.location.href = '/login'
    }
    // 其他错误向上抛出，由调用处处理
    return Promise.reject(error)
  }
)

/**
 * 认证相关API服务
 * 提供用户注册、登录和个人信息获取功能
 */
export const authAPI = {
  /**
   * 用户注册
   * @param {Object} userData - 用户注册信息
   * @param {string} userData.username - 用户名
   * @param {string} userData.email - 邮箱
   * @param {string} userData.password - 密码
   * @returns {Promise<axios.AxiosResponse>} 注册响应
   */
  register: (userData) => api.post('/auth/register', userData),
  
  /**
   * 用户登录
   * @param {Object} credentials - 登录凭证
   * @param {string} credentials.username - 用户名
   * @param {string} credentials.password - 密码
   * @returns {Promise<axios.AxiosResponse>} 登录响应，包含token和用户信息
   */
  login: (credentials) => api.post('/auth/login', credentials),
  
  /**
   * 获取当前用户信息
   * @returns {Promise<axios.AxiosResponse>} 用户信息响应
   */
  getProfile: () => api.get('/profile')
}

/**
 * ZeroTier网络管理API服务
 * 提供网络的CRUD操作功能
 */
export const networkAPI = {
  /**
   * 获取所有网络列表
   * @returns {Promise<axios.AxiosResponse>} 网络列表响应
   */
  getAllNetworks: () => api.get('/networks'),
  
  /**
   * 获取单个网络详情
   * @param {string} networkId - 网络ID
   * @returns {Promise<axios.AxiosResponse>} 网络详情响应
   */
  getNetwork: (networkId) => api.get(`/networks/${networkId}`),
  
  /**
   * 创建新网络
   * @param {Object} networkData - 网络配置数据
   * @param {string} networkData.name - 网络名称
   * @param {boolean} networkData.private - 是否私有网络
   * @param {Object} [networkData.settings] - 其他网络设置
   * @returns {Promise<axios.AxiosResponse>} 创建结果响应
   */
  createNetwork: (networkData) => api.post('/networks', networkData),
  
  /**
   * 更新网络配置
   * @param {string} networkId - 网络ID
   * @param {Object} networkData - 更新的网络配置
   * @returns {Promise<axios.AxiosResponse>} 更新结果响应
   */
  updateNetwork: (networkId, networkData) => api.put(`/networks/${networkId}`, networkData),
  
  /**
   * 删除网络
   * @param {string} networkId - 网络ID
   * @returns {Promise<axios.AxiosResponse>} 删除结果响应
   */
  deleteNetwork: (networkId) => api.delete(`/networks/${networkId}`)
}

/**
 * 网络成员管理API服务
 * 提供网络成员的查询和管理功能
 */
export const memberAPI = {
  /**
   * 获取网络的所有成员
   * @param {string} networkId - 网络ID
   * @returns {Promise<axios.AxiosResponse>} 成员列表响应
   */
  getMembers: (networkId) => api.get(`/networks/${networkId}/members`),
  
  /**
   * 获取单个成员详情
   * @param {string} networkId - 网络ID
   * @param {string} memberId - 成员ID
   * @returns {Promise<axios.AxiosResponse>} 成员详情响应
   */
  getMember: (networkId, memberId) => api.get(`/networks/${networkId}/members/${memberId}`),
  
  /**
   * 更新成员配置
   * @param {string} networkId - 网络ID
   * @param {string} memberId - 成员ID
   * @param {Object} memberData - 成员配置数据
   * @param {boolean} [memberData.authorized] - 是否授权
   * @param {string} [memberData.name] - 成员名称
   * @param {string} [memberData.description] - 成员描述
   * @returns {Promise<axios.AxiosResponse>} 更新结果响应
   */
  updateMember: (networkId, memberId, memberData) => 
    api.put(`/networks/${networkId}/members/${memberId}`, memberData),
  
  /**
   * 删除网络成员
   * @param {string} networkId - 网络ID
   * @param {string} memberId - 成员ID
   * @returns {Promise<axios.AxiosResponse>} 删除结果响应
   */
  deleteMember: (networkId, memberId) => 
    api.delete(`/networks/${networkId}/members/${memberId}`)
}

/**
 * 系统管理API服务
 * 提供系统状态、初始化和配置管理功能
 */
export const systemAPI = {
  /**
   * 获取系统状态
   * @returns {Promise<axios.AxiosResponse>} 系统状态响应
   */
  getStatus: () => api.get('/status'),
  
  /**
   * 获取系统设置状态（用于检测是否是首次运行）
   * @returns {Promise<axios.AxiosResponse>} 系统设置状态响应
   */
  getSetupStatus: () => api.get('/system/status'),
  
  /**
   * 配置数据库
   * @param {Object} dbConfig - 数据库配置信息
   * @param {string} dbConfig.type - 数据库类型
   * @param {string} dbConfig.host - 数据库主机
   * @param {number} dbConfig.port - 数据库端口
   * @param {string} dbConfig.name - 数据库名称
   * @param {string} dbConfig.username - 数据库用户名
   * @param {string} dbConfig.password - 数据库密码
   * @returns {Promise<axios.AxiosResponse>} 配置结果响应
   */
  configureDatabase: (dbConfig) => api.post('/system/database', dbConfig),
  
  /**
   * 初始化ZeroTier客户端
   * @returns {Promise<axios.AxiosResponse>} 初始化结果响应
   */
  initZeroTierClient: () => api.post('/system/zerotier/init'),
  
  /**
   * 测试ZeroTier连接
   * @returns {Promise<axios.AxiosResponse>} 连接测试结果响应
   */
  testZtConnection: () => api.post('/system/zerotier/test'),
  
  /**
   * 保存ZeroTier配置
   * @param {Object} ztConfig - ZeroTier配置信息
   * @param {string} ztConfig.apiUrl - ZeroTier API URL
   * @param {string} ztConfig.apiToken - ZeroTier API令牌
   * @returns {Promise<axios.AxiosResponse>} 保存结果响应
   */
  saveZtConfig: (ztConfig) => api.post('/system/zerotier/config', ztConfig),
  
  /**
   * 更新系统设置
   * @param {Object} settings - 系统设置信息
   * @returns {Promise<axios.AxiosResponse>} 更新结果响应
   */
  updateSettings: (settings) => api.put('/settings', settings),
  
  /**
   * 获取系统信息
   * @returns {Promise<axios.AxiosResponse>} 系统信息响应
   */
  getSystemInfo: () => api.get('/system/info'),
  
  /**
   * 设置系统初始化状态
   * @param {boolean} initialized - 是否已初始化
   * @returns {Promise<axios.AxiosResponse>} 设置结果响应
   */
  setInitialized: (initialized) => api.post('/system/initialized', { initialized })
}

export default api