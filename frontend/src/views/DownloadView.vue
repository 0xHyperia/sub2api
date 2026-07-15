<template>
  <div class="download-page" :data-theme="theme" @scroll.passive="handlePageScroll">
    <header class="download-header" :class="{ 'is-scrolled': headerScrolled }" aria-label="主导航">
      <div class="nav-shell">
        <RouterLink class="brand" to="/home" :aria-label="`${siteName} 首页`">
          <span class="brand-mark" aria-hidden="true"><img src="/home-experiment/logo.png" alt="" /></span>
          <span class="brand-name">{{ siteName }}</span>
        </RouterLink>

        <nav id="download-navigation" class="nav-links" :class="{ 'is-open': mobileNavOpen }" aria-label="页面导航">
          <RouterLink to="/home">首页</RouterLink>
          <RouterLink to="/key-usage">Key 用量</RouterLink>
          <RouterLink class="active" to="/download" aria-current="page">客户端下载</RouterLink>
        </nav>

        <div class="nav-actions">
          <button
            class="icon-button theme-button"
            type="button"
            :aria-label="theme === 'dark' ? '切换到浅色主题' : '切换到深色主题'"
            :title="theme === 'dark' ? '切换到浅色主题' : '切换到深色主题'"
            @click="toggleTheme"
          >
            <Icon :name="theme === 'dark' ? 'sun' : 'moon'" size="md" aria-hidden="true" />
          </button>
          <RouterLink class="primary-link" :to="entryPath">{{ entryLabel }}</RouterLink>
          <button
            class="icon-button menu-button"
            type="button"
            :aria-expanded="mobileNavOpen"
            aria-controls="download-navigation"
            :aria-label="mobileNavOpen ? '关闭页面导航' : '打开页面导航'"
            :title="mobileNavOpen ? '关闭导航' : '打开导航'"
            @click="mobileNavOpen = !mobileNavOpen"
          >
            <Icon :name="mobileNavOpen ? 'x' : 'menu'" size="md" aria-hidden="true" />
          </button>
        </div>
      </div>
    </header>

    <main>
      <section class="download-hero" aria-labelledby="download-title">
        <canvas ref="heroFlowCanvas" class="fluid-surface hero-flow" aria-hidden="true"></canvas>
        <div class="hero-content">
          <div class="app-logo-shell">
            <img class="app-logo" src="/home-experiment/logo.png" alt="ZeroBox" />
          </div>
          <p class="eyebrow"><span></span> ZEROBOX 桌面客户端</p>
          <h1 id="download-title">ZeroBox 客户端</h1>
          <p class="hero-copy">原生桌面与移动体验，让模型、会话和工作流始终触手可及。</p>

          <div class="release-line" aria-live="polite">
            <span class="release-version">{{ release?.tag_name || '正在同步' }}</span>
            <span class="release-separator" aria-hidden="true"></span>
            <span>{{ publishedDate }}</span>
            <button
              class="refresh-release"
              type="button"
              :disabled="isRefreshing"
              aria-label="重新同步最新版本"
              title="重新同步最新版本"
              @click="loadRelease(true)"
            >
              <Icon name="refresh" size="sm" :class="{ spinning: isRefreshing }" aria-hidden="true" />
            </button>
          </div>

          <div class="hero-actions">
            <a
              v-if="recommendedDownload"
              class="button primary"
              :href="recommendedDownload.asset.browser_download_url"
              rel="noopener noreferrer"
            >
              <Icon name="download" size="md" aria-hidden="true" />
              下载 {{ recommendedGroup?.name }} 版
            </a>
            <button v-else class="button primary" type="button" disabled>
              暂无匹配安装包
            </button>
            <button class="button secondary" type="button" @click="openDownloads">
              <Icon name="grid" size="md" aria-hidden="true" />
              查看其他版本
            </button>
            <button class="button secondary" type="button" @click="openReleaseNotes">
              <Icon name="document" size="md" aria-hidden="true" />
              查看更新日志
            </button>
          </div>

          <div class="sync-status" :class="`source-${releaseSource}`">
            <span aria-hidden="true"></span>
            {{ syncStatusText }}
          </div>
        </div>

        <div class="protocol-strip" aria-label="客户端发布信息">
          <div><span>发布通道</span><strong>稳定版</strong></div>
          <div><span>发布来源</span><strong>官方发布</strong></div>
          <div><span>更新方式</span><strong>自动同步</strong></div>
          <div class="protocol-pulse"><i></i><span>运行状态</span><strong>在线</strong></div>
        </div>
      </section>
    </main>

    <div v-if="downloadsOpen" class="modal-layer" role="presentation" @mousedown.self="closeDownloads">
      <section
        class="release-dialog downloads-dialog"
        role="dialog"
        aria-modal="true"
        aria-labelledby="downloads-dialog-title"
        tabindex="-1"
      >
        <header class="dialog-header">
          <div>
            <p class="eyebrow">客户端下载 · {{ release?.tag_name }}</p>
            <h2 id="downloads-dialog-title">选择其他平台</h2>
            <span>选择适合当前设备的安装包并直接下载</span>
          </div>
          <button ref="downloadsDialogCloseRef" class="icon-button" type="button" aria-label="关闭其他版本" title="关闭" @click="closeDownloads">
            <Icon name="x" size="md" aria-hidden="true" />
          </button>
        </header>

        <div class="downloads-dialog-body">
          <div v-if="downloads.length" class="download-options-grid">
            <article v-for="group in downloads" :key="group.id" class="download-option">
              <header class="download-option-header">
                <span class="platform-logo" :class="`logo-${group.id}`">
                  <PlatformLogo :platform="group.id" />
                </span>
                <div>
                  <h3>{{ group.name }}</h3>
                  <p>{{ group.description }}</p>
                </div>
              </header>

              <div class="download-option-links">
                <a
                  v-for="variant in group.variants"
                  :key="variant.asset.name"
                  :href="variant.asset.browser_download_url"
                  rel="noopener noreferrer"
                  :aria-label="`下载 ${group.name} ${variant.label}，${variant.hint}`"
                >
                  <span>
                    <strong>{{ variant.label }}</strong>
                    <small>{{ variant.hint }}<template v-if="variant.asset.size"> · {{ formatAssetSize(variant.asset.size) }}</template></small>
                  </span>
                  <Icon name="download" size="md" aria-hidden="true" />
                </a>
              </div>
            </article>
          </div>

          <div v-else class="downloads-empty">
            <Icon name="cloud" size="xl" aria-hidden="true" />
            <strong>暂未找到可用安装包</strong>
            <span>请稍后重新获取最新版本。</span>
          </div>
        </div>

        <footer class="dialog-footer">
          <span>安装包同步自官方稳定发布通道</span>
          <button type="button" @click="closeDownloads">关闭</button>
        </footer>
      </section>
    </div>

    <div v-if="releaseNotesOpen" class="modal-layer" role="presentation" @mousedown.self="closeReleaseNotes">
      <section
        class="release-dialog"
        role="dialog"
        aria-modal="true"
        aria-labelledby="release-dialog-title"
        tabindex="-1"
      >
        <header class="dialog-header">
          <div>
            <p class="eyebrow">更新日志 · {{ release?.tag_name }}</p>
            <h2 id="release-dialog-title">{{ release?.name || '最新更新' }}</h2>
            <span>{{ publishedDate }}</span>
          </div>
          <button ref="dialogCloseRef" class="icon-button" type="button" aria-label="关闭更新日志" title="关闭" @click="closeReleaseNotes">
            <Icon name="x" size="md" aria-hidden="true" />
          </button>
        </header>
        <div class="release-body" v-html="releaseNotesHtml"></div>
        <footer class="dialog-footer">
          <span>内容同步自官方最新发布</span>
          <button type="button" @click="closeReleaseNotes">关闭</button>
        </footer>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import DOMPurify from 'dompurify'
