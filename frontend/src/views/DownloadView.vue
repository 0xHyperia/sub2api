<template>
  <div class="software-page" :data-theme="theme" @scroll.passive="handlePageScroll">
    <header class="software-header" :class="{ 'is-scrolled': headerScrolled }" aria-label="主导航">
      <div class="nav-shell">
        <RouterLink class="brand" to="/home" :aria-label="`${siteName} 首页`">
          <span class="brand-mark" aria-hidden="true"><img src="/home-experiment/logo.png" alt="" /></span>
          <span class="brand-name">{{ siteName }}</span>
        </RouterLink>

        <nav id="software-navigation" class="nav-links" :class="{ 'is-open': mobileNavOpen }" aria-label="页面导航">
          <RouterLink to="/home">首页</RouterLink>
          <RouterLink to="/key-usage">Key 用量</RouterLink>
          <RouterLink class="active" to="/download" aria-current="page">软件中心</RouterLink>
        </nav>

        <div class="nav-actions">
          <button class="icon-button" type="button" :aria-label="theme === 'dark' ? '切换到浅色主题' : '切换到深色主题'" :title="theme === 'dark' ? '切换到浅色主题' : '切换到深色主题'" @click="toggleTheme">
            <Icon :name="theme === 'dark' ? 'sun' : 'moon'" size="md" aria-hidden="true" />
          </button>
          <RouterLink class="primary-link" :to="entryPath">{{ entryLabel }}</RouterLink>
          <button class="icon-button menu-button" type="button" :aria-expanded="mobileNavOpen" aria-controls="software-navigation" :aria-label="mobileNavOpen ? '关闭页面导航' : '打开页面导航'" @click="mobileNavOpen = !mobileNavOpen">
            <Icon :name="mobileNavOpen ? 'x' : 'menu'" size="md" aria-hidden="true" />
          </button>
        </div>
      </div>
    </header>

    <main>
      <section class="software-hero" aria-labelledby="software-title">
        <canvas ref="heroFlowCanvas" class="hero-flow" aria-hidden="true"></canvas>
        <div class="hero-inner">
          <article v-if="featuredApp" class="featured-product" :data-software-id="featuredApp.entry.id">
            <img class="featured-logo" :src="featuredApp.entry.logo" :alt="`${featuredApp.entry.name} 图标`" />
            <h1 id="software-title">{{ featuredApp.entry.name }}</h1>
            <p class="hero-copy">{{ featuredApp.entry.description }}</p>
            <div class="featured-meta" :aria-label="`${featuredApp.entry.name} 发布信息`">
              <strong>{{ featuredApp.result?.release.tag_name || '正在同步' }}</strong>
              <span aria-hidden="true"></span>
              <time>{{ publishedDate(featuredApp.result) }}</time>
              <span aria-hidden="true"></span>
              <span>{{ featuredApp.entry.supportedPlatforms.length }} 个平台</span>
            </div>
            <div class="hero-actions">
              <a v-if="featuredRecommended" class="button primary" :href="featuredRecommended.acceleratedUrl || featuredRecommended.originalUrl" rel="noopener noreferrer">
                <Icon name="download" size="md" aria-hidden="true" />
                下载 {{ featuredApp.entry.name }}
              </a>
              <button v-else class="button primary" type="button" @click="openDetails(featuredApp)">
                <Icon name="grid" size="md" aria-hidden="true" />
                选择 {{ featuredApp.entry.name }} 版本
              </button>
              <button class="button secondary" type="button" @click="openDetails(featuredApp)">
                <Icon name="grid" size="sm" aria-hidden="true" />
                全部版本
              </button>
              <a v-if="featuredApp.entry.repo === 'USA-Zero/ZeroAgent'" class="button secondary webui-button" href="https://agent.usa0.top" target="_blank" rel="noopener noreferrer">
                <Icon name="externalLink" size="sm" aria-hidden="true" />
                访问 WebUI
              </a>
            </div>
            <div class="featured-links">
              <a v-if="featuredRecommended?.acceleratedUrl" :href="featuredRecommended.originalUrl" rel="noopener noreferrer">GitHub 原始下载 <Icon name="externalLink" size="xs" aria-hidden="true" /></a>
              <button type="button" @click="openReleaseNotes(featuredApp)">查看更新日志</button>
              <a :href="featuredApp.entry.sourceUrl || `https://github.com/${featuredApp.entry.repo}`" target="_blank" rel="noopener noreferrer">{{ featuredApp.entry.sourceType === 'manual' ? '访问官网' : '来源仓库' }} <Icon name="externalLink" size="xs" aria-hidden="true" /></a>
            </div>
            <div class="platform-row" :aria-label="`${featuredApp.entry.name} 支持平台`">
              <button v-for="platform in featuredApp.entry.supportedPlatforms" :key="platform" type="button" :title="`查看 ${platformLabel(platform)} 安装包`" @click="openDetails(featuredApp, $event, platform)">
                <PlatformLogo :platform="platform" />{{ platformLabel(platform) }}
              </button>
            </div>
            <button class="catalog-jump" type="button" @click="scrollToCatalog">浏览软件中心 <Icon name="arrowDown" size="sm" aria-hidden="true" /></button>
          </article>
          <article v-else class="featured-product">
            <h1 id="software-title">软件中心</h1>
            <p class="hero-copy">暂时没有已发布的软件，请稍后再来查看。</p>
          </article>
        </div>

        <div class="hero-status" aria-label="软件中心状态">
          <div><span>收录应用</span><strong>{{ apps.length }} 款</strong></div>
          <div><span>下载线路</span><strong>加速 + GitHub</strong></div>
          <div><span>支持平台</span><strong>桌面与 Android</strong></div>
          <div><i></i><span>版本数据</span><strong>{{ readyCount }}/{{ apps.length }} 可用</strong></div>
        </div>
      </section>

      <section id="software-catalog" ref="catalogSection" class="catalog-section" aria-labelledby="catalog-title">
        <div class="section-inner">
          <header class="section-heading">
            <div>
              <p class="eyebrow">APPLICATION CATALOG</p>
              <h2 id="catalog-title">选择适合你的工具</h2>
              <p>按当前系统快速定位安装包，或进入应用详情选择架构与安装格式。</p>
            </div>
            <button class="refresh-all" type="button" :disabled="refreshingAll" @click="refreshAll">
              <Icon name="refresh" size="sm" :class="{ spinning: refreshingAll }" aria-hidden="true" />
              同步版本
            </button>
          </header>

          <div class="catalog-tools">
            <div class="search-box">
              <Icon name="search" size="sm" aria-hidden="true" />
              <input ref="searchInput" v-model="searchQuery" type="text" aria-label="搜索软件" placeholder="搜索软件或功能" autocomplete="off" @keydown.esc.stop="clearSearch" />
              <button v-if="searchQuery" class="search-clear" type="button" aria-label="清除搜索内容" title="清除搜索" @click="clearSearch"><Icon name="x" size="sm" aria-hidden="true" /></button>
            </div>
            <div class="platform-filter" role="group" aria-label="按平台筛选">
              <button v-for="option in platformOptions" :key="option.value" type="button" :class="{ active: selectedPlatform === option.value }" :aria-pressed="selectedPlatform === option.value" @click="selectedPlatform = option.value">
                {{ option.label }} <span>{{ platformAppCount(option.value) }}</span>
              </button>
            </div>
          </div>
          <div class="catalog-summary" role="status" aria-live="polite">
            <span>找到 <strong>{{ filteredApps.length }}</strong> 款软件<span v-if="device.platform"> · 当前设备 {{ platformLabel(device.platform) }}<template v-if="device.architecture"> {{ architectureLabel(device.architecture) }}</template></span></span>
            <button v-if="searchQuery || selectedPlatform !== 'all'" type="button" @click="resetFilters">清除筛选</button>
          </div>

          <div v-if="filteredApps.length" class="software-grid">
            <article v-for="app in filteredApps" :key="app.entry.id" class="software-card" :data-software-id="app.entry.id">
              <header class="card-header">
                <img :src="app.entry.logo" :alt="`${app.entry.name} 图标`" />
                <div class="card-title">
                  <div><h3>{{ app.entry.name }}</h3></div>
                  <a :href="app.entry.sourceUrl || `https://github.com/${app.entry.repo}`" target="_blank" rel="noopener noreferrer">{{ app.entry.repo || '官方网站' }} <Icon name="externalLink" size="xs" aria-hidden="true" /></a>
                </div>
              </header>
              <p class="card-description">{{ app.entry.description }}</p>
              <div class="card-platforms" :aria-label="`${app.entry.name} 支持平台`">
                <button v-for="platform in app.entry.supportedPlatforms" :key="platform" type="button" :title="`查看 ${app.entry.name} ${platformLabel(platform)} 安装包`" @click="openDetails(app, $event, platform)">
                  <PlatformLogo :platform="platform" /><span>{{ platformLabel(platform) }}</span>
                </button>
              </div>
              <div class="card-release">
                <span>{{ app.result?.release.tag_name || '正在同步版本' }}</span>
                <span>{{ publishedDate(app.result) }}</span>
              </div>
              <div class="card-actions">
                <a v-if="recommendedAsset(app)" class="card-download" :href="recommendedAsset(app)?.acceleratedUrl || recommendedAsset(app)?.originalUrl" rel="noopener noreferrer">
                  <Icon name="download" size="sm" aria-hidden="true" />
                  {{ recommendedAsset(app)?.acceleratedUrl ? '加速下载' : '立即下载' }}
                </a>
                <button v-else class="card-download" type="button" @click="openDetails(app)">
                  <Icon name="grid" size="sm" aria-hidden="true" />
                  选择版本
                </button>
                <a v-if="recommendedAsset(app)?.acceleratedUrl" class="card-original" :href="recommendedAsset(app)?.originalUrl" rel="noopener noreferrer" :aria-label="`${app.entry.name} GitHub 原始下载`" title="GitHub 原始下载">
                  <Icon name="externalLink" size="sm" aria-hidden="true" />
                </a>
                <button class="card-more" type="button" :aria-label="`查看 ${app.entry.name} 全部安装包`" @click="openDetails(app)">
                  <Icon name="chevronRight" size="sm" aria-hidden="true" />
                </button>
              </div>
              <footer>
                <button type="button" @click="openReleaseNotes(app)">查看更新日志</button>
                <button type="button" :disabled="app.state === 'loading'" @click="refreshApp(app)">
                  <Icon name="refresh" size="xs" :class="{ spinning: app.state === 'loading' }" aria-hidden="true" /> 单独刷新
                </button>
              </footer>
            </article>
          </div>

          <div v-else class="catalog-empty" role="status">
            <Icon name="search" size="xl" aria-hidden="true" />
            <strong>没有匹配的软件</strong>
            <span>尝试清除搜索词或切换平台筛选。</span>
            <button type="button" @click="resetFilters">清除筛选</button>
          </div>
        </div>
      </section>
    </main>

    <div v-if="detailsApp" class="modal-layer" role="presentation" @mousedown.self="closeDetails">
      <section class="software-dialog" role="dialog" aria-modal="true" aria-labelledby="software-dialog-title" tabindex="-1">
        <header class="dialog-header">
          <div><p class="eyebrow">客户端下载 · {{ detailsApp.result?.release.tag_name }}</p><h2 id="software-dialog-title">选择 {{ detailsApp.entry.name }} 安装版本</h2><span>选择适合当前设备的安装包，默认使用国内加速线路</span></div>
          <button ref="detailsCloseRef" class="icon-button" type="button" aria-label="关闭安装包选择" @click="closeDetails"><Icon name="x" size="md" aria-hidden="true" /></button>
        </header>
        <div class="dialog-platform-tabs" role="group" aria-label="筛选安装包平台">
          <button type="button" :class="{ active: detailPlatform === 'all' }" :aria-pressed="detailPlatform === 'all'" @click="detailPlatform = 'all'">全部</button>
          <button v-for="group in allDetailAssetGroups" :key="group.platform" type="button" :class="{ active: detailPlatform === group.platform }" :aria-pressed="detailPlatform === group.platform" @click="detailPlatform = group.platform"><PlatformLogo :platform="group.platform" />{{ platformLabel(group.platform) }}<span>{{ group.assets.length }}</span></button>
        </div>
        <div class="dialog-body">
          <section v-for="group in detailAssetGroups" :key="group.platform" class="asset-group" :class="{ 'is-current-platform': group.platform === device.platform }" :aria-labelledby="`asset-${group.platform}`">
            <header><span class="platform-icon"><PlatformLogo :platform="group.platform" /></span><div><div class="asset-group-title"><h3 :id="`asset-${group.platform}`">{{ platformLabel(group.platform) }}</h3><span v-if="group.platform === device.platform">当前设备</span></div><p>{{ platformDescription(group.platform) }} · {{ group.assets.length }} 个安装包</p></div></header>
            <div class="asset-list">
              <article v-for="asset in group.assets" :key="asset.name" class="asset-row" :class="{ 'is-recommended': detailsRecommendedAsset?.name === asset.name }">
                <div><div class="asset-name-line"><strong>{{ asset.label }}</strong><span v-if="detailsRecommendedAsset?.name === asset.name">推荐</span></div><span>{{ architectureLabel(asset.architecture) }} · {{ asset.format }} · {{ formatAssetSize(asset.size) }}</span></div>
                <div class="asset-actions">
                  <a v-if="asset.acceleratedUrl" class="accelerated-link" :href="asset.acceleratedUrl" rel="noopener noreferrer" :aria-label="`加速下载 ${asset.label}`" title="加速下载"><Icon name="download" size="sm" aria-hidden="true" /></a>
                  <a class="original-link" :href="asset.originalUrl" rel="noopener noreferrer" :aria-label="`下载 ${asset.label}`" :title="detailsApp.entry.sourceType === 'manual' ? '原始下载' : 'GitHub 原始下载'"><Icon :name="asset.acceleratedUrl ? 'externalLink' : 'download'" size="sm" aria-hidden="true" /></a>
                </div>
              </article>
            </div>
          </section>
          <div v-if="!detailAssetGroups.length" class="dialog-empty"><strong>暂未识别到安装包</strong><span>可以刷新版本数据或前往来源仓库查看。</span></div>
        </div>
        <footer class="dialog-footer"><span>下载项由软件发布方或本站管理员维护</span><button type="button" @click="closeDetails">关闭</button></footer>
      </section>
    </div>

    <div v-if="notesApp" class="modal-layer" role="presentation" @mousedown.self="closeReleaseNotes">
      <section class="software-dialog notes-dialog" role="dialog" aria-modal="true" aria-labelledby="notes-dialog-title" tabindex="-1">
        <header class="dialog-header">
          <div><p class="eyebrow">更新日志 · {{ notesApp.result?.release.tag_name }}</p><h2 id="notes-dialog-title">{{ notesApp.result?.release.name || notesApp.entry.name }}</h2><span>发布于 {{ publishedDate(notesApp.result) }}</span></div>
          <button ref="notesCloseRef" class="icon-button" type="button" aria-label="关闭更新日志" @click="closeReleaseNotes"><Icon name="x" size="md" aria-hidden="true" /></button>
        </header>
        <div class="release-body markdown-body" v-html="releaseNotesHtml"></div>
        <footer class="dialog-footer"><span>{{ notesApp.entry.sourceType === 'manual' ? '内容由本站管理员维护' : '内容同步自 GitHub Releases' }}</span><div><a :href="notesApp.result?.release.html_url" target="_blank" rel="noopener noreferrer">查看发布来源 <Icon name="externalLink" size="xs" aria-hidden="true" /></a><button type="button" @click="closeReleaseNotes">关闭</button></div></footer>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import DOMPurify from 'dompurify'
