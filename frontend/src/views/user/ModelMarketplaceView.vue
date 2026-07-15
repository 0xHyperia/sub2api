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

      <div class="grid min-w-0 gap-4 lg:grid-cols-[232px_minmax(0,1fr)]">
        <aside
          class="hidden h-fit rounded-panel border border-outline bg-surface p-3 lg:sticky lg:top-20 lg:block"
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
            class="rounded-panel border border-outline bg-surface p-2.5"
          >
            <div class="space-y-2.5 lg:hidden">
              <div class="flex min-w-0 items-center gap-2">
                <SearchInput
                  v-model="searchQuery"
                  :placeholder="t('modelMarketplace.searchPlaceholder')"
                  class="min-w-0 flex-1"
                />
                <button
                  type="button"
                  class="btn btn-secondary relative h-10 shrink-0 px-3"
                  :aria-label="t('modelMarketplace.filters.open')"
                  @click="openMobileFilters"
                >
                  <Icon name="filter" size="sm" />
                  <span>{{ t('modelMarketplace.filters.title') }}</span>
                  <span
                    v-if="activeFilterCount"
                    class="flex h-4 min-w-4 items-center justify-center rounded-full bg-brand px-1 font-mono text-[9px] font-semibold text-white"
                  >
                    {{ activeFilterCount }}
                  </span>
                </button>
              </div>
              <div class="flex min-w-0 items-center justify-between gap-3">
                <span class="shrink-0 text-xs text-foreground-subtle" aria-live="polite">
                  <strong class="font-semibold text-foreground">{{ filteredEntries.length }}</strong>
                  {{ t('modelMarketplace.stats.models') }}
                </span>
                <div class="flex min-w-0 items-center gap-2">
                  <label class="sr-only" for="marketplace-sort-mobile">{{ t('modelMarketplace.sort.label') }}</label>
                  <select id="marketplace-sort-mobile" v-model="sortMode" class="input h-9 min-w-0 max-w-40 py-1 text-xs">
                    <option value="name">{{ t('modelMarketplace.sort.name') }}</option>
                    <option value="price">{{ t('modelMarketplace.sort.price') }}</option>
                  </select>
                  <button
                    type="button"
                    class="btn btn-secondary btn-icon h-9 w-9 shrink-0"
                    :disabled="loading"
                    :title="t('common.refresh')"
                    :aria-label="t('common.refresh')"
                    @click="loadMarketplace"
                  >
                    <Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" />
                  </button>
                </div>
              </div>
              <div v-if="activeFilterChips.length" class="-mx-1 flex gap-1.5 overflow-x-auto px-1 pb-0.5" :aria-label="t('modelMarketplace.filters.active')">
                <button
                  v-for="chip in activeFilterChips"
                  :key="chip.key"
                  type="button"
                  class="inline-flex h-7 shrink-0 items-center gap-1 rounded-control border border-outline bg-surface-subtle px-2 text-[10px] text-foreground-muted"
                  :aria-label="t('modelMarketplace.filters.remove', { label: chip.label })"
                  @click="clearFilterChip(chip.key)"
                >
                  <span class="max-w-32 truncate">{{ chip.label }}</span>
                  <Icon name="x" size="xs" />
                </button>
              </div>
            </div>

            <div class="hidden min-w-0 gap-3 lg:flex xl:items-center xl:justify-between">
              <div class="flex min-w-0 flex-1 items-center gap-3">
                <SearchInput
                  v-model="searchQuery"
                  :placeholder="t('modelMarketplace.searchPlaceholder')"
                  class="w-full min-w-0 max-w-lg flex-1"
                />
                <div class="flex shrink-0 items-center gap-2.5 text-xs text-foreground-subtle" aria-live="polite">
                  <span><strong class="font-semibold text-foreground">{{ filteredEntries.length }}</strong> {{ t('modelMarketplace.stats.models') }}</span>
                  <span class="h-3 w-px bg-outline" aria-hidden="true"></span>
                  <span><strong class="font-semibold text-foreground">{{ providers.length }}</strong> {{ t('modelMarketplace.stats.providers') }}</span>
                </div>
              </div>
              <div class="flex flex-nowrap items-center gap-2">
                <label class="flex min-h-10 items-center gap-2 whitespace-nowrap text-xs font-medium text-foreground-muted">
                  <Toggle v-model="showEffectivePrices" :aria-label="t('modelMarketplace.effectivePrice')" />
                  <span>{{ t('modelMarketplace.effectivePrice') }}</span>
                </label>
                <label class="sr-only" for="marketplace-sort">{{ t('modelMarketplace.sort.label') }}</label>
                <select id="marketplace-sort" v-model="sortMode" class="input h-10 w-44 flex-none text-sm">
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
          </div>

          <div v-if="loading && entries.length === 0" class="grid gap-3 pt-4 md:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4" role="status">
            <div v-for="index in 8" :key="index" class="card h-[230px] animate-pulse bg-surface-subtle"></div>
            <span class="sr-only">{{ t('common.loading') }}</span>
          </div>

          <div
            v-else-if="filteredEntries.length === 0"
            class="flex min-h-80 flex-col items-center justify-center border-b border-outline py-12 text-center"
          >
            <span class="flex h-12 w-12 items-center justify-center rounded-panel border border-outline bg-surface-subtle text-foreground-subtle">
              <Icon name="inbox" size="lg" />
            </span>
            <h2 class="mt-4 text-base font-semibold text-foreground">{{ t('modelMarketplace.empty.title') }}</h2>
            <p class="mt-1 max-w-md text-sm text-foreground-subtle">{{ t('modelMarketplace.empty.description') }}</p>
          </div>

          <div v-else class="grid gap-3 pt-4 md:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4">
            <article
              v-for="entry in visibleEntries"
              :key="entry.key"
              data-testid="marketplace-model-card"
              tabindex="0"
              class="card group/card flex min-h-[230px] min-w-0 cursor-pointer flex-col p-3.5 transition-[border-color,box-shadow] duration-150 hover:border-outline-strong hover:shadow-card focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus"
              :aria-label="t('modelMarketplace.details.open', { model: entry.name })"
              @click="openDetails(entry)"
              @keydown.enter.self.prevent="openDetails(entry)"
            >
              <div class="grid grid-cols-[36px_minmax(0,1fr)_28px] items-center gap-2.5">
                <span
                  class="flex h-9 w-9 shrink-0 items-center justify-center rounded-panel border"
                  :class="platformBadgeClass(entry.platform)"
                  aria-hidden="true"
                >
                  <PlatformIcon :platform="entry.platform as GroupPlatform" size="md" />
                </span>
                <div class="min-w-0 flex-1">
                  <div class="min-w-0">
                    <h2 class="line-clamp-2 min-w-0 break-words text-sm font-semibold leading-5 text-foreground">
                      {{ entry.name }}
                      <span
                        v-if="entry.label"
                        class="ml-1 inline-flex h-[18px] max-w-24 items-center align-middle truncate rounded-control bg-black px-2 text-[10px] font-bold leading-none text-white"
                        :title="entry.label"
                      >{{ entry.label }}</span>
                    </h2>
                  </div>
                  <div class="mt-1 flex flex-wrap items-center gap-1.5">
                    <span class="text-[10px] font-semibold uppercase text-foreground-muted">
                      {{ providerLabel(entry.platform) }}
                    </span>
                    <span class="h-3 w-px bg-outline"></span>
                    <span class="text-[10px] font-medium text-foreground-subtle">
                      {{ billingModeLabel(cardBillingCategory(entry)) }}
                    </span>
                    <template v-if="modelMonitorEnabled">
                      <span class="h-3 w-px bg-outline"></span>
                      <span class="inline-flex min-w-0 items-center gap-1 text-[9px] font-medium" :class="monitorTextClass(entry.monitorStatus?.status)">
                        <span class="h-1.5 w-1.5 shrink-0 rounded-full bg-current"></span>
                        <span class="truncate">{{ monitorStatusLabel(entry.monitorStatus?.status) }}</span>
                      </span>
                    </template>
                  </div>
                </div>
                <button
                  type="button"
                  class="btn-ghost btn-icon h-7 w-7 shrink-0 self-center"
                  :title="t('modelMarketplace.copyModel')"
                  :aria-label="t('modelMarketplace.copyModel')"
                  @click.stop="copyModel(entry.name)"
                >
                  <Icon name="copy" size="sm" />
                </button>
              </div>

              <div v-if="modelMonitorEnabled && hasMonitorTimeline(entry)" class="mt-3 border-t border-outline pt-2.5">
                <div class="flex items-center justify-between gap-3 text-xs">
                  <span class="inline-flex items-center gap-1.5 font-medium" :class="monitorTextClass(entry.monitorStatus?.status)">
                    <span class="h-1.5 w-1.5 rounded-full bg-current"></span>{{ monitorStatusLabel(entry.monitorStatus?.status) }}
                  </span>
                  <span class="font-mono tabular-nums text-foreground-subtle">{{ monitorAvailability(entry) }}</span>
                </div>
                <ModelMonitorTimeline class="mt-2" :points="entry.monitorStatus?.timeline" />
              </div>

              <div v-if="entry.pricing" class="mt-3 min-h-20 border-t border-outline pt-3">
                <template v-if="cardBillingCategory(entry) === 'usage'">
                  <dl class="grid grid-cols-2 divide-x divide-outline">
                    <div v-for="row in cardPrimaryPriceRows(entry)" :key="row.key" class="min-w-0 px-3 first:pl-0 last:pr-0">
                      <dt class="text-[10px] font-medium text-foreground-subtle">{{ row.label }}</dt>
                      <dd class="mt-0.5 flex items-baseline gap-1.5">
                        <span class="font-mono text-lg font-semibold tabular-nums text-foreground">{{ row.value }}</span>
                        <span class="text-[10px] text-foreground-subtle">/ 1M</span>
                        <span v-if="row.baseValue" class="font-mono text-[10px] tabular-nums text-foreground-subtle line-through">{{ row.baseValue }}</span>
                      </dd>
                    </div>
                  </dl>
                  <div v-if="cardCachePrice(entry)" class="mt-2 truncate text-[10px] text-foreground-muted">{{ cardCachePrice(entry) }}</div>
                </template>
                <template v-else>
                  <div class="flex items-end justify-between gap-3">
                    <div>
                      <div class="text-[10px] font-medium text-foreground-subtle">{{ t('modelMarketplace.price.request') }}</div>
                      <div class="mt-0.5 flex items-baseline gap-2">
                        <span class="font-mono text-xl font-semibold tabular-nums text-foreground">{{ cardRequestPrice(entry).value }}</span>
                        <span class="text-[10px] text-foreground-subtle">{{ t('modelMarketplace.price.perRequest') }}</span>
                      </div>
                    </div>
                    <span v-if="cardRequestPrice(entry).baseValue" class="font-mono text-xs tabular-nums text-foreground-subtle line-through">{{ cardRequestPrice(entry).baseValue }}</span>
                  </div>
                </template>
              </div>
              <div v-else class="mt-3 flex min-h-20 items-center border-t border-outline pt-3 text-xs text-foreground-subtle">
                {{ t('modelMarketplace.noPricing') }}
              </div>

              <div class="mt-auto border-t border-outline pt-2.5">
                <div class="grid min-h-11 grid-cols-[minmax(0,1fr)_96px] divide-x divide-outline">
                  <div class="min-w-0 pr-3">
                    <div class="text-[9px] font-medium text-foreground-subtle">{{ t('modelMarketplace.details.billingGroup') }}</div>
                    <div class="mt-1 flex min-w-0 items-center gap-1.5">
                      <span class="truncate text-xs font-medium text-foreground">{{ activeEntryGroup(entry)?.name ?? '-' }}</span>
                      <span class="shrink-0 font-mono text-[10px] tabular-nums text-foreground-muted">&times;{{ formatRate(effectiveRate(entry)) }}</span>
                    </div>
                  </div>
                  <div
                    class="group/rate relative cursor-help pl-3 text-right outline-none"
                    tabindex="0"
                    :aria-label="t('modelMarketplace.realtimeRateHint', { cny: formatRate(officialUsdToCnyRate), usd: formatRate(balanceRechargeMultiplier), group: formatRate(effectiveRate(entry)), rate: formatRate(cardRealtimeRate(entry)) })"
                  >
                    <div class="text-[9px] font-medium text-foreground-subtle">{{ t('modelMarketplace.realtimeRate') }}</div>
                    <div class="mt-0.5 inline-flex items-center gap-1 font-mono text-sm font-semibold tabular-nums text-foreground">
                      {{ formatRate(cardRealtimeRate(entry)) }}&times;
                      <Icon name="infoCircle" size="xs" class="text-foreground-subtle" />
                    </div>
                    <div class="marketplace-rate-tooltip" role="tooltip">
                      <div class="flex items-center justify-between gap-4 border-b border-outline pb-2">
                        <span class="text-xs font-semibold text-foreground">{{ t('modelMarketplace.realtimeRate') }}</span>
                        <span class="font-mono text-sm font-semibold tabular-nums text-foreground">{{ formatRate(cardRealtimeRate(entry)) }}&times;</span>
                      </div>
                      <dl class="mt-2 space-y-1.5 text-[10px]">
                        <div class="flex items-center justify-between gap-4">
                          <dt class="text-foreground-subtle">{{ t('modelMarketplace.realtimeRateOfficial') }}</dt>
                          <dd class="font-mono tabular-nums text-foreground">1 USD = {{ formatRate(officialUsdToCnyRate) }} CNY</dd>
                        </div>
                        <div class="flex items-center justify-between gap-4">
                          <dt class="text-foreground-subtle">{{ t('modelMarketplace.realtimeRateRecharge') }}</dt>
                          <dd class="font-mono tabular-nums text-foreground">1 CNY = {{ formatRate(balanceRechargeMultiplier) }} USD</dd>
                        </div>
                        <div class="flex items-center justify-between gap-4">
                          <dt class="text-foreground-subtle">{{ t('modelMarketplace.realtimeRateGroup') }}</dt>
                          <dd class="font-mono tabular-nums text-foreground">{{ formatRate(effectiveRate(entry)) }}&times;</dd>
                        </div>
                      </dl>
                      <div class="mt-2 border-t border-dashed border-outline pt-2 text-left font-mono text-[10px] tabular-nums text-foreground-muted">
                        {{ formatRate(effectiveRate(entry)) }} / {{ formatRate(balanceRechargeMultiplier) }} / {{ formatRate(officialUsdToCnyRate) }} = {{ formatRate(cardRealtimeRate(entry)) }}&times;
                      </div>
                    </div>
                  </div>
                </div>
                <div v-if="entry.groups.length > 1" class="mt-2 flex min-w-0 items-center gap-1.5 border-t border-dashed border-outline pt-2">
                  <button
                    v-for="group in cardGroups(entry)"
                    :key="group.id"
                    type="button"
                    class="inline-flex min-w-0 items-center gap-1 rounded-control border px-2 py-1 text-[10px] transition-colors"
                    :class="group.id === activeEntryGroup(entry)?.id ? 'border-brand/30 bg-brand-subtle text-brand' : 'border-outline bg-surface text-foreground-muted hover:border-outline-strong hover:text-foreground'"
                    :title="`${group.name} · ×${formatRate(group.effectiveRate)}`"
                    @click.stop="selectEntryGroup(entry, group.id)"
                  >
                    <span class="max-w-24 truncate">{{ group.name }}</span>
                    <span class="shrink-0 font-mono">×{{ formatRate(group.effectiveRate) }}</span>
                  </button>
                  <button
                    v-if="entry.groups.length > cardGroups(entry).length"
                    type="button"
                    class="inline-flex h-6 shrink-0 items-center rounded-control border border-outline px-2 text-[10px] font-medium text-foreground-subtle hover:border-outline-strong hover:text-foreground"
                    :aria-label="t('modelMarketplace.details.moreGroups', { count: entry.groups.length - cardGroups(entry).length })"
                    @click.stop="openDetails(entry)"
                  >
                    +{{ entry.groups.length - cardGroups(entry).length }}
                  </button>
                </div>
              </div>
            </article>
          </div>

          <div v-if="hasMore" ref="loadMoreSentinel" class="mt-5 flex min-h-12 items-center justify-center border-t border-outline pt-4">
            <button type="button" class="btn btn-secondary" @click="loadMore">{{ t('modelMarketplace.loadMore') }}</button>
          </div>
        </section>
      </div>
    </div>

    <ModelMarketplaceFilterDrawer
      :open="mobileFilterOpen"
      :providers="providers"
      :groups="groups"
      :billing-options="billingOptions"
      :entries-count="entries.length"
      :selected-provider="selectedProvider"
      :selected-group="selectedGroup"
      :selected-billing="selectedBilling"
      :show-effective-prices="showEffectivePrices"
      :result-count="mobileFilterResultCount"
      @close="mobileFilterOpen = false"
      @draft-change="updateMobileFilterDraft"
      @apply="applyMobileFilters"
    />

    <ModelMarketplaceDetailDrawer
      :entry="detailEntry"
      :groups="detailEntry ? sortedEntryGroups(detailEntry, userGroupRates) : []"
      :active-group="detailEntry ? activeEntryGroup(detailEntry) : null"
      :show-effective-prices="showEffectivePrices"
      :monitor-enabled="modelMonitorEnabled"
      @close="detailEntry = null"
      @select-group="detailEntry && selectEntryGroup(detailEntry, $event)"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import SearchInput from '@/components/common/SearchInput.vue'
