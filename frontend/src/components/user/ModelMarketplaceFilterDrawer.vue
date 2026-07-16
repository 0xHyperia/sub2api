<template>
  <Teleport to="body">
    <Transition name="marketplace-filter-sheet">
      <div
        v-if="open"
        class="fixed inset-0 z-50 flex items-end bg-black/25 backdrop-blur-[1px] lg:hidden"
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
              <h2 :id="titleId" class="text-sm font-semibold text-foreground">{{ t('modelMarketplace.filters.title') }}</h2>
              <p class="mt-0.5 text-[10px] text-foreground-subtle">
                {{ t('modelMarketplace.filters.selectedCount', { count: selectedCount }) }}
              </p>
            </div>
            <button
              ref="closeButton"
              type="button"
              class="btn btn-ghost btn-icon h-9 w-9"
              :aria-label="t('common.close')"
              @click="emit('close')"
            >
              <Icon name="x" size="sm" />
            </button>
          </header>

          <div class="min-h-0 flex-1 overflow-y-auto overscroll-contain px-4 py-1">
            <section class="border-b border-outline py-4" :aria-labelledby="`${titleId}-provider`">
              <h3 :id="`${titleId}-provider`" class="text-xs font-semibold text-foreground">
                {{ t('modelMarketplace.filters.provider') }}
              </h3>
              <div class="mt-3 grid grid-cols-2 gap-2">
                <button
                  type="button"
                  class="marketplace-sheet-option"
                  :class="draftProvider === 'all' ? 'marketplace-sheet-option-active' : ''"
                  :aria-pressed="draftProvider === 'all'"
                  @click="draftProvider = 'all'"
                >
                  <span class="truncate">{{ t('modelMarketplace.filters.allProviders') }}</span>
                  <span class="marketplace-sheet-count">{{ entriesCount }}</span>
                </button>
                <button
                  v-for="provider in providers"
                  :key="provider.value"
                  type="button"
                  class="marketplace-sheet-option"
                  :class="draftProvider === provider.value ? 'marketplace-sheet-option-active' : ''"
                  :aria-pressed="draftProvider === provider.value"
                  @click="draftProvider = provider.value"
                >
                  <span class="flex min-w-0 items-center gap-1.5">
                    <PlatformIcon :platform="provider.value as GroupPlatform" size="sm" :class="platformIconClass(provider.value)" />
                    <span class="truncate">{{ provider.label }}</span>
                  </span>
                  <span class="marketplace-sheet-count">{{ provider.count }}</span>
                </button>
              </div>
            </section>

            <section class="border-b border-outline py-4" :aria-labelledby="`${titleId}-capability`">
              <h3 :id="`${titleId}-capability`" class="text-xs font-semibold text-foreground">
                {{ t('modelMarketplace.filters.capability') }}
              </h3>
              <div class="mt-3 grid grid-cols-2 gap-2">
                <button
                  v-for="capability in capabilityOptions"
                  :key="capability.value"
                  type="button"
                  class="marketplace-sheet-option"
                  :class="draftCapability === capability.value ? 'marketplace-sheet-option-active' : ''"
                  :aria-pressed="draftCapability === capability.value"
                  @click="draftCapability = capability.value"
                >
                  <span class="flex min-w-0 items-center gap-1.5">
                    <Icon v-if="capability.icon" :name="capability.icon" size="xs" />
                    <span class="truncate">{{ capability.label }}</span>
                  </span>
                  <span class="marketplace-sheet-count">{{ capability.count }}</span>
                </button>
              </div>
            </section>

            <section class="border-b border-outline py-4" :aria-labelledby="`${titleId}-group`">
              <div class="flex items-center justify-between gap-3">
                <h3 :id="`${titleId}-group`" class="text-xs font-semibold text-foreground">
                  {{ t('modelMarketplace.filters.group') }}
                </h3>
                <span class="text-[10px] text-foreground-subtle">{{ groups.length }}</span>
              </div>
              <div class="mt-3 divide-y divide-outline border-y border-outline">
                <button
                  type="button"
                  class="marketplace-sheet-row"
                  :class="draftGroup === 'all' ? 'text-brand' : 'text-foreground'"
                  :aria-pressed="draftGroup === 'all'"
                  @click="draftGroup = 'all'"
                >
                  <span>{{ t('modelMarketplace.filters.allGroups') }}</span>
                  <Icon v-if="draftGroup === 'all'" name="check" size="sm" />
                </button>
                <button
                  v-for="group in groups"
                  :key="group.id"
                  type="button"
                  class="marketplace-sheet-row"
                  :class="draftGroup === String(group.id) ? 'text-brand' : 'text-foreground'"
                  :aria-pressed="draftGroup === String(group.id)"
                  @click="draftGroup = String(group.id)"
                >
                  <span class="min-w-0 truncate">{{ group.name }}</span>
                  <span class="flex shrink-0 items-center gap-2">
                    <span class="font-mono text-[10px] text-foreground-subtle">×{{ formatRate(group.effectiveRate) }}</span>
                    <Icon v-if="draftGroup === String(group.id)" name="check" size="sm" />
                  </span>
                </button>
              </div>
            </section>

            <section class="border-b border-outline py-4" :aria-labelledby="`${titleId}-billing`">
              <h3 :id="`${titleId}-billing`" class="text-xs font-semibold text-foreground">
                {{ t('modelMarketplace.filters.billing') }}
              </h3>
              <div class="mt-3 grid grid-cols-2 gap-2">
                <button
                  v-for="mode in billingOptions"
                  :key="mode.value"
                  type="button"
                  class="marketplace-sheet-option"
                  :class="draftBilling === mode.value ? 'marketplace-sheet-option-active' : ''"
                  :aria-pressed="draftBilling === mode.value"
                  @click="draftBilling = mode.value"
                >
                  <span class="truncate">{{ mode.label }}</span>
                  <span class="marketplace-sheet-count">{{ mode.count }}</span>
                </button>
              </div>
            </section>

            <section class="py-4" :aria-labelledby="`${titleId}-display`">
              <h3 :id="`${titleId}-display`" class="text-xs font-semibold text-foreground">
                {{ t('modelMarketplace.filters.display') }}
              </h3>
              <label class="mt-3 flex min-h-11 items-center justify-between gap-4 border-y border-outline py-2">
                <span class="text-sm text-foreground-muted">{{ t('modelMarketplace.effectivePrice') }}</span>
                <Toggle v-model="draftShowEffectivePrices" :aria-label="t('modelMarketplace.effectivePrice')" />
              </label>
            </section>
          </div>

          <footer class="grid grid-cols-[104px_minmax(0,1fr)] gap-2 border-t border-outline bg-surface-raised px-4 pb-[max(16px,env(safe-area-inset-bottom))] pt-3">
            <button type="button" class="btn btn-secondary h-11" @click="resetDraft">
              {{ t('modelMarketplace.filters.reset') }}
            </button>
            <button type="button" class="btn btn-primary h-11" @click="applyDraft">
              {{ t('modelMarketplace.filters.showResults', { count: resultCount }) }}
            </button>
          </footer>
        </section>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import type { GroupPlatform } from '@/types'
