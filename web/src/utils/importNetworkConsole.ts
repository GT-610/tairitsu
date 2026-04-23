import type { ImportableNetworkCandidate, ImportableNetworksResponse } from '../services/api'

const emptySummary = {
  total: 0,
  available: 0,
  managed: 0,
  blocked: 0,
}

export function normalizeImportableNetworksResponse(payload: unknown): ImportableNetworksResponse {
  if (!payload || typeof payload !== 'object') {
    return { candidates: [], summary: emptySummary }
  }

  const data = payload as Partial<ImportableNetworksResponse>
  const candidates = Array.isArray(data.candidates) ? data.candidates : []

  if (data.summary && typeof data.summary === 'object') {
    return {
      candidates,
      summary: {
        total: typeof data.summary.total === 'number' ? data.summary.total : candidates.length,
        available: typeof data.summary.available === 'number' ? data.summary.available : countByStatus(candidates, 'available'),
        managed: typeof data.summary.managed === 'number' ? data.summary.managed : countByStatus(candidates, 'managed'),
        blocked: typeof data.summary.blocked === 'number' ? data.summary.blocked : countByStatus(candidates, 'blocked'),
      },
    }
  }

  return {
    candidates,
    summary: {
      total: candidates.length,
      available: countByStatus(candidates, 'available'),
      managed: countByStatus(candidates, 'managed'),
      blocked: countByStatus(candidates, 'blocked'),
    },
  }
}

function countByStatus(candidates: ImportableNetworkCandidate[], status: ImportableNetworkCandidate['status']) {
  return candidates.filter((candidate) => candidate.status === status).length
}
