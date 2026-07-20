<template>
  <div class="min-h-[100dvh] bg-canvas text-foreground">
    <header class="border-b border-outline bg-surface/95">
      <div class="mx-auto flex max-w-5xl items-center justify-between gap-4 px-4 py-3.5 sm:px-6">
        <RouterLink to="/home" class="flex min-w-0 items-center gap-3 rounded-control focus:outline-none focus-visible:ring-2 focus-visible:ring-focus/40">
          <template v-if="settings">
          <span class="flex h-10 w-10 flex-shrink-0 items-center justify-center overflow-hidden rounded-panel border border-outline bg-surface-raised shadow-card">
            <img
              :src="siteLogo || '/logo.svg'"
              :alt="t('legal.siteLogoAlt', { siteName })"
              class="h-full w-full object-contain"
            />
          </span>
          <span class="truncate text-base font-semibold text-foreground">{{ siteName }}</span>
          </template>
          <template v-else>
            <span class="h-10 w-10 flex-shrink-0 animate-pulse rounded-panel bg-surface-subtle" aria-hidden="true"></span>
            <span class="h-5 w-28 animate-pulse rounded-control bg-surface-subtle" aria-hidden="true"></span>
          </template>
        </RouterLink>
        <RouterLink to="/login" class="btn btn-primary btn-sm flex-shrink-0">
          {{ t('home.login') }}
        </RouterLink>
      </div>
    </header>

    <main class="mx-auto max-w-4xl px-4 py-8 sm:px-6 lg:py-10" :aria-busy="loading">
      <div
        v-if="loading"
        class="flex min-h-[320px] items-center justify-center text-sm text-foreground-muted"
        role="status"
        aria-live="polite"
      >
        <span class="h-6 w-6 animate-spin rounded-full border-2 border-outline-strong border-t-foreground" aria-hidden="true"></span>
        <span class="ml-3">{{ t('legal.loading') }}</span>
      </div>

      <section
        v-else-if="loadError"
        class="rounded-panel border border-danger/20 bg-danger-subtle p-5 text-danger-foreground sm:p-6"
        role="alert"
        aria-labelledby="legal-load-error-title"
      >
        <div class="flex items-start gap-3">
          <Icon name="exclamationCircle" size="md" class="mt-0.5 flex-shrink-0" aria-hidden="true" />
          <div class="min-w-0">
            <h1 id="legal-load-error-title" class="text-lg font-semibold">{{ t('legal.loadFailed') }}</h1>
            <p class="mt-2 text-sm leading-6">{{ t('legal.retryLater') }}</p>
            <button type="button" class="btn btn-secondary mt-4" @click="loadDocument">
              <Icon name="refresh" size="sm" class="mr-1.5" aria-hidden="true" />
              {{ t('legal.retry') }}
            </button>
          </div>
        </div>
      </section>

      <section
        v-else-if="!currentDocument"
        class="rounded-panel border border-outline bg-surface p-5 shadow-card sm:p-6"
        aria-labelledby="legal-not-found-title"
      >
        <div class="flex items-start gap-3">
          <span class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-panel bg-surface-subtle text-foreground-muted" aria-hidden="true">
            <Icon name="document" size="sm" />
          </span>
          <div>
            <h1 id="legal-not-found-title" class="text-lg font-semibold text-foreground">{{ t('legal.notFound') }}</h1>
            <p class="mt-2 text-sm leading-6 text-foreground-muted">{{ t('legal.notFoundDescription') }}</p>
          </div>
        </div>
      </section>

      <article v-else>
        <div class="mb-8 border-b border-outline pb-6">
          <div class="flex items-start gap-4">
            <span class="flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-panel bg-info-subtle text-info-foreground" aria-hidden="true">
              <Icon :name="documentIcon" size="md" />
            </span>
            <div class="min-w-0">
              <p class="text-sm font-medium text-info-foreground">{{ documentTypeLabel }}</p>
              <h1 class="mt-2 break-words text-2xl font-semibold text-foreground sm:text-3xl">
                {{ currentDocument.title }}
              </h1>
              <p v-if="updatedAt" class="mt-3 text-sm text-foreground-subtle">
                {{ t('legal.updatedAt', { date: updatedAt }) }}
              </p>
            </div>
          </div>
        </div>

        <div v-if="hasContent" class="legal-document-content" v-html="renderedHtml"></div>
        <div
          v-else
          class="rounded-panel border border-dashed border-outline-strong bg-surface px-6 py-14 text-center text-sm text-foreground-subtle"
        >
          {{ t('legal.empty') }}
        </div>
      </article>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { getLocale } from '@/i18n'
