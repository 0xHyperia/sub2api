<template>
  <AppLayout>
    <div class="custom-page-layout">
      <div class="card flex-1 min-h-0 overflow-hidden">
        <div
          v-if="loading"
          class="flex h-full items-center justify-center py-12 text-sm text-foreground-muted"
          role="status"
          aria-live="polite"
        >
          <span class="h-6 w-6 animate-spin rounded-full border-2 border-outline-strong border-t-foreground" aria-hidden="true"></span>
          <span class="ml-3">{{ t('common.loading') }}</span>
        </div>

        <div
          v-else-if="!menuItem"
          class="flex h-full items-center justify-center p-10 text-center"
        >
          <div class="max-w-md">
            <div
              class="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-surface-subtle"
            >
              <Icon name="link" size="lg" class="text-foreground-subtle" />
            </div>
            <h3 class="text-lg font-semibold text-foreground">
              {{ t('customPage.notFoundTitle') }}
            </h3>
            <p class="mt-2 text-sm text-foreground-muted">
              {{ t('customPage.notFoundDesc') }}
            </p>
          </div>
        </div>

        <!-- Markdown mode with TOC -->
        <div v-else-if="isMarkdownMode" class="relative flex h-full overflow-hidden">
          <button
            v-show="tocVisible && tocItems.length > 0"
            type="button"
            class="toc-backdrop"
            :aria-label="t('customPage.closeTableOfContents')"
            @click="closeToc()"
          ></button>

          <!-- TOC Sidebar -->
          <aside
            v-show="tocVisible && tocItems.length > 0"
            id="custom-page-toc"
            class="toc-sidebar"
            :aria-label="t('customPage.tableOfContents')"
          >
            <div class="toc-header">
              <span class="toc-title">{{ t('customPage.tableOfContents') }}</span>
              <button
                type="button"
                class="toc-close-btn"
                :aria-label="t('customPage.closeTableOfContents')"
                @click="closeToc()"
              >
                <Icon name="chevronLeft" size="sm" aria-hidden="true" />
              </button>
            </div>
            <nav class="toc-nav" :aria-label="t('customPage.tableOfContents')">
              <a
                v-for="item in tocItems"
                :key="item.id"
                :href="'#' + item.id"
                class="toc-item"
                :class="[
                  `toc-level-${item.level}`,
                  { 'toc-active': activeHeadingId === item.id }
                ]"
                :aria-current="activeHeadingId === item.id ? 'location' : undefined"
                @click.prevent="scrollToHeading(item.id)"
              >
                {{ item.text }}
              </a>
            </nav>
          </aside>

          <!-- TOC Toggle Button (when collapsed) -->
          <button
            ref="tocToggleButton"
            v-show="!tocVisible && tocItems.length > 0"
            type="button"
            class="toc-toggle-btn"
            aria-controls="custom-page-toc"
            :aria-expanded="tocVisible"
            @click="openToc"
          >
            <Icon name="menu" size="sm" aria-hidden="true" />
            <span class="ml-1 text-xs">{{ t('customPage.tableOfContents') }}</span>
          </button>

          <!-- Content -->
          <div
            v-if="markdownError"
            class="flex h-full flex-1 items-center justify-center p-6 text-center"
            role="alert"
          >
            <div class="max-w-md">
              <span
                class="mx-auto flex h-10 w-10 items-center justify-center rounded-panel bg-danger-subtle text-danger-foreground"
                aria-hidden="true"
              >
                <Icon name="exclamationCircle" size="md" />
              </span>
              <h3 class="mt-4 text-lg font-semibold text-foreground">{{ markdownErrorTitle }}</h3>
              <p class="mt-2 text-sm leading-6 text-foreground-muted">{{ markdownErrorDescription }}</p>
              <button type="button" class="btn btn-secondary mt-5" @click="retryMarkdownPage">
                <Icon name="refresh" size="sm" class="mr-1.5" aria-hidden="true" />
                {{ t('customPage.retry') }}
              </button>
            </div>
          </div>
          <div
            v-else
            ref="markdownContainer"
            class="markdown-page-content h-full flex-1 overflow-auto p-5 sm:p-6 md:p-10"
            v-html="renderedHtml"
            @scroll="onContentScroll"
          ></div>
        </div>

        <!-- URL not configured -->
        <div v-else-if="!isValidUrl" class="flex h-full items-center justify-center p-10 text-center">
          <div class="max-w-md">
            <div
              class="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-surface-subtle"
            >
              <Icon name="link" size="lg" class="text-foreground-subtle" />
            </div>
            <h3 class="text-lg font-semibold text-foreground">
              {{ t('customPage.notConfiguredTitle') }}
            </h3>
            <p class="mt-2 text-sm text-foreground-muted">
              {{ t('customPage.notConfiguredDesc') }}
            </p>
          </div>
        </div>

        <!-- Iframe embed mode -->
        <div v-else class="custom-embed-shell">
          <a
            v-if="!menuItem?.hide_open_button"
            :href="embeddedUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="btn btn-secondary btn-sm custom-open-fab"
          >
            <Icon name="externalLink" size="sm" class="mr-1.5" :stroke-width="2" />
            {{ t('customPage.openInNewTab') }}
          </a>
          <div
            v-if="embedLoading && !embedError"
            class="custom-embed-state"
            role="status"
            aria-live="polite"
          >
            <span class="h-6 w-6 animate-spin rounded-full border-2 border-outline-strong border-t-foreground" aria-hidden="true"></span>
            <span class="mt-3 text-sm text-foreground-muted">{{ t('customPage.loadingEmbed') }}</span>
          </div>
          <div v-if="embedError" class="custom-embed-state p-6 text-center" role="alert">
            <span
              class="flex h-10 w-10 items-center justify-center rounded-panel bg-danger-subtle text-danger-foreground"
              aria-hidden="true"
            >
              <Icon name="exclamationCircle" size="md" />
            </span>
            <h3 class="mt-4 text-lg font-semibold text-foreground">{{ t('customPage.embedLoadFailed') }}</h3>
            <p class="mt-2 max-w-md text-sm leading-6 text-foreground-muted">{{ t('customPage.embedLoadFailedDesc') }}</p>
            <button type="button" class="btn btn-secondary mt-5" @click="retryEmbeddedPage">
              <Icon name="refresh" size="sm" class="mr-1.5" aria-hidden="true" />
              {{ t('customPage.retry') }}
            </button>
          </div>
          <iframe
            v-show="!embedError"
            :key="embedFrameKey"
            :src="embeddedUrl"
            :title="t('customPage.embedTitle', { title: menuItem.label })"
            class="custom-embed-frame"
            referrerpolicy="strict-origin-when-cross-origin"
            allowfullscreen
            @load="handleEmbedLoad"
            @error="handleEmbedError"
          ></iframe>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import { useAuthStore } from '@/stores/auth'
