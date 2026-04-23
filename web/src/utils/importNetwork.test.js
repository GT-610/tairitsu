import { describe, expect, test } from 'bun:test'
import { buildImportResultFeedback, groupImportCandidates } from './importNetwork'

describe('importNetwork utils', () => {
  test('groups candidates by status', () => {
    const grouped = groupImportCandidates([
      { network_id: '1', status: 'available', can_import: true, reason_code: 'unregistered', reason_message: 'ok' },
      { network_id: '2', status: 'managed', can_import: false, reason_code: 'already_managed', reason_message: 'managed' },
      { network_id: '3', status: 'blocked', can_import: false, reason_code: 'controller_read_failed', reason_message: 'blocked' },
    ])

    expect(grouped.available).toHaveLength(1)
    expect(grouped.managed).toHaveLength(1)
    expect(grouped.blocked).toHaveLength(1)
  })

  test('builds success feedback for full success', () => {
    const feedback = buildImportResultFeedback({
      target_owner: { id: 'owner-1', username: 'alice' },
      summary: { requested: 2, imported: 2, failed: 0, skipped: 0 },
      imported: [{ network_id: '1' }, { network_id: '2' }],
      failed: [],
      skipped: [],
    })

    expect(feedback.severity).toBe('success')
    expect(feedback.text).toContain('2')
    expect(feedback.text).toContain('alice')
  })

  test('builds error feedback for all-failed result', () => {
    const feedback = buildImportResultFeedback({
      target_owner: { id: 'owner-1', username: 'alice' },
      summary: { requested: 2, imported: 0, failed: 2, skipped: 0 },
      imported: [],
      failed: [{ network_id: '1' }, { network_id: '2' }],
      skipped: [],
    })

    expect(feedback.severity).toBe('error')
    expect(feedback.text).toContain('失败 2 个')
  })
})
