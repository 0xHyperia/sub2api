import {
	genericSoftwareAssetRules,
  normalizeSoftwareAssets,
  type GitHubRelease,
  type GitHubReleaseAsset,
  type SoftwareAssetRule,
  type SoftwareAssetVariant,
  type SoftwareCatalogEntry
} from '@/utils/clientRelease'
import { apiClient } from '@/api/client'

const RELEASE_CACHE_TTL = 15 * 60 * 1000
const RELEASE_CACHE_PREFIX = 'software_center_release_v2:'

function githubAsset(repo: string, tag: string, name: string, size: number): GitHubReleaseAsset {
  return {
    name,
    size,
    browser_download_url: `https://github.com/${repo}/releases/download/${tag}/${name}`
  }
}

function fallbackRelease(
  repo: string,
  tag: string,
  name: string,
  publishedAt: string,
  body: string,
  assets: Array<[string, number]>
): GitHubRelease {
  return {
    tag_name: tag,
    name,
    published_at: publishedAt,
    html_url: `https://github.com/${repo}/releases/tag/${tag}`,
    body,
    assets: assets.map(([assetName, size]) => githubAsset(repo, tag, assetName, size))
  }
}

const zeroAgentRules: SoftwareAssetRule[] = [
  { pattern: /Windows-x64-Setup\.exe$/i, platform: 'windows', architecture: 'x64', format: 'EXE', kind: 'installer', label: 'Windows 安装程序', priority: 1 },
  { pattern: /Windows-x64\.msi$/i, platform: 'windows', architecture: 'x64', format: 'MSI', kind: 'installer', label: 'Windows MSI', priority: 2 },
  { pattern: /Windows-x64-portable\.zip$/i, platform: 'windows', architecture: 'x64', format: 'ZIP', kind: 'portable', label: 'Windows 便携版', priority: 3 },
  { pattern: /macOS-aarch64\.dmg$/i, platform: 'macos', architecture: 'arm64', format: 'DMG', kind: 'installer', label: 'Apple 芯片安装包', priority: 1 },
  { pattern: /macOS-x64\.dmg$/i, platform: 'macos', architecture: 'x64', format: 'DMG', kind: 'installer', label: 'Intel Mac 安装包', priority: 1 },
  { pattern: /macOS-aarch64\.app\.tar\.gz$/i, platform: 'macos', architecture: 'arm64', format: 'TAR.GZ', kind: 'archive', label: 'Apple 芯片应用归档', priority: 3 },
  { pattern: /macOS-x64\.app\.tar\.gz$/i, platform: 'macos', architecture: 'x64', format: 'TAR.GZ', kind: 'archive', label: 'Intel Mac 应用归档', priority: 3 },
  { pattern: /Linux-x86_64\.AppImage$/i, platform: 'linux', architecture: 'x64', format: 'AppImage', kind: 'portable', label: 'Linux 通用版', priority: 1 },
  { pattern: /Linux-x86_64\.deb$/i, platform: 'linux', architecture: 'x64', format: 'DEB', kind: 'installer', label: 'Debian / Ubuntu', priority: 2 },
  { pattern: /Linux-x86_64\.rpm$/i, platform: 'linux', architecture: 'x64', format: 'RPM', kind: 'installer', label: 'Fedora / RHEL', priority: 3 },
  { pattern: /Android-arm64\.apk$/i, platform: 'android', architecture: 'arm64', format: 'APK', kind: 'installer', label: 'Android ARM64', priority: 1 }
]

