export function isPublicRegistrationEnabled(allowPublicRegistration: boolean | null | undefined): boolean {
  return allowPublicRegistration !== false
}

export function getRegistrationClosedMessage(allowPublicRegistration: boolean | null | undefined): string {
  return isPublicRegistrationEnabled(allowPublicRegistration) ? '' : '公开注册已关闭，请联系管理员创建账户'
}
