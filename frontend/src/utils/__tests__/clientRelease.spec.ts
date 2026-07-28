import { describe, expect, it } from 'vitest'
import { SOFTWARE_CATALOG } from '@/api/clientRelease'
import {
  buildAcceleratedDownloadUrl,
  formatAssetSize,
  normalizeSoftwareAssets,
  selectRecommendedAsset,
  type GitHubReleaseAsset
} from '../clientRelease'

function asset(repo: string, tag: string, name: string, size = 1024): GitHubReleaseAsset {
  return {
    name,
    size,
    browser_download_url: `https://github.com/${repo}/releases/download/${tag}/${name}`
  }
}

describe('software release asset normalization', () => {
  it('recognizes ZeroAgent installers and filters updater metadata', () => {
    const entry = SOFTWARE_CATALOG.find((item) => item.id === 'zeroagent')!
    const assets = normalizeSoftwareAssets(entry, [
      asset(entry.repo, 'v0.3.15', 'latest.json'),
      asset(entry.repo, 'v0.3.15', 'ZeroAgent-v0.3.15-Windows-x64-Setup.exe'),
      asset(entry.repo, 'v0.3.15', 'ZeroAgent-v0.3.15-macOS-aarch64.dmg'),
      asset(entry.repo, 'v0.3.15', 'ZeroAgent-v0.3.15-Linux-x86_64.rpm'),
      asset(entry.repo, 'v0.3.15', 'ZeroAgent-v0.3.15-Android-arm64.apk')
    ])

    expect(assets.map((item) => [item.platform, item.architecture, item.format])).toEqual([
      ['windows', 'x64', 'EXE'],
      ['macos', 'arm64', 'DMG'],
      ['linux', 'x64', 'RPM'],
      ['android', 'arm64', 'APK']
    ])
  })

  it('recognizes CC Switch alternatives without exposing signatures', () => {
    const entry = SOFTWARE_CATALOG.find((item) => item.id === 'cc-switch')!
    const assets = normalizeSoftwareAssets(entry, [
      asset(entry.repo, 'v3.18.0', 'CC-Switch-v3.18.0-Windows.msi'),
      asset(entry.repo, 'v3.18.0', 'CC-Switch-v3.18.0-Windows-arm64-Portable.zip'),
      asset(entry.repo, 'v3.18.0', 'CC-Switch-v3.18.0-macOS.dmg'),
      asset(entry.repo, 'v3.18.0', 'CC-Switch-v3.18.0-Linux-arm64.AppImage'),
      asset(entry.repo, 'v3.18.0', 'CC-Switch-v3.18.0-Linux-arm64.AppImage.sig')
    ])

    expect(assets).toHaveLength(4)
    expect(assets.find((item) => item.platform === 'macos')?.architecture).toBe('universal')
    expect(assets.some((item) => item.name.endsWith('.sig'))).toBe(false)
  })

  it('keeps Codex++ Windows and both macOS architectures', () => {
    const entry = SOFTWARE_CATALOG.find((item) => item.id === 'codex-plus-plus')!
    const assets = normalizeSoftwareAssets(entry, [
      asset(entry.repo, 'v1.2.43', 'CodexPlusPlus-1.2.43-windows-x64-setup.exe'),
      asset(entry.repo, 'v1.2.43', 'CodexPlusPlus-1.2.43-macos-arm64.dmg'),
      asset(entry.repo, 'v1.2.43', 'CodexPlusPlus-1.2.43-macos-x64.dmg'),
      asset(entry.repo, 'v1.2.43', 'latest.json')
    ])

    expect(assets.map((item) => item.architecture)).toEqual(['x64', 'arm64', 'x64'])
  })
})

describe('accelerated download URL validation', () => {
  const repo = 'USA-Zero/ZeroAgent'

  it('wraps only an expected GitHub release asset', () => {
    const original = 'https://github.com/USA-Zero/ZeroAgent/releases/download/v1.0.0/app.exe'
    expect(buildAcceleratedDownloadUrl(original, repo)).toBe(`https://ghfast.top/${original}`)
  })

  it.each([
    'http://github.com/USA-Zero/ZeroAgent/releases/download/v1/app.exe',
    'https://github.com/other/repo/releases/download/v1/app.exe',
    'https://github.com/USA-Zero/ZeroAgent/archive/refs/tags/v1.zip',
    'https://example.com/USA-Zero/ZeroAgent/releases/download/v1/app.exe',
    'javascript:alert(1)'
  ])('rejects unsafe or unrelated URL %s', (url) => {
    expect(buildAcceleratedDownloadUrl(url, repo)).toBeNull()
  })
})

describe('device recommendation', () => {
  const entry = SOFTWARE_CATALOG.find((item) => item.id === 'codex-plus-plus')!
  const assets = normalizeSoftwareAssets(entry, entry.fallbackRelease.assets)

  it('does not guess an architecture when it cannot be detected', () => {
    expect(selectRecommendedAsset(assets, { platform: 'macos', architecture: null })).toBeNull()
  })

  it('selects the highest-priority package for a known architecture', () => {
    expect(selectRecommendedAsset(assets, { platform: 'macos', architecture: 'arm64' })?.format).toBe('DMG')
    expect(selectRecommendedAsset(assets, { platform: 'windows', architecture: 'x64' })?.format).toBe('EXE')
  })
})

describe('formatAssetSize', () => {
  it('formats package sizes without noisy precision', () => {
    expect(formatAssetSize(42.25 * 1024 * 1024)).toBe('42.3 MB')
    expect(formatAssetSize(132.8 * 1024 * 1024)).toBe('133 MB')
    expect(formatAssetSize(0)).toBe('安装包')
  })
})
