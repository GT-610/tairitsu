import type { ImportableNetworkCandidate, ImportNetworksResponse } from '../services/api'

export function groupImportCandidates(candidates: ImportableNetworkCandidate[]) {
  return {
    available: candidates.filter((candidate) => candidate.status === 'available'),
    managed: candidates.filter((candidate) => candidate.status === 'managed'),
    blocked: candidates.filter((candidate) => candidate.status === 'blocked'),
  }
}

export function buildImportResultFeedback(result: ImportNetworksResponse) {
  if (result.summary.imported > 0 && result.summary.failed === 0 && result.summary.skipped === 0) {
    return {
      severity: 'success' as const,
      text: `已将 ${result.summary.imported} 个网络分配给 ${result.target_owner.username}，控制器接管列表已刷新。`,
    }
  }

  if (result.summary.imported > 0) {
    return {
      severity: 'warning' as const,
      text: `已导入 ${result.summary.imported} 个网络，失败 ${result.summary.failed} 个，跳过 ${result.summary.skipped} 个。`,
    }
  }

  return {
    severity: 'error' as const,
    text: `没有成功导入任何网络，失败 ${result.summary.failed} 个，跳过 ${result.summary.skipped} 个。`,
  }
}