import { marked } from 'marked'
import {
  getSoftwareCenterReleases,
  getSoftwareRelease,
  SOFTWARE_CATALOG,
  type SoftwareReleaseResult,
  type SoftwareReleaseState
} from '@/api/clientRelease'
import Icon from '@/components/icons/Icon.vue'
import PlatformLogo from '@/components/icons/PlatformLogo.vue'
import { useTheme } from '@/composables/useTheme'
import { useAppStore, useAuthStore } from '@/stores'
import {
  architectureLabel,
  detectClientDevice,
  formatAssetSize,
  platformLabel,
  selectRecommendedAsset,
  type ClientPlatform,
  type SoftwareAssetVariant,
  type SoftwareCatalogEntry
} from '@/utils/clientRelease'

interface SoftwareAppView {
  entry: SoftwareCatalogEntry
  state: SoftwareReleaseState
  result: SoftwareReleaseResult | null
}

const appStore = useAppStore()
const authStore = useAuthStore()
const { resolvedTheme: theme, toggleTheme } = useTheme()
const device = detectClientDevice()

const appState = ref<Record<string, SoftwareAppView>>(Object.fromEntries(
  SOFTWARE_CATALOG.map((entry) => [entry.id, { entry, state: 'loading', result: null }])
))
const searchQuery = ref('')
const selectedPlatform = ref<'all' | ClientPlatform>('all')
const refreshingAll = ref(false)
const mobileNavOpen = ref(false)
const headerScrolled = ref(false)
const detailsApp = ref<SoftwareAppView | null>(null)
const notesApp = ref<SoftwareAppView | null>(null)
const detailPlatform = ref<'all' | ClientPlatform>('all')
const detailsCloseRef = ref<HTMLButtonElement | null>(null)
const notesCloseRef = ref<HTMLButtonElement | null>(null)
const heroFlowCanvas = ref<HTMLCanvasElement | null>(null)
const catalogSection = ref<HTMLElement | null>(null)
const searchInput = ref<HTMLInputElement | null>(null)
let dialogTrigger: HTMLElement | null = null
const flowCleanupCallbacks: Array<() => void> = []

