<template>
  <AppLayout>
    <div class="space-y-4">
      <div
        v-if="loadError"
        class="flex flex-col gap-3 rounded-panel border border-danger/30 bg-danger-subtle p-4 text-sm text-danger-foreground sm:flex-row sm:items-center sm:justify-between"
        role="alert"
      >
        <span>{{ t('modelMarketplace.loadError') }}</span>
        <button type="button" class="btn btn-secondary btn-sm" @click="loadMarketplace">
          <Icon name="refresh" size="sm" />
          {{ t('common.retry') }}
        </button>
      </div>

      <div class="grid min-w-0 gap-4 lg:grid-cols-[220px_minmax(0,1fr)]">
        <aside
          class="h-fit border-b border-outline pb-4 lg:sticky lg:top-20 lg:border-b-0 lg:border-r lg:pb-0 lg:pr-4"
          :aria-label="t('modelMarketplace.filters.title')"
        >
          <div class="flex items-center justify-between">
            <h2 class="flex items-center gap-2 text-sm font-semibold text-foreground">
              <Icon name="filter" size="sm" />
              {{ t('modelMarketplace.filters.title') }}
            </h2>
            <button
              type="button"
              class="rounded-control px-2 py-1 text-xs text-foreground-muted transition-colors hover:bg-surface-subtle hover:text-foreground"
              @click="resetFilters"
            >
              {{ t('modelMarketplace.filters.reset') }}
            </button>
          </div>

          <div class="marketplace-filter-sections">
          <section class="marketplace-filter-section" :aria-labelledby="'marketplace-provider-filter'">
            <h3 id="marketplace-provider-filter" class="text-xs font-semibold uppercase text-foreground-subtle">
              {{ t('modelMarketplace.filters.provider') }}
            </h3>
            <div class="mt-2 space-y-1">
              <button
                type="button"
                class="marketplace-filter-option"
                :class="selectedProvider === 'all' ? 'marketplace-filter-option-active' : ''"
                :aria-pressed="selectedProvider === 'all'"
                @click="selectedProvider = 'all'"
              >
                <span>{{ t('modelMarketplace.filters.allProviders') }}</span>
                <span class="marketplace-filter-count">{{ entries.length }}</span>
              </button>
              <button
                v-for="provider in providers"
                :key="provider.value"
                type="button"
                class="marketplace-filter-option"
                :class="selectedProvider === provider.value ? 'marketplace-filter-option-active' : ''"
                :aria-pressed="selectedProvider === provider.value"
                @click="selectedProvider = provider.value"
              >
                <span class="flex min-w-0 items-center gap-2">
                  <PlatformIcon :platform="provider.value as GroupPlatform" size="sm" />
                  <span class="truncate">{{ provider.label }}</span>
                </span>
                <span class="marketplace-filter-count">{{ provider.count }}</span>
              </button>
            </div>
          </section>

          <section class="marketplace-filter-section" :aria-labelledby="'marketplace-group-filter'">
            <h3 id="marketplace-group-filter" class="text-xs font-semibold uppercase text-foreground-subtle">
              {{ t('modelMarketplace.filters.group') }}
            </h3>
            <div class="mt-2 max-h-52 space-y-1 overflow-y-auto pr-1">
              <button
                type="button"
                class="marketplace-filter-option"
                :class="selectedGroup === 'all' ? 'marketplace-filter-option-active' : ''"
                :aria-pressed="selectedGroup === 'all'"
                @click="selectedGroup = 'all'"
              >
                <span>{{ t('modelMarketplace.filters.allGroups') }}</span>
                <span class="marketplace-filter-count">{{ groups.length }}</span>
              </button>
              <button
                v-for="group in groups"
                :key="group.id"
                type="button"
                class="marketplace-filter-option"
                :class="selectedGroup === String(group.id) ? 'marketplace-filter-option-active' : ''"
                :aria-pressed="selectedGroup === String(group.id)"
                @click="selectedGroup = String(group.id)"
              >
                <span class="min-w-0 truncate">{{ group.name }}</span>
                <span class="marketplace-filter-count">×{{ formatRate(group.effectiveRate) }}</span>
              </button>
            </div>
          </section>

          <section class="marketplace-filter-section" :aria-labelledby="'marketplace-billing-filter'">
            <h3 id="marketplace-billing-filter" class="text-xs font-semibold uppercase text-foreground-subtle">
              {{ t('modelMarketplace.filters.billing') }}
            </h3>
            <div class="mt-2 space-y-1">
              <button
                v-for="mode in billingOptions"
                :key="mode.value"
                type="button"
                class="marketplace-filter-option"
                :class="selectedBilling === mode.value ? 'marketplace-filter-option-active' : ''"
                :aria-pressed="selectedBilling === mode.value"
                @click="selectedBilling = mode.value"
              >
                <span>{{ mode.label }}</span>
                <span class="marketplace-filter-count">{{ mode.count }}</span>
              </button>
            </div>
          </section>
          </div>
        </aside>

        <section class="min-w-0" :aria-label="t('modelMarketplace.results')">
          <div
            data-testid="marketplace-toolbar"
            class="flex flex-col gap-3 border-b border-outline pb-3 xl:flex-row xl:items-center xl:justify-between"
          >
            <div class="flex min-w-0 flex-1 flex-col gap-2 sm:flex-row sm:items-center">
              <SearchInput
                v-model="searchQuery"
                :placeholder="t('modelMarketplace.searchPlaceholder')"
                class="w-full min-w-0 sm:max-w-lg sm:flex-1"
              />
              <div class="flex shrink-0 items-center gap-2.5 text-xs text-foreground-subtle" aria-live="polite">
                <span><strong class="font-semibold text-foreground">{{ filteredEntries.length }}</strong> {{ t('modelMarketplace.stats.models') }}</span>
                <span class="h-3 w-px bg-outline" aria-hidden="true"></span>
                <span><strong class="font-semibold text-foreground">{{ providers.length }}</strong> {{ t('modelMarketplace.stats.providers') }}</span>
              </div>
            </div>
            <div class="grid grid-cols-[minmax(0,1fr)_40px] items-center gap-2 sm:flex sm:flex-nowrap">
              <label class="col-span-2 flex min-h-10 items-center gap-2 whitespace-nowrap text-xs font-medium text-foreground-muted sm:col-auto">
                <Toggle
                  v-model="showEffectivePrices"
                  :aria-label="t('modelMarketplace.effectivePrice')"
                />
                <span>{{ t('modelMarketplace.effectivePrice') }}</span>
              </label>
              <label class="sr-only" for="marketplace-sort">{{ t('modelMarketplace.sort.label') }}</label>
              <select id="marketplace-sort" v-model="sortMode" class="input h-10 min-w-0 text-sm sm:w-44 sm:flex-none">
                <option value="name">{{ t('modelMarketplace.sort.name') }}</option>
                <option value="price">{{ t('modelMarketplace.sort.price') }}</option>
              </select>
              <button
                type="button"
                class="btn btn-secondary btn-icon"
                :disabled="loading"
                :title="t('common.refresh')"
                :aria-label="t('common.refresh')"
                @click="loadMarketplace"
              >
                <Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" />
              </button>
            </div>
          </div>

          <div v-if="loading && entries.length === 0" class="grid gap-4 pt-4 md:grid-cols-2 2xl:grid-cols-3" role="status">
            <div v-for="index in 6" :key="index" class="card h-64 animate-pulse bg-surface-subtle"></div>
            <span class="sr-only">{{ t('common.loading') }}</span>
          </div>

          <div
            v-else-if="pagedEntries.length === 0"
            class="flex min-h-80 flex-col items-center justify-center border-b border-outline py-12 text-center"
          >
            <span class="flex h-12 w-12 items-center justify-center rounded-panel border border-outline bg-surface-subtle text-foreground-subtle">
              <Icon name="inbox" size="lg" />
            </span>
            <h2 class="mt-4 text-base font-semibold text-foreground">{{ t('modelMarketplace.empty.title') }}</h2>
            <p class="mt-1 max-w-md text-sm text-foreground-subtle">{{ t('modelMarketplace.empty.description') }}</p>
          </div>

          <div v-else class="grid gap-4 pt-4 md:grid-cols-2 2xl:grid-cols-3">
            <article
              v-for="entry in pagedEntries"
              :key="entry.key"
              class="card flex min-h-64 min-w-0 flex-col p-4 transition-colors hover:border-outline-strong hover:shadow-card"
            >
              <div class="flex items-start gap-3">
                <span
                  class="flex h-10 w-10 shrink-0 items-center justify-center rounded-panel border"
                  :class="platformBadgeClass(entry.platform)"
                  aria-hidden="true"
                >
                  <PlatformIcon :platform="entry.platform as GroupPlatform" size="lg" />
                </span>
                <div class="min-w-0 flex-1">
                  <div class="flex min-w-0 items-start gap-2">
                    <h2 class="min-w-0 flex-1 break-words text-sm font-semibold leading-5 text-foreground">
                      {{ entry.name }}
                    </h2>
                    <button
                      type="button"
                      class="btn-ghost btn-icon h-8 w-8 shrink-0"
                      :title="t('modelMarketplace.copyModel')"
                      :aria-label="t('modelMarketplace.copyModel')"
                      @click="copyModel(entry.name)"
                    >
                      <Icon name="copy" size="sm" />
                    </button>
                  </div>
                  <div class="mt-1 flex flex-wrap items-center gap-1.5">
                    <span class="rounded-control bg-surface-subtle px-2 py-0.5 text-[10px] font-semibold uppercase text-foreground-muted">
                      {{ providerLabel(entry.platform) }}
                    </span>
                    <span class="rounded-control bg-brand-subtle px-2 py-0.5 text-[10px] font-semibold text-brand">
                      {{ billingModeLabel(entry.pricing?.billing_mode) }}
                    </span>
                    <span class="rounded-control bg-warning-subtle px-2 py-0.5 text-[10px] font-semibold text-warning-foreground">
                      ×{{ formatRate(effectiveRate(entry)) }}
                    </span>
                  </div>
                </div>
              </div>

              <p class="mt-3 border-t border-outline pt-3 text-xs leading-5 text-foreground-subtle">
                {{ t('modelMarketplace.availableInGroups', { count: entry.groups.length }) }}
              </p>

              <dl v-if="priceRows(entry).length > 0" class="mt-2 grid grid-cols-2 gap-x-4 gap-y-3 pb-3">
                <div v-for="row in priceRows(entry)" :key="row.key" class="min-w-0">
                  <dt class="truncate text-[11px] font-medium text-foreground-subtle">{{ row.label }}</dt>
                  <dd class="mt-1 flex flex-col gap-0.5 break-words font-mono text-base font-semibold text-foreground">
                    <span>{{ row.value }}</span>
                    <span class="font-sans text-[10px] font-normal text-foreground-subtle">{{ row.unit }}</span>
                  </dd>
                </div>
              </dl>
              <div v-else class="mt-2 flex min-h-16 items-center justify-center pb-3 text-xs text-foreground-subtle">
                {{ t('modelMarketplace.noPricing') }}
              </div>

              <div class="mt-auto border-t border-outline pt-3">
                <div class="flex flex-wrap gap-1.5">
                  <GroupBadge
                    v-for="group in visibleGroups(entry)"
                    :key="group.id"
                    :name="group.name"
                    :platform="group.platform as GroupPlatform"
                    :subscription-type="group.subscription_type as SubscriptionType"
                    :rate-multiplier="group.rate_multiplier"
                    :user-rate-multiplier="userGroupRates[group.id] ?? null"
                    always-show-rate
                  />
                  <span
                    v-if="entry.groups.length > visibleGroups(entry).length"
                    class="inline-flex items-center rounded-control border border-outline px-2 py-0.5 text-[10px] font-medium text-foreground-subtle"
                  >
                    +{{ entry.groups.length - visibleGroups(entry).length }}
                  </span>
                </div>
                <div class="mt-3 flex items-center justify-end gap-3 text-[11px] text-foreground-subtle">
                  <span v-if="showEffectivePrices">{{ t('modelMarketplace.effectivePriceShort') }}</span>
                  <span v-else>{{ t('modelMarketplace.basePriceShort') }}</span>
                </div>
              </div>
            </article>
          </div>

          <nav
            v-if="totalPages > 1"
            class="mt-5 flex items-center justify-between border-t border-outline pt-4"
            :aria-label="t('pagination.pageOf', { page: currentPage, total: totalPages })"
          >
            <p class="text-xs text-foreground-subtle">
              {{ t('modelMarketplace.pagination', { page: currentPage, total: totalPages }) }}
            </p>
            <div class="flex items-center gap-1">
              <button
                type="button"
                class="btn btn-secondary btn-icon"
                :disabled="currentPage <= 1"
                :aria-label="t('pagination.previous')"
                @click="currentPage -= 1"
              >
                <Icon name="chevronLeft" size="sm" />
              </button>
              <button
                type="button"
                class="btn btn-secondary btn-icon"
                :disabled="currentPage >= totalPages"
                :aria-label="t('pagination.next')"
                @click="currentPage += 1"
              >
                <Icon name="chevronRight" size="sm" />
              </button>
            </div>
          </nav>
        </section>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import SearchInput from '@/components/common/SearchInput.vue'