import { useAdminSettingsStore } from '@/stores/adminSettings'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { buildApiUrl } from '@/api/client'
import { buildEmbeddedUrl, detectTheme } from '@/utils/embedded-url'
import { marked } from 'marked'
import DOMPurify from 'dompurify'

interface TocItem {
  id: string
  text: string
  level: number
}

const { t, locale } = useI18n()
const route = useRoute()
const appStore = useAppStore()
const authStore = useAuthStore()
const adminSettingsStore = useAdminSettingsStore()

const settingsLoading = ref(false)
const markdownLoading = ref(false)
const pageTheme = ref<'light' | 'dark'>('light')
const renderedHtml = ref('')
const markdownContainer = ref<HTMLElement | null>(null)
const tocItems = ref<TocItem[]>([])
const tocVisible = ref(typeof window !== 'undefined' ? window.innerWidth > 768 : true)
const tocToggleButton = ref<HTMLButtonElement | null>(null)
const activeHeadingId = ref('')
const markdownError = ref<'not-found' | 'load-failed' | ''>('')
const embedLoading = ref(false)
const embedError = ref(false)
const embedFrameKey = ref(0)
let themeObserver: MutationObserver | null = null
let markdownRequestId = 0

const loading = computed(() => settingsLoading.value || markdownLoading.value)
const markdownErrorTitle = computed(() =>
  markdownError.value === 'not-found'
    ? t('customPage.contentNotFoundTitle')
    : t('customPage.contentLoadFailedTitle')
)
const markdownErrorDescription = computed(() =>
  markdownError.value === 'not-found'
    ? t('customPage.contentNotFoundDesc')
    : t('customPage.contentLoadFailedDesc')
)

