import type { SetupStatus } from '../services/api'

export function getInitialSetupWizardStep(status: SetupStatus): number {
  if (!status.zerotierConfigured && !status.databaseConfigured && !status.hasAdmin) {
    return 0
  }
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
