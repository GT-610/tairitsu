import axios from 'axios'

// Type definitions for API responses and requests

export interface User {
  id: string;
  username: string;
  role: 'admin' | 'user';
  createdAt: string;
  updatedAt: string;
}

export interface Network {
  id: string;
  name: string;
  description?: string;
  config: NetworkConfig;
  members: Member[];
  status: string;
  createdAt: string;
  updatedAt: string;
}

export interface NetworkSummary {
  id: string;
  name: string;
  description?: string;
  owner_id: string;
  member_count: number;
  created_at: string;
  updated_at: string;
}

export interface NetworkConfig {
  private: boolean;
  allowPassiveBridging: boolean;
  enableBroadcast: boolean;
  mtu: number;
  multicastLimit: number;
  v4AssignMode: {
    zt: boolean;
  };
  v6AssignMode: {
    zt: boolean;
    '6plane': boolean;
    rfc4193: boolean;
  };
  routes: Route[];
  ipAssignmentPools: IpAssignmentPool[];
}

export interface Route {
  target: string;
  via?: string;
}

export interface IpAssignmentPool {
  ipRangeStart: string;
  ipRangeEnd: string;
}

export interface Member {
  id: string;
  networkId: string;
  nodeId: string;
  name?: string;
  description?: string;
  authorized: boolean;
  activeBridge: boolean;
  ipAssignments: string[];
  lastSeen: string;
  createdAt: string;
  updatedAt: string;
}

export interface ImportableNetworkSummary {
  network_id: string;
  reason: string;
  is_importable: boolean;
}

export interface ImportNetworksResponse {
  message: string;
  imported_ids: string[];
}

export interface SystemStatus {
  version: string;
  address: string;
  uptime: number;
  zeroTierStatus: 'online' | 'offline' | 'error';
  databaseStatus: 'connected' | 'disconnected' | 'error';
}

// System statistics interface
export interface SystemStats {
  cpuUsage: number;
  memoryUsage: number;
  timestamp: number;
  osName: string;
  platform: string;
  platformVersion: string;
  kernelVersion: string;
}

export interface SetupStatus {
  initialized: boolean;
  databaseConfigured: boolean;
  zeroTierConfigured: boolean;
  adminCreated: boolean;
}

export interface IdentityInfo {
  success: boolean;
  message: string;
  identityPublic?: string;
  identityPath?: string;
}

export interface GeneratePlanetResponse {
  success: boolean;
  message: string;
  planetData?: number[];
  planetId: number;
  birthTime: number;
  cHeader?: string;
}

export interface GeneratePlanetRequest {
  identityPublic: string;
  endpoints: string[];
  comments?: string;
  outputPath?: string;
}

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
  register: (data: { username: string; password: string }) => api.post<{ user: User; token: string }>('/auth/register', data),
  // User login
  login: (data: { username: string; password: string }) => api.post<{ user: User; token: string }>('/auth/login', data),
  // Get user profile
  getProfile: () => api.get<User>('/profile'),
  // Update user password
  updatePassword: (data: { current_password: string; new_password: string; confirm_password: string }) => api.put<void>('/profile/password', data)
}

// User management APIs
export const userAPI = {
  // Get all users
  getAllUsers: () => api.get<User[]>('/users'),
  // Update user role
  updateUserRole: (userId: string, role: 'admin' | 'user') => api.put<User>(`/users/${userId}/role`, { role })
}

// ZeroTier network related APIs
export const networkAPI = {
  // Get all networks (from database, lightweight)
  getAllNetworks: () => api.get<NetworkSummary[]>('/networks'),
  // Get a single network (with full details from ZeroTier API)
  getNetwork: (networkId: string) => api.get<Network>(`/networks/${networkId}`),
  // Create a network
  createNetwork: (data: { name: string; description?: string }) => api.post<Network>('/networks', data),
  // Update a network
  updateNetwork: (networkId: string, data: Partial<NetworkConfig>) => api.put<Network>(`/networks/${networkId}`, data),
  // Delete a network
  deleteNetwork: (networkId: string) => api.delete<void>(`/networks/${networkId}`),
  // Get importable networks (admin only)
  getImportableNetworks: () => api.get<ImportableNetworkSummary[]>('/admin/networks/importable'),
  // Import specified networks (admin only)
  importNetworks: (networkIds: string[]) => api.post<ImportNetworksResponse>('/admin/networks/import', { network_ids: networkIds })
}

// Member related APIs
export const memberAPI = {
  // Get network members
  getMembers: (networkId: string) => api.get<Member[]>(`/networks/${networkId}/members`),
  // Get a single member
  getMember: (networkId: string, memberId: string) => api.get<Member>(`/networks/${networkId}/members/${memberId}`),
  // Update a member
  updateMember: (networkId: string, memberId: string, data: { authorized?: boolean; name?: string; description?: string }) => api.put<Member>(`/networks/${networkId}/members/${memberId}`, data),
  // Delete a member
  deleteMember: (networkId: string, memberId: string) => api.delete<void>(`/networks/${networkId}/members/${memberId}`)
}

// System status API
export const statusAPI = {
  getStatus: () => api.get<SystemStatus>('/status')
}

// System related APIs
export const systemAPI = {
  // Get system status
  getStatus: () => api.get<SystemStatus>('/status'),
  // Get system setup status (used to check if it's first run)
  getSetupStatus: () => api.get<SetupStatus>('/system/status'),
  // Configure database
  configureDatabase: (config: any) => api.post('/system/database', config),
  // Reload routes
  reloadRoutes: () => api.post('/system/reload'),
  // Initialize ZeroTier client
  initZeroTierClient: () => api.post('/system/zerotier/init'),
  // Test ZeroTier connection
  testZtConnection: () => api.post('/system/zerotier/test'),
  // Save ZeroTier configuration
  saveZtConfig: (config: any) => api.post('/system/zerotier/config', config),
  // Update system settings
  updateSettings: (settings: any) => api.put('/settings', settings),
  // Set system initialization status
  setInitialized: (initialized: boolean) => api.post('/system/initialized', { initialized }),
  // Initialize admin account creation step
  initializeAdminCreation: () => api.post('/system/admin/init'),
  // Get system statistics (CPU, memory usage)
  getSystemStats: () => api.get<SystemStats>('/system/stats')
}

// Planet related APIs (admin only)
export const planetAPI = {
  // Get identity.public from ZeroTier data directory
  getIdentity: (ztPath?: string) => api.get<IdentityInfo>('/admin/planet/identity', {
    params: { path: ztPath || '/var/lib/zerotier-one' }
  }),
  // Generate signing keys
  generateSigningKeys: (ztPath?: string) => api.post<{ success: boolean; message: string; previousKey: string; currentKey: string }>('/admin/planet/keys', null, {
    params: { path: ztPath || '/var/lib/zerotier-one' }
  }),
  // Generate custom planet file
  generatePlanet: (data: GeneratePlanetRequest) => api.post<GeneratePlanetResponse>('/admin/planet/generate', data)
}

export default api