import Toggle from '@/components/common/Toggle.vue'
import PlatformIcon from '@/components/common/PlatformIcon.vue'
import Icon from '@/components/icons/Icon.vue'
import ModelMonitorTimeline from '@/components/user/ModelMonitorTimeline.vue'
import ModelMarketplaceDetailDrawer from '@/components/user/ModelMarketplaceDetailDrawer.vue'
import ModelMarketplaceFilterDrawer from '@/components/user/ModelMarketplaceFilterDrawer.vue'
import userChannelsAPI from '@/api/channels'
import userGroupsAPI from '@/api/groups'
import type { GroupPlatform } from '@/types'
import { useAppStore } from '@/stores/app'
import { usePaymentStore } from '@/stores/payment'
import { useClipboard } from '@/composables/useClipboard'
import { extractApiErrorMessage } from '@/utils/apiError'
import { platformBadgeClass } from '@/utils/platformColors'
import {
  billingCategory,
  buildMarketplaceEntries,
  buildMarketplaceGroups,
  compareMarketplaceDisplayOrder,
  DEFAULT_USD_TO_CNY_RATE,
  effectiveRateForEntry,
  primaryPrice,
  realtimeRate,
  scaledPrice,
  sortedEntryGroups,
  type MarketplaceModelEntry,
} from './modelMarketplace'

