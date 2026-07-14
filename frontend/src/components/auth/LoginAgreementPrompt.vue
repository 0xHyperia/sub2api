<template>
  <div
    v-if="mode === 'checkbox' && documents.length > 0"
    class="px-0.5"
  >
    <div class="flex items-start gap-2">
      <input
        id="login-agreement-consent"
        type="checkbox"
        :checked="accepted"
        class="mt-[2px] h-5 w-5 flex-shrink-0 rounded border-outline-strong text-brand focus:ring-focus bg-canvas"
        @change="handleCheckboxChange"
      />
      <div class="min-w-0 flex-1">
        <p class="text-[13px] leading-5 text-foreground-muted">
          <label
            for="login-agreement-consent"
            class="cursor-pointer text-foreground-muted"
          >
            {{ t('legal.loginAgreementPrompt.checkboxPrefix') }}
          </label>
          <template v-for="(doc, index) in documents" :key="doc.id || doc.title">
            <RouterLink
              :to="documentRoute(doc)"
              target="_blank"
              rel="noopener noreferrer"
              class="font-medium text-brand underline-offset-4 transition hover:text-brand hover:underline"
            >
              {{ doc.title }}
            </RouterLink>
            <span v-if="index < documents.length - 1">{{ t('legal.loginAgreementPrompt.documentSeparator') }}</span>
          </template>
        </p>
      </div>
    </div>
  </div>

  <div
    v-else-if="!accepted && documents.length > 0"
    class="rounded-panel border border-brand/20 bg-brand-subtle p-3 text-sm text-brand"
  >
    <div class="flex items-start gap-3">
      <Icon name="shield" size="sm" class="mt-0.5 flex-shrink-0 text-brand" />
      <div class="min-w-0 flex-1">
        <p class="font-medium">{{ t('legal.loginAgreementPrompt.noticeTitle') }}</p>
        <p class="mt-1 text-brand/80">
          {{ t('legal.loginAgreementPrompt.noticeDescription') }}
        </p>
      </div>
      <button
        type="button"
        class="min-h-10 flex-shrink-0 rounded-control bg-brand px-3 py-2 text-xs font-medium text-brand-foreground transition hover:bg-brand-hover"
        @click="emit('open')"
      >
        {{ t('legal.loginAgreementPrompt.viewTerms') }}
      </button>
    </div>
  </div>

  <Teleport to="body">
    <Transition name="agreement-fade">
      <div
        v-if="dialogVisible"
        ref="overlayRef"
        class="fixed inset-0 z-[140] flex items-center justify-center overflow-y-auto bg-foreground/60 p-4 backdrop-blur-sm"
        role="dialog"
        aria-modal="true"
        aria-labelledby="login-agreement-title"
        aria-describedby="login-agreement-description"
        @keydown="handleDialogKeydown"
      >
        <div ref="dialogRef" class="w-full max-w-[600px] overflow-hidden rounded-panel shadow-modal ring-1 ring-black/10 bg-canvas ring-outline" tabindex="-1">
          <div class="border-b border-outline px-6 py-6 bg-canvas">
            <div class="flex items-start gap-4">
              <span class="flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-panel text-brand ring-1 ring-focus/20 bg-brand-subtle">
                <Icon name="shield" size="md" />
              </span>
              <div class="min-w-0 flex-1">
                <div class="flex flex-wrap items-center gap-2">
                  <h2 id="login-agreement-title" class="text-xl font-bold tracking-normal text-foreground">
                    {{ t('legal.loginAgreementPrompt.dialogTitle') }}
                  </h2>
                  <span
                    v-if="updatedAt"
                    class="rounded-full px-2.5 py-1 text-xs font-medium text-foreground-muted bg-surface"
                  >
                    {{ updatedAt }}
                  </span>
                </div>
                <p id="login-agreement-description" class="mt-2 text-sm leading-6 text-foreground-muted">
                  {{
                    t('legal.loginAgreementPrompt.dialogDescription', {
                      date: updatedAt || t('legal.loginAgreementPrompt.recently'),
                    })
                  }}
                </p>
              </div>
            </div>
          </div>

          <div class="max-h-[58vh] overflow-y-auto px-6 py-5">
            <div class="mb-3 flex items-center justify-between gap-3">
              <p class="text-sm font-semibold text-foreground">{{ t('legal.loginAgreementPrompt.relatedDocuments') }}</p>
            </div>
            <div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
              <RouterLink
                v-for="(doc, index) in documents"
                :key="doc.id || doc.title"
                :to="documentRoute(doc)"
                target="_blank"
                rel="noopener noreferrer"
                class="group flex min-h-[72px] w-full items-center gap-3 rounded-panel border border-outline px-4 py-3 text-left transition hover:shadow-sm bg-surface hover:border-brand/30 hover:bg-surface-subtle"
              >
                <span class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-panel text-foreground-muted ring-1 ring-outline transition group-hover:ring-focus/20 bg-canvas group-hover:bg-brand-subtle group-hover:text-brand">
                  <Icon :name="documentIcon(index, doc.title)" size="sm" />
                </span>
                <span class="min-w-0 flex-1">
                  <span class="block truncate text-sm font-semibold text-foreground">{{ doc.title }}</span>
                </span>
                <span class="flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-full text-foreground-subtle transition group-hover:bg-brand-subtle group-hover:text-brand">
                  <Icon name="externalLink" size="sm" />
                </span>
              </RouterLink>
            </div>
          </div>

          <div class="border-t border-outline px-6 py-4 bg-canvas/60">
            <div class="grid grid-cols-2 gap-3">
              <button
                type="button"
                class="min-h-11 rounded-panel border border-outline px-4 py-3 text-sm font-semibold text-foreground-muted transition hover:bg-surface"
                @click="rejectDialog"
              >
                {{ t('legal.loginAgreementPrompt.reject') }}
              </button>
              <button
                type="button"
                class="min-h-11 rounded-panel bg-brand px-4 py-3 text-sm font-semibold text-brand-foreground shadow-sm shadow-brand/20 transition hover:bg-brand-hover"
                @click="emit('accept')"
              >
                {{ t('legal.loginAgreementPrompt.accept') }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { LoginAgreementDocument } from '@/types'

const { t } = useI18n()

const props = withDefaults(defineProps<{
  accepted: boolean
  documents: LoginAgreementDocument[]
  mode: 'modal' | 'checkbox' | string
  updatedAt?: string
  visible: boolean
}>(), {
  updatedAt: ''
})

const emit = defineEmits<{
  accept: []
  reject: []
  open: []
}>()

const dialogVisible = computed(() => props.visible && documents.value.length > 0)
const documents = computed(() => props.documents.filter((doc) => doc.title.trim()))
const updatedAt = computed(() => props.updatedAt || '')
const accepted = computed(() => props.accepted)
const mode = computed(() => props.mode === 'checkbox' ? 'checkbox' : 'modal')
const overlayRef = ref<HTMLElement | null>(null)
const dialogRef = ref<HTMLElement | null>(null)
let previouslyFocusedElement: HTMLElement | null = null
let previousBodyOverflow = ''

function documentRoute(doc: LoginAgreementDocument) {
  return {
    name: 'LegalDocument',
    params: {
      documentId: doc.id || doc.title,
    },
  }
}

function handleCheckboxChange(event: Event): void {
  const checked = (event.target as HTMLInputElement).checked
  if (checked) {
    emit('accept')
  } else {
    emit('reject')
  }
}

function rejectDialog(): void {
  emit('reject')
}

function getFocusableElements(): HTMLElement[] {
  if (!overlayRef.value) return []
  return Array.from(
    overlayRef.value.querySelectorAll<HTMLElement>(
      'a[href], button:not([disabled]), input:not([disabled]), [tabindex]:not([tabindex="-1"])'
    )
  )
}

function handleDialogKeydown(event: KeyboardEvent): void {
  if (event.key === 'Escape') {
    event.preventDefault()
    rejectDialog()
    return
  }
  if (event.key !== 'Tab') return

  const focusableElements = getFocusableElements()
  if (!focusableElements.length) {
    event.preventDefault()
    dialogRef.value?.focus()
    return
  }

  const firstElement = focusableElements[0]
  const lastElement = focusableElements[focusableElements.length - 1]
  if (event.shiftKey && (document.activeElement === firstElement || document.activeElement === dialogRef.value)) {
    event.preventDefault()
    lastElement.focus()
  } else if (!event.shiftKey && document.activeElement === lastElement) {
    event.preventDefault()
    firstElement.focus()
  }
}

function restoreDialogState(): void {
  document.body.style.overflow = previousBodyOverflow
  previouslyFocusedElement?.focus()
  previouslyFocusedElement = null
}

watch(dialogVisible, async (visible) => {
  if (visible) {
    previouslyFocusedElement = document.activeElement as HTMLElement | null
    previousBodyOverflow = document.body.style.overflow
    document.body.style.overflow = 'hidden'
    await nextTick()
    dialogRef.value?.focus()
  } else if (previouslyFocusedElement) {
    restoreDialogState()
  }
}, { flush: 'post' })

onBeforeUnmount(() => {
  if (previouslyFocusedElement) restoreDialogState()
})

function documentIcon(index: number, title: string): 'document' | 'shield' | 'globe' | 'cog' {
  const normalizedTitle = title.toLowerCase()
  if (
    normalizedTitle.includes('policy') ||
    normalizedTitle.includes('privacy') ||
    title.includes('政策') ||
    title.includes('隐私')
  ) {
    return 'shield'
  }
  if (
    normalizedTitle.includes('country') ||
    normalizedTitle.includes('region') ||
    title.includes('国家') ||
    title.includes('地区')
  ) {
    return 'globe'
  }
  if (index === 3) {
    return 'cog'
  }
  return 'document'
}
</script>

<style scoped>
.agreement-fade-enter-active,
.agreement-fade-leave-active {
  transition: opacity 0.18s ease;
}

.agreement-fade-enter-from,
.agreement-fade-leave-to {
  opacity: 0;
}

.agreement-fade-enter-active > div,
.agreement-fade-leave-active > div {
  transition: transform 0.18s ease, opacity 0.18s ease;
}

.agreement-fade-enter-from > div,
.agreement-fade-leave-to > div {
  opacity: 0;
  transform: translateY(8px) scale(0.98);
}
</style>
