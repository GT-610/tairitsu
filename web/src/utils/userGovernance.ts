import type { CreateUserResponse, DeleteUserResponse, ResetUserPasswordResponse } from '../services/api'

export function buildCreateUserSuccessMessage(result: CreateUserResponse): string {
  return `已创建用户 ${result.user.username}，请立即通过其他方式告知其临时密码`
}

export function buildResetPasswordSuccessMessage(result: ResetUserPasswordResponse): string {
  return `已为 ${result.user.username} 生成新的临时密码，并吊销 ${result.revoked_sessions} 个现有会话`
}

export function buildDeleteUserSuccessMessage(result: DeleteUserResponse): string {
  return `已删除用户 ${result.user.username}，并转移 ${result.transferred_networks} 个网络`
}
