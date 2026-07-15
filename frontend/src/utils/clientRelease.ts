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

export interface ClientDownloadVariant {
  label: string
  hint: string
  asset: GitHubReleaseAsset
}

export interface ClientDownloadGroup {
  id: ClientPlatform
  name: string
  shortName: string
  description: string
  variants: ClientDownloadVariant[]
}

function findAsset(assets: GitHubReleaseAsset[], pattern: RegExp): GitHubReleaseAsset | undefined {
  return assets.find((asset) => pattern.test(asset.name) && !asset.name.endsWith('.blockmap'))
}

export function selectClientDownloads(assets: GitHubReleaseAsset[]): ClientDownloadGroup[] {
  const groups: ClientDownloadGroup[] = []

  const windows = findAsset(assets, /-Setup\.exe$/i)
  if (windows) {
    groups.push({
      id: 'windows',
      name: 'Windows',
      shortName: 'WIN',
      description: '适用于 Windows 10 / 11，标准安装程序。',
      variants: [{ label: '下载安装程序', hint: 'x64 · EXE', asset: windows }]
    })
  }

  const macArm = findAsset(assets, /-arm64\.dmg$/i)
  const macIntel = findAsset(assets, /ZeroBox-[\d.]+\.dmg$/i)
  const macVariants: ClientDownloadVariant[] = []
  if (macArm) macVariants.push({ label: 'Apple 芯片', hint: 'M1 及更新机型 · DMG', asset: macArm })
  if (macIntel) macVariants.push({ label: 'Intel 芯片', hint: 'Intel Mac · DMG', asset: macIntel })
  if (macVariants.length > 0) {
    groups.push({
      id: 'macos',
      name: 'macOS',
      shortName: 'MAC',
      description: '同时提供 Apple Silicon 与 Intel 版本。',
      variants: macVariants
    })
  }

  const linuxDeb = findAsset(assets, /-amd64\.deb$/i)
  const linuxAppImage = findAsset(assets, /-x86_64\.AppImage$/i)
  const linuxVariants: ClientDownloadVariant[] = []
  if (linuxDeb) linuxVariants.push({ label: 'Debian / Ubuntu', hint: 'x64 · DEB', asset: linuxDeb })
  if (linuxAppImage) linuxVariants.push({ label: '通用 Linux', hint: 'x64 · AppImage', asset: linuxAppImage })
  if (linuxVariants.length > 0) {
    groups.push({
      id: 'linux',
      name: 'Linux',
      shortName: 'LNX',
      description: '支持主流 Debian 系发行版与 AppImage。',
      variants: linuxVariants
    })
  }

  const android = findAsset(assets, /-[\d.]+-android\.apk$/i)
  if (android) {
    groups.push({
      id: 'android',
      name: 'Android',
      shortName: 'APK',
      description: '适用于主流 Android 手机和平板设备。',
      variants: [{ label: '下载 Android 版', hint: '正式版 · APK', asset: android }]
    })
  }

  return groups
}

export function detectClientPlatform(): ClientPlatform | null {
  if (typeof navigator === 'undefined') return null
  const value = `${navigator.platform || ''} ${navigator.userAgent || ''}`.toLowerCase()
  if (/android/.test(value)) return 'android'
  if (/win/.test(value)) return 'windows'
  if (/mac|iphone|ipad/.test(value)) return 'macos'
  if (/linux|x11/.test(value)) return 'linux'
  return null
}

export function formatAssetSize(bytes: number): string {
  if (!Number.isFinite(bytes) || bytes <= 0) return '安装包'
  const megabytes = bytes / 1024 / 1024
  return megabytes >= 100 ? `${Math.round(megabytes)} MB` : `${megabytes.toFixed(1)} MB`
}

