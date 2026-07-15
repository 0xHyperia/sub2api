<template>
  <Teleport to="body">
    <Transition name="marketplace-drawer">
      <div
        v-if="entry"
        class="fixed inset-0 z-50 bg-black/25 backdrop-blur-[1px]"
        role="dialog"
        aria-modal="true"
        :aria-labelledby="titleId"
        @click.self="emit('close')"
      >
        <aside ref="drawerPanel" tabindex="-1" class="absolute inset-y-0 right-0 flex w-full max-w-xl flex-col border-l border-outline bg-surface-raised shadow-floating">
          <header class="grid min-h-[72px] grid-cols-[40px_minmax(0,1fr)_40px] items-center gap-3 border-b border-outline px-5 py-3">
            <span class="flex h-9 w-9 shrink-0 items-center justify-center rounded-panel border" :class="platformBadgeClass(entry.platform)">
              <PlatformIcon :platform="entry.platform as GroupPlatform" size="md" />
            </span>
            <div class="min-w-0 flex-1">
              <h2 :id="titleId" class="min-w-0 break-words text-base font-semibold text-foreground">
                {{ entry.name }}
                <span v-if="entry.label" class="ml-1 inline-flex h-5 max-w-28 items-center align-middle truncate rounded-control bg-black px-2 text-[11px] font-bold leading-none text-white" :title="entry.label">{{ entry.label }}</span>
              </h2>
              <p class="mt-0.5 text-xs text-foreground-muted">{{ providerLabel }} · {{ billingLabel }}</p>
            </div>
            <button ref="closeButton" type="button" class="btn btn-ghost btn-icon shrink-0" :aria-label="t('common.close')" @click="emit('close')">
              <Icon name="x" size="md" />
            </button>
          </header>

          <div class="min-h-0 flex-1 overflow-y-auto px-5">
            <section v-if="monitorEnabled" class="border-b border-outline py-5">
              <div class="flex items-center justify-between gap-3">
                <h3 class="text-sm font-semibold text-foreground">{{ t('modelMarketplace.details.monitor') }}</h3>
                <span class="inline-flex items-center gap-2 text-xs font-medium" :class="monitorTextClass(entry.monitorStatus?.status)">
                  <span class="h-2 w-2 rounded-full bg-current"></span>
                  {{ monitorStatusLabel(entry.monitorStatus?.status) }}
                </span>
              </div>
              <dl class="mt-3 grid grid-cols-2 divide-x divide-outline border-y border-outline py-3">
                <div class="pr-4">
                  <dt class="text-[10px] font-medium text-foreground-subtle">{{ t('modelMarketplace.details.availability7d') }}</dt>
                  <dd class="mt-1 font-mono text-base font-semibold tabular-nums text-foreground">{{ monitorAvailability }}</dd>
                </div>
                <div class="pl-4">
                  <dt class="text-[10px] font-medium text-foreground-subtle">{{ t('modelMarketplace.details.latestLatency') }}</dt>
                  <dd class="mt-1 font-mono text-base font-semibold tabular-nums text-foreground">{{ latestLatency }}</dd>
                </div>
              </dl>
              <div class="mt-4 flex items-center justify-between text-[10px] text-foreground-subtle">
                <span>{{ t('modelMarketplace.details.history') }}</span>
                <span>{{ lastCheckedAt }}</span>
              </div>
              <ModelMonitorTimeline class="mt-1.5" :points="entry.monitorStatus?.timeline" />
              <div class="mt-1.5 flex items-center justify-between text-[9px] text-foreground-subtle">
                <span>{{ t('modelMarketplace.details.older') }}</span>
                <span>{{ t('modelMarketplace.details.newest') }}</span>
              </div>
            </section>

            <section class="py-5" :class="entry.groups.length ? 'border-b border-outline' : ''">
              <div class="mb-4 flex items-center justify-between gap-4">
                <h3 class="text-sm font-semibold text-foreground">{{ t('modelMarketplace.details.pricing') }}</h3>
                <label v-if="groups.length" class="flex items-center gap-2 text-xs text-foreground-muted">
                  <span class="sr-only">{{ t('modelMarketplace.details.billingGroup') }}</span>
                  <select :value="activeGroup?.id" class="input h-9 max-w-56 py-1 text-xs" @change="selectGroupFromMenu">
                    <option v-for="group in groups" :key="group.id" :value="group.id">{{ group.name }} · ×{{ formatRate(group.effectiveRate) }}</option>
                  </select>
                </label>
              </div>

              <template v-if="pricingRows.length">
                <dl v-if="billingCategory(entry.pricing) === 'usage'" class="grid grid-cols-2 divide-x divide-outline border-y border-outline py-4">
                  <div v-for="row in primaryPricingRows" :key="row.key" class="px-4 first:pl-0 last:pr-0">
                    <dt class="text-xs font-medium text-foreground-muted">{{ row.label }}</dt>
                    <dd class="mt-1.5 flex flex-wrap items-baseline gap-x-1.5 gap-y-1">
                      <span class="font-mono text-xl font-semibold tabular-nums text-foreground">{{ row.value }}</span>
                      <span class="text-[10px] text-foreground-subtle">/ 1M</span>
                    </dd>
                    <div v-if="row.baseValue" class="mt-1 text-[10px] text-foreground-subtle">
                      {{ t('modelMarketplace.details.basePrice') }} <span class="font-mono line-through">{{ row.baseValue }}</span>
                    </div>
                  </div>
                </dl>
                <dl v-if="secondaryPricingRows.length" class="mt-3 divide-y divide-outline border-y border-outline">
                  <div v-for="row in secondaryPricingRows" :key="row.key" class="flex items-center justify-between gap-4 py-2.5">
                    <dt class="text-xs text-foreground-muted">{{ row.label }}</dt>
                    <dd class="flex items-baseline gap-1.5 text-right">
                      <span class="font-mono text-sm font-semibold tabular-nums text-foreground">{{ row.value }}</span>
                      <span class="text-[9px] text-foreground-subtle">/ 1M</span>
                      <span v-if="row.baseValue" class="font-mono text-[10px] tabular-nums text-foreground-subtle line-through">{{ row.baseValue }}</span>
                    </dd>
                  </div>
                </dl>
                <dl v-if="billingCategory(entry.pricing) === 'request'" class="border-y border-outline py-4">
                  <div v-for="row in primaryPricingRows" :key="row.key" class="flex items-end justify-between gap-4">
                    <div>
                      <dt class="text-xs font-medium text-foreground-muted">{{ row.label }}</dt>
                      <dd class="mt-1.5 flex items-baseline gap-2">
                        <span class="font-mono text-2xl font-semibold tabular-nums text-foreground">{{ row.value }}</span>
                        <span class="text-xs text-foreground-subtle">{{ t('modelMarketplace.price.perRequest') }}</span>
                      </dd>
                    </div>
                    <span v-if="row.baseValue" class="font-mono text-xs text-foreground-subtle line-through">{{ row.baseValue }}</span>
                  </div>
                </dl>
              </template>
              <p v-else class="border-y border-outline py-5 text-sm text-foreground-subtle">{{ t('modelMarketplace.noPricing') }}</p>

              <div v-if="entry.pricing?.intervals.length" class="mt-5">
                <h4 class="mb-2 text-xs font-semibold uppercase text-foreground-subtle">{{ t('modelMarketplace.details.tieredPricing') }}</h4>
                <div class="overflow-x-auto border-y border-outline">
                  <table class="w-full min-w-96 text-left text-xs">
                    <thead class="text-foreground-subtle"><tr><th class="py-2 pr-3">{{ t('modelMarketplace.details.range') }}</th><th class="py-2 pr-3">{{ t('modelMarketplace.price.input') }}</th><th class="py-2">{{ t('modelMarketplace.price.output') }}</th></tr></thead>
                    <tbody class="divide-y divide-outline text-foreground-muted">
                      <tr v-for="(interval, index) in entry.pricing.intervals" :key="index">
                        <td class="py-2 pr-3">{{ intervalLabel(interval.min_tokens, interval.max_tokens) }}</td>
                        <td class="py-2 pr-3 font-mono">{{ intervalPrice(interval.input_price) }}</td>
                        <td class="py-2 font-mono">{{ intervalPrice(interval.output_price) }}</td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>
            </section>

            <section v-if="entry.groups.length" class="py-5">
              <div class="mb-3 flex items-center justify-between gap-3">
                <h3 class="text-sm font-semibold text-foreground">{{ t('modelMarketplace.details.groups') }}</h3>
                <span class="text-xs text-foreground-subtle">{{ entry.groups.length }}</span>
              </div>
              <div class="divide-y divide-outline border-y border-outline">
                <button
                  v-for="group in groups"
                  :key="group.id"
                  type="button"
                  class="flex min-h-12 w-full items-center justify-between gap-4 py-2.5 text-left transition-colors hover:bg-surface-subtle focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-focus"
                  @click="emit('selectGroup', group.id)"
                >
                  <span class="min-w-0 truncate text-sm font-medium" :class="group.id === activeGroup?.id ? 'text-brand' : 'text-foreground'">{{ group.name }}</span>
                  <span class="flex shrink-0 items-center gap-2">
                    <span class="font-mono text-xs tabular-nums text-foreground-muted">×{{ formatRate(group.effectiveRate) }}</span>
                    <Icon v-if="group.id === activeGroup?.id" name="check" size="sm" class="text-brand" />
                  </span>
                </button>
              </div>
            </section>
          </div>
        </aside>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import type { GroupPlatform } from '@/types'