import { marked } from 'marked'
import { getLatestClientRelease, type ClientReleaseResult } from '@/api/clientRelease'
import Icon from '@/components/icons/Icon.vue'
import PlatformLogo from '@/components/icons/PlatformLogo.vue'
import { useTheme } from '@/composables/useTheme'
import { useAppStore, useAuthStore } from '@/stores'
import {
  detectClientPlatform,
  formatAssetSize,
  selectClientDownloads,
  type ClientDownloadVariant
} from '@/utils/clientRelease'

const appStore = useAppStore()
const authStore = useAuthStore()
const { resolvedTheme: theme, toggleTheme } = useTheme()

const releaseResult = ref<ClientReleaseResult | null>(null)
const isRefreshing = ref(false)
const mobileNavOpen = ref(false)
const headerScrolled = ref(false)
const releaseNotesOpen = ref(false)
const downloadsOpen = ref(false)
const dialogCloseRef = ref<HTMLButtonElement | null>(null)
const downloadsDialogCloseRef = ref<HTMLButtonElement | null>(null)
const heroFlowCanvas = ref<HTMLCanvasElement | null>(null)
const detectedPlatform = ref(detectClientPlatform())
let notesTrigger: HTMLElement | null = null
let downloadsTrigger: HTMLElement | null = null
const flowCleanupCallbacks: Array<() => void> = []

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'USA-零')
const entryPath = computed(() => authStore.isAuthenticated ? (authStore.isAdmin ? '/admin/dashboard' : '/dashboard') : '/login')
const entryLabel = computed(() => authStore.isAuthenticated ? '进入控制台' : '开始接入')
const release = computed(() => releaseResult.value?.release || null)
const releaseSource = computed(() => releaseResult.value?.source || 'github')
const downloads = computed(() => selectClientDownloads(release.value?.assets || []))
const recommendedGroup = computed(() => downloads.value.find((group) => group.id === detectedPlatform.value) || downloads.value[0])
const recommendedDownload = computed<ClientDownloadVariant | undefined>(() => recommendedGroup.value?.variants[0])
const publishedDate = computed(() => {
  if (!release.value?.published_at) return '获取最新版本中'
  return new Intl.DateTimeFormat('zh-CN', { year: 'numeric', month: 'long', day: 'numeric' }).format(new Date(release.value.published_at))
})
const syncStatusText = computed(() => ({
  github: '已同步官方最新发布',
  cache: '已载入最近同步版本',
  fallback: '发布服务繁忙，已载入可用版本'
}[releaseSource.value]))
const releaseNotesHtml = computed(() => {
  const source = release.value?.body || '本次发布暂未提供更新说明。'
  const rendered = marked.parse(source, { async: false }) as string
  return sanitizeReleaseNotes(rendered)
})

