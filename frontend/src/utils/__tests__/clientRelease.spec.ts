import { describe, expect, it } from 'vitest'
import { formatAssetSize, selectClientDownloads, type GitHubReleaseAsset } from '../clientRelease'

function asset(name: string, size = 0): GitHubReleaseAsset {
  return {
    name,
    size,
    browser_download_url: `https://github.com/tkxs/USA0Box/releases/download/v1.2.3/${name}`
  }
}

describe('selectClientDownloads', () => {
  const assets = [
    asset('latest.yml'),
    asset('ZeroBox-1.2.3-Setup.exe.blockmap'),
    asset('ZeroBox-1.2.3-Setup.exe'),
    asset('ZeroBox-1.2.3-arm64.dmg'),
    asset('ZeroBox-1.2.3.dmg'),
    asset('ZeroBox-1.2.3-amd64.deb'),
    asset('ZeroBox-1.2.3-x86_64.AppImage'),
    asset('ZeroBox-1.2.3-arm64.AppImage'),
    asset('ZeroBox-1.2.3-android-test.apk'),
    asset('ZeroBox-android-update.apk'),
    asset('ZeroBox-1.2.3-android.apk')
  ]

  it('keeps only mainstream installable release assets', () => {
    const groups = selectClientDownloads(assets)

    expect(groups.map((group) => group.id)).toEqual(['windows', 'macos', 'linux', 'android'])
    expect(groups.find((group) => group.id === 'macos')?.variants.map((variant) => variant.asset.name)).toEqual([
      'ZeroBox-1.2.3-arm64.dmg',
      'ZeroBox-1.2.3.dmg'
    ])
    expect(groups.find((group) => group.id === 'android')?.variants[0].asset.name).toBe('ZeroBox-1.2.3-android.apk')
  })

  it('omits platforms that have no matching installer', () => {
    expect(selectClientDownloads([asset('ZeroBox-1.2.3-Setup.exe')]).map((group) => group.id)).toEqual(['windows'])
  })
})

describe('formatAssetSize', () => {
  it('formats package sizes without noisy precision', () => {
    expect(formatAssetSize(42.25 * 1024 * 1024)).toBe('42.3 MB')
    expect(formatAssetSize(132.8 * 1024 * 1024)).toBe('133 MB')
    expect(formatAssetSize(0)).toBe('安装包')
  })
})