import Toggle from '@/components/common/Toggle.vue'
import GroupBadge from '@/components/common/GroupBadge.vue'
import PlatformIcon from '@/components/common/PlatformIcon.vue'
import Icon from '@/components/icons/Icon.vue'
import userChannelsAPI from '@/api/channels'
import userGroupsAPI from '@/api/groups'
import type { GroupPlatform, SubscriptionType } from '@/types'
import { useAppStore } from '@/stores/app'
import { useClipboard } from '@/composables/useClipboard'
import { extractApiErrorMessage } from '@/utils/apiError'
import { platformBadgeClass } from '@/utils/platformColors'
import {
  buildMarketplaceEntries,
  buildMarketplaceGroups,
  effectiveRateForEntry,
  primaryPrice,
  scaledPrice,
  type MarketplaceModelEntry,
} from './modelMarketplace'

const PAGE_SIZE = 18

const { t } = useI18n()
const appStore = useAppStore()
const { copyToClipboard } = useClipboard()

const catalog = ref<Awaited<ReturnType<typeof userChannelsAPI.getMarketplace>>>([])
const userGroupRates = ref<Record<number, number>>({})
const loading = ref(false)
const loadError = ref(false)
const searchQuery = ref('')
const selectedProvider = ref('all')
const selectedGroup = ref('all')
const selectedBilling = ref('all')
const sortMode = ref<'name' | 'price'>('name')
const showEffectivePrices = ref(true)
const currentPage = ref(1)