const catalogOrder = ref<string[]>(SOFTWARE_CATALOG.map((entry) => entry.id))
const apps = computed(() => catalogOrder.value.map((id) => appState.value[id]).filter(Boolean))
const featuredApp = computed(() => apps.value.find((app) => app.entry.featured) || apps.value[0])
const featuredRecommended = computed(() => recommendedAsset(featuredApp.value))
const readyCount = computed(() => apps.value.filter((app) => app.result).length)
const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'USA-零')
const entryPath = computed(() => authStore.isAuthenticated ? (authStore.isAdmin ? '/admin/dashboard' : '/dashboard') : '/login')
const entryLabel = computed(() => authStore.isAuthenticated ? '进入控制台' : '开始接入')
const platformOptions: Array<{ value: 'all' | ClientPlatform; label: string }> = [
  { value: 'all', label: '全部' },
  { value: 'windows', label: 'Windows' },
  { value: 'macos', label: 'macOS' },
  { value: 'linux', label: 'Linux' },
  { value: 'android', label: 'Android' }
]

const filteredApps = computed(() => {
  const query = searchQuery.value.trim().toLowerCase()
  return apps.value.filter((app) => {
    const platformMatches = selectedPlatform.value === 'all' || app.entry.supportedPlatforms.includes(selectedPlatform.value)
    const queryMatches = !query || `${app.entry.name} ${app.entry.repo} ${app.entry.description}`.toLowerCase().includes(query)
    return platformMatches && queryMatches
  })
})

const allDetailAssetGroups = computed(() => {
  if (!detailsApp.value?.result) return []
  return detailsApp.value.entry.supportedPlatforms.map((platform) => ({
    platform,
    assets: detailsApp.value?.result?.assets.filter((asset) => asset.platform === platform) || []
  })).filter((group) => group.assets.length > 0).sort((left, right) =>
    Number(right.platform === device.platform) - Number(left.platform === device.platform)
  )
})
const detailAssetGroups = computed(() => detailPlatform.value === 'all'
  ? allDetailAssetGroups.value
  : allDetailAssetGroups.value.filter((group) => group.platform === detailPlatform.value))
const detailsRecommendedAsset = computed(() => recommendedAsset(detailsApp.value || undefined))

const releaseNotesHtml = computed(() => {
  const source = notesApp.value?.result?.release.body || '本次发布暂未提供更新说明。'
  const rendered = marked.parse(source, { async: false, gfm: true, breaks: false }) as string
  const sanitized = DOMPurify.sanitize(rendered, { USE_PROFILES: { html: true } })
  if (typeof document === 'undefined') return sanitized
  const template = document.createElement('template')
  template.innerHTML = sanitized
  const repo = notesApp.value?.entry.repo
  const tag = notesApp.value?.result?.release.tag_name
  template.content.querySelectorAll<HTMLAnchorElement>('a[href]').forEach((link) => {
    const href = link.getAttribute('href') || ''
    if (repo && tag && href && !href.startsWith('#') && !/^[a-z][a-z\d+.-]*:/i.test(href)) {
      link.href = new URL(href, `https://github.com/${repo}/blob/${tag}/`).href
    }
    link.target = '_blank'
    link.rel = 'noopener noreferrer'
  })
  template.content.querySelectorAll<HTMLImageElement>('img[src]').forEach((image) => {
    const sourceUrl = image.getAttribute('src') || ''
    if (repo && tag && sourceUrl && !/^[a-z][a-z\d+.-]*:/i.test(sourceUrl)) {
      image.src = new URL(sourceUrl, `https://raw.githubusercontent.com/${repo}/${tag}/`).href
    }
    image.loading = 'lazy'
  })
  return template.innerHTML
})

function platformDescription(platform: ClientPlatform): string {
  return ({
    windows: '适用于 Windows 10 / 11',
    macos: '支持 Apple 芯片与 Intel Mac',
    linux: '支持主流 Linux 发行版',
    android: '适用于 Android 手机和平板'
  })[platform]
}

