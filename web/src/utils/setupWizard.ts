import type { SetupStatus } from '../services/api'

export const setupWizardDatabaseStepCopy = {
  description: '当前仅支持 SQLite 单实例部署。数据库文件将由程序使用默认路径管理。',
  supportAlert: 'MySQL 与 PostgreSQL 相关抽象暂时保留，但当前不在支持范围内。',
  databaseTypeHelperText: '当前固定为 SQLite',
  databasePathHelperText: '留空则使用默认值 data/tairitsu.db',
}

export function getInitialSetupWizardStep(status: SetupStatus): number {
  if (!status.zerotierConfigured) {
    return 1
  }
  if (!status.databaseConfigured) {
    return 2
  }
  if (!status.hasAdmin) {
    return 3
  }
  return 4
}

export function isSetupStepSaved(status: SetupStatus, step: number): boolean {
  switch (step) {
    case 1:
      return status.zerotierConfigured
    case 2:
      return status.databaseConfigured
    case 3:
      return status.hasAdmin
    case 4:
      return status.initialized
    default:
      return false
  }
}