const entries = computed(() => buildMarketplaceEntries(catalog.value))
const groups = computed(() => buildMarketplaceGroups(entries.value, userGroupRates.value))

const providers = computed(() => {
  const counts = new Map<string, number>()
  for (const entry of entries.value) counts.set(entry.platform, (counts.get(entry.platform) ?? 0) + 1)
  return [...counts.entries()]
    .map(([value, count]) => ({ value, count, label: providerLabel(value) }))
    .sort((a, b) => a.label.localeCompare(b.label))
})

const billingOptions = computed(() => {
  const counts = new Map<string, number>()
  for (const entry of entries.value) {
    const mode = entry.pricing?.billing_mode || 'unpriced'
    counts.set(mode, (counts.get(mode) ?? 0) + 1)
  }
  return [
    { value: 'all', label: t('modelMarketplace.filters.allBilling'), count: entries.value.length },
    ...[...counts.entries()].map(([value, count]) => ({ value, count, label: billingModeLabel(value) })),
  ]
})

const selectedGroupId = computed(() => selectedGroup.value === 'all' ? null : Number(selectedGroup.value))

const filteredEntries = computed(() => {
  const query = searchQuery.value.trim().toLowerCase()
  const result = entries.value.filter((entry) => {
    if (selectedProvider.value !== 'all' && entry.platform !== selectedProvider.value) return false
    if (selectedGroupId.value != null && !entry.groups.some((group) => group.id === selectedGroupId.value)) return false
    const mode = entry.pricing?.billing_mode || 'unpriced'
    if (selectedBilling.value !== 'all' && mode !== selectedBilling.value) return false
    if (!query) return true
    return entry.name.toLowerCase().includes(query)
      || entry.groups.some((group) => group.name.toLowerCase().includes(query))
  })

  return result.sort((a, b) => {
    if (sortMode.value === 'price') {
      const aPrice = primaryPrice(a.pricing)
      const bPrice = primaryPrice(b.pricing)
      const aValue = aPrice == null ? Number.POSITIVE_INFINITY : aPrice * effectiveRate(a)
      const bValue = bPrice == null ? Number.POSITIVE_INFINITY : bPrice * effectiveRate(b)
      if (aValue !== bValue) return aValue - bValue
    }
    return a.name.localeCompare(b.name)
  })
})