function sanitizeReleaseNotes(content: string): string {
  const sanitized = DOMPurify.sanitize(content, { USE_PROFILES: { html: true } })
  if (typeof document === 'undefined') return sanitized

  const template = document.createElement('template')
  template.innerHTML = sanitized
  template.content.querySelectorAll<HTMLAnchorElement>('a[href]').forEach((link) => {
    try {
      const hostname = new URL(link.href, window.location.origin).hostname.toLowerCase()
      if (hostname === 'github.com' || hostname.endsWith('.github.com')) {
        link.replaceWith(document.createTextNode(link.textContent || ''))
      }
    } catch {
      link.removeAttribute('href')
    }
  })
  return template.innerHTML
}

async function loadRelease(forceRefresh = false): Promise<void> {
  isRefreshing.value = true
  try {
    releaseResult.value = await getLatestClientRelease(forceRefresh)
  } finally {
    isRefreshing.value = false
  }
}

function openReleaseNotes(event?: Event): void {
  notesTrigger = event?.currentTarget instanceof HTMLElement ? event.currentTarget : document.activeElement as HTMLElement
  releaseNotesOpen.value = true
}

function closeReleaseNotes(): void {
  releaseNotesOpen.value = false
  nextTick(() => notesTrigger?.focus())
}

function openDownloads(event?: Event): void {
  downloadsTrigger = event?.currentTarget instanceof HTMLElement ? event.currentTarget : document.activeElement as HTMLElement
  downloadsOpen.value = true
}

function closeDownloads(): void {
  downloadsOpen.value = false
  nextTick(() => downloadsTrigger?.focus())
}

function handleKeydown(event: KeyboardEvent): void {
  if (event.key !== 'Escape') return
  if (downloadsOpen.value) closeDownloads()
  else if (releaseNotesOpen.value) closeReleaseNotes()
}

function handlePageScroll(event: Event): void {
  headerScrolled.value = (event.currentTarget as HTMLElement).scrollTop > 24
}

