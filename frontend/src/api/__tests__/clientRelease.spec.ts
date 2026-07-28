import { beforeEach, describe, expect, it, vi } from 'vitest'

const { apiGet } = vi.hoisted(() => ({ apiGet: vi.fn() }))
vi.mock('@/api/client', () => ({ apiClient: { get: apiGet } }))

import { getSoftwareCenterReleases, getSoftwareRelease, SOFTWARE_CATALOG } from '../clientRelease'

function apiItem(index = 0) {
  const entry = SOFTWARE_CATALOG[index]
  return {
    id: index + 1,
    repository: entry.repo,
    repository_url: `https://github.com/${entry.repo}`,
    name: entry.name,
    description: entry.description,
    logo_url: entry.logo,
    featured: entry.featured,
    enabled: true,
    sort_order: (index + 1) * 10,
    supported_platforms: entry.supportedPlatforms,
    release: { ...entry.fallbackRelease, tag_name: 'v9.9.9' }
  }
}

describe('software center release loading', () => {
  beforeEach(() => { localStorage.clear(); apiGet.mockReset() })

  it('loads all software through the backend and caches releases locally', async () => {
    apiGet.mockResolvedValue({ data: SOFTWARE_CATALOG.map((_, index) => apiItem(index)) })
    const results = await getSoftwareCenterReleases()
    expect(apiGet).toHaveBeenCalledWith('/software-center')
    expect(results).toHaveLength(3)
    expect(results[0].release.tag_name).toBe('v9.9.9')
    expect(localStorage.getItem('software_center_release_v2:usa-zero/zeroagent')).toContain('v9.9.9')
  })

  it('maps a dynamically published repository using generic asset rules', async () => {
    apiGet.mockResolvedValue({ data: [{ ...apiItem(), id: 99, repository: 'acme/tool', repository_url: 'https://github.com/acme/tool', name: 'Tool', logo_url: 'https://example.com/tool.png', supported_platforms: ['windows'], release: { ...apiItem().release, assets: [{ name: 'Tool-windows-x64.exe', size: 42, browser_download_url: 'https://github.com/acme/tool/releases/download/v1.0.0/Tool-windows-x64.exe' }] } }] })
    const [result] = await getSoftwareCenterReleases()
    expect(result.entry.id).toBe('99')
    expect(result.assets[0].platform).toBe('windows')
    expect(result.assets[0].acceleratedUrl).toContain('ghfast.top/https://github.com/acme/tool/releases/download/')
  })

  it('uses manually configured assets without adding a GitHub proxy', async () => {
    apiGet.mockResolvedValue({ data: [{
      id: 100, source_type: 'manual', source_url: 'https://chatgpt.com/download/', name: 'Codex App',
      description: 'Official app', logo_url: 'https://example.com/icon.png', featured: false, enabled: true,
      sort_order: 40, supported_platforms: ['windows'],
      release: { tag_name: '1.0.0', name: 'Codex App 1.0.0', body: 'Notes', published_at: '2026-07-29T00:00:00Z', html_url: 'https://chatgpt.com/download/', assets: [{ name: 'Codex.exe', size: 0, browser_download_url: 'https://downloads.example.com/Codex.exe' }] },
      asset_variants: [{ name: 'Codex.exe', label: 'Windows 安装程序', platform: 'windows', architecture: 'x64', format: 'EXE', kind: 'installer', url: 'https://downloads.example.com/Codex.exe' }]
    }] })
    const [result] = await getSoftwareCenterReleases()
    expect(result.entry.repo).toBe('')
    expect(result.entry.sourceType).toBe('manual')
    expect(result.entry.sourceUrl).toBe('https://chatgpt.com/download/')
    expect(result.assets[0].originalUrl).toBe('https://downloads.example.com/Codex.exe')
    expect(result.assets[0].acceleratedUrl).toBeNull()
  })

  it('uses bundled snapshots when the backend is unavailable', async () => {
    apiGet.mockRejectedValue(new Error('offline'))
    const results = await getSoftwareCenterReleases()
    expect(results.map((result) => result.state)).toEqual(['fallback', 'fallback', 'fallback'])
    expect(results[0].release.tag_name).toBe('v0.3.15')
  })

  it('uses a fresh local cache for a single application without a network request', async () => {
    const entry = SOFTWARE_CATALOG[0]
    localStorage.setItem(`software_center_release_v2:${entry.repo.toLowerCase()}`, JSON.stringify({ fetchedAt: Date.now(), release: entry.fallbackRelease }))
    const result = await getSoftwareRelease(entry)
    expect(result.state).toBe('stale')
    expect(apiGet).not.toHaveBeenCalled()
  })

  it('keeps the complete ZeroAgent release notes in the bundled snapshot', () => {
    const release = SOFTWARE_CATALOG[0].fallbackRelease
    expect(release.body).toContain('## 更新内容')
    expect(release.body).toContain('crypto.randomUUID()')
    expect(release.body).toContain('## 在线使用')
  })
})