const totalPages = computed(() => Math.max(1, Math.ceil(filteredEntries.value.length / PAGE_SIZE)))
const pagedEntries = computed(() => {
  const start = (currentPage.value - 1) * PAGE_SIZE
  return filteredEntries.value.slice(start, start + PAGE_SIZE)
})

watch(
  [searchQuery, selectedProvider, selectedGroup, selectedBilling, sortMode],
  () => { currentPage.value = 1 },
)
watch(totalPages, (pages) => {
  if (currentPage.value > pages) currentPage.value = pages
})

function providerLabel(platform: string): string {
  const known = t(`modelMarketplace.providers.${platform}`)
  return known === `modelMarketplace.providers.${platform}` ? platform.toUpperCase() : known
}

function billingModeLabel(mode?: string): string {
  if (!mode) return t('modelMarketplace.billing.unpriced')
  const known = t(`modelMarketplace.billing.${mode}`)
  return known === `modelMarketplace.billing.${mode}` ? mode : known
}

function effectiveRate(entry: MarketplaceModelEntry): number {
  return effectiveRateForEntry(entry, selectedGroupId.value, userGroupRates.value)
}

function formatRate(rate: number): string {
  return Number(rate.toFixed(4)).toString()
}

function displayPrice(value: number | null, scale: number, entry: MarketplaceModelEntry): string {
  const rate = showEffectivePrices.value ? effectiveRate(entry) : 1
  return scaledPrice(value, scale, rate)
}