import { sanitizeUrl } from '@/utils/url'
import { useAppStore } from '@/stores/app'
import type { LoginAgreementDocument } from '@/types'
import zhAdminCompliance from '../../../../docs/legal/admin-compliance.zh.md?raw'
import enAdminCompliance from '../../../../docs/legal/admin-compliance.en.md?raw'

type LegalDocumentIcon = 'document' | 'shield' | 'globe' | 'cog'

const route = useRoute()
const { t } = useI18n()
const appStore = useAppStore()
const settings = computed(() => appStore.cachedPublicSettings)
const loading = ref(!settings.value)
const loadError = ref(false)

marked.setOptions({
  breaks: true,
  gfm: true,
})

const documentId = computed(() => String(route.params.documentId || ''))
const isAdminComplianceDocument = computed(() => documentId.value === 'admin-compliance')
const documents = computed(() => settings.value?.login_agreement_documents ?? [])
const siteName = computed(() => settings.value?.site_name || 'Sub2API')
const siteLogo = computed(() => sanitizeUrl(settings.value?.site_logo || '', {
  allowRelative: true,
  allowDataUrl: true,
}))
const updatedAt = computed(() =>
  isAdminComplianceDocument.value ? '' : settings.value?.login_agreement_updated_at || ''
)
const documentTypeLabel = computed(() =>
  isAdminComplianceDocument.value ? t('legal.adminCompliance') : t('legal.loginAgreement')
)

const currentDocument = computed<LoginAgreementDocument | null>(() => {
  if (isAdminComplianceDocument.value) {
    return {
      id: 'admin-compliance',
      title: t('adminCompliance.title'),
      content_md: getLocale() === 'zh' ? zhAdminCompliance : enAdminCompliance
    }
  }
  const id = documentId.value
  if (!id) return null
  return documents.value.find((doc) => doc.id === id) ?? null
})

const hasContent = computed(() => Boolean(currentDocument.value?.content_md?.trim()))

const renderedHtml = computed(() => {
  const content = currentDocument.value?.content_md?.trim() || ''
  if (!content) return ''
  return DOMPurify.sanitize(marked.parse(content) as string)
})

const documentIcon = computed<LegalDocumentIcon>(() => {
  const title = currentDocument.value?.title || ''
  if (title.includes('政策') || title.includes('隐私')) return 'shield'
  if (title.includes('国家') || title.includes('地区')) return 'globe'
  if (title.includes('特定')) return 'cog'
  return 'document'
})

async function loadDocument(): Promise<void> {
  loading.value = true
  loadError.value = false
  const loadedSettings = await appStore.fetchPublicSettings()
  if (!loadedSettings) {
    loadError.value = true
  }
  loading.value = false
}

onMounted(() => {
  void loadDocument()
})
</script>

<style scoped>
.legal-document-content {
  overflow-wrap: anywhere;
  color: inherit;
  line-height: 1.75;
}

.legal-document-content :deep(h1) {
  @apply mb-4 mt-8 border-b border-outline pb-3 text-3xl font-semibold;
}

.legal-document-content :deep(h2) {
  @apply mb-3 mt-7 text-2xl font-semibold;
}

.legal-document-content :deep(h3) {
  @apply mb-2 mt-6 text-xl font-semibold;
}

.legal-document-content :deep(h4) {
  @apply mb-2 mt-5 text-lg font-semibold;
}

.legal-document-content :deep(p) {
  @apply mb-4 text-foreground-muted;
}

.legal-document-content :deep(a) {
  @apply text-info-foreground underline underline-offset-4 hover:no-underline;
}

.legal-document-content :deep(ul) {
  @apply mb-4 list-disc pl-6;
}

.legal-document-content :deep(ol) {
  @apply mb-4 list-decimal pl-6;
}

.legal-document-content :deep(li) {
  @apply mb-1 text-foreground-muted;
}

.legal-document-content :deep(blockquote) {
  @apply my-5 border-l-4 border-outline-strong pl-4 text-foreground-muted;
}

.legal-document-content :deep(code) {
  @apply rounded-control bg-surface-subtle px-1.5 py-0.5 font-mono text-sm;
}

.legal-document-content :deep(pre) {
  @apply my-5 overflow-x-auto rounded-panel bg-foreground p-4 text-surface;
}

.legal-document-content :deep(pre code) {
  @apply bg-transparent p-0 text-inherit;
}

.legal-document-content :deep(table) {
  @apply my-5 block w-full overflow-x-auto border-collapse;
}

.legal-document-content :deep(th) {
  @apply border border-outline bg-surface-subtle px-3 py-2 text-left font-semibold;
}

.legal-document-content :deep(td) {
  @apply border border-outline px-3 py-2;
}

.legal-document-content :deep(img) {
  @apply my-5 h-auto max-w-full rounded-panel;
}

.legal-document-content :deep(hr) {
  @apply my-7 border-outline;
}
</style>
