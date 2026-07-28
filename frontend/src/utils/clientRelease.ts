export interface GitHubReleaseAsset {
  name: string
  size: number
  browser_download_url: string
  download_count?: number
}

export interface GitHubRelease {
  tag_name: string
  name: string
  published_at: string
  html_url: string
  body: string
  assets: GitHubReleaseAsset[]
}

export type ClientPlatform = 'windows' | 'macos' | 'linux' | 'android'
export type ClientArchitecture = 'x64' | 'arm64' | 'universal'
export type SoftwarePackageKind = 'installer' | 'portable' | 'archive'

export interface SoftwareAssetRule {
  pattern: RegExp
  platform: ClientPlatform
  architecture: ClientArchitecture
  format: string
  kind: SoftwarePackageKind
  label: string
  priority: number
}

export interface SoftwareCatalogEntry {
  id: string
  name: string
  repo: string
  sourceType?: 'github' | 'manual'
  sourceUrl?: string
  description: string
  logo: string
  featured: boolean
  supportedPlatforms: ClientPlatform[]
  assetRules: SoftwareAssetRule[]
  fallbackRelease: GitHubRelease
}

export const genericSoftwareAssetRules: SoftwareAssetRule[] = [
  { pattern: /(?:windows|win)[^/]*(?:arm64|aarch64)[^/]*\.exe$/i, platform: 'windows', architecture: 'arm64', format: 'EXE', kind: 'installer', label: 'Windows ARM64 安装程序', priority: 1 },
  { pattern: /(?:windows|win)[^/]*(?:arm64|aarch64)[^/]*\.msi$/i, platform: 'windows', architecture: 'arm64', format: 'MSI', kind: 'installer', label: 'Windows ARM64 MSI', priority: 2 },
  { pattern: /(?:windows|win)[^/]*\.(?:exe|msi)$/i, platform: 'windows', architecture: 'x64', format: '安装包', kind: 'installer', label: 'Windows 安装程序', priority: 1 },
  { pattern: /(?:windows|win)[^/]*\.zip$/i, platform: 'windows', architecture: 'x64', format: 'ZIP', kind: 'portable', label: 'Windows 便携版', priority: 3 },
  { pattern: /(?:macos|darwin|osx)[^/]*(?:arm64|aarch64)[^/]*\.(?:dmg|pkg)$/i, platform: 'macos', architecture: 'arm64', format: '安装包', kind: 'installer', label: 'Apple 芯片安装包', priority: 1 },
  { pattern: /(?:macos|darwin|osx)[^/]*(?:x64|x86_64|amd64)[^/]*\.(?:dmg|pkg)$/i, platform: 'macos', architecture: 'x64', format: '安装包', kind: 'installer', label: 'Intel Mac 安装包', priority: 1 },
  { pattern: /(?:macos|darwin|osx)[^/]*\.(?:dmg|pkg)$/i, platform: 'macos', architecture: 'universal', format: '安装包', kind: 'installer', label: 'macOS 安装包', priority: 1 },
  { pattern: /(?:macos|darwin|osx)[^/]*\.(?:zip|tar\.gz)$/i, platform: 'macos', architecture: 'universal', format: '归档', kind: 'archive', label: 'macOS 应用归档', priority: 3 },
  { pattern: /linux[^/]*(?:arm64|aarch64)[^/]*\.appimage$/i, platform: 'linux', architecture: 'arm64', format: 'AppImage', kind: 'portable', label: 'Linux ARM64 通用版', priority: 1 },
  { pattern: /linux[^/]*(?:arm64|aarch64)[^/]*\.(?:deb|rpm)$/i, platform: 'linux', architecture: 'arm64', format: '安装包', kind: 'installer', label: 'Linux ARM64 安装包', priority: 2 },
  { pattern: /linux[^/]*\.appimage$/i, platform: 'linux', architecture: 'x64', format: 'AppImage', kind: 'portable', label: 'Linux x64 通用版', priority: 1 },
  { pattern: /linux[^/]*\.(?:deb|rpm)$/i, platform: 'linux', architecture: 'x64', format: '安装包', kind: 'installer', label: 'Linux x64 安装包', priority: 2 },
  { pattern: /(?:android[^/]*|[^/]*)\.apk$/i, platform: 'android', architecture: 'arm64', format: 'APK', kind: 'installer', label: 'Android 安装包', priority: 1 }
]