function priceRows(entry: MarketplaceModelEntry) {
  const pricing = entry.pricing
  if (!pricing) return []
  if (pricing.billing_mode === 'token') {
    return [
      { key: 'input', label: t('modelMarketplace.price.input'), value: displayPrice(pricing.input_price, 1_000_000, entry), unit: t('modelMarketplace.price.perMillion') },
      { key: 'output', label: t('modelMarketplace.price.output'), value: displayPrice(pricing.output_price, 1_000_000, entry), unit: t('modelMarketplace.price.perMillion') },
      { key: 'cache-read', label: t('modelMarketplace.price.cacheRead'), value: displayPrice(pricing.cache_read_price, 1_000_000, entry), unit: t('modelMarketplace.price.perMillion') },
      { key: 'cache-write', label: t('modelMarketplace.price.cacheWrite'), value: displayPrice(pricing.cache_write_price, 1_000_000, entry), unit: t('modelMarketplace.price.perMillion') },
    ].filter((row) => row.value !== '-')
  }
  if (pricing.billing_mode === 'image') {
    return [{ key: 'image', label: t('modelMarketplace.price.image'), value: displayPrice(pricing.per_request_price ?? pricing.image_output_price, 1, entry), unit: t('modelMarketplace.price.perRequest') }]
  }
  return [{ key: 'request', label: t('modelMarketplace.price.request'), value: displayPrice(pricing.per_request_price, 1, entry), unit: t('modelMarketplace.price.perRequest') }]
}

