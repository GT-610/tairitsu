import { describe, expect, test } from 'bun:test'
import { normalizeImportableNetworksResponse } from './importNetworkConsole'

describe('importNetworkConsole utils', () => {
  test('returns empty fallback for invalid payload', () => {
    expect(normalizeImportableNetworksResponse(null)).toEqual({
      candidates: [],
      summary: { total: 0, available: 0, managed: 0, blocked: 0 },
    })
  })

  test('derives summary counts when backend summary is missing', () => {
    const normalized = normalizeImportableNetworksResponse({
      candidates: [
        { network_id: '1', status: 'available', can_import: true, reason_code: 'unregistered', reason_message: 'ok' },
        { network_id: '2', status: 'managed', can_import: false, reason_code: 'already_managed', reason_message: 'managed' },
        { network_id: '3', status: 'blocked', can_import: false, reason_code: 'controller_read_failed', reason_message: 'blocked' },
      ],
    })

    expect(normalized.summary).toEqual({
      total: 3,
      available: 1,
      managed: 1,
      blocked: 1,
    })
  })

  test('preserves provided summary and fills only missing numeric fields', () => {
    const normalized = normalizeImportableNetworksResponse({
      candidates: [
        { network_id: '1', status: 'available', can_import: true, reason_code: 'unregistered', reason_message: 'ok' },
        { network_id: '2', status: 'managed', can_import: false, reason_code: 'already_managed', reason_message: 'managed' },
      ],
      summary: {
        total: 10,
        available: 5,
      },
    })

    expect(normalized.summary).toEqual({
      total: 10,
      available: 5,
      managed: 1,
      blocked: 0,
    })
  })
})