const ccSwitchRules: SoftwareAssetRule[] = [
  { pattern: /Windows\.msi$/i, platform: 'windows', architecture: 'x64', format: 'MSI', kind: 'installer', label: 'Windows 安装程序', priority: 1 },
  { pattern: /Windows-Portable\.zip$/i, platform: 'windows', architecture: 'x64', format: 'ZIP', kind: 'portable', label: 'Windows 便携版', priority: 2 },
  { pattern: /Windows-arm64\.msi$/i, platform: 'windows', architecture: 'arm64', format: 'MSI', kind: 'installer', label: 'Windows ARM64', priority: 1 },
  { pattern: /Windows-arm64-Portable\.zip$/i, platform: 'windows', architecture: 'arm64', format: 'ZIP', kind: 'portable', label: 'Windows ARM64 便携版', priority: 2 },
  { pattern: /macOS\.dmg$/i, platform: 'macos', architecture: 'universal', format: 'DMG', kind: 'installer', label: 'macOS 安装包', priority: 1 },
  { pattern: /macOS\.zip$/i, platform: 'macos', architecture: 'universal', format: 'ZIP', kind: 'archive', label: 'macOS ZIP', priority: 2 },
  { pattern: /macOS\.tar\.gz$/i, platform: 'macos', architecture: 'universal', format: 'TAR.GZ', kind: 'archive', label: 'macOS TAR.GZ', priority: 3 },
  { pattern: /Linux-x86_64\.AppImage$/i, platform: 'linux', architecture: 'x64', format: 'AppImage', kind: 'portable', label: 'Linux x64 通用版', priority: 1 },
  { pattern: /Linux-x86_64\.deb$/i, platform: 'linux', architecture: 'x64', format: 'DEB', kind: 'installer', label: 'Linux x64 DEB', priority: 2 },
  { pattern: /Linux-x86_64\.rpm$/i, platform: 'linux', architecture: 'x64', format: 'RPM', kind: 'installer', label: 'Linux x64 RPM', priority: 3 },
  { pattern: /Linux-arm64\.AppImage$/i, platform: 'linux', architecture: 'arm64', format: 'AppImage', kind: 'portable', label: 'Linux ARM64 通用版', priority: 1 },
  { pattern: /Linux-arm64\.deb$/i, platform: 'linux', architecture: 'arm64', format: 'DEB', kind: 'installer', label: 'Linux ARM64 DEB', priority: 2 },
  { pattern: /Linux-arm64\.rpm$/i, platform: 'linux', architecture: 'arm64', format: 'RPM', kind: 'installer', label: 'Linux ARM64 RPM', priority: 3 }
]

const codexPlusPlusRules: SoftwareAssetRule[] = [
  { pattern: /windows-x64-setup\.exe$/i, platform: 'windows', architecture: 'x64', format: 'EXE', kind: 'installer', label: 'Windows 安装程序', priority: 1 },
  { pattern: /windows-x64\.zip$/i, platform: 'windows', architecture: 'x64', format: 'ZIP', kind: 'portable', label: 'Windows 便携版', priority: 2 },
  { pattern: /macos-arm64\.dmg$/i, platform: 'macos', architecture: 'arm64', format: 'DMG', kind: 'installer', label: 'Apple 芯片安装包', priority: 1 },
  { pattern: /macos-arm64\.zip$/i, platform: 'macos', architecture: 'arm64', format: 'ZIP', kind: 'archive', label: 'Apple 芯片 ZIP', priority: 2 },
  { pattern: /macos-x64\.dmg$/i, platform: 'macos', architecture: 'x64', format: 'DMG', kind: 'installer', label: 'Intel Mac 安装包', priority: 1 },
  { pattern: /macos-x64\.zip$/i, platform: 'macos', architecture: 'x64', format: 'ZIP', kind: 'archive', label: 'Intel Mac ZIP', priority: 2 }
]