const menuItemId = computed(() => route.params.id as string)

const menuItem = computed(() => {
  const id = menuItemId.value
  const publicItems = appStore.cachedPublicSettings?.custom_menu_items ?? []
  const found = publicItems.find((item) => item.id === id) ?? null
  if (found) return found
  if (authStore.isAdmin) {
    return adminSettingsStore.customMenuItems.find((item) => item.id === id) ?? null
  }
  return null
})

const markdownSlug = computed(() => {
  const item = menuItem.value
  if (!item) return ''
  if (item.page_slug) return item.page_slug
  if (item.url?.startsWith('md:')) return item.url.slice(3)
  return ''
})

const isMarkdownMode = computed(() => !!markdownSlug.value)

const embeddedUrl = computed(() => {
  if (!menuItem.value || isMarkdownMode.value) return ''
  return buildEmbeddedUrl(
    menuItem.value.url,
    authStore.user?.id,
    authStore.token,
    pageTheme.value,
    locale.value,
  )
})

const isValidUrl = computed(() => {
  if (isMarkdownMode.value) return false
  const url = embeddedUrl.value
  return url.startsWith('http://') || url.startsWith('https://')
})

function generateHeadingId(text: string, index: number): string {
  const base = text
    .toLowerCase()
    .replace(/[^\w一-鿿]+/g, '-')
    .replace(/^-+|-+$/g, '')
  return base ? `${base}-${index}` : `heading-${index}`
}

