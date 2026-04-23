import axios from 'axios'
import { toError } from './errors'

// Type definitions for API responses and requests

export interface User {
  id: string;
  username: string;
  role: 'admin' | 'user';
  createdAt: string;
  updatedAt: string;
}

export interface RegisterResponse {
  user: User;
  message: string;
}

export interface TransferAdminResponse {
  message: string;
  user: User;
}

export interface ResetUserPasswordResponse {
  message: string;
  user: User;
  temporary_password: string;
  revoked_sessions: number;
}

export interface CreateUserResponse {
  message: string;
  user: User;
  temporary_password: string;
}

export interface DeleteUserResponse {
  message: string;
  user: User;
  transferred_networks: number;
  revoked_sessions: number;
}

export interface UserSession {
  id: string;
  userAgent: string;
  ipAddress: string;
  rememberMe: boolean;
  lastSeenAt: string;
  expiresAt: string;
  revokedAt?: string | null;
  createdAt: string;
  updatedAt: string;
  current: boolean;
}

export interface UpdatePasswordResponse {
  message: string;
  revoked_other_sessions: number;
}

export interface Network {
  id: string;
  name: string;
  description?: string;
  db_description?: string;
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
  authorized_member_count: number;
  pending_member_count: number;
  created_at: string;
  updated_at: string;
}

export interface NetworkConfig {
  private: boolean;
  allowPassiveBridging?: boolean;
  enableBroadcast: boolean;
  mtu?: number;
  multicastLimit?: number;
  dns?: DNSConfig;
  v4AssignMode?: {
    zt: boolean;
  };
  v6AssignMode?: {
    zt: boolean;
    '6plane': boolean;
    rfc4193: boolean;
  };
  routes?: Route[];
  ipAssignmentPools?: IpAssignmentPool[];
}

export interface NetworkUpdateRequest {
  name?: string;
  description?: string;
  private?: boolean;
  allowPassiveBridging?: boolean;
  enableBroadcast?: boolean;
  mtu?: number;
  multicastLimit?: number;
  dns?: DNSConfig;
  v4AssignMode?: {
    zt: boolean;
  };
  v6AssignMode?: {
    zt: boolean;
    '6plane': boolean;
    rfc4193: boolean;
  };
  routes?: Route[];
  ipAssignmentPools?: IpAssignmentPool[];
}

export interface NetworkMetadataUpdateRequest {
  name: string;
  description?: string;
}

export interface Route {
  target: string;
  via?: string;
}

export interface IpAssignmentPool {
  ipRangeStart: string;
  ipRangeEnd: string;
}

export interface DNSConfig {
  domain?: string;
  servers?: string[];
}

export interface Member {
  id: string;
  networkId?: string;
  nodeId?: string;
  name?: string;
  description?: string;
  authorized?: boolean;
  activeBridge?: boolean;
  ipAssignments?: string[];
  lastSeen?: string | number;
  createdAt?: string | number;
  updatedAt?: string | number;
  clientVersion?: string;
  online?: boolean;
  address?: string;
  config?: {
    authorized?: boolean;
    activeBridge?: boolean;
    ipAssignments?: string[];
    noAutoAssignIps?: boolean;
  };
  noAutoAssignIps?: boolean;
  vMajor?: number;
  vMinor?: number;
  vRev?: number;
}

export interface ImportableNetworkCandidate {
  network_id: string;
  name?: string;
  description?: string;
  controller_status?: string;
  member_count?: number;
  status: 'available' | 'managed' | 'blocked';
  can_import: boolean;
  reason_code: string;
  reason_message: string;
  owner_id?: string;
  owner_username?: string;
}

export interface ImportableNetworksResponse {
  candidates: ImportableNetworkCandidate[];
  summary: {
    total: number;
    available: number;
    managed: number;
    blocked: number;
  };
}

export interface ImportNetworkResultItem {
  network_id: string;
  name?: string;
  owner_id?: string;
  owner_username?: string;
  reason_code?: string;
  reason_message?: string;
}

export interface ImportNetworksResponse {
  target_owner: {
    id: string;
    username: string;
  };
  summary: {
    requested: number;
    imported: number;
    failed: number;
    skipped: number;
  };
  imported: ImportNetworkResultItem[];
  failed: ImportNetworkResultItem[];
  skipped: ImportNetworkResultItem[];
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
  hasDatabase?: boolean;
  hasAdmin?: boolean;
  allowPublicRegistration: boolean;
  ztStatus?: {
    version: string;
    address: string;
    online: boolean;
    tcpFallbackAvailable?: boolean;
    apiReady?: boolean;
  };
}

export interface DatabaseSetupConfig {
  type: 'sqlite';
  path?: string;
  host?: string;
  port?: number;
  user?: string;
  pass?: string;
  name?: string;
}

export interface ZeroTierSetupConfig {
  controllerUrl: string;
  tokenPath: string;
}

export interface RuntimeSettings {
  allow_public_registration: boolean;
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
    return Promise.reject(toError(error))
  }
)

// Response interceptor
api.interceptors.response.use(
  response => {
    return response
  },
  error => {
    return Promise.reject(toError(error))
  }
)