import type { MarketplaceGroupOption } from '@/views/user/modelMarketplace'
import Icon from '@/components/icons/Icon.vue'
import PlatformIcon from '@/components/common/PlatformIcon.vue'
import Toggle from '@/components/common/Toggle.vue'
import { platformIconClass } from '@/utils/platformColors'

interface FilterOption {
  value: string
  label: string
  count: number
}

type CapabilityIconName =
  | 'eye'
  | 'image'
  | 'microphone'
  | 'video'
  | 'cog'
  | 'brain'
  | 'database'
  | 'globe'
  | 'document'
  | 'terminal'
  | 'sparkles'
  | 'speaker'
  | 'arrowsUpDown'
  | 'sort'
  | 'codeBracket'
  | 'edit'
  | 'signal'
  | 'chatBubble'
  | 'link'
  | 'cube'
  | 'badge'

interface CapabilityFilterOption extends FilterOption {
  icon: CapabilityIconName | null
}

const props = defineProps<{
  open: boolean
  providers: FilterOption[]
  groups: MarketplaceGroupOption[]
  billingOptions: FilterOption[]
  capabilityOptions: CapabilityFilterOption[]
  entriesCount: number
  selectedProvider: string
  selectedGroup: string
  selectedBilling: string
  selectedCapability: string
  showEffectivePrices: boolean
  resultCount: number
}>()

