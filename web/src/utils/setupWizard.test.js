import { describe, expect, test } from 'bun:test'
import { getInitialSetupWizardStep, getResumableSetupWizardStep } from './setupWizard'

/**
 * Builds a complete setup-status payload, overriding only what a case cares about.
 *
 * @param {Partial<import('../services/api').SetupStatus>} overrides
 * @returns {import('../services/api').SetupStatus}
 */
function status(overrides) {
  return {
    initialized: false,
    databaseConfigured: false,
    setupAuthorized: true,
    hasAdmin: false,
    zerotierConfigured: false,
    adminCreationPrepared: false,
    allowPublicRegistration: true,
    ...overrides,
  }
}

describe('setupWizard', () => {
  test('derives the first incomplete setup step from backend status', () => {
    expect(getInitialSetupWizardStep(status({}))).toBe(0)
    expect(getInitialSetupWizardStep(status({ databaseConfigured: true }))).toBe(1)
    expect(getInitialSetupWizardStep(status({ zerotierConfigured: true }))).toBe(2)
    expect(getInitialSetupWizardStep(status({
      databaseConfigured: true,
      zerotierConfigured: true,
    }))).toBe(3)
    expect(getInitialSetupWizardStep(status({
      databaseConfigured: true,
      zerotierConfigured: true,
      hasAdmin: true,
    }))).toBe(4)
  })

  // The setup endpoints reject callers without the token, so nothing that needs
  // the stored configuration may be shown before it is accepted.
  test('keeps an unauthorized caller on the token step', () => {
    expect(getResumableSetupWizardStep(status({
      setupAuthorized: false,
      databaseConfigured: true,
      zerotierConfigured: true,
      hasAdmin: true,
    }))).toBe(0)
  })

  test('resumes one step past the configured progress once authorized', () => {
    expect(getResumableSetupWizardStep(status({}))).toBe(1)
    expect(getResumableSetupWizardStep(status({ zerotierConfigured: true }))).toBe(2)
    expect(getResumableSetupWizardStep(status({
      zerotierConfigured: true,
      databaseConfigured: true,
    }))).toBe(3)
    expect(getResumableSetupWizardStep(status({
      zerotierConfigured: true,
      databaseConfigured: true,
      hasAdmin: true,
    }))).toBe(4)
  })
})