function startFluidSurface(canvas: HTMLCanvasElement | null): void {
  if (!canvas) return
  const context = canvas.getContext('2d')
  if (!context) return
  const surfaceCanvas: HTMLCanvasElement = canvas
  const drawingContext: CanvasRenderingContext2D = context

  const reducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches
  let animationFrame: number | undefined
  let visible = true
  let startedAt = performance.now()
  let lastDrawAt = 0
  const frameInterval = 1000 / 24

  const palettes = {
    heroLight: ['rgba(37, 99, 235, .075)', 'rgba(14, 116, 144, .065)', 'rgba(71, 85, 105, .055)', 'rgba(217, 119, 6, .035)'],
    heroDark: ['rgba(96, 165, 250, .11)', 'rgba(45, 212, 191, .075)', 'rgba(203, 213, 225, .055)', 'rgba(251, 191, 36, .04)']
  }

  function resize(): void {
    const bounds = surfaceCanvas.getBoundingClientRect()
    const ratio = Math.min(window.devicePixelRatio || 1, 1.25)
    surfaceCanvas.width = Math.max(1, Math.round(bounds.width * ratio))
    surfaceCanvas.height = Math.max(1, Math.round(bounds.height * ratio))
    drawingContext.setTransform(ratio, 0, 0, ratio, 0, 0)
  }

  function draw(now = performance.now()): void {
    const width = surfaceCanvas.clientWidth
    const height = surfaceCanvas.clientHeight
    if (!width || !height) return

    const elapsed = reducedMotion ? 0 : (now - startedAt) / 1000
    drawingContext.clearRect(0, 0, width, height)

    const palette = theme.value === 'dark' ? palettes.heroDark : palettes.heroLight
    drawingContext.globalCompositeOperation = theme.value === 'dark' ? 'screen' : 'multiply'

    const bandCount = 7
    for (let band = 0; band < bandCount; band += 1) {
      const progress = band / Math.max(1, bandCount - 1)
      const baseline = height * (0.06 + progress * 0.88)
      const amplitude = height * (0.035 + (band % 3) * 0.012)
      const thickness = height * (0.055 + (band % 2) * 0.025)
      const frequency = 1.25 + band * 0.17
      const phase = band * 0.94 + elapsed * (0.12 + band * 0.008)
      const step = Math.max(12, Math.floor(width / 90))

      const waveY = (x: number, offset = 0): number => {
        const normalized = x / width
        return baseline
          + Math.sin(normalized * Math.PI * 2 * frequency + phase + offset) * amplitude
          + Math.sin(normalized * Math.PI * 2 * .46 - phase * .62 + band) * amplitude * .48
      }

      drawingContext.beginPath()
      drawingContext.moveTo(0, waveY(0))
      for (let x = step; x <= width + step; x += step) drawingContext.lineTo(x, waveY(x))
      for (let x = width + step; x >= 0; x -= step) {
        drawingContext.lineTo(x, waveY(x, .42) + thickness)
      }
      drawingContext.closePath()
      drawingContext.fillStyle = palette[band % palette.length]
      drawingContext.fill()

      drawingContext.beginPath()
      drawingContext.moveTo(0, waveY(0))
      for (let x = step; x <= width + step; x += step) drawingContext.lineTo(x, waveY(x))
      drawingContext.strokeStyle = palette[(band + 1) % palette.length].replace(/\.[0-9]+\)$/, '.18)')
      drawingContext.lineWidth = band % 3 === 0 ? 1.1 : .65
      drawingContext.stroke()
    }

    drawingContext.globalCompositeOperation = 'source-over'
  }

  function animate(now: number): void {
    animationFrame = undefined
    if (!visible) return
    if (now - lastDrawAt >= frameInterval) {
      draw(now)
      lastDrawAt = now
    }
    animationFrame = window.requestAnimationFrame(animate)
  }

  const resizeObserver = new ResizeObserver(() => {
    resize()
    if (reducedMotion || !animationFrame) draw()
  })
  resizeObserver.observe(surfaceCanvas)
  resize()

  const intersectionObserver = new IntersectionObserver(([entry]) => {
    visible = entry?.isIntersecting !== false
    if (visible && !reducedMotion && !animationFrame) {
      startedAt = performance.now()
      animationFrame = window.requestAnimationFrame(animate)
    }
    if (!visible && animationFrame) {
      window.cancelAnimationFrame(animationFrame)
      animationFrame = undefined
    }
  }, { threshold: 0.01 })
  intersectionObserver.observe(surfaceCanvas)

  draw()
  if (!reducedMotion) animationFrame = window.requestAnimationFrame(animate)
  flowCleanupCallbacks.push(() => {
    if (animationFrame) window.cancelAnimationFrame(animationFrame)
    resizeObserver.disconnect()
    intersectionObserver.disconnect()
  })
}