function platformAppCount(platform: 'all' | ClientPlatform): number {
  return platform === 'all' ? apps.value.length : apps.value.filter((app) => app.entry.supportedPlatforms.includes(platform)).length
}

function recommendedAsset(app: SoftwareAppView | undefined): SoftwareAssetVariant | null {
  if (!app?.result) return null
  return selectRecommendedAsset(app.result.assets, device)
}

function publishedDate(result: SoftwareReleaseResult | null | undefined): string {
  if (!result?.release.published_at) return '获取中'
  const date = new Date(result.release.published_at)
  if (Number.isNaN(date.getTime())) return '日期未知'
  return new Intl.DateTimeFormat('zh-CN', { year: 'numeric', month: 'short', day: 'numeric' }).format(date)
}

function applyResult(result: SoftwareReleaseResult): void {
  appState.value[result.entry.id] = { entry: result.entry, state: result.state, result }
}

async function loadAll(forceRefresh = false): Promise<void> {
  if (forceRefresh) apps.value.forEach((app) => { appState.value[app.entry.id] = { ...app, state: 'loading' } })
  const results = await getSoftwareCenterReleases(forceRefresh)
  catalogOrder.value = results.map((result) => result.entry.id)
  appState.value = Object.fromEntries(results.map((result) => [result.entry.id, {
    entry: result.entry,
    state: result.state,
    result
  }]))
}

async function refreshAll(): Promise<void> {
  refreshingAll.value = true
  try { await loadAll(true) } finally { refreshingAll.value = false }
}

async function refreshApp(app: SoftwareAppView): Promise<void> {
  appState.value[app.entry.id] = { ...app, state: 'loading' }
  applyResult(await getSoftwareRelease(app.entry, true))
}

function rememberTrigger(event?: Event): void {
  dialogTrigger = event?.currentTarget instanceof HTMLElement ? event.currentTarget : document.activeElement as HTMLElement
}

function openDetails(app: SoftwareAppView | undefined, event?: Event, platform: 'all' | ClientPlatform = 'all'): void {
  if (!app) return
  rememberTrigger(event)
  detailPlatform.value = platform
  detailsApp.value = appState.value[app.entry.id]
}

function closeDetails(): void {
  detailsApp.value = null
  detailPlatform.value = 'all'
  nextTick(() => dialogTrigger?.focus())
}

function openReleaseNotes(app: SoftwareAppView, event?: Event): void {
  rememberTrigger(event)
  notesApp.value = appState.value[app.entry.id]
}

function closeReleaseNotes(): void {
  notesApp.value = null
  nextTick(() => dialogTrigger?.focus())
}

function resetFilters(): void {
  searchQuery.value = ''
  selectedPlatform.value = 'all'
}

function clearSearch(): void {
  searchQuery.value = ''
  nextTick(() => searchInput.value?.focus())
}

function scrollToCatalog(): void {
  catalogSection.value?.scrollIntoView({ behavior: window.matchMedia('(prefers-reduced-motion: reduce)').matches ? 'auto' : 'smooth' })
}

function handleKeydown(event: KeyboardEvent): void {
  if (event.key !== 'Escape') return
  if (detailsApp.value) closeDetails()
  else if (notesApp.value) closeReleaseNotes()
  else if (mobileNavOpen.value) mobileNavOpen.value = false
}

function handlePageScroll(event: Event): void {
  headerScrolled.value = (event.currentTarget as HTMLElement).scrollTop > 24
}

function startFluidSurface(canvas: HTMLCanvasElement | null): void {
  const context = canvas?.getContext('2d')
  if (!canvas || !context) return
  const surface = canvas
  const drawingContext = context
  const reducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches
  let frame: number | undefined
  let visible = true
  let startedAt = performance.now()
  let lastDrawAt = 0

  function resize(): void {
    const bounds = surface.getBoundingClientRect()
    const ratio = Math.min(window.devicePixelRatio || 1, 1.25)
    surface.width = Math.max(1, Math.round(bounds.width * ratio))
    surface.height = Math.max(1, Math.round(bounds.height * ratio))
    drawingContext.setTransform(ratio, 0, 0, ratio, 0, 0)
  }

  function draw(now = performance.now()): void {
    const width = surface.clientWidth
    const height = surface.clientHeight
    if (!width || !height) return
    const elapsed = reducedMotion ? 0 : (now - startedAt) / 1000
    const colors = theme.value === 'dark'
      ? ['rgba(96,165,250,.10)', 'rgba(45,212,191,.07)', 'rgba(203,213,225,.05)']
      : ['rgba(37,99,235,.07)', 'rgba(14,116,144,.055)', 'rgba(71,85,105,.045)']
    drawingContext.clearRect(0, 0, width, height)
    drawingContext.globalCompositeOperation = theme.value === 'dark' ? 'screen' : 'multiply'
    for (let band = 0; band < 7; band += 1) {
      const baseline = height * (.06 + band / 6 * .88)
      const amplitude = height * (.035 + band % 3 * .012)
      const thickness = height * (.055 + band % 2 * .025)
      const phase = band * .94 + elapsed * (.12 + band * .008)
      const step = Math.max(12, Math.floor(width / 90))
      const waveY = (x: number, offset = 0) => baseline
        + Math.sin(x / width * Math.PI * 2 * (1.25 + band * .17) + phase + offset) * amplitude
        + Math.sin(x / width * Math.PI * .92 - phase * .62 + band) * amplitude * .48
      drawingContext.beginPath()
      drawingContext.moveTo(0, waveY(0))
      for (let x = step; x <= width + step; x += step) drawingContext.lineTo(x, waveY(x))
      for (let x = width + step; x >= 0; x -= step) drawingContext.lineTo(x, waveY(x, .42) + thickness)
      drawingContext.closePath()
      drawingContext.fillStyle = colors[band % colors.length]
      drawingContext.fill()
    }
    drawingContext.globalCompositeOperation = 'source-over'
  }

  function animate(now: number): void {
    frame = undefined
    if (!visible) return
    if (now - lastDrawAt >= 1000 / 24) { draw(now); lastDrawAt = now }
    frame = window.requestAnimationFrame(animate)
  }

  const resizeObserver = new ResizeObserver(() => { resize(); draw() })
  const intersectionObserver = new IntersectionObserver(([entry]) => {
    visible = entry?.isIntersecting !== false
    if (visible && !reducedMotion && !frame) { startedAt = performance.now(); frame = window.requestAnimationFrame(animate) }
    if (!visible && frame) { window.cancelAnimationFrame(frame); frame = undefined }
  }, { threshold: .01 })
  resizeObserver.observe(canvas)
  intersectionObserver.observe(canvas)
  resize()
  draw()
  if (!reducedMotion) frame = window.requestAnimationFrame(animate)
  flowCleanupCallbacks.push(() => {
    if (frame) window.cancelAnimationFrame(frame)
    resizeObserver.disconnect()
    intersectionObserver.disconnect()
  })
}

watch([detailsApp, notesApp], async ([details, notes]) => {
  document.body.style.overflow = details || notes ? 'hidden' : ''
  if (!details && !notes) return
  await nextTick()
  if (details) detailsCloseRef.value?.focus()
  else notesCloseRef.value?.focus()
})

