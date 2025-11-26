import axios from 'axios'

// Create axios instance
const api = axios.create({
  baseURL: '/api',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// Request interceptor
api.interceptors.request.use(
  config => {
    // Prioritize getting token from localStorage, if not present then from sessionStorage
    const token = localStorage.getItem('token') || sessionStorage.getItem('token')
    if (token) {
      config.headers['Authorization'] = `Bearer ${token}`
    }
    return config
  },
  error => {
    return Promise.reject(error)
  }
)

// Response interceptor
api.interceptors.response.use(
  response => {
    return response
  },
  error => {
    return Promise.reject(error)
  }
)

// Authentication related APIs
export const authAPI = {
  // User registration
  register: (data) => api.post('/auth/register', data),
  // User login
  login: (data) => api.post('/auth/login', data),
  // Get user profile
  getProfile: () => api.get('/profile')
}

// ZeroTier network related APIs
export const networkAPI = {
  // Get all networks
  getAllNetworks: () => api.get('/networks'),
  // Get a single network
  getNetwork: (networkId) => api.get(`/networks/${networkId}`),
  // Create a network
  createNetwork: (data) => api.post('/networks', data),
  // Update a network
  updateNetwork: (networkId, data) => api.put(`/networks/${networkId}`, data),
  // Delete a network
  deleteNetwork: (networkId) => api.delete(`/networks/${networkId}`)
}

// Member related APIs
export const memberAPI = {
  // Get network members
  getMembers: (networkId) => api.get(`/networks/${networkId}/members`),
  // Get a single member
  getMember: (networkId, memberId) => api.get(`/networks/${networkId}/members/${memberId}`),
  // Update a member
  updateMember: (networkId, memberId, data) => api.put(`/networks/${networkId}/members/${memberId}`, data),
  // Delete a member
  deleteMember: (networkId, memberId) => api.delete(`/networks/${networkId}/members/${memberId}`)
}

// System status API
export const statusAPI = {
  getStatus: () => api.get('/status')
}

// System related APIs
export const systemAPI = {
  // Get system status
  getStatus: () => api.get('/status'),
  // Get system setup status (used to check if it's first run)
  getSetupStatus: () => api.get('/system/status'),
  // Configure database
  configureDatabase: (config) => api.post('/system/database', config),
  // Reload routes
  reloadRoutes: () => api.post('/system/reload'),
  // Initialize ZeroTier client
  initZeroTierClient: () => api.post('/system/zerotier/init'),
  // Test ZeroTier connection
  testZtConnection: () => api.post('/system/zerotier/test'),
  // Save ZeroTier configuration
  saveZtConfig: (config) => api.post('/system/zerotier/config', config),
  // Update system settings
  updateSettings: (settings) => api.put('/settings', settings),
  // Get system information
  getSystemInfo: () => api.get('/system/info'),
  // Set system initialization status
  setInitialized: (initialized) => api.post('/system/initialized', { initialized }),
  // Initialize admin account creation step
  initializeAdminCreation: () => api.post('/system/admin/init')
}

export default api