import { describe, expect, test } from 'bun:test'
import { setupWizardDatabaseStepCopy } from './setupWizard'

describe('setupWizard copy', () => {
  test('keeps SQLite-only guidance explicit', () => {
    expect(setupWizardDatabaseStepCopy.description).toContain('SQLite')
    expect(setupWizardDatabaseStepCopy.supportAlert).toContain('MySQL')
    expect(setupWizardDatabaseStepCopy.databaseTypeHelperText).toContain('SQLite')
  })
})
