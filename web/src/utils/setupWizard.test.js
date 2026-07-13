import { describe, expect, test } from 'bun:test'
import { getInitialSetupWizardStep } from './setupWizard'

describe('setupWizard', () => {
  test('derives the first incomplete setup step from backend status', () => {
    expect(getInitialSetupWizardStep({
      initialized: false,
      hasDatabase: false,
      databaseConfigured: false,
      hasAdmin: false,
      zerotierConfigured: false,
      adminCreationPrepared: false,
      allowPublicRegistration: true,
    })).toBe(0)

    expect(getInitialSetupWizardStep({
      initialized: false,
      hasDatabase: true,
      databaseConfigured: true,
      hasAdmin: false,
      zerotierConfigured: true,
      adminCreationPrepared: true,
      allowPublicRegistration: true,
    })).toBe(3)
  })
})