const emit = defineEmits<{
  close: []
  apply: [filters: { provider: string; group: string; billing: string; capability: string; showEffectivePrices: boolean }]
  'draft-change': [filters: { provider: string; group: string; billing: string; capability: string }]
}>()

const { t } = useI18n()
const closeButton = ref<HTMLButtonElement | null>(null)
const sheetPanel = ref<HTMLElement | null>(null)
const draftProvider = ref('all')
const draftGroup = ref('all')
const draftBilling = ref('all')
const draftCapability = ref('all')
const draftShowEffectivePrices = ref(true)
const titleId = `model-marketplace-filters-${Math.random().toString(36).slice(2)}`
let previousFocus: HTMLElement | null = null
let previousBodyOverflow = ''

const selectedCount = computed(() => [draftProvider.value, draftGroup.value, draftBilling.value, draftCapability.value].filter(value => value !== 'all').length)

function formatRate(value: number) { return Number(value.toFixed(4)).toString() }

function resetDraft() {
  draftProvider.value = 'all'
  draftGroup.value = 'all'
  draftBilling.value = 'all'
  draftCapability.value = 'all'
  draftShowEffectivePrices.value = true
}

function applyDraft() {
  emit('apply', {
    provider: draftProvider.value,
    group: draftGroup.value,
    billing: draftBilling.value,
    capability: draftCapability.value,
    showEffectivePrices: draftShowEffectivePrices.value,
  })
}

function handleKeydown(event: KeyboardEvent) {
  if (!props.open) return
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

watch([draftProvider, draftGroup, draftBilling, draftCapability], () => {
  emit('draft-change', { provider: draftProvider.value, group: draftGroup.value, billing: draftBilling.value, capability: draftCapability.value })
})

watch(() => props.open, async (open) => {
  if (open) {
    draftProvider.value = props.selectedProvider
    draftGroup.value = props.selectedGroup
    draftBilling.value = props.selectedBilling
    draftCapability.value = props.selectedCapability
    draftShowEffectivePrices.value = props.showEffectivePrices
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
.marketplace-sheet-option {
  display: flex;
  min-height: 40px;
  min-width: 0;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  border: 1px solid rgb(var(--color-border));
  border-radius: var(--radius-sm);
  padding: 7px 10px;
  color: rgb(var(--color-foreground-muted));
  font-size: 12px;
  text-align: left;
  transition: background-color var(--duration-fast), border-color var(--duration-fast), color var(--duration-fast);
}

.marketplace-sheet-option-active {
  border-color: rgb(var(--color-brand) / 0.35);
  background: rgb(var(--color-brand-subtle));
  color: rgb(var(--color-brand));
  font-weight: 600;
}

.marketplace-sheet-count {
  flex-shrink: 0;
  color: rgb(var(--color-foreground-subtle));
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 10px;
}

.marketplace-sheet-row {
  display: flex;
  min-height: 44px;
  width: 100%;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 8px 2px;
  font-size: 13px;
  text-align: left;
}

.marketplace-filter-sheet-enter-active,
.marketplace-filter-sheet-leave-active { transition: opacity 180ms ease; }
.marketplace-filter-sheet-enter-active section,
.marketplace-filter-sheet-leave-active section { transition: transform 180ms ease; }
.marketplace-filter-sheet-enter-from,
.marketplace-filter-sheet-leave-to { opacity: 0; }
.marketplace-filter-sheet-enter-from section,
.marketplace-filter-sheet-leave-to section { transform: translateY(100%); }
</style>
