import { describe, expect, test } from 'bun:test'
import { getInitialSetupWizardStep, isSetupStepSaved, setupWizardDatabaseStepCopy } from './setupWizard'

describe('setupWizard copy', () => {
  test('keeps SQLite-only guidance explicit', () => {
    expect(setupWizardDatabaseStepCopy.description).toContain('SQLite')
    expect(setupWizardDatabaseStepCopy.supportAlert).toContain('MySQL')
    expect(setupWizardDatabaseStepCopy.databaseTypeHelperText).toContain('SQLite')
  })

  test('derives the first incomplete setup step from backend status', () => {
    expect(getInitialSetupWizardStep({
      initialized: false,
      hasDatabase: false,
      databaseConfigured: false,
      hasAdmin: false,
      zerotierConfigured: false,
      adminCreationPrepared: false,
      allowPublicRegistration: true,
    })).toBe(1)

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

  test('marks setup steps as saved from backend state', () => {
    const status = {
      initialized: false,
      hasDatabase: true,
      databaseConfigured: true,
      hasAdmin: true,
      zerotierConfigured: true,
      adminCreationPrepared: true,
      allowPublicRegistration: true,
    }

    expect(isSetupStepSaved(status, 1)).toBe(true)
    expect(isSetupStepSaved(status, 2)).toBe(true)
    expect(isSetupStepSaved(status, 3)).toBe(true)
    expect(isSetupStepSaved(status, 4)).toBe(false)
  })
})