const BATCH_SIZE = 18

const { t } = useI18n()
const appStore = useAppStore()
const paymentStore = usePaymentStore()
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
const visibleCount = ref(BATCH_SIZE)
const loadMoreSentinel = ref<HTMLElement | null>(null)
const selectedEntryGroups = ref<Record<string, number>>({})
const detailEntry = ref<MarketplaceModelEntry | null>(null)
const mobileFilterOpen = ref(false)
const draftProvider = ref('all')
const draftGroup = ref('all')
const draftBilling = ref('all')
let loadMoreObserver: IntersectionObserver | null = null

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
    const mode = billingCategory(entry.pricing)
    counts.set(mode, (counts.get(mode) ?? 0) + 1)
  }
  return [
    { value: 'all', label: t('modelMarketplace.filters.allBilling'), count: entries.value.length },
    ...[...counts.entries()].map(([value, count]) => ({ value, count, label: billingModeLabel(value) })),
  ]
})

const selectedGroupId = computed(() => selectedGroup.value === 'all' ? null : Number(selectedGroup.value))

const activeFilterCount = computed(() => [selectedProvider.value, selectedGroup.value, selectedBilling.value].filter(value => value !== 'all').length)
const activeFilterChips = computed(() => {
  const chips: Array<{ key: 'provider' | 'group' | 'billing'; label: string }> = []
  if (selectedProvider.value !== 'all') {
    chips.push({ key: 'provider', label: providers.value.find(provider => provider.value === selectedProvider.value)?.label ?? selectedProvider.value })
  }
  if (selectedGroup.value !== 'all') {
    chips.push({ key: 'group', label: groups.value.find(group => String(group.id) === selectedGroup.value)?.name ?? selectedGroup.value })
  }
  if (selectedBilling.value !== 'all') {
    chips.push({ key: 'billing', label: billingOptions.value.find(mode => mode.value === selectedBilling.value)?.label ?? selectedBilling.value })
  }
  return chips
})

