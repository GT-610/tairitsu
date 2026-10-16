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

// The wizard's first step collects the setup token, so a step derived from the
// stored configuration sits one further along. Without a valid token the server
// withholds that configuration, and the only honest place to resume is the
// token step itself.
export function getResumableSetupWizardStep(status: SetupStatus): number {
  if (!status.setupAuthorized) {
    return 0
  }
  return Math.max(1, getInitialSetupWizardStep(status))
}
