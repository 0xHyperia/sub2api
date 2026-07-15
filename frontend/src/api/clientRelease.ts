import type { GitHubRelease } from '@/utils/clientRelease'

const RELEASE_API_URL = 'https://api.github.com/repos/tkxs/USA0Box/releases/latest'
const RELEASE_CACHE_KEY = 'usa0box_latest_release_v1'
const RELEASE_CACHE_TTL = 15 * 60 * 1000

interface CachedRelease {
  fetchedAt: number
  release: GitHubRelease
}

const fallbackRelease: GitHubRelease = {
  tag_name: 'v0.1.6',
  name: 'ZeroBox 0.1.6',
  published_at: '2026-07-13T07:46:17Z',
  html_url: 'https://github.com/tkxs/USA0Box/releases/tag/v0.1.6',
  body: `本次更新重点修复移动端应用内更新、帮助引导与会话默认设置，并统一 ZeroBox 官方支持入口。

## 移动端更新

- 修复 Android 更新下载任务正常进行时误报 \`Missing download id\` 的问题。
- 下载任务 ID 改用字符串跨 JavaScript/Java 桥接传输，并兼容旧数字格式。
- 更新弹窗在下载期间保持正确状态并显示实时进度。

## 帮助与引导

- 帮助内容重写为 ZeroBox 专属说明，覆盖 ZeroBox AI、模型提供方、系统提示、知识库、图片生成和应用更新。
- 官网统一为 \`https://usa0.top\`，并区分帮助中心、版本更新、问题反馈和 GitHub 入口。
- 修复“离开引导”按钮无法继续导航的问题。
- 修复引导回答已经完成，发送按钮仍一直显示“停止”的问题。

## 会话默认值

- 新会话名称默认为“会话”；用户手动修改或自动生成名称后不再覆盖。
- 默认系统提示调整为“你是一个专业AI助手为用户解答各种问题”。
- 仅迁移仍在使用旧默认提示的设置，不覆盖用户自定义内容。`,
  assets: [
    ['ZeroBox-0.1.6-Setup.exe', 0],
    ['ZeroBox-0.1.6-arm64.dmg', 0],
    ['ZeroBox-0.1.6.dmg', 0],
    ['ZeroBox-0.1.6-amd64.deb', 0],
    ['ZeroBox-0.1.6-x86_64.AppImage', 0],
    ['ZeroBox-0.1.6-android.apk', 0]
  ].map(([name, size]) => ({
    name: String(name),
    size: Number(size),
    browser_download_url: `https://github.com/tkxs/USA0Box/releases/download/v0.1.6/${name}`
  }))
}

function readCachedRelease(): CachedRelease | null {
  if (typeof localStorage === 'undefined') return null
  try {
    const parsed = JSON.parse(localStorage.getItem(RELEASE_CACHE_KEY) || '') as CachedRelease
    if (!parsed?.release?.tag_name || !Array.isArray(parsed.release.assets)) return null
    return parsed
  } catch {
    return null
  }
}

function cacheRelease(release: GitHubRelease): void {
  if (typeof localStorage === 'undefined') return
  try {
    localStorage.setItem(RELEASE_CACHE_KEY, JSON.stringify({ fetchedAt: Date.now(), release }))
  } catch {
    // Private browsing and restricted storage modes may reject localStorage writes.
  }
}

function isGitHubRelease(value: unknown): value is GitHubRelease {
  const release = value as Partial<GitHubRelease> | null
  return Boolean(
    release?.tag_name
    && release?.html_url
    && typeof release.body === 'string'
    && Array.isArray(release.assets)
  )
}

export interface ClientReleaseResult {
  release: GitHubRelease
  source: 'github' | 'cache' | 'fallback'
}

export async function getLatestClientRelease(forceRefresh = false): Promise<ClientReleaseResult> {
  const cached = readCachedRelease()
  if (!forceRefresh && cached && Date.now() - cached.fetchedAt < RELEASE_CACHE_TTL) {
    return { release: cached.release, source: 'cache' }
  }

  try {
    const response = await fetch(RELEASE_API_URL, {
      headers: { Accept: 'application/vnd.github+json' }
    })
    if (!response.ok) throw new Error(`GitHub release request failed: ${response.status}`)
    const release: unknown = await response.json()
    if (!isGitHubRelease(release)) throw new Error('GitHub returned an invalid release payload')
    cacheRelease(release)
    return { release, source: 'github' }
  } catch {
    if (cached) return { release: cached.release, source: 'cache' }
    return { release: fallbackRelease, source: 'fallback' }
  }
}