onMounted(() => {
  authStore.checkAuth()
  if (!appStore.publicSettingsLoaded) void appStore.fetchPublicSettings()
  void loadAll()
  window.addEventListener('keydown', handleKeydown)
  startFluidSurface(heroFlowCanvas.value)
})

onUnmounted(() => {
  document.body.style.overflow = ''
  window.removeEventListener('keydown', handleKeydown)
  flowCleanupCallbacks.splice(0).forEach((cleanup) => cleanup())
})
</script>

<style scoped>
.software-page {
  color-scheme: light;
  --background: #fff;
  --foreground: #101828;
  --muted: #f5f7fa;
  --muted-2: #edf1f5;
  --muted-foreground: #667085;
  --border: #d9e1ea;
  --border-strong: #bcc8d6;
  --accent: #2563eb;
  --accent-strong: #1d4ed8;
  --green: #059669;
  --amber: #d97706;
  --panel: rgba(255,255,255,.94);
  --surface: #fff;
  --shadow: 0 24px 60px rgba(15,23,42,.13);
  --soft-shadow: 0 14px 36px rgba(15,23,42,.07);
  min-width: 320px;
  height: 100dvh;
  overflow-x: hidden;
  overflow-y: auto;
  background: var(--background);
  color: var(--foreground);
  font-family: "Public Sans", Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
  letter-spacing: 0;
  isolation: isolate;
}
.software-page[data-theme="dark"] {
  color-scheme: dark;
  --background: #101418;
  --foreground: #eef4fb;
  --muted: #171d23;
  --muted-2: #1d252d;
  --muted-foreground: #9caab9;
  --border: #2d3946;
  --border-strong: #405064;
  --accent: #60a5fa;
  --accent-strong: #93c5fd;
  --panel: rgba(16,20,24,.94);
  --surface: #161c22;
  --shadow: 0 24px 60px rgba(0,0,0,.4);
  --soft-shadow: 0 14px 36px rgba(0,0,0,.24);
}
.software-page * { box-sizing: border-box; }
.software-page a { color: inherit; text-decoration: none; }
.software-page button, .software-page input { color: inherit; font: inherit; }
.software-page :where(a,button,input,[tabindex]):focus-visible { outline: 2px solid var(--accent); outline-offset: 3px; }
.sr-only { position: absolute; width: 1px; height: 1px; padding: 0; margin: -1px; overflow: hidden; clip: rect(0,0,0,0); white-space: nowrap; border: 0; }
.software-header { position: absolute; inset: 0 0 auto; z-index: 30; padding: 18px; pointer-events: none; transition: padding .3s ease; }
.nav-shell { max-width: 1180px; min-height: 56px; margin: 0 auto; padding: 7px; display: flex; align-items: center; gap: 16px; border: 1px solid transparent; background: transparent; pointer-events: auto; transition: .35s ease; }
.software-header.is-scrolled { position: fixed; padding: 14px 18px; }
.software-header.is-scrolled .nav-shell { border-color: var(--border); border-radius: 999px; background: color-mix(in srgb,var(--panel) 90%,transparent); box-shadow: var(--soft-shadow); backdrop-filter: blur(18px) saturate(135%); }
.brand { display: inline-flex; align-items: center; gap: 9px; min-width: max-content; padding: 4px 10px 4px 5px; font-weight: 750; }
.brand-mark { width: 36px; height: 36px; overflow: hidden; border: 1px solid var(--border); border-radius: 50%; background: var(--background); }
.brand-mark img { width: 100%; height: 100%; display: block; object-fit: cover; }
.nav-links { flex: 1; display: flex; align-items: center; justify-content: center; gap: 2px; color: var(--muted-foreground); font-size: 14px; font-weight: 600; }
.nav-links a { min-height: 42px; padding: 8px 13px; display: inline-flex; align-items: center; border-radius: 999px; }
.nav-links a:hover,.nav-links a:focus-visible { background: var(--muted); color: var(--foreground); }
.nav-links a.active { color: var(--foreground); font-weight: 700; }
.nav-actions { margin-left: auto; display: flex; align-items: center; gap: 8px; }
.icon-button { width: 42px; height: 42px; padding: 0; border: 1px solid var(--border); border-radius: 50%; background: var(--background); display: inline-grid; place-items: center; cursor: pointer; transition: border-color .18s ease,background .18s ease,transform .18s ease; }
.primary-link { min-height: 42px; padding: 9px 15px; border-radius: 999px; background: var(--foreground); color: var(--background)!important; display: inline-flex; align-items: center; font-size: 14px; font-weight: 700; transition: transform .18s ease,box-shadow .18s ease; }
.menu-button { display: none; }
.software-hero { position: relative; min-height: 100svh; padding: 126px 18px 92px; overflow: hidden; display: grid; place-items: center; isolation: isolate; }
.hero-flow { position: absolute; z-index: -3; top: -12%; left: -12%; width: 124%; height: 124%; opacity: .72; filter: blur(6px) saturate(108%); pointer-events: none; }
.hero-inner { width: min(820px,100%); display: flex; align-items: center; justify-content: center; }
.eyebrow { margin: 0; color: var(--muted-foreground); font: 700 12px/1.4 ui-monospace,SFMono-Regular,Menlo,Consolas,monospace; text-transform: uppercase; }
.featured-product { width: 100%; min-width: 0; display: flex; flex-direction: column; align-items: center; text-align: center; }
.featured-logo { width: 112px; height: 112px; display: block; object-fit: contain; filter: drop-shadow(0 18px 24px rgba(15,23,42,.16)); }
.featured-product h1 { margin: 14px 0 0; font-size: clamp(54px,6.4vw,78px); line-height: 1; font-weight: 780; }
.hero-copy { max-width: 650px; margin: 20px auto 0; color: var(--muted-foreground); font-size: 16px; line-height: 1.75; }
.featured-meta { min-height: 26px; margin-top: 16px; display: flex; flex-wrap: wrap; align-items: center; justify-content: center; gap: 9px; color: var(--muted-foreground); font: 650 11px/1.3 ui-monospace,SFMono-Regular,Menlo,monospace; }
.featured-meta strong { color: var(--foreground); }
.featured-meta>span[aria-hidden] { width: 1px; height: 12px; background: var(--border-strong); }
.hero-actions { margin-top: 25px; display: flex; flex-wrap: wrap; justify-content: center; gap: 9px; }
.button { min-height: 48px; padding: 11px 18px; border: 1px solid var(--border-strong); border-radius: 8px; display: inline-flex; align-items: center; justify-content: center; gap: 9px; cursor: pointer; font-size: 14px; font-weight: 720; transition: border-color .18s ease,background .18s ease,box-shadow .18s ease,transform .18s ease; }
.button.primary { border-color: var(--foreground); background: var(--foreground); color: var(--background); box-shadow: 0 14px 28px rgba(15,23,42,.14); }
.button.secondary { background: var(--surface); }
.webui-button { border-color: color-mix(in srgb,var(--accent) 45%,var(--border-strong))!important; color: var(--accent-strong)!important; }
.state-loading { color: var(--muted-foreground)!important; }
.platform-row { margin-top: 20px; display: flex; flex-wrap: wrap; justify-content: center; gap: 7px; }
.platform-row>button { min-height: 30px; padding: 5px 8px; border: 1px solid var(--border); border-radius: 6px; background: var(--surface); display: inline-flex; align-items: center; gap: 5px; color: var(--muted-foreground); cursor: pointer; font-size: 10px; font-weight: 700; transition: border-color .18s ease,color .18s ease,transform .18s ease; }
.platform-row svg { width: 16px; height: 16px; }
.featured-links { margin-top: 16px; display: flex; flex-wrap: wrap; align-items: center; justify-content: center; gap: 6px 18px; }
.featured-links button,.featured-links a { padding: 4px 0; border: 0; background: transparent; display: inline-flex; align-items: center; gap: 4px; color: var(--muted-foreground); cursor: pointer; font-size: 11px; font-weight: 700; }
.featured-links :is(button,a):hover { color: var(--foreground); }
.catalog-jump { margin-top: 22px; padding: 6px 0; border: 0; background: transparent; display: inline-flex; align-items: center; gap: 6px; color: var(--muted-foreground)!important; cursor: pointer; font-size: 11px; font-weight: 700; transition: color .18s ease,transform .18s ease; }
.hero-status { position: absolute; bottom: 22px; left: 50%; width: min(820px,calc(100% - 36px)); transform: translateX(-50%); display: grid; grid-template-columns: repeat(4,1fr); border: 1px solid var(--border); border-radius: 8px; background: color-mix(in srgb,var(--panel) 90%,transparent); backdrop-filter: blur(12px); }
.hero-status>div { min-width: 0; padding: 11px 15px; display: flex; align-items: center; gap: 8px; border-right: 1px solid var(--border); }
.hero-status>div:last-child { border-right: 0; }
.hero-status span { color: var(--muted-foreground); font: 600 9px/1.2 ui-monospace,SFMono-Regular,Menlo,monospace; }
.hero-status strong { overflow: hidden; font: 650 11px/1.2 ui-monospace,SFMono-Regular,Menlo,monospace; text-overflow: ellipsis; white-space: nowrap; }
.hero-status i { width: 6px; height: 6px; border-radius: 50%; background: var(--green); animation: pulse 2s ease-in-out infinite; }
.catalog-section { min-height: 100svh; padding: 82px 18px 96px; border-top: 1px solid var(--border); background: var(--muted); }
.section-inner { width: min(1120px,100%); margin: 0 auto; }
.section-heading { display: flex; align-items: flex-end; justify-content: space-between; gap: 32px; }
.section-heading h2 { margin: 10px 0 0; font-size: clamp(30px,3.5vw,42px); line-height: 1.1; }
.section-heading>div>p:last-child { max-width: 650px; margin: 15px 0 0; color: var(--muted-foreground); line-height: 1.7; }
.refresh-all { min-height: 42px; padding: 9px 13px; border: 1px solid var(--border-strong); border-radius: 7px; background: var(--surface); display: inline-flex; align-items: center; gap: 8px; cursor: pointer; font-size: 12px; font-weight: 700; transition: border-color .18s ease,background .18s ease,transform .18s ease; }
.refresh-all:disabled { cursor: wait; opacity: .65; }
.catalog-tools { margin-top: 36px; display: flex; align-items: center; gap: 12px; }
.search-box { min-width: 240px; min-height: 46px; flex: 1; padding: 0 6px 0 14px; border: 1px solid var(--border); border-radius: 7px; background: var(--surface); display: flex; align-items: center; gap: 10px; color: var(--muted-foreground); transition: border-color .18s ease, box-shadow .18s ease, background .18s ease; }
.search-box:hover { border-color: var(--border-strong); }
.search-box:focus-within { border-color: var(--accent); box-shadow: 0 0 0 3px color-mix(in srgb,var(--accent) 16%,transparent); background: var(--background); }
.search-box input { width: 100%; min-width: 0; min-height: 44px; padding: 0; border: 0!important; outline: 0!important; box-shadow: none!important; background: transparent; color: var(--foreground); }
.search-box input::-webkit-search-cancel-button { display: none; appearance: none; }
.search-clear { width: 36px; height: 36px; flex: 0 0 36px; padding: 0; border: 0; border-radius: 6px; background: transparent; display: inline-grid; place-items: center; color: var(--muted-foreground)!important; cursor: pointer; }
.search-clear:hover { background: var(--muted); color: var(--foreground)!important; }
.platform-filter { min-height: 46px; padding: 4px; border: 1px solid var(--border); border-radius: 7px; background: var(--surface); display: flex; gap: 3px; }
.platform-filter button { min-height: 38px; padding: 7px 11px; border: 0; border-radius: 6px; background: transparent; color: var(--muted-foreground); cursor: pointer; font-size: 11px; font-weight: 700; transition: background .18s ease,color .18s ease; }
.platform-filter button>span { min-width: 17px; height: 17px; margin-left: 3px; padding: 0 4px; border-radius: 999px; background: var(--muted); display: inline-grid; place-items: center; font-size: 8px; }
.platform-filter button.active { background: var(--foreground); color: var(--background); }
.platform-filter button.active>span { background: color-mix(in srgb,var(--background) 16%,transparent); }
.catalog-summary { min-height: 34px; padding: 8px 2px 0; display: flex; align-items: center; justify-content: space-between; gap: 12px; color: var(--muted-foreground); font-size: 11px; }
.catalog-summary strong { color: var(--foreground); }
.catalog-summary button { padding: 3px 0; border: 0; background: transparent; color: var(--accent); cursor: pointer; font-size: 11px; font-weight: 700; }
.software-grid { margin-top: 18px; display: grid; grid-template-columns: repeat(3,minmax(0,1fr)); gap: 14px; }
.software-card { min-width: 0; min-height: 390px; padding: 20px; border: 1px solid var(--border); border-radius: 8px; background: var(--surface); display: flex; flex-direction: column; transition: border-color .2s ease,box-shadow .2s ease,transform .2s ease; }
.card-header { display: flex; align-items: center; gap: 13px; }
.card-header>img { width: 58px; height: 58px; flex: 0 0 58px; border-radius: 8px; object-fit: contain; }
.card-title { min-width: 0; flex: 1; }
.card-title>div { display: flex; align-items: center; justify-content: space-between; gap: 8px; }
.card-title h3 { min-width: 0; margin: 0; overflow: hidden; font-size: 19px; text-overflow: ellipsis; white-space: nowrap; }
.card-title>a { max-width: 100%; margin-top: 5px; display: inline-flex; align-items: center; gap: 3px; color: var(--muted-foreground); font-size: 10px; }
.card-description { min-height: 64px; margin: 18px 0 0; color: var(--muted-foreground); font-size: 13px; line-height: 1.65; }
.card-platforms { margin-top: 18px; display: flex; flex-wrap: wrap; gap: 6px; }
.card-platforms>button { min-height: 27px; padding: 4px 7px; border: 1px solid var(--border); border-radius: 5px; background: transparent; display: inline-flex; align-items: center; gap: 4px; color: var(--muted-foreground); cursor: pointer; font-size: 9px; font-weight: 700; transition: border-color .18s ease,color .18s ease; }
.card-platforms svg { width: 14px; height: 14px; }
.card-release { margin-top: 18px; padding: 13px 0; border-block: 1px solid var(--border); display: flex; justify-content: space-between; gap: 12px; font: 600 10px/1.3 ui-monospace,SFMono-Regular,Menlo,monospace; }
.card-release span:last-child { color: var(--muted-foreground); }
.card-actions { margin-top: auto; padding-top: 18px; display: grid; grid-template-columns: 1fr 42px 42px; gap: 7px; }
.card-download,.card-original,.card-more { min-height: 42px; border: 1px solid var(--foreground); border-radius: 7px; background: var(--foreground); color: var(--background)!important; display: inline-flex; align-items: center; justify-content: center; gap: 7px; cursor: pointer; font-size: 12px; font-weight: 750; transition: border-color .18s ease,background .18s ease,transform .18s ease; }
.card-original,.card-more { width: 42px; padding: 0; border-color: var(--border-strong); background: var(--surface); color: var(--foreground)!important; }
.software-card>footer { margin-top: 12px; display: flex; justify-content: space-between; gap: 8px; }
.software-card>footer button { padding: 3px 0; border: 0; background: transparent; display: inline-flex; align-items: center; gap: 4px; color: var(--muted-foreground); cursor: pointer; font-size: 10px; }
.catalog-empty { min-height: 330px; margin-top: 18px; border: 1px dashed var(--border-strong); border-radius: 8px; background: var(--surface); display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 9px; color: var(--muted-foreground); text-align: center; }
.catalog-empty strong { color: var(--foreground); }
.catalog-empty button { min-height: 38px; margin-top: 8px; padding: 7px 12px; border: 1px solid var(--border); border-radius: 6px; background: var(--background); cursor: pointer; }
.modal-layer { position: fixed; inset: 0; z-index: 100; padding: 24px; display: grid; place-items: center; background: rgba(5,10,16,.68); backdrop-filter: blur(8px); }
.software-dialog { width: min(900px,100%); max-height: min(820px,calc(100dvh - 48px)); overflow: hidden; border: 1px solid var(--border-strong); border-radius: 8px; background: var(--surface); box-shadow: var(--shadow); display: flex; flex-direction: column; }
.notes-dialog { width: min(760px,100%); }
.dialog-header { flex: 0 0 auto; padding: 22px 24px; border-bottom: 1px solid var(--border); display: flex; align-items: flex-start; justify-content: space-between; gap: 20px; background: var(--muted); }
.dialog-header .eyebrow { color: var(--accent); }
.dialog-header h2 { margin: 6px 0 0; font-size: 24px; font-weight: 720; }
.dialog-header>div>span { margin-top: 6px; display: block; color: var(--muted-foreground); font-size: 11px; }
.dialog-platform-tabs { flex: 0 0 auto; min-height: 54px; padding: 8px 20px; overflow-x: auto; border-bottom: 1px solid var(--border); background: var(--surface); display: flex; align-items: center; gap: 6px; scrollbar-width: thin; }
.dialog-platform-tabs button { min-height: 36px; padding: 7px 11px; border: 1px solid transparent; border-radius: 6px; background: transparent; display: inline-flex; align-items: center; gap: 6px; color: var(--muted-foreground); cursor: pointer; font-size: 11px; font-weight: 700; white-space: nowrap; transition: border-color .18s ease,background .18s ease,color .18s ease; }
.dialog-platform-tabs button svg { width: 15px; height: 15px; }
.dialog-platform-tabs button>span { min-width: 17px; height: 17px; padding: 0 4px; border-radius: 999px; background: var(--muted); display: inline-grid; place-items: center; font-size: 8px; }
.dialog-platform-tabs button.active { border-color: var(--border-strong); background: var(--muted); color: var(--foreground); }
.dialog-body { min-height: 0; padding: 20px; overflow-y: auto; display: grid; grid-template-columns: repeat(2,minmax(0,1fr)); align-items: start; gap: 12px; }
.dialog-body>.asset-group:only-child { grid-column: 1/-1; }
.asset-group { min-width: 0; padding: 16px; border: 1px solid var(--border); border-radius: 8px; background: var(--muted); transition: border-color .18s ease; }
.asset-group.is-current-platform { border-color: color-mix(in srgb,var(--accent) 38%,var(--border)); }
.asset-group>header { min-height: 50px; display: flex; align-items: center; gap: 12px; }
.platform-icon { width: 48px; height: 48px; flex: 0 0 48px; border: 1px solid var(--border); border-radius: 8px; background: var(--surface); display: grid; place-items: center; }
.platform-icon svg { width: 27px; height: 27px; }
.asset-group h3 { margin: 0; font-size: 18px; font-weight: 650; }
.asset-group-title { display: flex; align-items: center; gap: 8px; }
.asset-group-title>span { padding: 3px 6px; border-radius: 4px; background: color-mix(in srgb,var(--accent) 10%,transparent); color: var(--accent); font-size: 8px; font-weight: 750; }
.asset-group header p { margin: 3px 0 0; color: var(--muted-foreground); font-size: 10px; line-height: 1.4; }
.asset-list { margin-top: 14px; display: grid; gap: 7px; }
.asset-row { min-width: 0; min-height: 56px; padding: 8px 9px 8px 12px; border: 1px solid var(--border); border-radius: 7px; background: var(--surface); display: flex; align-items: center; justify-content: space-between; gap: 10px; transition: border-color .18s ease,box-shadow .18s ease; }
.asset-row.is-recommended { border-color: color-mix(in srgb,var(--accent) 38%,var(--border)); }
.asset-row>div:first-child { min-width: 0; display: grid; gap: 4px; }
.asset-name-line { min-width: 0; display: flex; align-items: center; gap: 7px; }
.asset-row strong { overflow: hidden; font-size: 12px; text-overflow: ellipsis; white-space: nowrap; }
.asset-row span { color: var(--muted-foreground); font-size: 9px; }
.asset-name-line>span { flex: 0 0 auto; padding: 2px 5px; border-radius: 4px; background: color-mix(in srgb,var(--accent) 10%,transparent); color: var(--accent); font-size: 8px; font-weight: 750; }
.asset-actions { flex: 0 0 auto; display: flex; align-items: center; gap: 2px; }
.asset-actions a { width: 36px; height: 36px; padding: 0; border: 0; border-radius: 6px; display: inline-grid; place-items: center; color: var(--muted-foreground); }
.asset-actions a:hover { background: var(--muted-2); color: var(--foreground); }
.asset-actions .accelerated-link { color: var(--accent); }
.dialog-empty { grid-column: 1/-1; min-height: 260px; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 7px; color: var(--muted-foreground); }
.dialog-footer { flex: 0 0 auto; min-height: 60px; padding: 11px 22px; border-top: 1px solid var(--border); display: flex; flex-direction: row; align-items: center; justify-content: space-between; gap: 16px; color: var(--muted-foreground); font-size: 10px; }
.dialog-footer>div { display: flex; align-items: center; gap: 12px; }
.dialog-footer a { display: inline-flex; align-items: center; gap: 4px; color: var(--accent); font-weight: 700; }
.dialog-footer button { min-height: 38px; padding: 7px 12px; border: 1px solid var(--border); border-radius: 6px; background: var(--background); cursor: pointer; font-size: 11px; font-weight: 700; }
.release-body { min-height: 0; padding: 26px 28px; overflow-y: auto; color: var(--foreground); font-size: 14px; line-height: 1.65; overflow-wrap: anywhere; }
.markdown-body :deep(*) { box-sizing: border-box; }
.markdown-body :deep(> :first-child) { margin-top: 0!important; }
.markdown-body :deep(> :last-child) { margin-bottom: 0!important; }
.markdown-body :deep(h1),.markdown-body :deep(h2),.markdown-body :deep(h3),.markdown-body :deep(h4) { margin: 24px 0 16px; color: var(--foreground); font-weight: 650; line-height: 1.25; }
.markdown-body :deep(h1),.markdown-body :deep(h2) { padding-bottom: 8px; border-bottom: 1px solid var(--border); }
.markdown-body :deep(h1) { font-size: 26px; }
.markdown-body :deep(h2) { font-size: 21px; }
.markdown-body :deep(h3) { font-size: 17px; }
.markdown-body :deep(h4) { font-size: 14px; }
.markdown-body :deep(p),.markdown-body :deep(blockquote),.markdown-body :deep(ul),.markdown-body :deep(ol),.markdown-body :deep(pre),.markdown-body :deep(table) { margin: 0 0 16px; }
.markdown-body :deep(ul),.markdown-body :deep(ol) { padding-left: 2em; }
.markdown-body :deep(li+li) { margin-top: 5px; }
.markdown-body :deep(li>p) { margin-top: 12px; }
.markdown-body :deep(a) { color: var(--accent); text-decoration: underline; text-underline-offset: 3px; }
.markdown-body :deep(blockquote) { padding: 0 1em; border-left: 4px solid var(--border-strong); color: var(--muted-foreground); }
.markdown-body :deep(hr) { height: 1px; margin: 24px 0; border: 0; background: var(--border); }
.markdown-body :deep(code) { padding: 2px 5px; border-radius: 4px; background: var(--muted-2); color: var(--foreground); font: 85%/1.45 ui-monospace,SFMono-Regular,Menlo,monospace; }
.markdown-body :deep(pre) { padding: 16px; overflow: auto; border: 1px solid var(--border); border-radius: 6px; background: var(--muted); }
.markdown-body :deep(pre code) { padding: 0; background: transparent; font-size: 12px; }
.markdown-body :deep(table) { width: max-content; max-width: 100%; border-spacing: 0; overflow: auto; display: block; }
.markdown-body :deep(th),.markdown-body :deep(td) { padding: 7px 12px; border: 1px solid var(--border); }
.markdown-body :deep(th) { background: var(--muted); font-weight: 650; }
.markdown-body :deep(tr:nth-child(2n)) { background: var(--muted); }
.markdown-body :deep(img) { max-width: 100%; height: auto; border-radius: 6px; }
.button:active,.icon-button:active,.primary-link:active,.refresh-all:active,.card-download:active,.card-original:active,.card-more:active,.platform-row>button:active,.catalog-jump:active { transform: translateY(1px); }
@media (hover:hover) {
  .icon-button:hover { border-color: var(--border-strong); background: var(--muted); transform: translateY(-1px); }
  .primary-link:hover { box-shadow: var(--soft-shadow); transform: translateY(-1px); }
  .button:hover { border-color: var(--foreground); box-shadow: 0 10px 24px rgba(15,23,42,.1); transform: translateY(-1px); }
  .platform-row>button:hover,.card-platforms>button:hover { border-color: var(--accent); color: var(--foreground); transform: translateY(-1px); }
  .catalog-jump:hover { color: var(--foreground)!important; transform: translateY(2px); }
  .refresh-all:hover { border-color: var(--foreground); transform: translateY(-1px); }
  .platform-filter button:not(.active):hover,.dialog-platform-tabs button:not(.active):hover { background: var(--muted); color: var(--foreground); }
  .software-card:hover { border-color: var(--border-strong); box-shadow: var(--soft-shadow); transform: translateY(-2px); }
  .asset-row:hover { border-color: var(--border-strong); box-shadow: 0 5px 14px rgba(15,23,42,.05); }
  .card-download:hover,.card-original:hover,.card-more:hover { transform: translateY(-1px); }
}
.spinning { animation: spin .8s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }
@keyframes pulse { 50% { opacity: .35; transform: scale(.75); } }
@media (prefers-reduced-motion: reduce) { .software-page *,.software-page *::before,.software-page *::after { animation-duration: .01ms!important; animation-iteration-count: 1!important; transition-duration: .01ms!important; } }
@media (max-width: 900px) {
  .featured-product { width: min(720px,100%); margin: 0 auto; }
  .software-grid { grid-template-columns: repeat(2,minmax(0,1fr)); }
}
@media (max-width: 760px) {
  .software-header { padding: 12px 10px; }
  .software-header.is-scrolled { padding: 10px; }
  .software-header.is-scrolled .nav-shell { border-radius: 8px; }
  .nav-shell { gap: 8px; }
  .nav-links { position: absolute; top: calc(100% + 8px); left: 0; right: 0; padding: 8px; border: 1px solid var(--border); border-radius: 8px; background: var(--panel); box-shadow: var(--shadow); display: none; align-items: stretch; flex-direction: column; }
  .nav-links.is-open { display: flex; }
  .nav-links a { width: 100%; justify-content: flex-start; }
  .menu-button { display: inline-grid; }
  .hero-status { grid-template-columns: repeat(2,1fr); }
  .hero-status>div:nth-child(2) { border-right: 0; }
  .hero-status>div:nth-child(-n+2) { border-bottom: 1px solid var(--border); }
  .catalog-tools { align-items: stretch; flex-direction: column; }
  .platform-filter { overflow-x: auto; }
  .platform-filter button { flex: 0 0 auto; }
  .catalog-summary { align-items: flex-start; }
  .dialog-body { grid-template-columns: 1fr; }
}
@media (max-width: 620px) {
  .brand-name { font-size: 14px; }
  .primary-link { display: none; }
  .software-hero { padding: 112px 14px 112px; }
  .featured-logo { width: 92px; height: 92px; }
  .featured-product h1 { font-size: 50px; }
  .hero-copy { max-width: 350px; font-size: 14px; }
  .featured-meta { gap: 7px; font-size: 10px; }
  .hero-actions { width: 100%; }
  .hero-actions .button { flex: 1 1 145px; }
  .hero-actions .button.primary { flex-basis: 100%; }
  .featured-links { gap: 5px 14px; }
  .catalog-section { padding: 72px 14px; }
  .section-heading { align-items: flex-start; flex-direction: column; }
  .software-grid { grid-template-columns: 1fr; }
  .software-card { min-height: 380px; }
  .modal-layer { padding: 10px; }
  .software-dialog { max-height: calc(100dvh - 20px); }
  .dialog-header,.release-body { padding: 17px; }
  .dialog-header h2 { font-size: 21px; }
  .dialog-platform-tabs { padding-inline: 12px; }
  .dialog-body { padding: 12px; }
  .asset-group { padding: 12px; }
  .platform-icon { width: 42px; height: 42px; flex-basis: 42px; }
  .platform-icon svg { width: 24px; height: 24px; }
  .asset-group h3 { font-size: 16px; }
  .asset-row { padding-left: 10px; }
  .dialog-footer { min-height: 82px; padding: 10px 17px; flex-direction: column-reverse; justify-content: center; gap: 7px; text-align: center; }
}
</style>