import Icon from '@/components/icons/Icon.vue'
import PlatformIcon from '@/components/common/PlatformIcon.vue'
import ModelMonitorTimeline from '@/components/user/ModelMonitorTimeline.vue'
import { platformBadgeClass } from '@/utils/platformColors'
import { billingCategory, scaledPrice, type MarketplaceGroupOption, type MarketplaceModelEntry } from '@/views/user/modelMarketplace'

const props = defineProps<{
  entry: MarketplaceModelEntry | null
  groups: MarketplaceGroupOption[]
  activeGroup: MarketplaceGroupOption | null
  showEffectivePrices: boolean
  monitorEnabled: boolean
}>()

const emit = defineEmits<{
  close: []
  selectGroup: [groupId: number]
}>()

const { t, locale } = useI18n()
const closeButton = ref<HTMLButtonElement | null>(null)
const drawerPanel = ref<HTMLElement | null>(null)
const titleId = `model-marketplace-detail-${Math.random().toString(36).slice(2)}`
let previousFocus: HTMLElement | null = null
let previousBodyOverflow = ''

const rate = computed(() => props.showEffectivePrices ? props.activeGroup?.effectiveRate ?? 1 : 1)
const providerLabel = computed(() => props.entry ? labelOrFallback(`modelMarketplace.providers.${props.entry.platform}`, props.entry.platform.toUpperCase()) : '')
const billingLabel = computed(() => props.entry ? t(`modelMarketplace.billing.${billingCategory(props.entry.pricing)}`) : '')
const monitorAvailability = computed(() => {
  const value = props.entry?.monitorStatus?.availability_7d
  return value == null ? '—' : `${value.toFixed(2)}%`
})
const latestLatency = computed(() => props.entry?.monitorStatus?.latency_ms == null ? '—' : `${props.entry.monitorStatus.latency_ms} ms`)
const lastCheckedAt = computed(() => {
  const value = props.entry?.monitorStatus?.last_checked_at
  if (!value) return t('modelMarketplace.monitor.noData')
  return new Intl.DateTimeFormat(locale.value, { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' }).format(new Date(value))
})

const pricingRows = computed(() => {
  const pricing = props.entry?.pricing
  if (!pricing) return []
  const rows = pricing.billing_mode === 'token'
    ? [
        ['input', t('modelMarketplace.price.input'), pricing.input_price, 1_000_000],
        ['output', t('modelMarketplace.price.output'), pricing.output_price, 1_000_000],
        ['cache-read', t('modelMarketplace.price.cacheRead'), pricing.cache_read_price, 1_000_000],
        ['cache-write', t('modelMarketplace.price.cacheWrite'), pricing.cache_write_price, 1_000_000],
      ] as const
    : [['request', t('modelMarketplace.price.request'), pricing.per_request_price ?? pricing.image_output_price, 1]] as const
  return rows.filter(([, , value]) => value != null).map(([key, label, value, scale]) => ({
    key,
    label,
    value: scaledPrice(value, scale, rate.value),
    baseValue: props.showEffectivePrices && rate.value !== 1 ? scaledPrice(value, scale, 1) : '',
  }))
})
const primaryPricingRows = computed(() => pricingRows.value.filter(row => row.key === 'input' || row.key === 'output' || row.key === 'request'))
const secondaryPricingRows = computed(() => pricingRows.value.filter(row => row.key === 'cache-read' || row.key === 'cache-write'))

function labelOrFallback(key: string, fallback: string) { const value = t(key); return value === key ? fallback : value }
function formatRate(value: number) { return Number(value.toFixed(4)).toString() }
function monitorStatusLabel(status?: string) { return status ? t(`modelMarketplace.monitor.${status}`) : t('modelMarketplace.monitor.unknown') }
function monitorTextClass(status?: string) { return status === 'operational' ? 'text-success-foreground' : status === 'degraded' ? 'text-warning-foreground' : status === 'failed' || status === 'error' ? 'text-danger-foreground' : 'text-foreground-subtle' }
function intervalLabel(min: number, max: number | null) { return max == null ? `${min.toLocaleString()}+` : `${min.toLocaleString()}–${max.toLocaleString()}` }
function intervalPrice(value: number | null) { return scaledPrice(value, 1_000_000, rate.value) }
function selectGroupFromMenu(event: Event) { emit('selectGroup', Number((event.target as HTMLSelectElement).value)) }
function handleKeydown(event: KeyboardEvent) {
  if (!props.entry) return
  if (event.key === 'Escape') {
    event.preventDefault()
    emit('close')
    return
  }
  if (event.key !== 'Tab' || !drawerPanel.value) return
  const focusable = [...drawerPanel.value.querySelectorAll<HTMLElement>('button:not([disabled]),a[href],[tabindex]:not([tabindex="-1"])')]
    .filter(element => getComputedStyle(element).visibility !== 'hidden')
  if (!focusable.length) {
    event.preventDefault()
    drawerPanel.value.focus()
    return
  }
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

watch(() => props.entry, async (entry) => {
  if (entry) {
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
.marketplace-drawer-enter-active,
.marketplace-drawer-leave-active { transition: opacity 180ms ease; }
.marketplace-drawer-enter-active aside,
.marketplace-drawer-leave-active aside { transition: transform 180ms ease; }
.marketplace-drawer-enter-from,
.marketplace-drawer-leave-to { opacity: 0; }
.marketplace-drawer-enter-from aside,
.marketplace-drawer-leave-to aside { transform: translateX(100%); }
</style>
