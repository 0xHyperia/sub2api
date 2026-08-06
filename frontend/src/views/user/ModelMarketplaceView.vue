<template>
  <AppLayout>
    <div class="space-y-4">
      <div
        v-if="loadError"
        class="flex flex-col gap-3 rounded-panel border border-danger/30 bg-danger-subtle p-4 text-sm text-danger-foreground sm:flex-row sm:items-center sm:justify-between"
        role="alert"
      >
        <span>{{ t('modelMarketplace.loadError') }}</span>
        <button type="button" class="btn btn-secondary btn-sm" @click="loadMarketplace()">
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
                  <PlatformIcon :platform="provider.value as GroupPlatform" size="sm" :class="platformIconClass(provider.value)" />
                  <span class="truncate">{{ provider.label }}</span>
                </span>
                <span class="marketplace-filter-count">{{ provider.count }}</span>
              </button>
            </div>
          </section>

          <section class="marketplace-filter-section" :aria-labelledby="'marketplace-capability-filter'">
            <h3 id="marketplace-capability-filter" class="text-xs font-semibold uppercase text-foreground-subtle">
              {{ t('modelMarketplace.filters.capability') }}
            </h3>
            <div
              data-testid="marketplace-capability-list"
              class="mt-2 max-h-52 space-y-1 overflow-y-auto pr-1"
            >
              <button
                v-for="capability in capabilityOptions"
                :key="capability.value"
                type="button"
                class="marketplace-filter-option"
                :class="selectedCapability === capability.value ? 'marketplace-filter-option-active' : ''"
                :aria-pressed="selectedCapability === capability.value"
                @click="selectedCapability = capability.value"
              >
                <span class="flex min-w-0 items-center gap-1.5">
                  <Icon v-if="capability.icon" :name="capability.icon" size="xs" class="shrink-0" />
                  <span class="truncate">{{ capability.label }}</span>
                </span>
                <span class="marketplace-filter-count">{{ capability.count }}</span>
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
                    class="flex h-4 min-w-4 items-center justify-center rounded-full bg-brand px-1 font-mono text-[9px] font-semibold text-brand-foreground"
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
                    @click="loadMarketplace()"
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
                  @click="loadMarketplace()"
                >
                  <Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" />
                </button>
              </div>
            </div>
          </div>

          <div v-if="loading && entries.length === 0" class="grid gap-3 pt-4 md:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4" role="status">
            <div v-for="index in 8" :key="index" class="card h-[152px] animate-pulse bg-surface-subtle"></div>
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
              class="card group/card flex min-h-[152px] min-w-0 cursor-pointer flex-col p-3 transition-[border-color,box-shadow] duration-150 hover:border-outline-strong hover:shadow-card focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus"
              :aria-label="t('modelMarketplace.details.open', { model: entry.name })"
              @click="openDetails(entry)"
              @keydown.enter.self.prevent="openDetails(entry)"
            >
              <div class="grid grid-cols-[24px_minmax(0,1fr)_52px] items-center gap-2">
                <span class="flex h-6 w-6 shrink-0 items-center justify-center text-foreground" aria-hidden="true">
                  <PlatformIcon :platform="entry.platform as GroupPlatform" size="lg" :class="platformIconClass(entry.platform)" />
                </span>
                <div class="min-w-0 flex-1">
                  <div class="flex min-w-0 items-center gap-1.5">
                    <h2 class="truncate text-sm font-semibold leading-5 text-foreground">
                      {{ entry.name }}
                    </h2>
                    <span
                      v-if="entry.label"
                      class="inline-flex h-4 max-w-20 shrink-0 items-center truncate rounded-control bg-black px-1.5 text-[9px] font-bold leading-none text-white"
                      :title="entry.label"
                    >{{ entry.label }}</span>
                  </div>
                </div>
                <div class="flex items-center justify-end gap-1">
                  <button
                    type="button"
                    class="inline-flex h-6 w-6 shrink-0 items-center justify-center text-foreground-subtle transition-[color,transform] duration-150 hover:-translate-y-px hover:text-foreground focus-visible:-translate-y-px focus-visible:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus"
                    :title="t('modelMarketplace.copyModel')"
                    :aria-label="t('modelMarketplace.copyModel')"
                    @click.stop="copyModel(entry.name)"
                  >
                    <Icon name="copy" size="xs" />
                  </button>
                  <button
                    type="button"
                    class="inline-flex h-6 w-6 shrink-0 items-center justify-center text-foreground-subtle transition-[color,transform] duration-150 hover:-translate-y-px hover:text-foreground focus-visible:-translate-y-px focus-visible:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus"
                    :title="t('modelMarketplace.details.open', { model: entry.name })"
                    :aria-label="t('modelMarketplace.details.open', { model: entry.name })"
                    @click.stop="openDetails(entry)"
                  >
                    <Icon name="eye" size="xs" />
                  </button>
                </div>
              </div>

              <div class="mt-1.5 flex min-h-5 items-center justify-between gap-2">
                <div class="flex min-w-0 items-center gap-2 text-foreground-subtle">
                  <span class="max-w-20 truncate text-[9px] font-medium uppercase" :title="providerLabel(entry.platform)">
                    {{ providerLabel(entry.platform) }}
                  </span>
                  <span class="shrink-0 text-[9px]">{{ billingModeLabel(entry.pricing?.billing_mode) }}</span>
                  <span class="h-3 w-px shrink-0 bg-outline" aria-hidden="true"></span>
                  <span class="flex min-w-0 items-center gap-1.5" :aria-label="t('modelMarketplace.capabilities.label')">
                  <span
                    v-for="capability in visibleCardCapabilityBadges(entry)"
                    :key="capability.key"
                    class="group/capability relative inline-flex h-5 w-5 shrink-0 cursor-help items-center justify-center text-foreground-subtle transition-[color,transform] duration-150 hover:-translate-y-px hover:text-foreground focus-visible:-translate-y-px focus-visible:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus"
                    tabindex="0"
                    :aria-label="capability.label"
                    :aria-describedby="`capability-${entry.key}-${capability.key}`"
                  >
                    <Icon :name="capability.icon" size="xs" aria-hidden="true" />
                    <span
                      :id="`capability-${entry.key}-${capability.key}`"
                      class="marketplace-capability-tooltip"
                      role="tooltip"
                    >{{ capability.label }}</span>
                  </span>
                  <span
                    v-if="hiddenCardCapabilityBadges(entry).length"
                    class="group/capability relative inline-flex h-5 min-w-5 shrink-0 cursor-help items-center justify-center px-0.5 text-[9px] font-semibold text-foreground-subtle outline-none transition-colors hover:text-foreground focus-visible:text-foreground focus-visible:ring-2 focus-visible:ring-focus"
                    tabindex="0"
                    :aria-label="hiddenCapabilityLabel(entry)"
                    :aria-describedby="`capability-${entry.key}-more`"
                    @click.stop
                  >
                    +{{ hiddenCardCapabilityBadges(entry).length }}
                    <span
                      :id="`capability-${entry.key}-more`"
                      class="marketplace-capability-tooltip marketplace-capability-tooltip-list"
                      role="tooltip"
                    >
                      <span
                        v-for="capability in hiddenCardCapabilityBadges(entry)"
                        :key="capability.key"
                        class="flex items-center gap-1.5"
                      >
                        <Icon :name="capability.icon" size="xs" aria-hidden="true" />
                        <span>{{ capability.label }}</span>
                      </span>
                    </span>
                  </span>
                  </span>
                </div>
                <div
                  v-if="entry.monitorStatus"
                  class="inline-flex shrink-0 items-end gap-2"
                  :title="monitorCompactLabel(entry)"
                  :aria-label="monitorCompactLabel(entry)"
                >
                  <span class="flex h-3 items-end gap-[3px]" aria-hidden="true">
                    <span
                      v-for="(status, index) in monitorSignalPoints(entry)"
                      :key="index"
                      class="marketplace-status-bar w-1 rounded-full"
                      :class="monitorSignalClass(status)"
                    ></span>
                  </span>
                </div>
              </div>

              <div v-if="entry.pricing" class="mt-1 flex min-h-6 items-center">
                <template v-if="cardBillingCategory(entry) === 'usage'">
                  <dl class="flex min-w-0 items-baseline divide-x divide-outline overflow-hidden">
                    <div v-for="row in cardPrimaryPriceRows(entry)" :key="row.key" class="flex min-w-0 items-baseline gap-1.5 px-2 first:pl-0 last:pr-0">
                      <dt class="shrink-0 text-[9px] font-medium text-foreground-subtle">{{ row.label }}</dt>
                      <dd class="flex min-w-0 items-baseline gap-1">
                        <span class="font-mono text-sm font-semibold tabular-nums text-foreground">{{ row.value }}</span>
                        <span v-if="row.baseValue" class="truncate font-mono text-[9px] tabular-nums text-foreground-subtle line-through">{{ row.baseValue }}</span>
                      </dd>
                    </div>
                    <span class="shrink-0 pl-2 text-[9px] text-foreground-subtle">/ 1M</span>
                  </dl>
                </template>
                <template v-else-if="entry.pricing.billing_mode === 'image'">
                  <dl class="flex min-w-0 items-baseline divide-x divide-outline overflow-hidden">
                    <div v-for="row in cardImagePriceRows(entry)" :key="row.key" class="flex min-w-0 items-baseline gap-1.5 px-2 first:pl-0 last:pr-0">
                      <dt class="shrink-0 text-[9px] font-medium text-foreground-subtle">{{ row.tier }}</dt>
                      <dd class="flex min-w-0 items-baseline gap-1">
                        <span class="font-mono text-sm font-semibold tabular-nums text-foreground">{{ row.value }}</span>
                        <span v-if="row.baseValue" class="truncate font-mono text-[9px] tabular-nums text-foreground-subtle line-through">{{ row.baseValue }}</span>
                      </dd>
                    </div>
                    <span class="shrink-0 pl-2 text-[9px] text-foreground-subtle">{{ t('modelMarketplace.price.perImage') }}</span>
                  </dl>
                </template>
                <template v-else>
                  <div class="flex items-baseline gap-1.5">
                    <span class="text-[9px] font-medium text-foreground-subtle">{{ t('modelMarketplace.price.request') }}</span>
                    <span class="font-mono text-sm font-semibold tabular-nums text-foreground">{{ cardRequestPrice(entry).value }}</span>
                    <span v-if="cardRequestPrice(entry).baseValue" class="font-mono text-[9px] tabular-nums text-foreground-subtle line-through">{{ cardRequestPrice(entry).baseValue }}</span>
                    <span class="text-[9px] text-foreground-subtle">{{ t('modelMarketplace.price.perRequest') }}</span>
                  </div>
                </template>
              </div>
              <div v-else class="mt-1 flex min-h-6 items-center text-[10px] text-foreground-subtle">
                {{ t('modelMarketplace.noPricing') }}
              </div>

              <div class="mt-auto flex min-h-8 items-center gap-2 border-t border-outline pt-2">
                <ModelMarketplaceCardGroups
                  :groups="cardGroups(entry)"
                  :active-group-id="activeEntryGroup(entry)?.id"
                  @select="selectEntryGroup(entry, $event)"
                />
                <div
                  class="group/rate relative flex shrink-0 cursor-help items-center gap-1.5 outline-none"
                  tabindex="0"
                  :aria-label="t('modelMarketplace.realtimeRateHint', { cny: formatExchangeRate(officialUsdToCnyRate), usd: formatRate(balanceRechargeMultiplier), group: formatRate(effectiveRate(entry)), rate: formatRate(cardRealtimeRate(entry)) })"
                >
                  <span class="text-[9px] font-medium text-warning-foreground">{{ t('modelMarketplace.realtimeRate') }}</span>
                  <span class="inline-flex items-center gap-1 font-mono text-xs font-semibold tabular-nums text-warning-foreground">
                    {{ formatRate(cardRealtimeRate(entry)) }}&times;
                    <Icon name="infoCircle" size="xs" />
                  </span>
                  <div class="marketplace-rate-tooltip" role="tooltip">
                      <div class="flex items-center justify-between gap-4 border-b border-outline pb-2">
                        <span class="text-xs font-semibold text-foreground">{{ t('modelMarketplace.realtimeRate') }}</span>
                        <span class="font-mono text-sm font-semibold tabular-nums text-warning-foreground">{{ formatRate(cardRealtimeRate(entry)) }}&times;</span>
                      </div>
                      <dl class="mt-2 space-y-1.5 text-[10px]">
                        <div class="flex items-center justify-between gap-4">
                          <dt class="text-foreground-subtle">{{ t('modelMarketplace.realtimeRateOfficial') }}</dt>
                          <dd class="font-mono tabular-nums text-foreground">1 USD = {{ formatExchangeRate(officialUsdToCnyRate) }} CNY</dd>
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
                        {{ formatRate(effectiveRate(entry)) }} / {{ formatRate(balanceRechargeMultiplier) }} / {{ formatExchangeRate(officialUsdToCnyRate) }} = {{ formatRate(cardRealtimeRate(entry)) }}&times;
                      </div>
                  </div>
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
      :capability-options="capabilityOptions"
      :entries-count="entries.length"
      :selected-provider="selectedProvider"
      :selected-group="selectedGroup"
      :selected-billing="selectedBilling"
      :selected-capability="selectedCapability"
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
      :monitor-resolution="monitorResolution"
      :performance-loading="resolutionLoading"
      :show-detailed-performance="appStore.cachedPublicSettings?.model_marketplace_performance_visible !== false"
      @close="detailEntry = null"
      @select-group="detailEntry && selectEntryGroup(detailEntry, $event)"
      @update:monitor-resolution="setMonitorResolution"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { formatExchangeRate } from '@/utils/currency'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import SearchInput from '@/components/common/SearchInput.vue'
