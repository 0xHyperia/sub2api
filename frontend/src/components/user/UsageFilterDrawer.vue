<template>
  <Teleport to="body">
    <Transition name="usage-filter-sheet">
      <div
        v-if="open"
        class="fixed inset-0 z-50 flex items-end bg-black/25 backdrop-blur-[1px] sm:hidden"
        role="dialog"
        aria-modal="true"
        :aria-labelledby="titleId"
        @click.self="emit('close')"
      >
        <section
          ref="sheetPanel"
          tabindex="-1"
          class="flex max-h-[88dvh] w-full flex-col overflow-hidden rounded-t-panel border border-b-0 border-outline bg-surface-raised shadow-floating"
        >
          <header class="grid min-h-14 grid-cols-[40px_minmax(0,1fr)_40px] items-center border-b border-outline px-4">
            <span class="flex h-8 w-8 items-center justify-center rounded-control bg-surface-subtle text-foreground-muted" aria-hidden="true">
              <Icon name="filter" size="sm" />
            </span>
            <div class="min-w-0">
              <h2 :id="titleId" class="text-sm font-semibold text-foreground">{{ t('common.filter') }}</h2>
              <p class="mt-0.5 text-[10px] text-foreground-subtle">{{ t('common.selectedCount', { count: activeCount }) }}</p>
            </div>
            <button ref="closeButton" type="button" class="btn btn-ghost btn-icon h-9 w-9" :aria-label="t('common.close')" @click="emit('close')">
              <Icon name="x" size="sm" />
            </button>
          </header>

          <div class="min-h-0 flex-1 overflow-y-auto overscroll-contain px-4 py-4">
            <div v-if="mode === 'usage'" class="grid gap-4">
              <div>
                <label class="input-label">{{ t('usage.apiKeyFilter') }}</label>
                <Select v-model="usageDraft.api_key_id" :options="apiKeyOptions" />
              </div>
              <div>
                <label class="input-label">{{ t('usage.model') }}</label>
                <Select v-model="usageDraft.model" :options="modelOptions" searchable />
              </div>
              <div>
                <label class="input-label">{{ t('admin.usage.group') }}</label>
                <Select v-model="usageDraft.group_id" :options="groupOptions" searchable />
              </div>
              <div class="grid grid-cols-2 gap-3">
                <div class="min-w-0">
                  <label class="input-label">{{ t('usage.type') }}</label>
                  <Select v-model="usageDraft.request_type" :options="requestTypeOptions" />
                </div>
                <div class="min-w-0">
                  <label class="input-label">{{ t('admin.usage.billingType') }}</label>
                  <Select v-model="usageDraft.billing_type" :options="billingTypeOptions" />
                </div>
              </div>
              <div>
                <label class="input-label">{{ t('admin.usage.billingMode') }}</label>
                <Select v-model="usageDraft.billing_mode" :options="billingModeOptions" />
              </div>
            </div>

            <div v-else class="grid gap-4">
              <div>
                <label class="input-label">{{ t('usage.errors.keyName') }}</label>
                <Select v-model="errorDraft.api_key_id" :options="errorKeyOptions" />
              </div>
              <div>
                <label class="input-label">{{ t('usage.errors.model') }}</label>
                <Select v-model="errorDraft.model" :options="errorModelOptions" searchable creatable clearable :placeholder="t('usage.errors.modelPlaceholder')" />
              </div>
              <div class="grid grid-cols-2 gap-3">
                <div class="min-w-0">
                  <label class="input-label">{{ t('usage.errors.category') }}</label>
                  <Select v-model="errorDraft.category" :options="errorCategoryOptions" />
                </div>
                <div class="min-w-0">
                  <label class="input-label">{{ t('usage.errors.status') }}</label>
                  <Select v-model="errorDraft.status_code" :options="errorStatusOptions" />
                </div>
              </div>
            </div>
          </div>

          <footer class="grid grid-cols-[104px_minmax(0,1fr)] gap-2 border-t border-outline bg-surface-raised px-4 pb-[max(16px,env(safe-area-inset-bottom))] pt-3">
            <button type="button" class="btn btn-secondary h-11" @click="resetDraft">{{ t('common.reset') }}</button>
            <button type="button" class="btn btn-primary h-11" @click="applyDraft">{{ t('common.confirm') }}</button>
          </footer>
        </section>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import type { SelectOption } from '@/components/common/Select.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'