function isRelativeMarkdownAsset(src: string): boolean {
  const trimmed = src.trim()
  if (!trimmed || /^[a-z][a-z0-9+.-]*:/i.test(trimmed) || trimmed.startsWith('//') || trimmed.startsWith('/')) {
    return false
  }
  const [pathPart] = trimmed.split(/([?#].*)/, 2)
  return pathPart
    .split('/')
    .filter((part) => part && part !== '.')
    .every((part) => part !== '..' && !part.includes('\\'))
}

function buildPageImageUrl(slug: string, src: string): string {
  const trimmed = src.trim()
  const [pathPart, suffix = ''] = trimmed.split(/([?#].*)/, 2)
  const encodedPath = pathPart
    .split('/')
    .filter((part) => part && part !== '.')
    .map((part) => encodeURIComponent(part))
    .join('/')
  return buildApiUrl(`/pages/${encodeURIComponent(slug)}/images/${encodedPath}${suffix}`)
}

async function fetchAndRenderMarkdown(slug: string) {
  const requestId = ++markdownRequestId
  markdownLoading.value = true
  markdownError.value = ''
  renderedHtml.value = ''
  tocItems.value = []
  activeHeadingId.value = ''
  try {
    const resp = await fetch(buildApiUrl(`/pages/${encodeURIComponent(slug)}`), {
      headers: authStore.token ? { Authorization: `Bearer ${authStore.token}` } : {},
    })
    if (requestId !== markdownRequestId) return
    if (!resp.ok) {
      markdownError.value = resp.status === 404 ? 'not-found' : 'load-failed'
      return
    }
    let raw = await resp.text()
    if (requestId !== markdownRequestId) return

    raw = raw.replace(
      /!\[([^\]]*)\]\(([^)]+)\)/g,
      (match, alt, src) => isRelativeMarkdownAsset(src) ? `![${alt}](${buildPageImageUrl(slug, src)})` : match
    )

    const html = marked.parse(raw) as string
    const sanitized = DOMPurify.sanitize(html, {
      ADD_TAGS: ['iframe'],
      ADD_ATTR: ['allowfullscreen', 'frameborder', 'src'],
    })

    // Inject IDs into headings and build TOC
    const toc: TocItem[] = []
    let headingIndex = 0
    const withIds = sanitized.replace(
      /<(h[1-4])[^>]*>(.*?)<\/h[1-4]>/gi,
      (_, tag: string, content: string) => {
        const level = parseInt(tag[1])
        const text = content.replace(/<[^>]+>/g, '').trim()
        const id = generateHeadingId(text, headingIndex++)
        toc.push({ id, text, level })
        return `<${tag} id="${id}" tabindex="-1">${content}</${tag}>`
      }
    )

    if (requestId !== markdownRequestId) return
    renderedHtml.value = withIds
    tocItems.value = toc
  } catch {
    if (requestId === markdownRequestId) {
      markdownError.value = 'load-failed'
    }
  } finally {
    if (requestId === markdownRequestId) {
      markdownLoading.value = false
      if (!markdownError.value) {
        await nextTick()
        injectCopyButtons()
      }
    }
  }
}

function retryMarkdownPage(): void {
  if (markdownSlug.value) {
    void fetchAndRenderMarkdown(markdownSlug.value)
  }
}

function openToc(): void {
  tocVisible.value = true
}

function closeToc(restoreFocus = true): void {
  tocVisible.value = false
  if (restoreFocus) {
    void nextTick(() => tocToggleButton.value?.focus())
  }
}

function scrollToHeading(id: string): void {
  const container = markdownContainer.value
  if (!container) return
  const el = container.querySelector<HTMLElement>(`#${CSS.escape(id)}`)
  if (el) {
    el.scrollIntoView({ behavior: 'smooth', block: 'start' })
    activeHeadingId.value = id
    if (window.innerWidth <= 640) {
      closeToc(false)
      el.focus({ preventScroll: true })
    }
  }
}

let scrollRafId = 0
function onContentScroll() {
  if (scrollRafId) return
  scrollRafId = requestAnimationFrame(() => {
    scrollRafId = 0
    const container = markdownContainer.value
    if (!container || tocItems.value.length === 0) return

    const containerRect = container.getBoundingClientRect()
    let current = ''

    for (const item of tocItems.value) {
      const el = container.querySelector(`#${CSS.escape(item.id)}`) as HTMLElement | null
      if (el) {
        const elRect = el.getBoundingClientRect()
        if (elRect.top - containerRect.top <= 100) {
          current = item.id
        }
      }
    }
    activeHeadingId.value = current
  })
}

function injectCopyButtons() {
  const container = markdownContainer.value
  if (!container) return

  container.querySelectorAll('pre').forEach((pre) => {
    if (pre.querySelector('.copy-btn')) return
    const btn = document.createElement('button')
    btn.type = 'button'
    btn.className = 'copy-btn'
    btn.textContent = t('customPage.copyCode')
    btn.setAttribute('aria-label', t('customPage.copyCode'))
    btn.addEventListener('click', async () => {
      const code = pre.querySelector('code')?.textContent ?? pre.textContent ?? ''
      try {
        await navigator.clipboard.writeText(code)
        btn.textContent = t('customPage.copiedCode')
        setTimeout(() => { btn.textContent = t('customPage.copyCode') }, 2000)
      } catch {
        btn.textContent = t('customPage.copyCodeFailed')
        setTimeout(() => { btn.textContent = t('customPage.copyCode') }, 2000)
      }
    })
    pre.style.position = 'relative'
    pre.appendChild(btn)
  })
}

watch(markdownSlug, (slug) => {
  if (slug) {
    void fetchAndRenderMarkdown(slug)
  } else {
    markdownRequestId += 1
    markdownLoading.value = false
    markdownError.value = ''
    renderedHtml.value = ''
    tocItems.value = []
  }
}, { immediate: true })

watch(embeddedUrl, (url) => {
  embedError.value = false
  embedLoading.value = Boolean(url)
  embedFrameKey.value += 1
}, { immediate: true })

function handleEmbedLoad(): void {
  embedLoading.value = false
  embedError.value = false
}

function handleEmbedError(): void {
  embedLoading.value = false
  embedError.value = true
}

function retryEmbeddedPage(): void {
  embedError.value = false
  embedLoading.value = true
  embedFrameKey.value += 1
}

function handleDocumentKeydown(event: KeyboardEvent): void {
  if (event.key === 'Escape' && tocVisible.value && window.innerWidth <= 640) {
    event.preventDefault()
    closeToc()
  }
}

onMounted(async () => {
  pageTheme.value = detectTheme()

  if (typeof document !== 'undefined') {
    themeObserver = new MutationObserver(() => {
      pageTheme.value = detectTheme()
    })
    themeObserver.observe(document.documentElement, {
      attributes: true,
      attributeFilter: ['class'],
    })
    document.addEventListener('keydown', handleDocumentKeydown)
  }

  if (appStore.publicSettingsLoaded) return
  settingsLoading.value = true
  try {
    await appStore.fetchPublicSettings()
  } finally {
    settingsLoading.value = false
  }
})

onUnmounted(() => {
  markdownRequestId += 1
  document.removeEventListener('keydown', handleDocumentKeydown)
  if (scrollRafId) {
    cancelAnimationFrame(scrollRafId)
    scrollRafId = 0
  }
  if (themeObserver) {
    themeObserver.disconnect()
    themeObserver = null
  }
})
</script>

<style scoped>
.custom-page-layout {
  display: flex;
  height: calc(100dvh - var(--app-header-height, 64px) - (var(--page-gutter, 32px) * 2));
  min-height: 360px;
  flex-direction: column;
}

.toc-sidebar {
  display: flex;
  height: 100%;
  flex-direction: column;
  width: min(240px, 30%);
  min-width: 160px;
  max-width: 280px;
  overflow: hidden;
  border-right: 1px solid var(--ui-border);
  background: var(--ui-surface-subtle);
}

.toc-backdrop {
  display: none;
}

@media (max-width: 640px) {
  .custom-page-layout {
    height: calc(100dvh - var(--app-header-height, 64px) - (var(--page-gutter, 16px) * 2));
  }

  .toc-backdrop {
    position: absolute;
    inset: 0;
    z-index: 15;
    display: block;
    border: 0;
    background: rgb(15 23 42 / 38%);
  }

  .toc-sidebar {
    position: absolute;
    inset: 0 auto 0 0;
    z-index: 20;
    width: min(82%, 280px);
    min-width: 0;
    max-width: none;
    height: 100%;
    box-shadow: var(--ui-shadow-floating);
  }
}

.toc-header {
  display: flex;
  min-height: 48px;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px 8px 16px;
  border-bottom: 1px solid var(--ui-border);
}

.toc-title {
  color: var(--ui-text);
  font-size: 14px;
  font-weight: 600;
}

.toc-close-btn {
  display: inline-flex;
  width: 40px;
  height: 40px;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-md);
  color: var(--ui-text-subtle);
  transition: color var(--duration-fast), background-color var(--duration-fast);
}

.toc-close-btn:hover {
  background: var(--ui-surface);
  color: var(--ui-text);
}

.toc-close-btn:focus-visible,
.toc-toggle-btn:focus-visible {
  outline: 2px solid var(--ui-focus);
  outline-offset: 2px;
}

.toc-nav {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
}

.toc-item {
  display: block;
  overflow: hidden;
  padding-top: 7px;
  padding-bottom: 7px;
  border-radius: var(--radius-sm);
  color: var(--ui-text-muted);
  font-size: 14px;
  text-overflow: ellipsis;
  white-space: nowrap;
  transition: color var(--duration-fast), background-color var(--duration-fast);
}

.toc-item:hover {
  background: var(--ui-surface);
  color: var(--ui-text);
}

.toc-item:focus-visible {
  outline: 2px solid var(--ui-focus);
  outline-offset: -2px;
}

.toc-item.toc-active {
  background: var(--ui-surface);
  color: var(--ui-text);
  box-shadow: inset 2px 0 0 var(--ui-focus);
  font-weight: 600;
}

.toc-level-1 { padding-left: 8px; }
.toc-level-2 { padding-left: 20px; }
.toc-level-3 { padding-left: 32px; }
.toc-level-4 { padding-left: 44px; }

.toc-toggle-btn {
  position: absolute;
  top: 8px;
  left: 8px;
  z-index: 10;
  display: flex;
  min-height: 36px;
  align-items: center;
  padding: 6px 10px;
  border: 1px solid var(--ui-border);
  border-radius: var(--radius-md);
  background: var(--ui-surface);
  color: var(--ui-text-muted);
  box-shadow: var(--ui-shadow-xs);
  cursor: pointer;
  transition: color var(--duration-fast), background-color var(--duration-fast);
}

.toc-toggle-btn:hover {
  background: var(--ui-surface-subtle);
  color: var(--ui-text);
}

.custom-embed-shell {
  position: relative;
  width: 100%;
  height: 100%;
  overflow: hidden;
  background: var(--ui-surface-subtle);
}

.custom-open-fab {
  position: absolute;
  top: 12px;
  right: 12px;
  z-index: 10;
  box-shadow: var(--ui-shadow-xs);
}

.custom-embed-state {
  position: absolute;
  inset: 0;
  z-index: 5;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-direction: column;
  background: var(--ui-surface);
}

.custom-embed-frame {
  display: block;
  margin: 0;
  width: 100%;
  height: 100%;
  border: 0;
  border-radius: 0;
  box-shadow: none;
  background: var(--ui-surface);
}
</style>

<style>
.markdown-page-content {
  line-height: 1.7;
  color: rgb(var(--color-foreground));
}
.markdown-page-content h1 { @apply text-3xl font-bold mt-8 mb-4 pb-2 border-b border-outline; }
.markdown-page-content h2 { @apply text-2xl font-bold mt-6 mb-3; }
.markdown-page-content h3 { @apply text-xl font-semibold mt-5 mb-2; }
.markdown-page-content h4 { @apply text-lg font-semibold mt-4 mb-2; }
.markdown-page-content p { @apply mb-4; }
.markdown-page-content ul { @apply list-disc pl-6 mb-4; }
.markdown-page-content ol { @apply list-decimal pl-6 mb-4; }
.markdown-page-content li { @apply mb-1; }
.markdown-page-content a { @apply text-info-foreground hover:underline underline-offset-4; }
.markdown-page-content blockquote { @apply border-l-4 border-outline-strong pl-4 italic text-foreground-muted my-4; }
.markdown-page-content img { @apply max-w-full h-auto rounded-panel my-4; }
.markdown-page-content table { @apply my-4 block w-full overflow-x-auto border-collapse; }
.markdown-page-content th { @apply border border-outline px-3 py-2 bg-surface-subtle font-semibold text-left; }
.markdown-page-content td { @apply border border-outline px-3 py-2; }
.markdown-page-content code { @apply bg-surface-subtle px-1.5 py-0.5 rounded text-sm font-mono; }
.markdown-page-content pre { @apply relative my-4 overflow-x-auto rounded-panel border border-outline bg-surface-subtle p-4 text-foreground; }
.markdown-page-content pre code { @apply bg-transparent p-0 text-inherit; }
.markdown-page-content hr { @apply my-6 border-outline; }

.markdown-page-content :is(h1, h2, h3, h4):focus-visible {
  outline: 2px solid rgb(var(--color-focus));
  outline-offset: 4px;
}

.copy-btn {
  position: absolute;
  top: 8px;
  right: 8px;
  padding: 4px 10px;
  font-size: 12px;
  border-radius: 4px;
  border: 1px solid rgb(var(--color-border-strong));
  background: rgb(var(--color-surface-raised));
  color: rgb(var(--color-foreground-muted));
  cursor: pointer;
  opacity: 1;
  transition: opacity 0.2s, background 0.2s;
  font-family: inherit;
}
.copy-btn:hover {
  background: rgb(var(--color-surface));
  color: rgb(var(--color-foreground));
}
.copy-btn:focus-visible {
  outline: 2px solid rgb(var(--color-focus));
  outline-offset: 2px;
}

@media (hover: hover) and (pointer: fine) {
  .copy-btn {
    opacity: 0;
  }

  pre:hover .copy-btn,
  pre:focus-within .copy-btn,
  .copy-btn:focus-visible {
    opacity: 1;
  }
}

@media (prefers-reduced-motion: reduce) {
  .copy-btn {
    transition-duration: 1ms;
  }
}
</style>