import Toggle from '@/components/common/Toggle.vue'
import PlatformIcon from '@/components/common/PlatformIcon.vue'
import Icon from '@/components/icons/Icon.vue'
import ModelMarketplaceCardGroups from '@/components/user/ModelMarketplaceCardGroups.vue'
import ModelMarketplaceDetailDrawer from '@/components/user/ModelMarketplaceDetailDrawer.vue'
import ModelMarketplaceFilterDrawer from '@/components/user/ModelMarketplaceFilterDrawer.vue'
import userChannelsAPI from '@/api/channels'
import userGroupsAPI from '@/api/groups'
import type { GroupPlatform } from '@/types'
import { useAppStore } from '@/stores/app'
import { usePaymentStore } from '@/stores/payment'
import { useClipboard } from '@/composables/useClipboard'
import { extractApiErrorMessage } from '@/utils/apiError'
import { platformIconClass } from '@/utils/platformColors'
import {
  billingCategory,
  buildMarketplaceEntries,
  buildMarketplaceGroups,
  compareMarketplaceDisplayOrder,
  compareMarketplaceModelRecency,
  compareMarketplaceProviders,
  DEFAULT_USD_TO_CNY_RATE,
  effectiveRateForEntry,
  imagePriceRows,
  inferMarketplaceModelCapabilities,
  MARKETPLACE_CARD_CAPABILITY_ORDER,
  MARKETPLACE_MODEL_CAPABILITIES,
  modelAvailabilityBarClass,
  primaryPrice,
  recentModelSuccessRates,
  realtimeRate,
  scaledPrice,
  sortedEntryGroups,
  type MarketplaceModelEntry,
  type MarketplaceModelCapability,
} from './modelMarketplace'