// Authentication related APIs
export const authAPI = {
  // User registration
  register: (data: { username: string; password: string }) => api.post<RegisterResponse>('/auth/register', data),
  // User login
  login: (data: { username: string; password: string; remember_me?: boolean }) => api.post<{ user: User; token: string; session: UserSession }>('/auth/login', data),
  // Logout current session
  logout: () => api.post<{ message: string }>('/auth/logout'),
  // Get user profile
  getProfile: () => api.get<User>('/profile'),
  // Get current user's sessions
  getSessions: () => api.get<{ sessions: UserSession[] }>('/profile/sessions'),
  // Revoke one session
  revokeSession: (sessionId: string) => api.delete<{ message: string }>(`/profile/sessions/${sessionId}`),
  // Revoke all other sessions
  revokeOtherSessions: () => api.delete<{ message: string; count: number }>('/profile/sessions/others'),
  // Update user password
  updatePassword: (data: { current_password: string; new_password: string; confirm_password: string; logout_other_sessions?: boolean }) => api.put<UpdatePasswordResponse>('/profile/password', data)
}

// User management APIs
export const userAPI = {
  // Get all users
  getAllUsers: () => api.get<User[]>('/users'),
  // Create one user as admin
  createUser: (data: { username: string }) => api.post<CreateUserResponse>('/users', data),
  // Delete one user as admin
  deleteUser: (userId: string) => api.delete<DeleteUserResponse>(`/users/${userId}`),
  // Transfer admin role to another user
  transferAdmin: (userId: string) => api.post<TransferAdminResponse>('/users/transfer-admin', { user_id: userId }),
  // Reset one user's password as admin
  resetPassword: (userId: string) => api.post<ResetUserPasswordResponse>(`/users/${userId}/reset-password`)
}

// ZeroTier network related APIs
export const networkAPI = {
  // Get all networks (from database, lightweight)
  getAllNetworks: () => api.get<NetworkSummary[]>('/networks'),
  // Get a single network (with full details from ZeroTier API)
  getNetwork: (networkId: string) => api.get<Network>(`/networks/${networkId}`),
  // Create a network
  createNetwork: (data: { name: string; description?: string }) => api.post<Network>('/networks', data),
  // Update a network (config only, goes to ZeroTier controller)
  updateNetwork: (networkId: string, data: NetworkUpdateRequest) => api.put<Network>(`/networks/${networkId}`, data),
  // Update network metadata (name and description, goes to database only for description, both for name)
  updateNetworkMetadata: (networkId: string, data: NetworkMetadataUpdateRequest) => api.put<Network>(`/networks/${networkId}/metadata`, data),
  // Delete a network
  deleteNetwork: (networkId: string) => api.delete<void>(`/networks/${networkId}`),
  // Get importable networks (admin only)
  getImportableNetworks: () => api.get<ImportableNetworksResponse>('/admin/networks/importable'),
  // Import specified networks (admin only)
  importNetworks: (networkIds: string[], ownerId: string) => api.post<ImportNetworksResponse>('/admin/networks/import', {
    network_ids: networkIds,
    owner_id: ownerId
  })
}

// Member related APIs
export const memberAPI = {
  // Get network members
  getMembers: (networkId: string) => api.get<Member[]>(`/networks/${networkId}/members`),
  // Get a single member
  getMember: (networkId: string, memberId: string) => api.get<Member>(`/networks/${networkId}/members/${memberId}`),
  // Update a member
  updateMember: (networkId: string, memberId: string, data: { authorized?: boolean; name?: string; activeBridge?: boolean; noAutoAssignIps?: boolean; ipAssignments?: string[] }) => api.put<Member>(`/networks/${networkId}/members/${memberId}`, data),
  // Delete a member
  deleteMember: (networkId: string, memberId: string) => api.delete<void>(`/networks/${networkId}/members/${memberId}`)
}

// System related APIs
export const systemAPI = {
  // Get system status
  getStatus: () => api.get<SystemStatus>('/status'),
  // Get system setup status (used to check if it's first run)
  getSetupStatus: () => api.get<SetupStatus>('/system/status'),
  // Configure database
  configureDatabase: (config: DatabaseSetupConfig) => api.post('/system/database', config),
  // Initialize ZeroTier client
  initZeroTierClient: () => api.post('/system/zerotier/init'),
  // Test ZeroTier connection
  testZtConnection: () => api.get('/system/zerotier/test'),
  // Save ZeroTier configuration
  saveZtConfig: (config: ZeroTierSetupConfig) => api.post('/system/zerotier/config', config),
  // Set system initialization status
  setInitialized: (initialized: boolean) => api.post('/system/initialized', { initialized }),
  // Initialize admin account creation step
  initializeAdminCreation: () => api.post('/system/admin/init'),
  // Get runtime settings (admin only)
  getRuntimeSettings: () => api.get<RuntimeSettings>('/system/settings'),
  // Update runtime settings (admin only)
  updateRuntimeSettings: (settings: RuntimeSettings) => api.put<{ message: string; settings: RuntimeSettings }>('/system/settings', settings),
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