export interface SoftwareAssetVariant {
  name: string
  label: string
  platform: ClientPlatform
  architecture: ClientArchitecture
  format: string
  kind: SoftwarePackageKind
  size: number
  priority: number
  originalUrl: string
  acceleratedUrl: string | null
}

export interface ClientDevice {
  platform: ClientPlatform | null
  architecture: Exclude<ClientArchitecture, 'universal'> | null
}

const auxiliaryAssetPattern = /(?:\.sig|\.blockmap|\.sha\d*|\.md5|\.json|\.ya?ml|\.txt|RELEASE_NOTES\.md)$/i

export function buildAcceleratedDownloadUrl(originalUrl: string, expectedRepo: string): string | null {
  try {
    const url = new URL(originalUrl)
    if (url.protocol !== 'https:' || url.hostname !== 'github.com' || url.port || url.username || url.password) {
      return null
    }

    const [owner, repo, segment, action, tag, ...assetPath] = url.pathname.split('/').filter(Boolean)
    if (`${owner}/${repo}`.toLowerCase() !== expectedRepo.toLowerCase()) return null
    if (segment !== 'releases' || action !== 'download' || !tag || assetPath.length === 0) return null

    return `https://ghfast.top/${url.href}`
  } catch {
    return null
  }
}

export function normalizeSoftwareAssets(
  entry: SoftwareCatalogEntry,
  assets: GitHubReleaseAsset[]
): SoftwareAssetVariant[] {
  const normalized: SoftwareAssetVariant[] = []

  for (const asset of assets) {
    if (auxiliaryAssetPattern.test(asset.name)) continue
    const rule = entry.assetRules.find((candidate) => candidate.pattern.test(asset.name))
    if (!rule) continue

    normalized.push({
      name: asset.name,
      label: rule.label,
      platform: rule.platform,
      architecture: rule.architecture,
      format: rule.format,
      kind: rule.kind,
      size: asset.size,
      priority: rule.priority,
      originalUrl: asset.browser_download_url,
      acceleratedUrl: entry.repo ? buildAcceleratedDownloadUrl(asset.browser_download_url, entry.repo) : null
    })
  }

  return normalized.sort((left, right) => {
    const platformOrder: ClientPlatform[] = ['windows', 'macos', 'linux', 'android']
    return platformOrder.indexOf(left.platform) - platformOrder.indexOf(right.platform)
      || left.priority - right.priority
      || left.name.localeCompare(right.name)
  })
}

export function detectClientDevice(): ClientDevice {
  if (typeof navigator === 'undefined') return { platform: null, architecture: null }
  const value = `${navigator.platform || ''} ${navigator.userAgent || ''}`.toLowerCase()

  let platform: ClientPlatform | null = null
  if (/android/.test(value)) platform = 'android'
  else if (/iphone|ipad|ipod/.test(value)) platform = null
  else if (/win/.test(value)) platform = 'windows'
  else if (/mac/.test(value)) platform = 'macos'
  else if (/linux|x11/.test(value)) platform = 'linux'

  let architecture: ClientDevice['architecture'] = null
  if (/arm64|aarch64/.test(value)) architecture = 'arm64'
  else if (/x86_64|x64|win64|wow64|amd64/.test(value)) architecture = 'x64'

  return { platform, architecture }
}

export function selectRecommendedAsset(
  assets: SoftwareAssetVariant[],
  device: ClientDevice
): SoftwareAssetVariant | null {
  if (!device.platform) return null
  const platformAssets = assets.filter((asset) => asset.platform === device.platform)
  const universal = platformAssets.find((asset) => asset.architecture === 'universal')
  if (universal) return universal
  if (!device.architecture) return null
  return platformAssets.find((asset) => asset.architecture === device.architecture) || null
}

export function formatAssetSize(bytes: number): string {
  if (!Number.isFinite(bytes) || bytes <= 0) return '安装包'
  const megabytes = bytes / 1024 / 1024
  return megabytes >= 100 ? `${Math.round(megabytes)} MB` : `${megabytes.toFixed(1)} MB`
}

export function platformLabel(platform: ClientPlatform): string {
  return ({ windows: 'Windows', macos: 'macOS', linux: 'Linux', android: 'Android' })[platform]
}

export function architectureLabel(architecture: ClientArchitecture): string {
  return ({ x64: 'x64', arm64: 'ARM64', universal: '通用' })[architecture]
}