watch([releaseNotesOpen, downloadsOpen], async ([notesOpen, otherDownloadsOpen]) => {
  document.body.style.overflow = notesOpen || otherDownloadsOpen ? 'hidden' : ''
  if (!notesOpen && !otherDownloadsOpen) return

  await nextTick()
  if (otherDownloadsOpen) downloadsDialogCloseRef.value?.focus()
  else dialogCloseRef.value?.focus()
})

onMounted(() => {
  authStore.checkAuth()
  if (!appStore.publicSettingsLoaded) void appStore.fetchPublicSettings()
  void loadRelease()
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
.download-page {
  color-scheme: light;
  --background: #ffffff;
  --foreground: #0f172a;
  --muted: #f4f7fb;
  --muted-2: #eef3f8;
  --muted-foreground: #667085;
  --border: #dbe3ee;
  --border-strong: #c7d3e2;
  --primary: #475569;
  --accent: #2563eb;
  --green: #059669;
  --amber: #d97706;
  --rose: #e11d48;
  --panel: rgba(255, 255, 255, 0.94);
  --surface: #ffffff;
  --shadow: 0 24px 60px rgba(15, 23, 42, 0.13);
  --soft-shadow: 0 14px 36px rgba(15, 23, 42, 0.07);
  --font: "Public Sans", Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
  min-width: 320px;
  height: 100dvh;
  overflow-x: hidden;
  overflow-y: auto;
  overscroll-behavior-y: contain;
  background: var(--background);
  color: var(--foreground);
  font-family: var(--font);
  letter-spacing: 0;
  isolation: isolate;
}

.download-page[data-theme="dark"] {
  color-scheme: dark;
  --background: #101418;
  --foreground: #eef4fb;
  --muted: #161c22;
  --muted-2: #1b232b;
  --muted-foreground: #9caab9;
  --border: #2d3946;
  --border-strong: #405064;
  --primary: #cbd5e1;
  --accent: #60a5fa;
  --panel: rgba(16, 20, 24, 0.94);
  --surface: #161c22;
  --shadow: 0 24px 60px rgba(0, 0, 0, 0.38);
  --soft-shadow: 0 14px 36px rgba(0, 0, 0, 0.24);
}

.download-page * { box-sizing: border-box; }
.download-page a { color: inherit; text-decoration: none; }
.download-page button { color: inherit; font: inherit; }
.download-page :where(a, button, [tabindex]):focus-visible { outline: 2px solid var(--accent); outline-offset: 3px; }

.download-header {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  z-index: 30;
  padding: 18px;
  pointer-events: none;
  transition: padding .32s ease;
}

.nav-shell {
  position: relative;
  max-width: 1180px;
  min-height: 56px;
  margin: 0 auto;
  padding: 7px;
  display: flex;
  align-items: center;
  gap: 16px;
  border: 1px solid transparent;
  border-radius: 0;
  background: transparent;
  box-shadow: 0 0 0 rgba(15, 23, 42, 0);
  backdrop-filter: blur(0);
  pointer-events: auto;
  transition: background .38s ease, border-color .38s ease, box-shadow .38s ease, border-radius .38s ease, backdrop-filter .38s ease;
}

.download-header.is-scrolled { position: fixed; padding: 14px 18px; }
.download-header.is-scrolled .nav-shell { border-color: color-mix(in srgb, var(--border) 86%, transparent); border-radius: 999px; background: color-mix(in srgb, var(--panel) 90%, transparent); box-shadow: var(--soft-shadow); backdrop-filter: blur(18px) saturate(135%); }

.brand { display: inline-flex; align-items: center; gap: 9px; min-width: max-content; padding: 4px 10px 4px 5px; font-weight: 750; }
.brand-mark { width: 36px; height: 36px; overflow: hidden; border: 1px solid var(--border); border-radius: 50%; background: var(--background); }
.brand-mark img { width: 100%; height: 100%; display: block; object-fit: cover; }
.nav-links { flex: 1; display: flex; align-items: center; justify-content: center; gap: 2px; color: var(--muted-foreground); font-size: 14px; font-weight: 600; }
.nav-links a { min-height: 42px; padding: 8px 13px; display: inline-flex; align-items: center; border-radius: 999px; }
.nav-links a:hover, .nav-links a:focus-visible { background: var(--muted); color: var(--foreground); }
.nav-links a.active { color: var(--foreground); font-weight: 700; }
.nav-actions { margin-left: auto; display: flex; align-items: center; gap: 8px; }
.icon-button { width: 42px; height: 42px; border: 1px solid var(--border); border-radius: 50%; background: var(--background); display: inline-grid; place-items: center; cursor: pointer; }
.primary-link { min-height: 42px; padding: 9px 15px; border-radius: 999px; background: var(--foreground); color: var(--background) !important; display: inline-flex; align-items: center; font-size: 14px; font-weight: 700; }
.menu-button { display: none; }

.download-hero {
  position: relative;
  min-height: 100svh;
  padding: 122px 18px 26px;
  overflow: hidden;
  display: grid;
  place-items: center;
  isolation: isolate;
  background: var(--background);
}

.fluid-surface { position: absolute; top: -12%; left: -12%; width: 124%; height: 124%; display: block; pointer-events: none; }
.hero-flow { z-index: -3; opacity: .7; filter: blur(6px) saturate(108%); }

.hero-content { width: min(800px, 100%); text-align: center; display: flex; flex-direction: column; align-items: center; }
.app-logo-shell { width: 96px; height: 96px; display: grid; place-items: center; background: transparent; }
.app-logo { width: 88px; height: 88px; display: block; object-fit: contain; background: transparent; filter: drop-shadow(0 18px 18px rgba(15, 23, 42, .2)); }
.download-page[data-theme="dark"] .app-logo { filter: drop-shadow(0 18px 22px rgba(0, 0, 0, .48)); }
.eyebrow { margin: 20px 0 0; color: var(--muted-foreground); font: 700 12px/1.4 ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; text-transform: uppercase; }
.hero-content .eyebrow { display: inline-flex; align-items: center; gap: 8px; }
.hero-content .eyebrow span { width: 7px; height: 7px; border-radius: 50%; background: var(--green); box-shadow: 0 0 0 5px color-mix(in srgb, var(--green) 12%, transparent); }
.hero-content h1 { margin: 14px 0 0; font-size: clamp(46px, 7vw, 76px); line-height: 1; font-weight: 780; }
.hero-copy { max-width: 620px; margin: 20px auto 0; color: var(--muted-foreground); font-size: 17px; line-height: 1.75; }

.release-line { margin-top: 20px; display: flex; align-items: center; justify-content: center; gap: 9px; color: var(--muted-foreground); font: 600 13px/1.4 ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; }
.release-version { color: var(--foreground); }
.release-separator { width: 1px; height: 14px; background: var(--border-strong); }
.refresh-release { width: 30px; height: 30px; border: 0; border-radius: 50%; background: transparent; display: inline-grid; place-items: center; cursor: pointer; }
.refresh-release:hover { background: var(--muted); color: var(--foreground); }
.refresh-release:disabled { cursor: wait; opacity: .65; }
.spinning { animation: spin .8s linear infinite; }

.hero-actions { margin-top: 26px; display: flex; flex-wrap: wrap; justify-content: center; gap: 10px; }
.button { min-height: 46px; padding: 11px 18px; border: 1px solid var(--border-strong); border-radius: 8px; display: inline-flex; align-items: center; justify-content: center; gap: 9px; cursor: pointer; font-size: 14px; font-weight: 720; }
.button.primary { border-color: var(--foreground); background: var(--foreground); color: var(--background); box-shadow: 0 14px 28px rgba(15, 23, 42, .14); }
.button.primary:hover { transform: translateY(-1px); }
.button.secondary { background: var(--surface); color: var(--foreground); }
.button.secondary:hover { border-color: var(--foreground); }
.sync-status { min-height: 28px; margin-top: 18px; display: flex; align-items: center; gap: 8px; color: var(--muted-foreground); font-size: 12px; }
.sync-status > span { width: 6px; height: 6px; border-radius: 50%; background: var(--green); }
.sync-status.source-fallback > span { background: var(--amber); }

.protocol-strip { position: absolute; bottom: 22px; left: 50%; width: min(760px, calc(100% - 36px)); transform: translateX(-50%); display: grid; grid-template-columns: repeat(4, 1fr); border: 1px solid var(--border); border-radius: 8px; background: color-mix(in srgb, var(--panel) 90%, transparent); backdrop-filter: blur(12px); }
.protocol-strip > div { min-width: 0; padding: 11px 15px; display: flex; align-items: center; gap: 9px; border-right: 1px solid var(--border); }
.protocol-strip > div:last-child { border-right: 0; }
.protocol-strip span { color: var(--muted-foreground); font: 600 9px/1.2 ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; }
.protocol-strip strong { min-width: 0; overflow: hidden; text-overflow: ellipsis; color: var(--foreground); font: 650 11px/1.2 ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; }
.protocol-pulse i { width: 6px; height: 6px; border-radius: 50%; background: var(--green); animation: pulse 2s ease-in-out infinite; }

.downloads-dialog { width: min(980px, 100%); }
.downloads-dialog-body { min-height: 0; padding: 20px; overflow-y: auto; background: var(--background); }
.download-options-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 12px; }
.download-option { min-width: 0; padding: 16px; border: 1px solid var(--border); border-radius: 8px; background: var(--muted); }
.download-option-header { display: flex; align-items: center; gap: 13px; }
.platform-logo { width: 52px; height: 52px; flex: 0 0 52px; border: 1px solid var(--border); border-radius: 8px; display: grid; place-items: center; background: var(--surface); box-shadow: inset 0 1px 0 color-mix(in srgb, var(--foreground) 6%, transparent); }
.platform-logo svg { width: 31px; height: 31px; display: block; fill: currentColor; }
.logo-windows { color: #2563eb; }
.logo-macos { color: var(--foreground); }
.logo-android { color: #3ddc84; }
.logo-android :deep(path) { fill: currentColor; }
.download-option-header h3 { margin: 0 0 4px; font-size: 17px; line-height: 1.2; }
.download-option-header p { margin: 0; color: var(--muted-foreground); font-size: 11px; line-height: 1.45; }
.download-option-links { margin-top: 14px; display: grid; gap: 7px; }
.download-option-links a { min-height: 54px; padding: 9px 11px; border: 1px solid var(--border); border-radius: 7px; background: var(--surface); display: flex; align-items: center; justify-content: space-between; gap: 12px; }
.download-option-links a:hover { border-color: var(--border-strong); background: var(--background); }
.download-option-links a > span { min-width: 0; display: grid; gap: 3px; }
.download-option-links strong { font-size: 12px; }
.download-option-links small { overflow: hidden; color: var(--muted-foreground); font-size: 10px; text-overflow: ellipsis; white-space: nowrap; }
.download-option-links svg { flex: 0 0 auto; color: var(--accent); }
.downloads-empty { min-height: 260px; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 8px; color: var(--muted-foreground); text-align: center; }
.downloads-empty strong { color: var(--foreground); }
.downloads-empty span { font-size: 12px; }

.modal-layer { position: fixed; inset: 0; z-index: 100; padding: 24px; display: grid; place-items: center; background: rgba(5, 10, 16, .68); backdrop-filter: blur(8px); }
.release-dialog { width: min(760px, 100%); max-height: min(820px, calc(100dvh - 48px)); overflow: hidden; border: 1px solid var(--border-strong); border-radius: 8px; background: var(--surface); box-shadow: var(--shadow); display: flex; flex-direction: column; }
.dialog-header { flex: 0 0 auto; padding: 22px 24px; border-bottom: 1px solid var(--border); display: flex; align-items: flex-start; justify-content: space-between; gap: 20px; background: var(--muted); }
.dialog-header .eyebrow { margin: 0 0 8px; color: var(--accent); }
.dialog-header h2 { margin: 0; font-size: 23px; line-height: 1.25; }
.dialog-header > div > span { margin-top: 7px; display: block; color: var(--muted-foreground); font-size: 12px; }
.release-body { min-height: 0; padding: 24px; overflow-y: auto; color: var(--muted-foreground); font-size: 14px; line-height: 1.78; }
.release-body :deep(h1), .release-body :deep(h2), .release-body :deep(h3) { margin: 28px 0 10px; color: var(--foreground); line-height: 1.3; }
.release-body :deep(h1:first-child), .release-body :deep(h2:first-child) { margin-top: 0; }
.release-body :deep(h1) { font-size: 22px; }
.release-body :deep(h2) { padding-bottom: 8px; border-bottom: 1px solid var(--border); font-size: 18px; }
.release-body :deep(h3) { font-size: 15px; }
.release-body :deep(p), .release-body :deep(ul), .release-body :deep(ol) { margin: 10px 0; }
.release-body :deep(ul), .release-body :deep(ol) { padding-left: 21px; }
.release-body :deep(code) { padding: 2px 5px; border: 1px solid var(--border); border-radius: 4px; background: var(--muted); color: var(--foreground); font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; font-size: .9em; overflow-wrap: anywhere; }
.release-body :deep(a) { color: var(--accent); text-decoration: underline; text-underline-offset: 3px; }
.dialog-footer { flex: 0 0 auto; min-height: 62px; padding: 12px 24px; border-top: 1px solid var(--border); display: flex; align-items: center; justify-content: space-between; gap: 18px; color: var(--muted-foreground); font-size: 11px; }
.dialog-footer button { min-height: 38px; padding: 8px 12px; border: 1px solid var(--border); border-radius: 6px; background: var(--background); color: var(--foreground); cursor: pointer; font-size: 12px; font-weight: 700; }
.dialog-footer button:hover { background: var(--muted); }

@keyframes spin { to { transform: rotate(360deg); } }
@keyframes pulse { 50% { opacity: .35; transform: scale(.75); } }

@media (prefers-reduced-motion: reduce) {
  .download-page *, .download-page *::before, .download-page *::after { animation-duration: .01ms !important; animation-iteration-count: 1 !important; transition-duration: .01ms !important; }
}

@media (max-width: 760px) {
  .download-header { padding: 12px 10px; }
  .download-header.is-scrolled { padding: 10px; }
  .download-header.is-scrolled .nav-shell { border-radius: 8px; }
  .nav-shell { gap: 8px; }
  .nav-links { position: absolute; top: calc(100% + 8px); left: 0; right: 0; padding: 8px; border: 1px solid var(--border); border-radius: 8px; background: var(--panel); box-shadow: var(--shadow); display: none; align-items: stretch; flex-direction: column; }
  .nav-links.is-open { display: flex; }
  .nav-links a { width: 100%; justify-content: flex-start; }
  .menu-button { display: inline-grid; }
  .download-hero { min-height: 100svh; padding-top: 110px; }
  .protocol-strip { grid-template-columns: repeat(2, 1fr); }
  .protocol-strip > div:nth-child(2) { border-right: 0; }
  .protocol-strip > div:nth-child(-n+2) { border-bottom: 1px solid var(--border); }
  .download-options-grid { grid-template-columns: 1fr; }
}

@media (max-width: 560px) {
  .brand-name { font-size: 14px; }
  .primary-link { display: none; }
  .download-hero { min-height: 100svh; padding-inline: 14px; }
  .app-logo-shell { width: 78px; height: 78px; }
  .app-logo { width: 74px; height: 74px; }
  .hero-content h1 { font-size: 43px; }
  .hero-copy { font-size: 15px; line-height: 1.65; }
  .hero-actions { width: 100%; }
  .hero-actions .button { flex: 1 1 150px; padding-inline: 12px; }
  .release-line { flex-wrap: wrap; }
  .modal-layer { padding: 10px; }
  .release-dialog { max-height: calc(100dvh - 20px); }
  .dialog-header, .release-body { padding: 18px; }
  .downloads-dialog-body { padding: 12px; }
  .download-option { padding: 13px; }
  .dialog-footer { padding: 11px 18px; }
}
</style>
