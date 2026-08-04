<template>
  <div v-if="homeContent" class="min-h-[100dvh] bg-canvas text-foreground">
    <div v-if="isHomeContentUrl" class="custom-home-frame-shell">
      <header class="custom-home-toolbar">
        <span class="min-w-0 truncate text-sm font-semibold text-foreground">
          {{ siteName }}
        </span>
        <span class="hidden text-xs text-foreground-subtle sm:inline">
          {{ t('home.customOverride') }}
        </span>
        <a
          :href="homeContentUrl"
          target="_blank"
          rel="noopener noreferrer"
          class="btn btn-secondary btn-sm ml-auto flex-shrink-0"
        >
          <Icon name="externalLink" size="sm" class="mr-1.5" aria-hidden="true" />
          {{ t('home.openCustomPage') }}
        </a>
      </header>

      <main class="relative min-h-0 flex-1 bg-surface-subtle">
        <div
          v-if="iframeLoading && !iframeFailed"
          class="absolute inset-0 z-10 flex items-center justify-center bg-surface"
          role="status"
          aria-live="polite"
        >
          <span class="h-6 w-6 animate-spin rounded-full border-2 border-outline-strong border-t-foreground" aria-hidden="true"></span>
          <span class="ml-3 text-sm text-foreground-muted">{{ t('home.loadingCustomPage') }}</span>
        </div>

        <div
          v-if="iframeFailed"
          class="absolute inset-0 z-20 flex items-center justify-center bg-canvas p-4"
          role="alert"
        >
          <div class="w-full max-w-md rounded-panel border border-outline bg-surface p-5 text-center shadow-card sm:p-6">
            <span
              class="mx-auto flex h-10 w-10 items-center justify-center rounded-panel bg-danger-subtle text-danger-foreground"
              aria-hidden="true"
            >
              <Icon name="exclamationCircle" size="md" />
            </span>
            <h1 class="mt-4 text-lg font-semibold text-foreground">{{ t('home.customPageLoadFailed') }}</h1>
            <p class="mt-2 text-sm leading-6 text-foreground-muted">{{ t('home.customPageLoadFailedHint') }}</p>
            <div class="mt-5 flex flex-col gap-2 sm:flex-row sm:justify-center">
              <button type="button" class="btn btn-secondary" @click="reloadIframe">
                <Icon name="refresh" size="sm" class="mr-1.5" aria-hidden="true" />
                {{ t('common.refresh') }}
              </button>
              <a :href="homeContentUrl" target="_blank" rel="noopener noreferrer" class="btn btn-primary">
                <Icon name="externalLink" size="sm" class="mr-1.5" aria-hidden="true" />
                {{ t('home.openCustomPage') }}
              </a>
            </div>
          </div>
        </div>

        <iframe
          v-show="!iframeFailed"
          :key="iframeKey"
          :src="homeContentUrl"
          :title="t('home.customPageTitle', { siteName })"
          class="h-full w-full border-0 bg-surface"
          :sandbox="HOME_IFRAME_SANDBOX"
          allow="clipboard-read; clipboard-write; fullscreen"
          referrerpolicy="strict-origin-when-cross-origin"
          allowfullscreen
          @load="handleIframeLoad"
          @error="handleIframeError"
        ></iframe>
      </main>
    </div>

    <main v-else class="custom-home-html" v-html="sanitizedHomeContent"></main>
  </div>

  <div
    v-else-if="compactHomeEnabled"
    data-testid="compact-home"
    class="flex min-h-[100dvh] flex-col bg-canvas text-foreground"
  >
    <header class="border-b border-outline px-4 py-4 sm:px-6">
      <nav class="mx-auto flex max-w-5xl flex-wrap items-center justify-between gap-3">
        <div class="flex min-w-0 flex-1 items-center gap-3">
          <img :src="siteLogo || '/logo.svg'" alt="" class="h-9 w-9 shrink-0 rounded-control object-contain" />
          <span class="min-w-0 truncate text-base font-semibold">{{ siteName }}</span>
        </div>
        <div class="flex shrink-0 items-center gap-2">
          <LocaleSwitcher />
          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="btn btn-ghost btn-icon h-10 w-10"
            :title="t('home.viewDocs')"
            :aria-label="t('home.viewDocs')"
          >
            <Icon name="book" size="md" />
          </a>
          <button
            type="button"
            class="btn btn-ghost btn-icon h-10 w-10"
            :title="isDark ? t('home.switchToLight') : t('home.switchToDark')"
            :aria-label="isDark ? t('home.switchToLight') : t('home.switchToDark')"
            @click="toggleTheme"
          >
            <Icon :name="isDark ? 'sun' : 'moon'" size="md" />
          </button>
          <router-link :to="isAuthenticated ? dashboardPath : '/login'" class="btn btn-primary min-h-10">
            {{ isAuthenticated ? t('home.dashboard') : t('home.login') }}
          </router-link>
        </div>
      </nav>
    </header>

    <main class="flex min-w-0 flex-1 items-center justify-center px-4 py-16 sm:px-6">
      <div class="min-w-0 max-w-2xl text-center">
        <img :src="siteLogo || '/logo.svg'" alt="" class="mx-auto mb-6 h-20 w-20 rounded-panel object-contain" />
        <h1 class="break-words text-3xl font-bold md:text-4xl">{{ siteName }}</h1>
        <p class="mt-4 whitespace-pre-wrap break-words text-base text-foreground-muted">{{ siteSubtitle }}</p>
        <router-link :to="isAuthenticated ? dashboardPath : '/login'" class="btn btn-primary mt-8 min-h-10 px-5">
          {{ isAuthenticated ? t('home.goToDashboard') : t('home.login') }}
        </router-link>
      </div>
    </main>

    <footer class="border-t border-outline px-4 py-5 text-center text-sm text-foreground-subtle">
      &copy; {{ currentYear }} {{ siteName }}
    </footer>
  </div>

  <HomeExperiment
    v-else
    :site-name="siteName"
    :site-subtitle="siteSubtitle"
    :is-authenticated="isAuthenticated"
    :dashboard-path="dashboardPath"
    :software-center-enabled="softwareCenterEnabled"
  />
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import DOMPurify from 'dompurify'
import { useAuthStore, useAppStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import Icon from '@/components/icons/Icon.vue'
import { useTheme } from '@/composables/useTheme'
import { sanitizeUrl } from '@/utils/url'
import HomeExperiment from './HomeExperiment.vue'

const { t } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()
const { isDark, toggleTheme } = useTheme()

const iframeLoading = ref(false)
const iframeFailed = ref(false)
const iframeKey = ref(0)
const HOME_IFRAME_SANDBOX = 'allow-downloads allow-forms allow-modals allow-popups allow-popups-to-escape-sandbox allow-presentation allow-same-origin allow-scripts'

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'USA-零')
const siteLogo = computed(() => sanitizeUrl(
  appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '',
  { allowRelative: true, allowDataUrl: true },
))
const siteSubtitle = computed(() =>
  appStore.cachedPublicSettings?.site_subtitle
  || '统一 OpenAI、Claude、Gemini 等不同接口，把多模型调用规范成一个稳定、可计量、可治理的标准 API。'
)
const docUrl = computed(() => sanitizeUrl(appStore.cachedPublicSettings?.doc_url || appStore.docUrl || ''))
const homeContent = computed(() => (appStore.cachedPublicSettings?.home_content || '').trim())
const compactHomeEnabled = computed(() => appStore.cachedPublicSettings?.compact_home_enabled === true)
const homeContentUrl = computed(() => sanitizeUrl(homeContent.value))
const isHomeContentUrl = computed(() => Boolean(homeContentUrl.value))
const sanitizedHomeContent = computed(() => sanitizeHomeHtml(homeContent.value))