const BATCH_SIZE = 18
const CARD_CAPABILITY_LIMIT = 5
const marketplaceCapabilityIcons = {
  vision: 'eye',
  image_input: 'image',
  audio_input: 'microphone',
  video_input: 'video',
  function_calling: 'cog',
  reasoning: 'brain',
  prompt_caching: 'database',
  web_search: 'globe',
  pdf_input: 'document',
  computer_use: 'terminal',
  image_generation: 'sparkles',
  audio_output: 'speaker',
  parallel_tools: 'arrowsUpDown',
  tool_choice: 'sort',
  structured_output: 'codeBracket',
  assistant_prefill: 'edit',
  streaming: 'signal',
  system_messages: 'chatBubble',
  url_context: 'link',
  image_embedding: 'cube',
  service_tier: 'badge',
} as const

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
const selectedCapability = ref('all')
const sortMode = ref<'name' | 'price'>('name')
const showEffectivePrices = ref(true)
const visibleCount = ref(BATCH_SIZE)
const loadMoreSentinel = ref<HTMLElement | null>(null)
const selectedEntryGroups = ref<Record<string, number>>({})
const detailEntry = ref<MarketplaceModelEntry | null>(null)
const monitorResolution = ref<'minute' | 'hour'>(localStorage.getItem('usa0:model-marketplace-resolution') === 'minute' ? 'minute' : 'hour')
const marketplaceRefreshing = ref(false)
const resolutionLoading = ref(false)
const mobileFilterOpen = ref(false)
const draftProvider = ref('all')
const draftGroup = ref('all')
const draftBilling = ref('all')
const draftCapability = ref('all')
let loadMoreObserver: IntersectionObserver | null = null
let marketplaceRefreshTimer: number | null = null