export const SOFTWARE_CATALOG: SoftwareCatalogEntry[] = [
  {
    id: 'zeroagent',
    name: 'ZeroAgent',
    repo: 'USA-Zero/ZeroAgent',
    description: '可扩展的 AI Agent 桌面客户端，覆盖对话、工具调用、WebUI 与跨平台工作流。',
    logo: '/software-center/zeroagent.png',
    featured: true,
    supportedPlatforms: ['windows', 'macos', 'linux', 'android'],
    assetRules: zeroAgentRules,
    fallbackRelease: fallbackRelease(
      'USA-Zero/ZeroAgent', 'v0.3.15', 'ZeroAgent v0.3.15', '2026-07-28T15:07:12Z',
      `## 更新内容

- 修复部分浏览器不支持 \`crypto.randomUUID()\` 时，登录后出现白屏或登录按钮无响应的问题。
- 统一桌面端、Linux 安装包和 Android APK 的正式下载与自动更新地址。
- 提供 macOS Apple Silicon、macOS Intel、Windows、Linux 和 Android arm64 安装包。

## 在线使用

浏览器访问 [agent.usa0.top](https://agent.usa0.top) 即可使用 ZeroAgent。

## 下载

- macOS：Apple Silicon 与 Intel 的 \`.dmg\` 安装包。
- Windows：安装版 \`.exe\`、\`.msi\` 和便携版 \`.zip\`。
- Linux：\`.AppImage\`、\`.deb\` 和 \`.rpm\`。
- Android：arm64 \`.apk\`。`,
      [
        ['ZeroAgent-v0.3.15-Android-arm64.apk', 31274693],
        ['ZeroAgent-v0.3.15-Linux-x86_64.AppImage', 100555256],
        ['ZeroAgent-v0.3.15-Linux-x86_64.deb', 28002000],
        ['ZeroAgent-v0.3.15-Linux-x86_64.rpm', 28011179],
        ['ZeroAgent-v0.3.15-macOS-aarch64.app.tar.gz', 26266964],
        ['ZeroAgent-v0.3.15-macOS-aarch64.dmg', 26663250],
        ['ZeroAgent-v0.3.15-macOS-x64.app.tar.gz', 26789401],
        ['ZeroAgent-v0.3.15-macOS-x64.dmg', 27330647],
        ['ZeroAgent-v0.3.15-Windows-x64-portable.zip', 24748109],
        ['ZeroAgent-v0.3.15-Windows-x64-Setup.exe', 19761254],
        ['ZeroAgent-v0.3.15-Windows-x64.msi', 25104384]
      ]
    )
  },
  {
    id: 'cc-switch',
    name: 'CC Switch',
    repo: 'farion1231/cc-switch',
    description: '跨平台 AI 编程助手管理工具，集中切换 Claude Code、Codex、Gemini 等供应商配置。',
    logo: '/software-center/cc-switch.png',
    featured: false,
    supportedPlatforms: ['windows', 'macos', 'linux'],
    assetRules: ccSwitchRules,
    fallbackRelease: fallbackRelease(
      'farion1231/cc-switch', 'v3.18.0', 'CC Switch v3.18.0', '2026-07-21T15:34:53Z',
      '新增 Grok Build 管理与 xAI 接入，并完善 Codex 用量、诊断日志和 Windows 切换体验。',
      [
        ['CC-Switch-v3.18.0-Linux-arm64.AppImage', 89229832],
        ['CC-Switch-v3.18.0-Linux-arm64.deb', 12396096],
        ['CC-Switch-v3.18.0-Linux-arm64.rpm', 12396820],
        ['CC-Switch-v3.18.0-Linux-x86_64.AppImage', 91621880],
        ['CC-Switch-v3.18.0-Linux-x86_64.deb', 12915920],
        ['CC-Switch-v3.18.0-Linux-x86_64.rpm', 12916477],
        ['CC-Switch-v3.18.0-macOS.dmg', 26699423],
        ['CC-Switch-v3.18.0-macOS.tar.gz', 27353597],
        ['CC-Switch-v3.18.0-macOS.zip', 26686100],
        ['CC-Switch-v3.18.0-Windows-arm64-Portable.zip', 12252953],
        ['CC-Switch-v3.18.0-Windows-arm64.msi', 12156928],
        ['CC-Switch-v3.18.0-Windows-Portable.zip', 12880212],
        ['CC-Switch-v3.18.0-Windows.msi', 12849152]
      ]
    )
  },
  {
    id: 'codex-plus-plus',
    name: 'Codex++',
    repo: 'BigPizzaV3/CodexPlusPlus',
    description: '面向 Codex 桌面应用的外部启动与管理工具，提供供应商切换、会话管理和界面增强。',
    logo: '/software-center/codex-plus-plus.png',
    featured: false,
    supportedPlatforms: ['windows', 'macos'],
    assetRules: codexPlusPlusRules,
    fallbackRelease: fallbackRelease(
      'BigPizzaV3/CodexPlusPlus', 'v1.2.43', 'Codex++ v1.2.43', '2026-07-26T09:52:02Z',
      '修复 macOS 重启、会话索引和供应商同步，并完善增强页面体验。',
      [
        ['CodexPlusPlus-1.2.43-macos-arm64.dmg', 34341178],
        ['CodexPlusPlus-1.2.43-macos-arm64.zip', 30439992],
        ['CodexPlusPlus-1.2.43-macos-x64.dmg', 35383007],
        ['CodexPlusPlus-1.2.43-macos-x64.zip', 30837047],
        ['CodexPlusPlus-1.2.43-windows-x64-setup.exe', 21342642],
        ['CodexPlusPlus-1.2.43-windows-x64.zip', 26818255]
      ]
    )
  }
]

interface CachedRelease {
  fetchedAt: number
  release: GitHubRelease
}

export type SoftwareReleaseState = 'loading' | 'ready' | 'stale' | 'fallback' | 'error'

export interface SoftwareReleaseResult {
  entry: SoftwareCatalogEntry
  release: GitHubRelease
  assets: SoftwareAssetVariant[]
  state: Exclude<SoftwareReleaseState, 'loading' | 'error'>
  error?: string
}

function cacheKey(entry: SoftwareCatalogEntry): string {
  return `${RELEASE_CACHE_PREFIX}${(entry.repo || entry.id).toLowerCase()}`
}

function readCachedRelease(entry: SoftwareCatalogEntry): CachedRelease | null {
  if (typeof localStorage === 'undefined') return null
  try {
    const parsed = JSON.parse(localStorage.getItem(cacheKey(entry)) || '') as CachedRelease
    if (!parsed?.fetchedAt || !isGitHubRelease(parsed.release)) return null
    return parsed
  } catch {
    return null
  }
}

function cacheRelease(entry: SoftwareCatalogEntry, release: GitHubRelease): void {
  if (typeof localStorage === 'undefined') return
  try {
    localStorage.setItem(cacheKey(entry), JSON.stringify({ fetchedAt: Date.now(), release }))
  } catch {
    // Restricted storage modes may reject writes; the live result remains usable.
  }
}