const mobileFilterResultCount = computed(() => {
  const query = searchQuery.value.trim().toLowerCase()
  const groupId = draftGroup.value === 'all' ? null : Number(draftGroup.value)
  return entries.value.filter((entry) => {
    if (draftProvider.value !== 'all' && entry.platform !== draftProvider.value) return false
    if (groupId != null && !entry.groups.some(group => group.id === groupId)) return false
    if (draftBilling.value !== 'all' && billingCategory(entry.pricing) !== draftBilling.value) return false
    if (!query) return true
    return entry.name.toLowerCase().includes(query)
      || entry.groups.some(group => group.name.toLowerCase().includes(query))
  }).length
})

const filteredEntries = computed(() => {
  const query = searchQuery.value.trim().toLowerCase()
  const result = entries.value.filter((entry) => {
    if (selectedProvider.value !== 'all' && entry.platform !== selectedProvider.value) return false
    if (selectedGroupId.value != null && !entry.groups.some((group) => group.id === selectedGroupId.value)) return false
    const mode = billingCategory(entry.pricing)
    if (selectedBilling.value !== 'all' && mode !== selectedBilling.value) return false
    if (!query) return true
    return entry.name.toLowerCase().includes(query)
      || entry.groups.some((group) => group.name.toLowerCase().includes(query))
  })

  return result.sort((a, b) => {
    const byDisplayOrder = compareMarketplaceDisplayOrder(a, b)
    if (byDisplayOrder !== 0) return byDisplayOrder
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

const visibleEntries = computed(() => filteredEntries.value.slice(0, visibleCount.value))
const hasMore = computed(() => visibleCount.value < filteredEntries.value.length)

watch(
  [searchQuery, selectedProvider, selectedGroup, selectedBilling, sortMode],
  () => { visibleCount.value = BATCH_SIZE },
)
watch(loadMoreSentinel, (node, previous) => { if (previous) loadMoreObserver?.unobserve(previous); if (node) loadMoreObserver?.observe(node) }, { flush: 'post' })

const modelMonitorEnabled = computed(() => appStore.cachedPublicSettings?.model_monitor_enabled === true)
const balanceRechargeMultiplier = computed(() => {
  const multiplier = paymentStore.config?.balance_recharge_multiplier
  return Number.isFinite(multiplier) && Number(multiplier) > 0 ? Number(multiplier) : 1
})
const officialUsdToCnyRate = computed(() => {
  const rate = paymentStore.config?.subscription_usd_to_cny_rate
  return Number.isFinite(rate) && Number(rate) > 0 ? Number(rate) : DEFAULT_USD_TO_CNY_RATE
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
  return activeEntryGroup(entry)?.effectiveRate ?? effectiveRateForEntry(entry, selectedGroupId.value, userGroupRates.value)
}

function cardRealtimeRate(entry: MarketplaceModelEntry): number {
  return realtimeRate(effectiveRate(entry), balanceRechargeMultiplier.value, officialUsdToCnyRate.value)
}

function formatRate(rate: number): string {
  return Number(rate.toFixed(4)).toString()
}

function displayPrice(value: number | null, scale: number, entry: MarketplaceModelEntry): string {
  const rate = showEffectivePrices.value ? effectiveRate(entry) : 1
  return scaledPrice(value, scale, rate)
}

function cardBillingCategory(entry: MarketplaceModelEntry) {
  return billingCategory(entry.pricing)
}

function activeEntryGroup(entry: MarketplaceModelEntry) {
  const sorted = sortedEntryGroups(entry, userGroupRates.value)
  if (selectedGroupId.value != null) return sorted.find(group => group.id === selectedGroupId.value) ?? sorted[0] ?? null
  const selected = selectedEntryGroups.value[entry.key]
  return sorted.find(group => group.id === selected) ?? sorted[0] ?? null
}

function selectEntryGroup(entry: MarketplaceModelEntry, groupId: number) {
  selectedEntryGroups.value = { ...selectedEntryGroups.value, [entry.key]: groupId }
}

function cardGroups(entry: MarketplaceModelEntry) {
  const sorted = sortedEntryGroups(entry, userGroupRates.value)
  if (selectedGroupId.value != null) return sorted.filter(group => group.id === selectedGroupId.value)
  const active = activeEntryGroup(entry)
  return active ? [active, ...sorted.filter(group => group.id !== active.id)].slice(0, 2) : sorted.slice(0, 2)
}

function cardPrice(value: number | null, scale: number, entry: MarketplaceModelEntry) {
  const display = displayPrice(value, scale, entry)
  const rate = effectiveRate(entry)
  return {
    value: display,
    baseValue: showEffectivePrices.value && rate !== 1 ? scaledPrice(value, scale, 1) : '',
  }
}

function cardPrimaryPriceRows(entry: MarketplaceModelEntry) {
  const pricing = entry.pricing
  if (!pricing) return []
  return [
    { key: 'input', label: t('modelMarketplace.price.input'), ...cardPrice(pricing.input_price, 1_000_000, entry) },
    { key: 'output', label: t('modelMarketplace.price.output'), ...cardPrice(pricing.output_price, 1_000_000, entry) },
  ].filter(row => row.value !== '-')
}

function cardRequestPrice(entry: MarketplaceModelEntry) {
  const pricing = entry.pricing
  return cardPrice(pricing?.per_request_price ?? pricing?.image_output_price ?? null, 1, entry)
}

function cardCachePrice(entry: MarketplaceModelEntry) {
  const pricing = entry.pricing
  if (!pricing) return ''
  const parts = []
  if (pricing.cache_read_price != null) parts.push(`${t('modelMarketplace.price.cacheReadShort')} ${displayPrice(pricing.cache_read_price, 1_000_000, entry)} / 1M`)
  if (pricing.cache_write_price != null) parts.push(`${t('modelMarketplace.price.cacheWriteShort')} ${displayPrice(pricing.cache_write_price, 1_000_000, entry)} / 1M`)
  return parts.join(' · ')
}

function openDetails(entry: MarketplaceModelEntry) {
  detailEntry.value = entry
}

function loadMore() { visibleCount.value = Math.min(visibleCount.value + BATCH_SIZE, filteredEntries.value.length) }
function monitorStatusLabel(status?: string) { return status ? t(`modelMarketplace.monitor.${status}`) : t('modelMarketplace.monitor.unknown') }
function monitorTextClass(status?: string) { return status === 'operational' ? 'text-success-foreground' : status === 'degraded' ? 'text-warning-foreground' : status === 'failed' || status === 'error' ? 'text-danger-foreground' : 'text-foreground-subtle' }
function monitorAvailability(entry: MarketplaceModelEntry) { const value = entry.monitorStatus?.availability_7d; return value == null ? t('modelMarketplace.monitor.noData') : t('modelMarketplace.monitor.availability', { value: value.toFixed(2) }) }
function hasMonitorTimeline(entry: MarketplaceModelEntry) { return (entry.monitorStatus?.timeline?.length ?? 0) > 0 }

function resetFilters() {
  searchQuery.value = ''
  selectedProvider.value = 'all'
  selectedGroup.value = 'all'
  selectedBilling.value = 'all'
  sortMode.value = 'name'
}

function openMobileFilters() {
  draftProvider.value = selectedProvider.value
  draftGroup.value = selectedGroup.value
  draftBilling.value = selectedBilling.value
  mobileFilterOpen.value = true
}

function updateMobileFilterDraft(filters: { provider: string; group: string; billing: string }) {
  draftProvider.value = filters.provider
  draftGroup.value = filters.group
  draftBilling.value = filters.billing
}

function applyMobileFilters(filters: { provider: string; group: string; billing: string; showEffectivePrices: boolean }) {
  selectedProvider.value = filters.provider
  selectedGroup.value = filters.group
  selectedBilling.value = filters.billing
  showEffectivePrices.value = filters.showEffectivePrices
  mobileFilterOpen.value = false
}

function clearFilterChip(key: 'provider' | 'group' | 'billing') {
  if (key === 'provider') selectedProvider.value = 'all'
  if (key === 'group') selectedGroup.value = 'all'
  if (key === 'billing') selectedBilling.value = 'all'
}

async function copyModel(name: string) {
  await copyToClipboard(name, t('modelMarketplace.copySuccess', { name }))
}

async function loadMarketplace() {
  loading.value = true
  visibleCount.value = BATCH_SIZE
  loadError.value = false
  try {
    const [catalogResponse, rates] = await Promise.all([
      userChannelsAPI.getMarketplace(),
      userGroupsAPI.getUserGroupRates().catch(() => ({} as Record<number, number>)),
      paymentStore.fetchConfig(true),
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

onMounted(() => {
  loadMoreObserver = typeof IntersectionObserver === 'undefined' ? null : new IntersectionObserver(entries => { if (entries.some(entry => entry.isIntersecting)) loadMore() }, { rootMargin: '240px' })
  if (loadMoreSentinel.value) loadMoreObserver?.observe(loadMoreSentinel.value)
  void loadMarketplace()
})
onBeforeUnmount(() => loadMoreObserver?.disconnect())
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

.marketplace-rate-tooltip {
  position: absolute;
  z-index: 30;
  right: -4px;
  bottom: calc(100% + 10px);
  width: min(284px, calc(100vw - 32px));
  visibility: hidden;
  border: 1px solid rgb(var(--color-border-strong));
  border-radius: var(--radius-md);
  background: rgb(var(--color-surface-raised));
  box-shadow: var(--shadow-floating);
  padding: 11px 12px;
  opacity: 0;
  transform: translateY(4px);
  transition: opacity 140ms ease, transform 140ms ease, visibility 140ms ease;
}

.marketplace-rate-tooltip::after {
  position: absolute;
  right: 18px;
  bottom: -5px;
  width: 9px;
  height: 9px;
  border-right: 1px solid rgb(var(--color-border-strong));
  border-bottom: 1px solid rgb(var(--color-border-strong));
  background: rgb(var(--color-surface-raised));
  content: '';
  transform: rotate(45deg);
}

.group\/rate:hover .marketplace-rate-tooltip,
.group\/rate:focus .marketplace-rate-tooltip,
.group\/rate:focus-within .marketplace-rate-tooltip {
  visibility: visible;
  opacity: 1;
  transform: translateY(0);
}
</style>