const entries = computed(() => buildMarketplaceEntries(catalog.value))
const groups = computed(() => buildMarketplaceGroups(entries.value, userGroupRates.value))

const providers = computed(() => {
  const counts = new Map<string, number>()
  for (const entry of entries.value) counts.set(entry.platform, (counts.get(entry.platform) ?? 0) + 1)
  return [...counts.entries()]
    .map(([value, count]) => ({ value, count, label: providerLabel(value) }))
    .sort((a, b) => compareMarketplaceProviders(a.value, b.value) || a.label.localeCompare(b.label))
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

const capabilityOptions = computed(() => {
  const counts = new Map<MarketplaceModelCapability, number>()
  for (const entry of entries.value) {
    for (const capability of inferMarketplaceModelCapabilities(entry)) {
      counts.set(capability, (counts.get(capability) ?? 0) + 1)
    }
  }
  return [
    { value: 'all', label: t('modelMarketplace.filters.allCapabilities'), count: entries.value.length, icon: null },
    ...MARKETPLACE_MODEL_CAPABILITIES.map(capability => ({
        value: capability,
        label: t(`modelMarketplace.capabilities.${capability}`),
        count: counts.get(capability) ?? 0,
        icon: marketplaceCapabilityIcons[capability],
      })),
  ]
})

const selectedGroupId = computed(() => selectedGroup.value === 'all' ? null : Number(selectedGroup.value))

const activeFilterCount = computed(() => [selectedProvider.value, selectedGroup.value, selectedBilling.value, selectedCapability.value].filter(value => value !== 'all').length)
const activeFilterChips = computed(() => {
  const chips: Array<{ key: 'provider' | 'group' | 'billing' | 'capability'; label: string }> = []
  if (selectedProvider.value !== 'all') {
    chips.push({ key: 'provider', label: providers.value.find(provider => provider.value === selectedProvider.value)?.label ?? selectedProvider.value })
  }
  if (selectedGroup.value !== 'all') {
    chips.push({ key: 'group', label: groups.value.find(group => String(group.id) === selectedGroup.value)?.name ?? selectedGroup.value })
  }
  if (selectedBilling.value !== 'all') {
    chips.push({ key: 'billing', label: billingOptions.value.find(mode => mode.value === selectedBilling.value)?.label ?? selectedBilling.value })
  }
  if (selectedCapability.value !== 'all') {
    chips.push({ key: 'capability', label: capabilityOptions.value.find(capability => capability.value === selectedCapability.value)?.label ?? selectedCapability.value })
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
    if (draftCapability.value !== 'all' && !inferMarketplaceModelCapabilities(entry).includes(draftCapability.value as MarketplaceModelCapability)) return false
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
    if (selectedCapability.value !== 'all' && !inferMarketplaceModelCapabilities(entry).includes(selectedCapability.value as MarketplaceModelCapability)) return false
    if (!query) return true
    return entry.name.toLowerCase().includes(query)
      || entry.groups.some((group) => group.name.toLowerCase().includes(query))
  })

  return result.sort((a, b) => {
    const byProvider = compareMarketplaceProviders(a.platform, b.platform)
    if (byProvider !== 0) return byProvider
    const byRecency = compareMarketplaceModelRecency(a.name, b.name)
    if (byRecency !== 0) return byRecency
    const byDisplayOrder = compareMarketplaceDisplayOrder(a, b)
    if (byDisplayOrder !== 0) return byDisplayOrder
    if (sortMode.value === 'price') {
      const aValue = entrySortPrice(a)
      const bValue = entrySortPrice(b)
      if (aValue !== bValue) return aValue - bValue
    }
    return a.name.localeCompare(b.name)
  })
})

const visibleEntries = computed(() => filteredEntries.value.slice(0, visibleCount.value))
const hasMore = computed(() => visibleCount.value < filteredEntries.value.length)

watch(
  [searchQuery, selectedProvider, selectedGroup, selectedBilling, selectedCapability, sortMode],
  () => { visibleCount.value = BATCH_SIZE },
)
watch(loadMoreSentinel, (node, previous) => { if (previous) loadMoreObserver?.unobserve(previous); if (node) loadMoreObserver?.observe(node) }, { flush: 'post' })

const balanceRechargeMultiplier = computed(() => {
  const multiplier = paymentStore.config?.balance_recharge_multiplier
  return Number.isFinite(multiplier) && Number(multiplier) > 0 ? Number(multiplier) : 1
})
const officialUsdToCnyRate = computed(() => {
  const rate = appStore.cachedPublicSettings?.currency_usd_to_cny_rate
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

function cardCapabilityBadges(entry: MarketplaceModelEntry) {
  const capabilities = new Set(inferMarketplaceModelCapabilities(entry))
  return MARKETPLACE_CARD_CAPABILITY_ORDER.filter(capability => capabilities.has(capability)).map(capability => ({
    key: capability,
    icon: marketplaceCapabilityIcons[capability],
    label: t(`modelMarketplace.capabilities.${capability}`),
  }))
}

function visibleCardCapabilityBadges(entry: MarketplaceModelEntry) {
  return cardCapabilityBadges(entry).slice(0, CARD_CAPABILITY_LIMIT)
}

function hiddenCardCapabilityBadges(entry: MarketplaceModelEntry) {
  return cardCapabilityBadges(entry).slice(CARD_CAPABILITY_LIMIT)
}

function hiddenCapabilityLabel(entry: MarketplaceModelEntry) {
  return hiddenCardCapabilityBadges(entry).map(capability => capability.label).join(', ')
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
  return sorted
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

function cardImagePriceRows(entry: MarketplaceModelEntry) {
  const group = activeEntryGroup(entry)
  return imagePriceRows(entry.pricing, group).map(row => ({
    ...row,
    value: scaledPrice(showEffectivePrices.value ? row.effectiveValue : row.rawValue, 1, 1),
    baseValue: showEffectivePrices.value && group && group.effectiveRate !== 1
      ? scaledPrice(row.rawValue, 1, 1)
      : '',
  }))
}

function entrySortPrice(entry: MarketplaceModelEntry): number {
  if (entry.pricing?.billing_mode === 'image') {
    return imagePriceRows(entry.pricing, activeEntryGroup(entry))[0]?.effectiveValue ?? Number.POSITIVE_INFINITY
  }
  const price = primaryPrice(entry.pricing)
  return price == null ? Number.POSITIVE_INFINITY : price * effectiveRate(entry)
}

function cardRequestPrice(entry: MarketplaceModelEntry) {
  const pricing = entry.pricing
  return cardPrice(pricing?.per_request_price ?? pricing?.image_output_price ?? null, 1, entry)
}

function openDetails(entry: MarketplaceModelEntry) {
  detailEntry.value = entry
}

function loadMore() { visibleCount.value = Math.min(visibleCount.value + BATCH_SIZE, filteredEntries.value.length) }
function monitorStatusLabel(rate?: number | null) {
  if (rate == null || !Number.isFinite(rate)) return t('modelMarketplace.monitor.unknown')
  if (rate >= 70) return t('modelMarketplace.monitor.operational')
  if (rate < 70) return t('modelMarketplace.monitor.failed')
  return t('modelMarketplace.monitor.unknown')
}
function monitorSignalPoints(entry: MarketplaceModelEntry) { return recentModelSuccessRates(entry.monitorStatus?.hourly_metrics?.buckets, 3) }
function monitorSignalClass(rate?: number | null) {
  return modelAvailabilityBarClass(rate)
}
function monitorCompactLabel(entry: MarketplaceModelEntry) {
  const rate = entry.monitorStatus?.hourly_metrics?.success_rate
  const parts = [monitorStatusLabel(rate)]
  if (rate != null) parts.push(t('modelMarketplace.monitor.availability', { value: rate.toFixed(2) }))
  return parts.join(', ')
}

function resetFilters() {
  searchQuery.value = ''
  selectedProvider.value = 'all'
  selectedGroup.value = 'all'
  selectedBilling.value = 'all'
  selectedCapability.value = 'all'
  sortMode.value = 'name'
}

function openMobileFilters() {
  draftProvider.value = selectedProvider.value
  draftGroup.value = selectedGroup.value
  draftBilling.value = selectedBilling.value
  draftCapability.value = selectedCapability.value
  mobileFilterOpen.value = true
}

function updateMobileFilterDraft(filters: { provider: string; group: string; billing: string; capability: string }) {
  draftProvider.value = filters.provider
  draftGroup.value = filters.group
  draftBilling.value = filters.billing
  draftCapability.value = filters.capability
}

function applyMobileFilters(filters: { provider: string; group: string; billing: string; capability: string; showEffectivePrices: boolean }) {
  selectedProvider.value = filters.provider
  selectedGroup.value = filters.group
  selectedBilling.value = filters.billing
  selectedCapability.value = filters.capability
  showEffectivePrices.value = filters.showEffectivePrices
  mobileFilterOpen.value = false
}

function clearFilterChip(key: 'provider' | 'group' | 'billing' | 'capability') {
  if (key === 'provider') selectedProvider.value = 'all'
  if (key === 'group') selectedGroup.value = 'all'
  if (key === 'billing') selectedBilling.value = 'all'
  if (key === 'capability') selectedCapability.value = 'all'
}

async function copyModel(name: string) {
  await copyToClipboard(name, t('modelMarketplace.copySuccess', { name }))
}

async function loadMarketplace(silent = false) {
  if (loading.value || marketplaceRefreshing.value) return false
  const selectedKey = detailEntry.value?.key
  silent ? marketplaceRefreshing.value = true : loading.value = true
  if (!silent) visibleCount.value = BATCH_SIZE
  loadError.value = false
  try {
    const [catalogResponse, rates] = await Promise.all([
      userChannelsAPI.getMarketplace({ resolution: monitorResolution.value }),
      userGroupsAPI.getUserGroupRates().catch(() => ({} as Record<number, number>)),
      paymentStore.fetchConfig(true),
      appStore.fetchPublicSettings(true),
    ])
    catalog.value = catalogResponse
    userGroupRates.value = rates
    if (selectedKey) detailEntry.value = entries.value.find(entry => entry.key === selectedKey) ?? null
    return true
  } catch (error) {
    if (!silent) {
      loadError.value = true
      appStore.showError(extractApiErrorMessage(error, t('modelMarketplace.loadError')))
    }
    return false
  } finally {
    loading.value = false
    marketplaceRefreshing.value = false
  }
}

async function setMonitorResolution(value: 'minute' | 'hour') {
  if (monitorResolution.value === value) return
  const previous = monitorResolution.value
  monitorResolution.value = value
  resolutionLoading.value = true
  try {
    const loaded = await loadMarketplace(true)
    if (!loaded) {
      monitorResolution.value = previous
      return
    }
    localStorage.setItem('usa0:model-marketplace-resolution', value)
  } finally {
    resolutionLoading.value = false
  }
}

onMounted(() => {
  loadMoreObserver = typeof IntersectionObserver === 'undefined' ? null : new IntersectionObserver(entries => { if (entries.some(entry => entry.isIntersecting)) loadMore() }, { rootMargin: '240px' })
  if (loadMoreSentinel.value) loadMoreObserver?.observe(loadMoreSentinel.value)
  void loadMarketplace()
  marketplaceRefreshTimer = window.setInterval(() => {
    if (document.visibilityState === 'visible') void loadMarketplace(true)
  }, 30_000)
})
onBeforeUnmount(() => {
  loadMoreObserver?.disconnect()
  if (marketplaceRefreshTimer != null) window.clearInterval(marketplaceRefreshTimer)
})
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

.marketplace-capability-tooltip {
  position: absolute;
  z-index: 30;
  bottom: calc(100% + 7px);
  left: 50%;
  visibility: hidden;
  border-radius: var(--radius-xs);
  background: #000;
  box-shadow: var(--shadow-floating);
  padding: 4px 7px;
  color: #fff;
  font-size: 10px;
  font-weight: 600;
  line-height: 1;
  white-space: nowrap;
  opacity: 0;
  pointer-events: none;
  transform: translate(-50%, 4px);
  transition: opacity 120ms ease, transform 120ms ease, visibility 120ms ease;
}

.marketplace-capability-tooltip::after {
  position: absolute;
  top: 100%;
  left: 50%;
  border: 4px solid transparent;
  border-top-color: #000;
  content: '';
  transform: translateX(-50%);
}

.marketplace-capability-tooltip-list {
  left: auto;
  right: -6px;
  display: grid;
  min-width: max-content;
  gap: 6px;
  padding: 8px 9px;
  line-height: 1.2;
  transform: translateY(4px);
}

.marketplace-capability-tooltip-list::after {
  right: 10px;
  left: auto;
  transform: none;
}

.group\/capability:hover .marketplace-capability-tooltip,
.group\/capability:focus .marketplace-capability-tooltip,
.group\/capability:focus-within .marketplace-capability-tooltip {
  visibility: visible;
  opacity: 1;
  transform: translate(-50%, 0);
}

.group\/capability:hover .marketplace-capability-tooltip-list,
.group\/capability:focus .marketplace-capability-tooltip-list,
.group\/capability:focus-within .marketplace-capability-tooltip-list {
  transform: translateY(0);
}

.marketplace-card-groups {
  scrollbar-width: none;
}

.marketplace-card-groups::-webkit-scrollbar {
  display: none;
}

.marketplace-status-bar:nth-child(1) {
  height: 8px;
}

.marketplace-status-bar:nth-child(2) {
  height: 10px;
}

.marketplace-status-bar:nth-child(3) {
  height: 12px;
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
