import React, { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react'
import { createTheme, ThemeProvider } from '@mui/material/styles'
import type { ThemeOptions } from '@mui/material/styles'
import { enUS, zhCN } from '@mui/material/locale'

export type Language = 'en' | 'zh-CN'
export type LanguagePreference = 'system' | Language

const STORAGE_KEY = 'tairitsu.language'

const baseTheme: ThemeOptions = {
  palette: {
    mode: 'dark',
    primary: {
      main: '#64b5f6',
    },
    secondary: {
      main: '#ff8a65',
    },
    background: {
      default: '#121212',
      paper: '#1e1e1e',
    },
  },
}

const en: Record<string, string> = {
  'language.system': 'Follow system',
  'language.en': 'English',
  'language.zh-CN': 'Simplified Chinese',
  'settings.language.title': 'Language',
  'settings.language.description': 'Choose the display language for this browser.',
  'settings.language.current': 'Current language',
  'common.loading': 'Loading...',
  'common.unknown': 'Unknown',
  'common.save': 'Save',
  'common.cancel': 'Cancel',
  'common.confirm': 'Confirm',
  'common.delete': 'Delete',
  'common.refresh': 'Refresh',
  'common.search': 'Search',
  'common.resetChanges': 'Reset changes',
  'common.processing': 'Processing...',
  'common.saving': 'Saving...',
  'common.removing': 'Removing...',
  'common.unavailable': 'Unavailable',
  'errors.auth.invalidToken': 'Invalid authentication token',
  'errors.auth.missingToken': 'Missing authentication token',
  'errors.auth.invalidFormat': 'Invalid authentication format',
  'errors.auth.required': 'Authentication required',
  'errors.auth.adminRequired': 'Administrator permission required',
  'errors.rateLimited': 'Too many requests. Please try again later.',
  'messages.passwordUpdated': 'Password updated successfully',
  'messages.logoutCurrent': 'Current session signed out',
  'messages.sessionRemoved': 'Session removed',
  'messages.otherSessionsRemoved': 'Other sessions removed',
}

const zh: Record<string, string> = {
  'language.system': '跟随系统',
  'language.en': 'English',
  'language.zh-CN': '简体中文',
  'settings.language.title': '语言',
  'settings.language.description': '选择此浏览器使用的显示语言。',
  'settings.language.current': '当前语言',
  'common.loading': '加载中...',
  'common.unknown': '未知',
  'common.save': '保存',
  'common.cancel': '取消',
  'common.confirm': '确认',
  'common.delete': '删除',
  'common.refresh': '刷新',
  'common.search': '搜索',
  'common.resetChanges': '重置更改',
  'common.processing': '处理中...',
  'common.saving': '保存中...',
  'common.removing': '移除中...',
  'common.unavailable': '不可用',
  'errors.auth.invalidToken': '无效的认证令牌',
  'errors.auth.missingToken': '缺少认证令牌',
  'errors.auth.invalidFormat': '认证格式无效',
  'errors.auth.required': '需要认证',
  'errors.auth.adminRequired': '需要管理员权限',
  'errors.rateLimited': '请求频率过高，请稍后再试',
  'messages.passwordUpdated': '密码修改成功',
  'messages.logoutCurrent': '已退出当前会话',
  'messages.sessionRemoved': '会话已移除',
  'messages.otherSessionsRemoved': '其他会话已移除',
}

const messageCodes: Record<string, { en: string; 'zh-CN': string }> = {
  'auth.missing_token': { en: 'Missing authentication token', 'zh-CN': '缺少认证令牌' },
  'auth.invalid_format': { en: 'Invalid authentication format', 'zh-CN': '认证格式无效' },
  'auth.invalid_token': { en: 'Invalid authentication token', 'zh-CN': '无效的认证令牌' },
  'auth.required': { en: 'Authentication required', 'zh-CN': '需要认证' },
  'auth.admin_required': { en: 'Administrator permission required', 'zh-CN': '需要管理员权限' },
  'auth.unauthorized': { en: 'Unauthorized access', 'zh-CN': '未授权访问' },
  'auth.logout_success': { en: 'Current session signed out', 'zh-CN': '已退出当前会话' },
  'auth.session_removed': { en: 'Session removed', 'zh-CN': '会话已移除' },
  'auth.other_sessions_removed': { en: 'Other sessions removed', 'zh-CN': '其他会话已移除' },
  'auth.password_updated': { en: 'Password updated successfully', 'zh-CN': '密码修改成功' },
  'auth.password_confirmation_mismatch': { en: 'The new password and confirmation do not match', 'zh-CN': '新密码与确认密码不匹配' },
  'auth.token_generation_failed': { en: 'Failed to generate token', 'zh-CN': '生成令牌失败' },
  'user.db_unavailable': { en: 'Database is not configured. Complete initial setup first.', 'zh-CN': '系统尚未配置数据库，请先完成初始设置' },
  'user.invalid_username': { en: 'Username is required', 'zh-CN': '用户名不能为空' },
  'user.username_exists': { en: 'Username already exists', 'zh-CN': '用户名已存在' },
  'user.invalid_credentials': { en: 'Username or password is incorrect', 'zh-CN': '用户名或密码错误' },
  'user.not_found': { en: 'User not found', 'zh-CN': '用户不存在' },
  'user.old_password_incorrect': { en: 'Current password is incorrect', 'zh-CN': '原密码错误' },
  'user.invalid_role': { en: 'Invalid role. Must be admin or user.', 'zh-CN': '无效的角色值，必须是admin或user' },
  'user.admin_access_denied': { en: 'The current user is not an administrator.', 'zh-CN': '当前用户不是管理员，无法执行该操作' },
  'user.invalid_admin_operation': { en: 'This administrator operation is not allowed', 'zh-CN': '该管理员操作不允许' },
  'user.public_registration_disabled': { en: 'Public registration is disabled. Contact an administrator to create an account.', 'zh-CN': '公开注册已关闭，请联系管理员创建账户' },
  'user.required': { en: 'User is required', 'zh-CN': '必须指定用户' },
  'session.not_found': { en: 'Session not found', 'zh-CN': '会话不存在' },
  'session.access_denied': { en: 'You do not have access to this session', 'zh-CN': '无权访问该会话' },
  'session.revoked': { en: 'Session is no longer valid. Please sign in again.', 'zh-CN': '会话已失效，请重新登录' },
  'session.expired': { en: 'Session expired. Please sign in again.', 'zh-CN': '会话已过期，请重新登录' },
  'network.not_found': { en: 'Network not found', 'zh-CN': '网络不存在' },
  'network.access_denied': { en: 'Network access denied', 'zh-CN': '无权限访问网络' },
  'network.member_access_denied': { en: 'Network member access denied', 'zh-CN': '无权限访问网络成员' },
  'network.viewer_access_denied': { en: 'Network viewer access denied', 'zh-CN': '无权限管理网络查看授权' },
  'network.viewer_target_invalid': { en: 'Only regular users can be granted network viewer access', 'zh-CN': '只能授权普通用户查看网络' },
  'network.import_access_denied': { en: 'Only administrators can import networks', 'zh-CN': '只有管理员可以导入网络' },
  'network.import_empty': { en: 'Network ID list is empty', 'zh-CN': '网络ID列表为空' },
  'network.import_owner_required': { en: 'Network owner is required', 'zh-CN': '必须指定网络所有者' },
  'network.import_owner_not_found': { en: 'Specified network owner was not found', 'zh-CN': '指定的网络所有者不存在' },
  'network.delete_success': { en: 'Network deleted successfully', 'zh-CN': '网络删除成功' },
  'network.viewer_added': { en: 'Read-only viewer access granted', 'zh-CN': '已授予只读查看权限' },
  'network.viewer_removed': { en: 'Read-only viewer access removed', 'zh-CN': '已移除只读查看权限' },
  'member.not_found': { en: 'Member not found', 'zh-CN': '成员不存在' },
  'member.delete_success': { en: 'Member deleted successfully', 'zh-CN': '成员删除成功' },
  'system.already_initialized': { en: 'The system is already initialized. This endpoint is only available during first-time setup.', 'zh-CN': '系统已初始化，当前接口仅在首次设置期间可用' },
  'system.setup_required': { en: 'System setup is required. Complete the setup wizard first.', 'zh-CN': '系统尚未初始化，请先完成设置向导' },
  'system.user_service_unavailable': { en: 'User service is unavailable', 'zh-CN': '用户服务不可用' },
  'system.rate_limited': { en: 'Too many requests. Please try again later.', 'zh-CN': '请求频率过高，请稍后再试' },
  'system.internal_error': { en: 'Internal server error', 'zh-CN': '服务器内部错误' },
  'system.settings_updated': { en: 'Instance settings updated successfully', 'zh-CN': '实例设置更新成功' },
  'system.database_configured': { en: 'Database configured successfully', 'zh-CN': '数据库配置成功' },
  'system.zerotier_initialized': { en: 'ZeroTier client initialized successfully', 'zh-CN': 'ZeroTier客户端初始化成功' },
  'system.zerotier_configured': { en: 'ZeroTier configuration saved successfully', 'zh-CN': 'ZeroTier配置保存成功' },
  'system.admin_creation_initialized': { en: 'Administrator account creation step initialized successfully', 'zh-CN': '管理员账户创建步骤初始化成功' },
  'system.initialized_updated': { en: 'Initialization state updated successfully', 'zh-CN': '初始化状态更新成功' },
  'system.stats_unavailable': { en: 'Unable to retrieve system resource statistics', 'zh-CN': '无法获取系统资源统计信息' },
}

const rawEn: Record<string, string> = {
  '加载中...': 'Loading...',
  '加载中…': 'Loading...',
  '网络': 'Networks',
  '个人信息': 'Profile',
  '设置': 'Settings',
  '管理员面板': 'Admin Dashboard',
  '用户管理': 'User Management',
  '获取用户数据失败': 'Failed to load user data',
  '创建用户失败': 'Failed to create user',
  '重置用户密码失败': 'Failed to reset user password',
  '删除用户失败': 'Failed to delete user',
  '实例治理设置已保存': 'Instance governance settings saved',
  '保存实例治理设置失败': 'Failed to save instance governance settings',
  '请选择新的管理员': 'Select the new administrator',
  '转让管理员身份失败': 'Failed to transfer administrator role',
  '当前系统仅保留一个管理员。你可以创建普通用户、重置普通用户密码，或将管理员身份转让给某个普通用户。': 'This system keeps a single administrator. You can create regular users, reset regular-user passwords, or transfer the administrator role to a regular user.',
  '创建用户': 'Create User',
  '实例治理': 'Instance Governance',
  '这里控制实例级账户与发布边界。当前支持公开注册策略配置。': 'Control instance-level account and exposure boundaries here. Public registration policy is currently supported.',
  '允许公开注册': 'Allow public registration',
  '关闭后，未登录用户将不能继续公开创建账号，但 setup 阶段的首个管理员创建逻辑不受影响。': 'When disabled, unauthenticated users can no longer create accounts publicly. First administrator creation during setup is not affected.',
  '重置': 'Reset',
  '管理员职责': 'Administrator Role',
  '当前系统保持单管理员模型。你可以在这里把管理员身份转让给某个普通用户，转让后自己会自动降为普通用户。': 'The system keeps a single-administrator model. You can transfer the administrator role to a regular user here; after transfer, your own account becomes a regular user.',
  '当前管理员：': 'Current administrator: ',
  '新的管理员': 'New administrator',
  '当前没有可接收管理员身份的普通用户。请先创建或保留至少一个普通用户账号。': 'There is no regular user who can receive the administrator role. Create or keep at least one regular user account first.',
  '转让中...': 'Transferring...',
  '转让管理员身份': 'Transfer administrator role',
  '角色': 'Role',
  '删除用户': 'Delete user',
  '重置密码': 'Reset password',
  '转让管理员': 'Transfer administrator',
  '当前管理员': 'Current administrator',
  '已是管理员': 'Already administrator',
  '新用户会自动创建为普通用户，并生成一个一次性临时密码。': 'The new user will be created as a regular user and receive a one-time temporary password.',
  '建议使用可识别、便于通知的用户名': 'Use an identifiable username that is easy to notify.',
  '创建中...': 'Creating...',
  '确认创建': 'Confirm creation',
  '一次性临时密码': 'One-time temporary password',
  '该密码只会展示这一次。关闭后将无法再次查看，请立即通过其他方式安全告知用户。': 'This password is shown only once. After closing, it cannot be viewed again. Share it securely with the user right away.',
  '新用户': 'New user',
  '目标用户': 'Target user',
  '临时密码': 'Temporary password',
  '用户首次收到密码后，应尽快登录并到“设置”中修改为自己的新密码。': 'After receiving the password, the user should sign in soon and change it in Settings.',
  '我已记录': 'I have recorded it',
  '删除 ': 'Delete ',
  '该用户': 'this user',
  ' 后，账号将不可恢复。': '. This account cannot be restored.',
  '该用户拥有的网络将自动转让给当前管理员，ZeroTier 控制器内网络本身不会被删除。': 'Networks owned by this user will be transferred to the current administrator. The networks themselves will not be deleted from the ZeroTier controller.',
  '该用户当前所有登录会话会立即失效，后续请求会被强制退出。': 'All current login sessions for this user will become invalid immediately, and subsequent requests will be forced to sign out.',
  '用户已删除': 'User deleted',
  '已被删除。': 'has been deleted.',
  '转移网络：': 'Transferred networks: ',
  '吊销会话：': 'Revoked sessions: ',
  '这些网络现在已归当前管理员所有，可在网络列表中继续管理。': 'These networks are now owned by the current administrator and can continue to be managed from the network list.',
  '我知道了': 'Got it',
  '重置用户密码': 'Reset user password',
  '将为 ': 'A new random password will be generated for ',
  ' 生成一个新的随机密码，并立即吊销该用户当前所有登录会话。': ', and all current login sessions for that user will be revoked immediately.',
  '新密码只会展示一次。请通过其他安全方式告知用户，并提醒其在登录后尽快到“设置”中修改密码。': 'The new password is shown only once. Share it through another secure channel and remind the user to change it in Settings after signing in.',
  '重置中...': 'Resetting...',
  '确认重置': 'Confirm reset',
  '已吊销会话：': 'Revoked sessions: ',
  ' 个。用户收到密码后应尽快登录并修改为自己的新密码。': '. The user should sign in soon and change it to their own password.',
  '导入网络': 'Import Networks',
  '生成 Planet（实验性）': 'Generate Planet (Experimental)',
  '退出登录': 'Sign out',
  '确定要退出登录吗？': 'Are you sure you want to sign out?',
  '取消': 'Cancel',
  '确认': 'Confirm',
  '返回首页': 'Back home',
  '页面未找到': 'Page not found',
  '“休斯顿，我们有麻烦了！”': '"Houston, we have a problem!"',
  '登录到 Tairitsu': 'Sign in to Tairitsu',
  '用户名': 'Username',
  '密码': 'Password',
  '确认密码': 'Confirm password',
  '登录': 'Sign in',
  '登录中...': 'Signing in...',
  '没有账户? 去注册': 'No account? Register',
  '忘记密码请联系管理员处理': 'Contact an administrator if you forgot your password',
  '注册到 Tairitsu': 'Register for Tairitsu',
  '注册': 'Register',
  '注册中...': 'Registering...',
  '已有账户? 登录': 'Already have an account? Sign in',
  '当前实例已关闭公开注册，请联系管理员创建账户或稍后再试。': 'Public registration is disabled for this instance. Contact an administrator to create an account or try again later.',
  '用户名不能为空': 'Username is required',
  '密码不能为空': 'Password is required',
  '密码长度不能少于6位': 'Password must be at least 6 characters',
  '请确认密码': 'Please confirm the password',
  '两次输入的密码不一致': 'The two passwords do not match',
  '注册成功，请登录': 'Registration succeeded. Please sign in.',
  '注册失败，请稍后重试': 'Registration failed. Please try again later.',
  '账户安全': 'Account Security',
  '当前账号': 'Current account',
  '当前角色': 'Current role',
  '未知用户': 'Unknown user',
  '管理员': 'Administrator',
  '普通用户': 'User',
  '加载设置失败': 'Failed to load settings',
  '密码修改失败，请稍后重试': 'Failed to update password. Please try again later.',
  '密码修改成功': 'Password updated successfully',
  '当前会话已退出': 'Current session signed out',
  '已移除该登录会话': 'Session removed',
  '移除登录会话失败': 'Failed to remove login session',
  '没有其他可移除的会话': 'No other sessions to remove',
  '移除其他会话失败': 'Failed to remove other sessions',
  '修改密码': 'Change password',
  '退出其他设备': 'Sign out other devices',
  '你可以在这里修改密码，并管理当前账户的登录会话。退出其他设备会吊销同一账户在其他浏览器或机器上的登录状态。': 'Change your password and manage login sessions for this account. Signing out other devices revokes this account on other browsers or machines.',
  '登录会话': 'Login Sessions',
  '当前页面展示的是服务端登记的登录会话。移除其他会话后，对应设备会在下一次请求时失效。': 'This page shows sessions registered on the server. Removed sessions become invalid on their next request.',
  '当前没有可展示的登录会话。': 'No login sessions to display.',
  '当前会话': 'Current session',
  '其他会话': 'Other session',
  '已移除': 'Removed',
  '已过期': 'Expired',
  '退出当前会话': 'Sign out current session',
  '移除此会话': 'Remove this session',
  '最近活跃：': 'Last active: ',
  '登录时间：': 'Signed in: ',
  '到期时间：': 'Expires: ',
  '客户端标识：': 'Client identity: ',
  '原密码': 'Current password',
  '新密码': 'New password',
  '再次确认新密码': 'Confirm new password',
  '请输入原密码': 'Enter the current password',
  '请输入新密码': 'Enter a new password',
  '新密码长度至少为6位': 'New password must be at least 6 characters',
  '请再次确认新密码': 'Confirm the new password again',
  '两次输入的新密码不一致': 'The two new passwords do not match',
  '密码长度至少 6 位': 'Password must be at least 6 characters',
  '修改密码后同时退出其他设备': 'Sign out other devices after changing password',
  '建议开启。保存后会保留当前会话，并吊销当前账户在其他浏览器或机器上的登录状态。': 'Recommended. Saving keeps the current session and revokes this account on other browsers or machines.',
  '确认修改': 'Confirm change',
  '修改中...': 'Changing...',
  '处理中...': 'Processing...',
  '移除中...': 'Removing...',
  '网络管理': 'Network Management',
  '刷新': 'Refresh',
  '创建网络': 'Create Network',
  '总网络数': 'Total networks',
  '我拥有': 'Owned by me',
  '共享给我': 'Shared with me',
  '共享给我的网络暂时无法加载，当前仅显示您拥有的网络': 'Networks shared with me could not be loaded. Only networks you own are shown.',
  '获取网络列表失败': 'Failed to load networks',
  '更新网络失败': 'Failed to update network',
  '创建网络失败': 'Failed to create network',
  '删除网络失败': 'Failed to delete network',
  '搜索': 'Search',
  '清除搜索': 'Clear search',
  '还没有任何网络': 'No networks yet',
  '您可以直接创建一个新网络，或前往“导入网络”把控制器中已有的网络登记到当前账号。': 'Create a new network, or import an existing controller network into the current account.',
  '导入现有网络': 'Import existing networks',
  '没有匹配的网络': 'No matching networks',
  '当前搜索条件未匹配到任何网络，请尝试缩短关键字或清空搜索后重试。': 'No networks match the current search. Shorten the keywords or clear the search and try again.',
  '清空搜索': 'Clear search',
  '名称': 'Name',
  '网络ID': 'Network ID',
  '成员统计': 'Member stats',
  '操作': 'Actions',
  '未命名网络': 'Unnamed network',
  '只读': 'Read-only',
  '详情': 'Details',
  '查看设备': 'View devices',
  '编辑网络': 'Edit Network',
  '网络名称': 'Network name',
  '网络描述': 'Network description',
  '更新': 'Update',
  '创建': 'Create',
  '确认删除网络': 'Confirm network deletion',
  '您确定要删除网络': 'Are you sure you want to delete network',
  '吗？此操作不可恢复，将永久删除该网络及其所有配置。': '? This action cannot be undone and will permanently delete this network and all of its configuration.',
  '确认删除': 'Confirm delete',
  '删除中...': 'Deleting...',
  '成员设备': 'Member Devices',
  '成员': 'Members',
  '保存': 'Save',
  '重置更改': 'Reset changes',
  '未保存': 'Unsaved',
  '删除网络': 'Delete Network',
  '此操作不可恢复。删除网络将断开所有连接的设备，并永久删除网络配置。': 'This action cannot be undone. Deleting the network disconnects all devices and permanently removes the network configuration.',
  '编辑成员': 'Edit member',
  '授权': 'Authorize',
  '拒绝设备': 'Deny device',
  '移除成员': 'Remove member',
  '确认移除成员': 'Confirm member removal',
  '移除': 'Remove',
  '设备详情': 'Device Details',
  '设备 ID': 'Device ID',
  '复制': 'Copy',
  '设备名称': 'Device name',
  '设备已授权': 'Device authorized',
  '设备待授权': 'Device pending authorization',
  '成员元信息': 'Member Metadata',
  '节点地址': 'Node address',
  '在线状态': 'Online status',
  '在线': 'Online',
  '离线': 'Offline',
  '身份标识': 'Identity',
  '加入时间': 'Join time',
  'ZeroTier 版本': 'ZeroTier version',
  '桥接模式': 'Bridge mode',
  '已启用': 'Enabled',
  '未启用': 'Disabled',
  '自动分配 IP': 'Auto-assign IP',
  '已禁止': 'Disabled',
  '允许': 'Allowed',
  '能力': 'Capabilities',
  '标签': 'Tags',
  'Peer 角色': 'Peer role',
  '当前路径': 'Current path',
  '延迟': 'Latency',
  '添加 IP': 'Add IP',
  '禁止自动分配 IP': 'Disable auto-assigned IPs',
  '允许以桥接设备身份接入': 'Allow bridged device access',
  '网络基本信息': 'Basic Network Info',
  'IPv4分配': 'IPv4 Assignment',
  'IPv6分配': 'IPv6 Assignment',
  '多播设置': 'Multicast Settings',
  'DNS 设置': 'DNS Settings',
  '托管路由': 'Managed Routes',
  '网络不存在': 'Network not found',
  '返回网络列表': 'Back to networks',
  '网络ID不能为空': 'Network ID is required',
  '获取网络详情失败': 'Failed to load network details',
  '只读授权信息暂时无法加载，其他网络内容仍可正常查看': 'Read-only access data could not be loaded, but other network content is still available.',
  '设备授权成功': 'Device authorized',
  '设备已拒绝': 'Device denied',
  '设备授权失败': 'Failed to authorize device',
  '设备拒绝失败': 'Failed to deny device',
  '成员信息更新成功': 'Member updated',
  '更新成员失败': 'Failed to update member',
  '成员移除成功': 'Member removed',
  '移除成员失败': 'Failed to remove member',
  '设备 ID 已复制': 'Device ID copied',
  '复制设备 ID 失败': 'Failed to copy device ID',
  '网络 ID 已复制': 'Network ID copied',
  '复制网络 ID 失败': 'Failed to copy network ID',
  '名称和描述保存成功': 'Name and description saved',
  '保存失败': 'Failed to save',
  '删除失败': 'Failed to delete',
  'IPv4配置保存成功': 'IPv4 configuration saved',
  'IPv6配置保存成功': 'IPv6 configuration saved',
  '托管路由保存成功': 'Managed routes saved',
  'DNS 设置保存成功': 'DNS settings saved',
  '多播设置保存成功': 'Multicast settings saved',
  '该地址池范围已存在': 'This address pool range already exists',
  '该 IPv6 地址池范围已存在': 'This IPv6 address pool range already exists',
  '该托管路由目标已存在': 'This managed route target already exists',
  'DNS 服务器必须是有效的 IPv4 或 IPv6 地址': 'DNS server must be a valid IPv4 or IPv6 address',
  '该 DNS 服务器已存在': 'This DNS server already exists',
  '只读查看权限已授予': 'Read-only access granted',
  '授予只读查看权限失败': 'Failed to grant read-only access',
  '只读查看权限已移除': 'Read-only access removed',
  '移除只读查看权限失败': 'Failed to remove read-only access',
  '只读查看授权': 'Read-only viewer access',
  '这里可以授予普通用户只读查看该网络成员设备的权限。被授权用户只能查看设备和元信息，不能修改任何内容。': 'Grant regular users read-only access to this network member devices. Authorized users can view devices and metadata, but cannot modify anything.',
  '选择用户': 'Select user',
  '授予只读查看': 'Grant read-only access',
  '当前还没有被授予只读查看权限的用户。': 'No users have read-only access yet.',
  '授权时间：': 'Granted at: ',
  '移除权限': 'Remove access',
  '授权第一个待审批成员': 'Authorize first pending member',
  '拒绝第一个待审批成员': 'Deny first pending member',
  '关闭': 'Close',
  '当前有 ': 'There are ',
  ' 个待授权设备。你可以在此快速审批，也可以在下方成员列表中逐个处理。': ' pending devices. You can review one quickly here, or handle them individually in the member list below.',
  '设备总数': 'Total devices',
  '已授权设备': 'Authorized devices',
  '待授权设备': 'Pending devices',
  '搜索设备名称或设备 ID': 'Search device name or device ID',
  '没有找到匹配的成员设备': 'No matching member devices found',
  '暂无设备连接': 'No connected devices yet',
  '未命名设备': 'Unnamed device',
  '已授权': 'Authorized',
  '待授权': 'Pending',
  '打开成员菜单：': 'Open member menu: ',
  '账户信息': 'Account Info',
  '创建时间': 'Created at',
  '更新时间': 'Updated at',
  '打开账户设置与修改密码': 'Open account settings and change password',
  '用户信息不可用': 'User info unavailable',
  '获取数据失败，请稍后重试': 'Failed to load data. Please try again later.',
  '无法获取网络总览统计信息': 'Unable to retrieve network overview statistics',
  '无法获取系统资源统计信息': 'Unable to retrieve system resource statistics',
  '重试': 'Retry',
  '控制器总览': 'Controller Overview',
  '网络总数': 'Total Networks',
  '成员总数': 'Total Members',
  '已授权成员': 'Authorized Members',
  '待授权成员': 'Pending Members',
  '系统健康监控': 'System Health',
  'CPU使用率': 'CPU Usage',
  '内存使用率': 'Memory Usage',
  '无法获取': 'Unavailable',
  '操作系统信息': 'Operating System',
  '控制器详情': 'Controller Details',
  '控制器地址': 'Controller Address',
  '版本': 'Version',
  '控制器状态': 'Controller Status',
  '数据库状态': 'Database Status',
  '错误': 'Error',
  '已连接': 'Connected',
  '未连接': 'Disconnected',
  '欢迎使用 Tairitsu': 'Welcome to Tairitsu',
  '配置 ZeroTier 控制器': 'Configure ZeroTier Controller',
  '配置数据库': 'Configure Database',
  '创建管理员账户': 'Create Administrator',
  '完成设置': 'Finish Setup',
  'Tairitsu 初始化向导': 'Tairitsu Setup Wizard',
  '开始初始化': 'Start setup',
  '下一步': 'Next',
  '返回': 'Back',
  '测试并保存': 'Test and save',
  '保存数据库配置': 'Save database config',
  '创建首个管理员': 'Create first administrator',
  '完成初始化': 'Finish setup',
  '已保存': 'Saved',
  '未知步骤': 'Unknown step',
  '数据库类型': 'Database type',
  'SQLite 文件路径': 'SQLite file path',
  '认证令牌文件路径': 'Auth token file path',
  '身份加载': 'Identity Loading',
  'Planet 配置': 'Planet Configuration',
  '高级模式': 'Advanced Mode',
  '读取身份': 'Read identity',
  '添加端点': 'Add endpoint',
  '端点': 'Endpoint',
  '需人工处理': 'Needs attention',
  '控制器接管台': 'Controller Claim Console',
  '查看首个网络': 'View first network',
  '查看网络': 'View networks',
  '该页面用于接管控制器中已存在但尚未纳入 Tairitsu 管理的网络，并将其分配给指定 owner。已接管网络会保留在控制器中，仅补齐 Tairitsu 侧登记和归属关系。': 'Use this page to claim networks that already exist in the controller but are not managed by Tairitsu, then assign them to an owner. Claimed networks stay in the controller; Tairitsu only records ownership and management metadata.',
  '控制器网络总数': 'Controller networks',
  '可接管': 'Claimable',
  '已接管': 'Claimed',
  '获取控制器接管台信息失败': 'Failed to load controller claim console',
  '请至少选择一个可接管网络': 'Select at least one claimable network',
  '请选择目标 owner': 'Select a target owner',
  '接管设置': 'Claim Settings',
  '先选择导入后的 owner。只有“可接管”状态的网络才会进入当前批次，已被其他 owner 接管或控制器读取异常的网络不会被强行导入。': 'Choose the owner after import first. Only claimable networks are included in this batch; networks claimed by another owner or with controller read errors will not be forced in.',
  '目标 owner': 'Target owner',
  '最近一次接管结果': 'Latest Claim Result',
  '请求 ': 'Requested ',
  ' 个，成功 ': ', succeeded ',
  ' 个，失败 ': ', failed ',
  ' 个，跳过 ': ', skipped ',
  ' 个。': '.',
  '已成功接管': 'Claimed successfully',
  '已分配给 ': 'Assigned to ',
  '失败项': 'Failed items',
  '可接管网络': 'Claimable Networks',
  '这些网络尚未纳入 Tairitsu 或尚未分配 owner，可直接选择并分配给目标 owner。': 'These networks are not yet managed by Tairitsu or have no owner. You can select them and assign them to the target owner.',
  '已选 ': 'Selected ',
  '当前没有待接管网络，控制器中的网络都已经被 Tairitsu 管理或需要人工处理。': 'No networks are waiting to be claimed. Controller networks are already managed by Tairitsu or need manual attention.',
  '取消全选': 'Clear selection',
  '全选可接管网络': 'Select all claimable',
  '接管中...': 'Claiming...',
  '接管所选网络': 'Claim selected networks',
  '状态': 'Status',
  '说明': 'Notes',
  '当前没有已接管网络。': 'No claimed networks yet.',
  '这些网络不会出现在当前导入批次中，通常是因为控制器详情读取失败或状态异常。请先排查控制器连通性，再刷新列表重试。': 'These networks are excluded from the current import batch, usually because controller details could not be read or the state is abnormal. Check controller connectivity, then refresh and try again.',
  '当前没有需要人工处理的网络。': 'No networks need manual attention.',
  '已被 Tairitsu 接管': 'Managed by Tairitsu',
  '同一批次中重复提交了该网络，已跳过重复项': 'This network was submitted more than once in the same batch and the duplicate was skipped',
  '网络不存在于 ZeroTier 控制器中': 'Network does not exist in the ZeroTier controller',
  '读取 ZeroTier 网络详情失败': 'Failed to read ZeroTier network details',
  '写入数据库失败': 'Failed to write to the database',
  '更新网络所有者失败': 'Failed to update network owner',
  '网络已由其他 owner 接管，已跳过': 'Network is already managed by another owner and was skipped',
  '目标 owner 已拥有该网络，已跳过重复导入': 'Target owner already owns this network and the duplicate import was skipped',
  '网络尚未登记到 Tairitsu，可直接接管': 'Network is not registered in Tairitsu and can be claimed directly',
  '网络已登记但尚未分配 owner，可继续接管': 'Network is registered but has no owner and can be claimed',
  '网络已由其他 owner 接管': 'Network is already managed by another owner',
  '读取控制器网络详情失败，暂时无法导入': 'Controller network details could not be read, so this network cannot be imported yet',
  '已跳过': 'Skipped',
  '导入失败': 'Import failed',
}

const rawZhDuplicateTargets: Record<string, string> = {
  'Loading...': '加载中...',
  'Password must be at least 6 characters': '密码长度至少 6 位',
  'Clear search': '清空搜索',
  'Disabled': '已禁用',
  'Revoked sessions: ': '吊销会话：',
  'Device authorized': '设备已授权',
}

function invertRawDictionary(dictionary: Record<string, string>, duplicateTargets: Record<string, string>): Record<string, string> {
  const inverted: Record<string, string> = {}
  const duplicates: string[] = []

  for (const [source, target] of Object.entries(dictionary)) {
    if (Object.prototype.hasOwnProperty.call(inverted, target)) {
      if (!Object.prototype.hasOwnProperty.call(duplicateTargets, target)) {
        duplicates.push(target)
      }
      continue
    }
    inverted[target] = source
  }

  if (duplicates.length > 0) {
    throw new Error(`Duplicate raw i18n target values: ${[...new Set(duplicates)].join(', ')}`)
  }

  return { ...inverted, ...duplicateTargets }
}

const rawZh = invertRawDictionary(rawEn, rawZhDuplicateTargets)

export function detectSystemLanguage(languages?: readonly string[]): Language {
  const candidates: readonly string[] = languages && languages.length > 0
    ? languages
    : [globalThis.navigator?.language, ...(globalThis.navigator?.languages ?? [])].filter(Boolean)

  return candidates.some((language) => {
    const normalized = language.toLowerCase()
    if (normalized === 'zh-tw' || normalized === 'zh-hk' || normalized === 'zh-mo' || normalized === 'zh-hant' || normalized.startsWith('zh-hant-')) {
      return false
    }
    return normalized === 'zh' || normalized === 'zh-cn' || normalized === 'zh-hans' || normalized.startsWith('zh-hans-') || normalized.startsWith('zh-')
  }) ? 'zh-CN' : 'en'
}

export function normalizeLanguagePreference(value: unknown): LanguagePreference {
  return value === 'en' || value === 'zh-CN' || value === 'system' ? value : 'system'
}

export function resolveLanguage(preference: LanguagePreference): Language {
  return preference === 'system' ? detectSystemLanguage() : preference
}

export function getStoredLanguagePreference(): LanguagePreference {
  try {
    return normalizeLanguagePreference(localStorage.getItem(STORAGE_KEY))
  } catch {
    return 'system'
  }
}

function storeLanguagePreference(preference: LanguagePreference) {
  try {
    localStorage.setItem(STORAGE_KEY, preference)
  } catch {
    // Storage may be unavailable in private or test environments.
  }
}

export function translateRawText(value: string, language: Language): string {
  const normalized = value.replace(/\s+/g, ' ').trim()
  if (!normalized) return value

  const dictionary = language === 'en' ? rawEn : rawZh
  const translated = dictionary[normalized]
  if (translated) {
    return value.split(normalized).join(translated)
  }

  const rawReplacementPairs: Array<[RegExp, string, RegExp, string]> = [
    [/共享来源：/g, 'Shared by: ', /Shared by: /g, '共享来源：'],
    [/网络ID:/g, 'Network ID:', /Network ID:/g, '网络ID:'],
    [/网络 ID/g, 'Network ID', /Network ID/g, '网络 ID'],
    [/已授权 /g, 'Authorized ', /Authorized /g, '已授权 '],
    [/待授权 /g, 'Pending ', /Pending /g, '待授权 '],
    [/ 台/g, ' devices', / devices/g, ' 台'],
    [/最近活跃：/g, 'Last active: ', /Last active: /g, '最近活跃：'],
    [/登录时间：/g, 'Signed in: ', /Signed in: /g, '登录时间：'],
    [/到期时间：/g, 'Expires: ', /Expires: /g, '到期时间：'],
    [/授权时间：/g, 'Granted at: ', /Granted at: /g, '授权时间：'],
    [/当前步骤状态：/g, 'Current step status: ', /Current step status: /g, '当前步骤状态：'],
    [/平台:/g, 'Platform:', /Platform:/g, '平台:'],
    [/内核:/g, 'Kernel:', /Kernel:/g, '内核:'],
    [/密码修改成功，并已移除其他会话 /g, 'Password updated successfully and removed ', /Password updated successfully and removed /g, '密码修改成功，并已移除其他会话 '],
    [/已移除其他会话 /g, 'Removed ', /Removed /g, '已移除其他会话 '],
    [/ 个/g, ' sessions', / sessions/g, ' 个'],
    [/IP：/g, 'IP: ', /IP: /g, 'IP：'],
    [/客户端标识：/g, 'Client identity: ', /Client identity: /g, '客户端标识：'],
    [/持久登录/g, 'Remembered login', /Remembered login/g, '持久登录'],
    [/会话登录/g, 'Session login', /Session login/g, '会话登录'],
    [/平板设备/g, 'Tablet', /Tablet/g, '平板设备'],
    [/移动设备/g, 'Mobile device', /Mobile device/g, '移动设备'],
    [/桌面设备/g, 'Desktop device', /Desktop device/g, '桌面设备'],
    [/未知浏览器/g, 'Unknown browser', /Unknown browser/g, '未知浏览器'],
    [/未知系统/g, 'Unknown platform', /Unknown platform/g, '未知系统'],
    [/未知设备/g, 'Unknown device', /Unknown device/g, '未知设备'],
    [/未知/g, 'Unknown', /Unknown/g, '未知'],
    [/已将 /g, 'Assigned ', /Assigned /g, '已将 '],
    [/ 个网络分配给 /g, ' networks to ', / networks to /g, ' 个网络分配给 '],
    [/，控制器接管列表已刷新。/g, '. The controller claim list has been refreshed.', /\. The controller claim list has been refreshed\./g, '，控制器接管列表已刷新。'],
    [/已导入 /g, 'Imported ', /Imported /g, '已导入 '],
    [/ 个网络，失败 /g, ' networks, failed ', / networks, failed /g, ' 个网络，失败 '],
    [/ 个，跳过 /g, ', skipped ', /, skipped /g, ' 个，跳过 '],
    [/没有成功导入任何网络，失败 /g, 'No networks were imported. Failed ', /No networks were imported\. Failed /g, '没有成功导入任何网络，失败 '],
    [/已创建用户 /g, 'Created user ', /Created user /g, '已创建用户 '],
    [/，请立即通过其他方式告知其临时密码/g, '. Share the temporary password through another channel right away', /\. Share the temporary password through another channel right away/g, '，请立即通过其他方式告知其临时密码'],
    [/已为 /g, 'Generated a new temporary password for ', /Generated a new temporary password for /g, '已为 '],
    [/ 生成新的临时密码，并吊销 /g, ' and revoked ', / and revoked /g, ' 生成新的临时密码，并吊销 '],
    [/ 个现有会话/g, ' existing sessions', / existing sessions/g, ' 个现有会话'],
    [/已删除用户 /g, 'Deleted user ', /Deleted user /g, '已删除用户 '],
    [/，并转移 /g, ' and transferred ', / and transferred /g, '，并转移 '],
    [/管理员身份已转让给 /g, 'Administrator role transferred to ', /Administrator role transferred to /g, '管理员身份已转让给 '],
  ]

  return rawReplacementPairs.reduce((result, [zhPattern, enText, enPattern, zhText]) => {
    return language === 'en' ? result.replace(zhPattern, enText) : result.replace(enPattern, zhText)
  }, value)
}

export function translateMessageCode(code: string, language = resolveLanguage(getStoredLanguagePreference())): string | null {
  return messageCodes[code]?.[language] ?? null
}

interface TranslationContextValue {
  language: Language
  preference: LanguagePreference
  setPreference: (preference: LanguagePreference) => void
  t: (key: string, params?: Record<string, string | number>) => string
  formatDateTime: (value: string | number | Date) => string
  translateText: (value: string) => string
}

const TranslationContext = createContext<TranslationContextValue | null>(null)

function interpolate(template: string, params?: Record<string, string | number>): string {
  if (!params) return template
  return Object.entries(params).reduce(
    (result, [key, value]) => result.split(`{{${key}}}`).join(String(value)),
    template,
  )
}

const legacyTranslationSelector = '[data-legacy-translate], .legacy-i18n'

function legacyTranslationRoots(root: ParentNode): Element[] {
  const roots: Element[] = []
  if (root instanceof Element && root.matches(legacyTranslationSelector)) {
    roots.push(root)
  }
  if (root instanceof Element || root instanceof Document) {
    roots.push(...Array.from(root.querySelectorAll(legacyTranslationSelector)))
  }
  return roots
}

function isInLegacyTranslationRoot(node: Node): boolean {
  return Boolean(node.parentElement?.closest(legacyTranslationSelector))
}

function translateDocument(root: ParentNode, language: Language) {
  const ignoredTags = new Set(['SCRIPT', 'STYLE', 'TEXTAREA'])
  for (const legacyRoot of legacyTranslationRoots(root)) {
    const walker = document.createTreeWalker(legacyRoot, NodeFilter.SHOW_TEXT)
    let node = walker.nextNode()
    while (node) {
      const parent = node.parentElement
      if (parent && !ignoredTags.has(parent.tagName)) {
        const nextValue = translateRawText(node.nodeValue ?? '', language)
        if (nextValue !== node.nodeValue) {
          node.nodeValue = nextValue
        }
      }
      node = walker.nextNode()
    }

    legacyRoot.querySelectorAll<HTMLElement>('[placeholder],[aria-label],[title]').forEach((element) => {
      for (const attr of ['placeholder', 'aria-label', 'title']) {
        const value = element.getAttribute(attr)
        if (value) {
          element.setAttribute(attr, translateRawText(value, language))
        }
      }
    })

    if (legacyRoot instanceof HTMLElement) {
      for (const attr of ['placeholder', 'aria-label', 'title']) {
        const value = legacyRoot.getAttribute(attr)
        if (value) {
          legacyRoot.setAttribute(attr, translateRawText(value, language))
        }
      }
    }
  }
}

function RuntimeTextTranslator({ language }: { language: Language }) {
  useEffect(() => {
    const root = document.getElementById('root') ?? document.body
    const scheduleIdle = (callback: () => void) => {
      const idleCallback = globalThis.requestIdleCallback
      if (idleCallback) {
        const id = idleCallback(callback)
        return () => globalThis.cancelIdleCallback?.(id)
      }
      const id = window.setTimeout(callback, 0)
      return () => window.clearTimeout(id)
    }

    let queuedMutations: MutationRecord[] = []
    let cancelScheduledWork: (() => void) | null = null

    const flushMutations = () => {
      const mutations = queuedMutations
      queuedMutations = []
      cancelScheduledWork = null

      try {
        for (const mutation of mutations) {
          mutation.addedNodes.forEach((node) => {
            if (node.nodeType === Node.TEXT_NODE && isInLegacyTranslationRoot(node)) {
              const translated = translateRawText(node.nodeValue ?? '', language)
              if (translated !== node.nodeValue) {
                node.nodeValue = translated
              }
            } else if (node instanceof Element) {
              translateDocument(node, language)
            }
          })
          if (mutation.type === 'characterData' && isInLegacyTranslationRoot(mutation.target)) {
            const translated = translateRawText(mutation.target.nodeValue ?? '', language)
            if (translated !== mutation.target.nodeValue) {
              mutation.target.nodeValue = translated
            }
          }
        }
      } catch (error) {
        console.error('Runtime text translation failed', error)
      }
    }

    const cancelInitialTranslation = scheduleIdle(() => {
      try {
        translateDocument(root, language)
      } catch (error) {
        console.error('Initial runtime text translation failed', error)
      }
    })
    const observer = new MutationObserver((mutations) => {
      try {
        queuedMutations.push(...mutations)
        if (!cancelScheduledWork) {
          cancelScheduledWork = scheduleIdle(flushMutations)
        }
      } catch (error) {
        console.error('Runtime text translation observer failed', error)
        cancelScheduledWork = null
      }
    })
    observer.observe(root, {
      childList: true,
      subtree: true,
      characterData: true,
    })
    return () => {
      observer.disconnect()
      cancelInitialTranslation()
      cancelScheduledWork?.()
    }
  }, [language])

  return null
}

export function LanguageProvider({ children }: { children: React.ReactNode }) {
  const [preference, setPreferenceState] = useState<LanguagePreference>(() => getStoredLanguagePreference())
  const [systemLanguageTick, setSystemLanguageTick] = useState(0)
  const language = useMemo(
    () => (preference === 'system' ? detectSystemLanguage() : preference),
    [preference, systemLanguageTick],
  )

  useEffect(() => {
    document.documentElement.lang = language === 'zh-CN' ? 'zh-CN' : 'en'
  }, [language])

  useEffect(() => {
    if (preference !== 'system') return

    const onLanguageChange = () => setSystemLanguageTick((current) => current + 1)
    window.addEventListener('languagechange', onLanguageChange)
    return () => {
      window.removeEventListener('languagechange', onLanguageChange)
    }
  }, [preference])

  const setPreference = useCallback((nextPreference: LanguagePreference) => {
    const normalized = normalizeLanguagePreference(nextPreference)
    storeLanguagePreference(normalized)
    setPreferenceState(normalized)
  }, [])

  const value = useMemo<TranslationContextValue>(() => {
    const dictionary = language === 'zh-CN' ? zh : en
    return {
      language,
      preference,
      setPreference,
      t: (key, params) => interpolate(dictionary[key] ?? en[key] ?? key, params),
      formatDateTime: (valueToFormat) => new Intl.DateTimeFormat(language === 'zh-CN' ? 'zh-CN' : 'en', {
        dateStyle: 'medium',
        timeStyle: 'short',
      }).format(new Date(valueToFormat)),
      translateText: (valueToTranslate) => translateRawText(valueToTranslate, language),
    }
  }, [language, preference, setPreference])

  const theme = useMemo(
    () => createTheme(baseTheme, language === 'zh-CN' ? zhCN : enUS),
    [language],
  )

  return (
    <TranslationContext.Provider value={value}>
      <ThemeProvider theme={theme}>
        <RuntimeTextTranslator language={language} />
        {children}
      </ThemeProvider>
    </TranslationContext.Provider>
  )
}

export function useTranslation() {
  const value = useContext(TranslationContext)
  if (!value) {
    throw new Error('useTranslation must be used within LanguageProvider')
  }
  return value
}