function isGitHubRelease(value: unknown): value is GitHubRelease {
  const release = value as Partial<GitHubRelease> | null
  return Boolean(
    release?.tag_name
    && release?.html_url
    && typeof release.body === 'string'
    && typeof release.published_at === 'string'
    && Array.isArray(release.assets)
  )
}

interface SoftwareCatalogAPIItem {
  id: number
  source_type?: 'github' | 'manual'
  source_url?: string
  repository?: string
  repository_url?: string
  name: string
  description: string
  logo_url: string
  featured: boolean
  enabled: boolean
  sort_order: number
  supported_platforms: string[]
  release?: GitHubRelease
  asset_variants?: Array<{
    name: string; label: string; platform: SoftwareAssetVariant['platform']; architecture: SoftwareAssetVariant['architecture'];
    format: string; kind: SoftwareAssetVariant['kind']; url: string; accelerated_url?: string; size?: number
  }>
}

function catalogEntryFromAPI(item: SoftwareCatalogAPIItem): SoftwareCatalogEntry | null {
  if (!item.release || !isGitHubRelease(item.release)) return null
  const builtIn = SOFTWARE_CATALOG.find((entry) => entry.repo?.toLowerCase() === item.repository?.toLowerCase())
  const supportedPlatforms = item.supported_platforms.filter((platform): platform is SoftwareCatalogEntry['supportedPlatforms'][number] =>
    ['windows', 'macos', 'linux', 'android'].includes(platform))
  return {
    id: String(item.id),
    name: item.name,
    repo: item.repository || '',
    sourceType: item.source_type || 'github',
    sourceUrl: item.source_url || item.repository_url,
    description: item.description,
    logo: builtIn?.logo || item.logo_url,
    featured: item.featured,
    supportedPlatforms,
    assetRules: builtIn?.assetRules || genericSoftwareAssetRules,
    fallbackRelease: builtIn?.fallbackRelease || item.release
  }
}

function resultFrom(
  entry: SoftwareCatalogEntry,
  release: GitHubRelease,
  state: SoftwareReleaseResult['state'],
  error?: string,
  assets?: SoftwareAssetVariant[]
): SoftwareReleaseResult {
  return { entry, release, assets: assets || normalizeSoftwareAssets(entry, release.assets), state, error }
}

export async function getSoftwareRelease(
  entry: SoftwareCatalogEntry,
  forceRefresh = false
): Promise<SoftwareReleaseResult> {
  const cached = readCachedRelease(entry)
  if (!forceRefresh && cached && Date.now() - cached.fetchedAt < RELEASE_CACHE_TTL) {
    return resultFrom(entry, cached.release, 'stale')
  }

  try {
    const results = await fetchSoftwareCenter()
    const result = results.find((candidate) => candidate.entry.id === entry.id)
    if (!result) throw new Error('软件已下架或暂未同步 Release')
    return result
  } catch (error) {
    const message = error instanceof Error ? error.message : 'Release request failed'
    if (cached) return resultFrom(entry, cached.release, 'stale', message)
    return resultFrom(entry, entry.fallbackRelease, 'fallback', message)
  }
}

export async function getSoftwareCenterReleases(_forceRefresh = false): Promise<SoftwareReleaseResult[]> {
  try {
    return await fetchSoftwareCenter()
  } catch (error) {
    return Promise.all(SOFTWARE_CATALOG.map((entry) => fallbackResult(entry, error)))
  }
}

async function fetchSoftwareCenter(): Promise<SoftwareReleaseResult[]> {
  const { data } = await apiClient.get<SoftwareCatalogAPIItem[]>('/software-center')
  return data.flatMap((item) => {
    const entry = catalogEntryFromAPI(item)
    if (!entry || !item.release) return []
    cacheRelease(entry, item.release)
    const manualAssets = item.source_type === 'manual' ? (item.asset_variants || []).map((asset, index) => ({
      name: asset.name, label: asset.label, platform: asset.platform, architecture: asset.architecture,
      format: asset.format, kind: asset.kind, size: asset.size || 0, priority: index + 1,
      originalUrl: asset.url, acceleratedUrl: asset.accelerated_url || null
    })) : undefined
    return [resultFrom(entry, item.release, 'ready', undefined, manualAssets)]
  })
}

function fallbackResult(entry: SoftwareCatalogEntry, error: unknown): SoftwareReleaseResult {
  const cached = readCachedRelease(entry)
  const message = error instanceof Error ? error.message : '软件中心暂时不可用'
  if (cached) return resultFrom(entry, cached.release, 'stale', message)
  return resultFrom(entry, entry.fallbackRelease, 'fallback', message)
}