interface UsageDraft {
  api_key_id: number | null
  model: string | null
  group_id: number | null
  request_type: string | null
  billing_type: number | null
  billing_mode: string | null
}

interface ErrorDraft {
  api_key_id: number | null
  model: string | null
  category: string
  status_code: number | null
}

const props = defineProps<{
  open: boolean
  mode: 'usage' | 'errors'
  usageFilter: UsageDraft
  errorFilter: ErrorDraft
  apiKeyOptions: SelectOption[]
  groupOptions: SelectOption[]
  modelOptions: SelectOption[]
  requestTypeOptions: SelectOption[]
  billingTypeOptions: SelectOption[]
  billingModeOptions: SelectOption[]
  errorKeyOptions: SelectOption[]
  errorModelOptions: SelectOption[]
  errorCategoryOptions: SelectOption[]
  errorStatusOptions: SelectOption[]
}>()

const emit = defineEmits<{
  close: []
  applyUsage: [draft: UsageDraft]
  applyErrors: [draft: ErrorDraft]
}>()

const { t } = useI18n()
const titleId = 'usage-mobile-filter-title'
const sheetPanel = ref<HTMLElement | null>(null)
const closeButton = ref<HTMLButtonElement | null>(null)
const usageDraft = reactive<UsageDraft>({ api_key_id: null, model: null, group_id: null, request_type: null, billing_type: null, billing_mode: null })
const errorDraft = reactive<ErrorDraft>({ api_key_id: null, model: null, category: '', status_code: null })
let previousFocus: HTMLElement | null = null
let previousBodyOverflow = ''

const syncDrafts = () => {
  Object.assign(usageDraft, props.usageFilter)
  Object.assign(errorDraft, props.errorFilter)
}

const activeCount = computed(() => {
  const values = props.mode === 'usage'
    ? Object.values(usageDraft)
    : Object.values(errorDraft)
  return values.filter(value => value !== null && value !== undefined && value !== '').length
})

const resetDraft = () => {
  if (props.mode === 'usage') {
    Object.assign(usageDraft, { api_key_id: null, model: null, group_id: null, request_type: null, billing_type: null, billing_mode: null })
  } else {
    Object.assign(errorDraft, { api_key_id: null, model: null, category: '', status_code: null })
  }
}

const applyDraft = () => {
  if (props.mode === 'usage') emit('applyUsage', { ...usageDraft })
  else emit('applyErrors', { ...errorDraft })
}

const handleKeydown = (event: KeyboardEvent) => {
  if (event.key === 'Escape') {
    event.preventDefault()
    emit('close')
    return
  }
  if (event.key !== 'Tab' || !sheetPanel.value) return
  const focusable = [...sheetPanel.value.querySelectorAll<HTMLElement>('button:not([disabled]),input:not([disabled]),select:not([disabled]),[tabindex]:not([tabindex="-1"])')]
    .filter(element => getComputedStyle(element).visibility !== 'hidden')
  if (!focusable.length) return
  const first = focusable[0]
  const last = focusable[focusable.length - 1]
  if (event.shiftKey && document.activeElement === first) {
    event.preventDefault()
    last.focus()
  } else if (!event.shiftKey && document.activeElement === last) {
    event.preventDefault()
    first.focus()
  }
}

watch(() => props.open, async (open) => {
  if (open) {
    syncDrafts()
    previousFocus = document.activeElement as HTMLElement
    previousBodyOverflow = document.body.style.overflow
    document.body.style.overflow = 'hidden'
    document.addEventListener('keydown', handleKeydown)
    await nextTick()
    closeButton.value?.focus()
  } else {
    document.removeEventListener('keydown', handleKeydown)
    document.body.style.overflow = previousBodyOverflow
    previousFocus?.focus()
    previousFocus = null
  }
})

onBeforeUnmount(() => {
  document.removeEventListener('keydown', handleKeydown)
  document.body.style.overflow = previousBodyOverflow
})
</script>

<style scoped>
.usage-filter-sheet-enter-active,
.usage-filter-sheet-leave-active { transition: opacity 180ms ease; }
.usage-filter-sheet-enter-active section,
.usage-filter-sheet-leave-active section { transition: transform 180ms ease; }
.usage-filter-sheet-enter-from,
.usage-filter-sheet-leave-to { opacity: 0; }
.usage-filter-sheet-enter-from section,
.usage-filter-sheet-leave-to section { transform: translateY(100%); }
</style>