const isAuthenticated = computed(() => authStore.isAuthenticated)
const dashboardPath = computed(() => authStore.isAdmin ? '/admin/dashboard' : '/dashboard')
const currentYear = computed(() => new Date().getFullYear())
const softwareCenterEnabled = computed(
  () => appStore.cachedPublicSettings?.software_center_enabled !== false
)

function sanitizeHomeHtml(content: string): string {
  const sanitized = DOMPurify.sanitize(content, {
    ADD_TAGS: ['iframe'],
    ADD_ATTR: ['allow', 'allowfullscreen', 'loading', 'referrerpolicy', 'sandbox', 'target'],
  })
  if (typeof document === 'undefined') return sanitized

  const template = document.createElement('template')
  template.innerHTML = sanitized
  template.content.querySelectorAll('iframe').forEach((frame) => {
    const safeSrc = sanitizeUrl(frame.getAttribute('src') || '')
    if (!safeSrc) {
      frame.remove()
      return
    }
    frame.setAttribute('src', safeSrc)
    frame.setAttribute('sandbox', HOME_IFRAME_SANDBOX)
    frame.setAttribute('referrerpolicy', 'strict-origin-when-cross-origin')
    frame.setAttribute('allow', 'clipboard-read; clipboard-write; fullscreen')
  })
  template.content.querySelectorAll<HTMLAnchorElement>('a[target="_blank"]').forEach((link) => {
    link.rel = 'noopener noreferrer'
  })
  return template.innerHTML
}

watch(homeContentUrl, (url) => {
  iframeFailed.value = false
  iframeLoading.value = Boolean(url)
  iframeKey.value += 1
}, { immediate: true })

function handleIframeLoad(): void {
  iframeLoading.value = false
  iframeFailed.value = false
}

function handleIframeError(): void {
  iframeLoading.value = false
  iframeFailed.value = true
}

function reloadIframe(): void {
  iframeFailed.value = false
  iframeLoading.value = true
  iframeKey.value += 1
}

onMounted(() => {
  authStore.checkAuth()
  if (!appStore.publicSettingsLoaded) {
    void appStore.fetchPublicSettings()
  }
})
</script>

<style scoped>
.custom-home-frame-shell {
  display: flex;
  min-height: 100dvh;
  flex-direction: column;
  overflow: hidden;
}

.custom-home-toolbar {
  display: flex;
  min-height: 52px;
  align-items: center;
  gap: 12px;
  padding: 6px var(--page-gutter);
  border-bottom: 1px solid var(--ui-border);
  background: var(--ui-surface);
}

.custom-home-html {
  min-height: 100dvh;
  min-width: 0;
  overflow-x: auto;
  overflow-wrap: anywhere;
}

.custom-home-html :deep(img),
.custom-home-html :deep(video),
.custom-home-html :deep(canvas) {
  max-width: 100%;
  height: auto;
}
</style>