function visibleGroups(entry: MarketplaceModelEntry) {
  if (selectedGroupId.value != null) return entry.groups.filter((group) => group.id === selectedGroupId.value)
  return entry.groups.slice(0, 3)
}

function resetFilters() {
  searchQuery.value = ''
  selectedProvider.value = 'all'
  selectedGroup.value = 'all'
  selectedBilling.value = 'all'
  sortMode.value = 'name'
}

async function copyModel(name: string) {
  await copyToClipboard(name, t('modelMarketplace.copySuccess', { name }))
}

async function loadMarketplace() {
  loading.value = true
  loadError.value = false
  try {
    const [catalogResponse, rates] = await Promise.all([
      userChannelsAPI.getMarketplace(),
      userGroupsAPI.getUserGroupRates().catch(() => ({} as Record<number, number>)),
    ])
    catalog.value = catalogResponse
    userGroupRates.value = rates
  } catch (error) {
    loadError.value = true
    appStore.showError(extractApiErrorMessage(error, t('modelMarketplace.loadError')))
  } finally {
    loading.value = false
  }
}

onMounted(loadMarketplace)
</script>

<style scoped>
.marketplace-filter-sections {
  display: grid;
  gap: 16px;
  margin-top: 16px;
}

.marketplace-filter-section {
  min-width: 0;
}

@media (max-width: 639px) {
  .marketplace-filter-section + .marketplace-filter-section {
    border-top: 1px solid rgb(var(--color-border));
    padding-top: 16px;
  }
}

@media (min-width: 640px) and (max-width: 1023px) {
  .marketplace-filter-sections {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .marketplace-filter-section + .marketplace-filter-section {
    border-left: 1px solid rgb(var(--color-border));
    padding-left: 16px;
  }
}

@media (min-width: 1024px) {
  .marketplace-filter-sections {
    display: block;
  }

  .marketplace-filter-section + .marketplace-filter-section {
    margin-top: 20px;
    border-top: 1px solid rgb(var(--color-border));
    padding-top: 20px;
  }
}

.marketplace-filter-option {
  display: flex;
  width: 100%;
  min-height: 36px;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  border: 1px solid transparent;
  border-radius: var(--radius-sm);
  padding: 6px 8px;
  color: rgb(var(--color-foreground-muted));
  font-size: 12px;
  text-align: left;
  transition: background-color var(--duration-fast), border-color var(--duration-fast), color var(--duration-fast);
}

.marketplace-filter-option:hover {
  background: rgb(var(--color-surface-subtle));
  color: rgb(var(--color-foreground));
}

.marketplace-filter-option-active {
  border-color: rgb(var(--color-border));
  background: rgb(var(--color-brand-subtle));
  color: rgb(var(--color-brand));
  font-weight: 600;
}

.marketplace-filter-count {
  flex-shrink: 0;
  border-radius: var(--radius-xs);
  background: rgb(var(--color-surface-subtle));
  padding: 1px 6px;
  color: rgb(var(--color-foreground-subtle));
  font-size: 10px;
}

.marketplace-filter-option-active .marketplace-filter-count {
  background: rgb(var(--color-surface));
  color: rgb(var(--color-brand));
}
</style>
