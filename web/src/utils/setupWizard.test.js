import { describe, expect, test } from 'bun:test'
import { getInitialSetupWizardStep, getResumableSetupWizardStep } from './setupWizard'

describe('setupWizard', () => {
  test('derives the first incomplete setup step from backend status', () => {
    expect(getInitialSetupWizardStep({
      initialized: false,
      databaseConfigured: false,
      hasAdmin: false,
      zerotierConfigured: false,
      adminCreationPrepared: false,
      allowPublicRegistration: true,
    })).toBe(0)

    expect(getInitialSetupWizardStep({
      initialized: false,
      databaseConfigured: true,
      hasAdmin: false,
      zerotierConfigured: true,
      adminCreationPrepared: true,
      allowPublicRegistration: true,
    })).toBe(3)

    expect(getInitialSetupWizardStep({
      initialized: false,
      databaseConfigured: true,
      hasAdmin: true,
      zerotierConfigured: false,
      adminCreationPrepared: true,
      allowPublicRegistration: true,
    })).toBe(1)

    expect(getInitialSetupWizardStep({
      initialized: false,
      databaseConfigured: false,
      hasAdmin: true,
      zerotierConfigured: true,
      adminCreationPrepared: true,
      allowPublicRegistration: true,
    })).toBe(2)

    expect(getInitialSetupWizardStep({
      initialized: false,
      databaseConfigured: true,
      hasAdmin: true,
      zerotierConfigured: true,
      adminCreationPrepared: true,
      allowPublicRegistration: true,
    })).toBe(4)
  })

  // The setup endpoints reject callers without the token, so nothing that needs
  // the stored configuration may be shown before it is accepted.
  test('keeps an unauthorized caller on the token step', () => {
    expect(getResumableSetupWizardStep({
      initialized: false,
      databaseConfigured: true,
      setupAuthorized: false,
      hasAdmin: true,
      zerotierConfigured: true,
      adminCreationPrepared: true,
      allowPublicRegistration: true,
    })).toBe(0)
  })

  test('resumes one step past the configured progress once authorized', () => {
    const base = {
      initialized: false,
      allowPublicRegistration: true,
      setupAuthorized: true,
    }

    expect(getResumableSetupWizardStep({
      ...base,
      databaseConfigured: false,
      hasAdmin: false,
      zerotierConfigured: false,
      adminCreationPrepared: false,
    })).toBe(1)

    expect(getResumableSetupWizardStep({
      ...base,
      databaseConfigured: false,
      hasAdmin: false,
      zerotierConfigured: true,
      adminCreationPrepared: false,
    })).toBe(2)

    expect(getResumableSetupWizardStep({
      ...base,
      databaseConfigured: true,
      hasAdmin: false,
      zerotierConfigured: true,
      adminCreationPrepared: true,
    })).toBe(3)

    expect(getResumableSetupWizardStep({
      ...base,
      databaseConfigured: true,
      hasAdmin: true,
      zerotierConfigured: true,
      adminCreationPrepared: true,
    })).toBe(4)
  })
})
