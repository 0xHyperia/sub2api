<template>
  <Teleport to="body">
    <Transition name="marketplace-detail">
      <div
        v-if="entry"
        class="fixed inset-0 z-50 bg-black/25 backdrop-blur-[1px]"
        role="dialog"
        aria-modal="true"
        :aria-labelledby="titleId"
        @click.self="emit('close')"
      >
        <aside
          ref="drawerPanel"
          tabindex="-1"
          class="absolute inset-y-0 right-0 flex h-full w-full max-w-5xl min-h-0 flex-col border-l border-outline bg-surface-raised shadow-floating outline-none"
        >
          <header class="z-20 shrink-0 border-b border-outline bg-surface-raised">
            <div class="flex min-h-[76px] w-full items-center gap-3 px-4 py-3 sm:px-6">
              <span class="flex h-8 w-8 shrink-0 items-center justify-center text-foreground" aria-hidden="true">
                <PlatformIcon :platform="entry.platform as GroupPlatform" size="lg" :class="platformIconClass(entry.platform)" />
              </span>
              <div class="min-w-0 flex-1">
                <div class="flex min-w-0 items-center gap-2">
                  <div class="flex min-w-0 items-center gap-2">
                    <h2 :id="titleId" class="truncate font-mono text-lg font-bold text-foreground sm:text-xl">{{ entry.name }}</h2>
                    <span
                      v-if="entry.label"
                      class="inline-flex h-5 max-w-28 shrink-0 items-center truncate rounded-control bg-black px-2 text-[10px] font-bold leading-none text-white"
                      :title="entry.label"
                    >{{ entry.label }}</span>
                  </div>
                  <button
                    type="button"
                    class="inline-flex h-7 w-7 shrink-0 items-center justify-center text-foreground-subtle transition-colors hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus"
                    :title="t('modelMarketplace.copyModel')"
                    :aria-label="t('modelMarketplace.copyModel')"
                    @click="copyModelName"
                  >
                    <Icon name="copy" size="sm" />
                  </button>
                </div>
                <p class="mt-0.5 text-xs text-foreground-muted">{{ providerLabel }} <span class="px-1 text-foreground-subtle">·</span> {{ billingLabel }}</p>
              </div>
              <button
                ref="closeButton"
                type="button"
                class="inline-flex h-9 w-9 shrink-0 items-center justify-center text-foreground-muted transition-colors hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus"
                :aria-label="t('common.close')"
                @click="emit('close')"
              >
                <Icon name="x" size="md" />
              </button>
            </div>
          </header>

          <div class="min-h-0 flex-1 overflow-y-auto">
            <main class="w-full px-4 py-5 sm:px-6">
              <nav
                class="grid h-10 rounded-panel bg-surface-subtle p-0.5"
                :class="detailTabs.length === 3 ? 'grid-cols-3' : 'grid-cols-2'"
                :aria-label="t('modelMarketplace.details.navigation')"
              >
                <button
                  v-for="tab in detailTabs"
                  :key="tab.value"
                  type="button"
                  class="inline-flex min-w-0 items-center justify-center gap-2 rounded-control px-3 text-xs font-medium text-foreground-muted transition-[background-color,color,box-shadow] sm:text-sm"
                  :class="activeTab === tab.value ? 'bg-surface-raised text-foreground shadow-sm' : 'hover:text-foreground'"
                  :aria-current="activeTab === tab.value ? 'page' : undefined"
                  @click="activeTab = tab.value"
                >
                  <Icon :name="tab.icon" size="xs" />
                  <span class="truncate">{{ tab.label }}</span>
                </button>
              </nav>

              <dl v-if="monitorEnabled" class="mt-4 grid grid-cols-1 divide-y divide-outline rounded-panel border border-outline sm:grid-cols-3 sm:divide-x sm:divide-y-0">
                <div class="px-4 py-3">
                  <dt class="flex items-center gap-2 text-[10px] font-medium text-foreground-subtle">
                    <Icon name="clock" size="xs" />
                    {{ t('modelMarketplace.details.latestLatency') }}
                  </dt>
                  <dd class="mt-1 font-mono text-base font-semibold tabular-nums text-foreground">{{ latestLatency }}</dd>
                </div>
                <div class="px-4 py-3">
                  <dt class="flex items-center gap-2 text-[10px] font-medium text-foreground-subtle">
                    <Icon name="chart" size="xs" />
                    {{ t('modelMarketplace.details.averageLatency') }}
                  </dt>
                  <dd class="mt-1 font-mono text-base font-semibold tabular-nums text-foreground">{{ averageLatency }}</dd>
                </div>
                <div class="px-4 py-3">
                  <dt class="flex items-center gap-2 text-[10px] font-medium text-foreground-subtle">
                    <Icon name="badge" size="xs" />
                    {{ t('modelMarketplace.details.availability7d') }}
                  </dt>
                  <dd class="mt-1 font-mono text-base font-semibold tabular-nums" :class="availabilityTextClass">{{ monitorAvailability }}</dd>
                </div>
              </dl>

              <section v-if="monitorEnabled" class="mt-4">
                <div class="flex items-center justify-between gap-4 text-[10px] text-foreground-subtle">
                  <h3 class="font-medium">{{ t('modelMarketplace.details.history') }}</h3>
                  <span>{{ lastCheckedAt }}</span>
                </div>
                <ModelMonitorTimeline class="mt-1.5" :points="entry.monitorStatus?.timeline" />
                <div class="mt-1.5 flex items-center justify-between text-[9px] text-foreground-subtle">
                  <span>{{ t('modelMarketplace.details.older') }}</span>
                  <span>{{ t('modelMarketplace.details.newest') }}</span>
                </div>
              </section>

              <div v-if="activeTab === 'overview'" class="mt-6 space-y-7">
                <section class="rounded-panel border border-outline p-4 sm:p-5">
                  <h3 class="text-sm font-semibold text-foreground">{{ t('modelMarketplace.details.pricing') }}</h3>

                  <template v-if="pricingRows.length">
                    <h4 class="mt-4 text-xs font-semibold text-foreground-muted">{{ t('modelMarketplace.details.basePricing') }}</h4>
                    <dl v-if="billingCategory(entry.pricing) === 'usage'" class="mt-3 grid gap-2 sm:grid-cols-2">
                      <div v-for="row in primaryBasePricingRows" :key="row.key" class="rounded-panel border border-outline px-4 py-3">
                        <dt class="text-xs text-foreground-muted">{{ row.label }}</dt>
                        <dd class="mt-1 flex items-baseline gap-2">
                          <span class="font-mono text-xl font-semibold tabular-nums text-foreground">{{ row.value }}</span>
                          <span class="text-[10px] text-foreground-subtle">/ 1M</span>
                        </dd>
                      </div>
                    </dl>
                    <dl v-if="secondaryBasePricingRows.length" class="mt-3 divide-y divide-outline rounded-panel border border-outline px-4">
                      <div v-for="row in secondaryBasePricingRows" :key="row.key" class="flex items-center justify-between gap-4 py-2.5">
                        <dt class="text-xs text-foreground-muted">{{ row.label }}</dt>
                        <dd class="flex items-baseline gap-1.5">
                          <span class="font-mono text-sm tabular-nums text-foreground">{{ row.value }}</span>
                          <span class="text-[9px] text-foreground-subtle">/ 1M</span>
                        </dd>
                      </div>
                    </dl>
                    <dl v-if="billingCategory(entry.pricing) === 'request'" class="mt-3 rounded-panel border border-outline px-4 py-3">
                      <div v-for="row in primaryBasePricingRows" :key="row.key" class="flex items-end justify-between gap-4">
                        <div>
                          <dt class="text-xs text-foreground-muted">{{ row.label }}</dt>
                          <dd class="mt-1 font-mono text-xl font-semibold tabular-nums text-foreground">{{ row.value }}</dd>
                        </div>
                        <span class="text-[10px] text-foreground-subtle">{{ entry.pricing?.billing_mode === 'image' ? t('modelMarketplace.price.perImage') : t('modelMarketplace.price.perRequest') }}</span>
                      </div>
                    </dl>
                  </template>
                  <p v-else class="mt-4 border-y border-outline py-5 text-sm text-foreground-subtle">{{ t('modelMarketplace.noPricing') }}</p>

                  <div v-if="groups.length && pricingRows.length" class="mt-5">
                    <h4 class="text-xs font-semibold text-foreground-muted">{{ t('modelMarketplace.details.groupPricing') }}</h4>
                    <div class="mt-2 space-y-2 sm:hidden">
                      <article
                        v-for="row in groupPricingRows"
                        :key="`mobile-price-${row.id}`"
                        class="rounded-panel border border-outline p-3"
                        :class="row.id === activeGroup?.id ? 'bg-surface-subtle ring-1 ring-inset ring-outline-strong' : ''"
                      >
                        <div class="flex items-center justify-between gap-3">
                          <h5 class="min-w-0 truncate text-xs font-semibold text-foreground" :title="row.name">{{ row.name }}</h5>
                          <span class="shrink-0 font-mono text-xs text-foreground-muted">{{ formatRate(row.effectiveRate) }}×</span>
                        </div>
                        <dl class="mt-2 grid grid-cols-2 gap-x-3 gap-y-2 border-t border-outline pt-2">
                          <div v-for="price in row.prices" :key="price.key" class="min-w-0">
                            <dt class="truncate text-[10px] text-foreground-subtle">{{ price.label }}</dt>
                            <dd class="mt-0.5 truncate font-mono text-xs tabular-nums text-foreground">{{ price.value }}</dd>
                          </div>
                        </dl>
                      </article>
                    </div>
                    <div class="mt-2 hidden overflow-x-auto rounded-panel border border-outline sm:block">
                      <table class="w-full min-w-[680px] text-left text-xs">
                        <thead class="border-b border-outline bg-surface-subtle text-foreground-muted">
                          <tr>
                            <th class="px-3 py-2.5 font-medium">{{ t('modelMarketplace.details.group') }}</th>
                            <th class="px-3 py-2.5 font-medium">{{ t('modelMarketplace.details.multiplier') }}</th>
                            <th v-for="row in pricingRows" :key="row.key" class="px-3 py-2.5 text-right font-medium">{{ row.label }}</th>
                          </tr>
                        </thead>
                        <tbody class="divide-y divide-outline">
                          <tr v-for="row in groupPricingRows" :key="row.id" :class="row.id === activeGroup?.id ? 'bg-surface-subtle' : ''">
                            <td class="max-w-60 px-3 py-3 font-medium text-foreground"><span class="block truncate" :title="row.name">{{ row.name }}</span></td>
                            <td class="px-3 py-3 font-mono text-foreground-muted">{{ formatRate(row.effectiveRate) }}×</td>
                            <td v-for="price in row.prices" :key="price.key" class="px-3 py-3 text-right font-mono tabular-nums text-foreground">{{ price.value }}</td>
                          </tr>
                        </tbody>
                      </table>
                    </div>
                    <p class="mt-2 text-[10px] text-foreground-subtle">{{ entry.pricing?.billing_mode === 'image' ? t('modelMarketplace.details.imagePricingHint') : t('modelMarketplace.details.pricePerMillionHint') }}</p>
                  </div>

                  <div v-if="entry.pricing?.intervals.length" class="mt-5">
                    <h4 class="text-xs font-semibold text-foreground-muted">{{ t('modelMarketplace.details.tieredPricing') }}</h4>
                    <dl class="mt-2 divide-y divide-outline rounded-panel border border-outline sm:hidden">
                      <div v-for="(interval, index) in entry.pricing.intervals" :key="`mobile-tier-${index}`" class="p-3">
                        <dt class="text-xs font-medium text-foreground">{{ intervalLabel(interval.min_tokens, interval.max_tokens) }}</dt>
                        <dd class="mt-2 grid grid-cols-2 gap-3">
                          <span>
                            <span class="block text-[10px] text-foreground-subtle">{{ t('modelMarketplace.price.input') }}</span>
                            <span class="mt-0.5 block font-mono text-xs text-foreground">{{ intervalPrice(interval.input_price) }}</span>
                          </span>
                          <span>
                            <span class="block text-[10px] text-foreground-subtle">{{ t('modelMarketplace.price.output') }}</span>
                            <span class="mt-0.5 block font-mono text-xs text-foreground">{{ intervalPrice(interval.output_price) }}</span>
                          </span>
                        </dd>
                      </div>
                    </dl>
                    <div class="mt-2 hidden overflow-x-auto rounded-panel border border-outline sm:block">
                      <table class="w-full min-w-96 text-left text-xs">
                        <thead class="border-b border-outline bg-surface-subtle text-foreground-muted">
                          <tr><th class="px-3 py-2">{{ t('modelMarketplace.details.range') }}</th><th class="px-3 py-2">{{ t('modelMarketplace.price.input') }}</th><th class="px-3 py-2">{{ t('modelMarketplace.price.output') }}</th></tr>
                        </thead>
                        <tbody class="divide-y divide-outline text-foreground-muted">
                          <tr v-for="(interval, index) in entry.pricing.intervals" :key="index">
                            <td class="px-3 py-2">{{ intervalLabel(interval.min_tokens, interval.max_tokens) }}</td>
                            <td class="px-3 py-2 font-mono">{{ intervalPrice(interval.input_price) }}</td>
                            <td class="px-3 py-2 font-mono">{{ intervalPrice(interval.output_price) }}</td>
                          </tr>
                        </tbody>
                      </table>
                    </div>
                  </div>
                </section>

                <section>
                  <h3 class="mb-3 text-sm font-semibold text-foreground">{{ t('modelMarketplace.details.modelInfo') }}</h3>
                  <dl class="grid overflow-hidden rounded-panel border border-outline sm:grid-cols-2">
                    <div class="border-b border-outline px-4 py-3 sm:border-r">
                      <dt class="text-[10px] text-foreground-subtle">{{ t('modelMarketplace.details.provider') }}</dt>
                      <dd class="mt-1 text-sm font-semibold text-foreground">{{ providerLabel }}</dd>
                    </div>
                    <div class="border-b border-outline px-4 py-3">
                      <dt class="text-[10px] text-foreground-subtle">{{ t('modelMarketplace.details.billingType') }}</dt>
                      <dd class="mt-1 text-sm font-semibold text-foreground">{{ billingLabel }}</dd>
                    </div>
                    <div class="border-b border-outline px-4 py-3 sm:border-b-0 sm:border-r">
                      <dt class="text-[10px] text-foreground-subtle">{{ t('modelMarketplace.details.groups') }}</dt>
                      <dd class="mt-2 flex flex-wrap gap-1.5">
                        <button
                          v-for="group in groups"
                          :key="group.id"
                          type="button"
                          class="rounded-control bg-surface-subtle px-2 py-1 text-[10px] text-foreground-muted transition-colors hover:text-foreground"
                          :class="group.id === activeGroup?.id ? 'font-semibold text-foreground ring-1 ring-inset ring-outline-strong' : ''"
                          @click="emit('selectGroup', group.id)"
                        >{{ group.name }}</button>
                      </dd>
                    </div>
                    <div class="px-4 py-3">
                      <dt class="text-[10px] text-foreground-subtle">{{ t('modelMarketplace.details.endpoints') }}</dt>
                      <dd class="mt-2 flex flex-wrap gap-1.5">
                        <span v-for="protocol in availableProtocols" :key="protocol" class="rounded-control bg-surface-subtle px-2 py-1 font-mono text-[10px] text-foreground-muted">{{ protocol }}</span>
                      </dd>
                    </div>
                  </dl>
                </section>
              </div>

              <div v-else-if="activeTab === 'performance' && monitorEnabled" class="mt-6 space-y-8">
                <section>
                  <div class="mb-3">
                    <h3 class="flex items-center gap-2 text-sm font-semibold text-foreground">
                      <Icon name="badge" size="sm" class="text-foreground-subtle" />
                      {{ t('modelMarketplace.details.groupPerformance') }}
                    </h3>
                    <p class="mt-0.5 text-xs text-foreground-subtle">{{ t('modelMarketplace.details.groupPerformanceHint') }}</p>
                  </div>
                  <div class="space-y-2 sm:hidden">
                    <article v-for="row in groupPerformanceRows" :key="`mobile-performance-${row.id}`" class="rounded-panel border border-outline p-3">
                      <h4 class="truncate text-xs font-semibold text-foreground" :title="row.name">{{ row.name }}</h4>
                      <dl class="mt-2 grid grid-cols-3 gap-2 border-t border-outline pt-2 text-center">
                        <div>
                          <dt class="text-[9px] text-foreground-subtle">{{ t('modelMarketplace.details.averageLatency') }}</dt>
                          <dd class="mt-1 font-mono text-xs tabular-nums text-foreground">{{ row.averageLatency }}</dd>
                        </div>
                        <div>
                          <dt class="text-[9px] text-foreground-subtle">{{ t('modelMarketplace.details.successRate') }}</dt>
                          <dd class="mt-1 font-mono text-xs font-semibold tabular-nums" :class="row.successRateClass">{{ row.successRate }}</dd>
                        </div>
                        <div>
                          <dt class="text-[9px] text-foreground-subtle">{{ t('modelMarketplace.details.samples') }}</dt>
                          <dd class="mt-1 font-mono text-xs tabular-nums text-foreground-muted">{{ row.samples }}</dd>
                        </div>
                      </dl>
                    </article>
                  </div>
                  <div class="hidden overflow-x-auto rounded-panel border border-outline sm:block">
                    <table class="w-full min-w-[620px] text-left text-xs">
                      <thead class="border-b border-outline bg-surface-subtle text-foreground-muted">
                        <tr>
                          <th class="px-3 py-2.5 font-medium">{{ t('modelMarketplace.details.group') }}</th>
                          <th class="px-3 py-2.5 text-right font-medium">{{ t('modelMarketplace.details.averageLatency') }}</th>
                          <th class="px-3 py-2.5 text-right font-medium">{{ t('modelMarketplace.details.successRate') }}</th>
                          <th class="px-3 py-2.5 text-right font-medium">{{ t('modelMarketplace.details.samples') }}</th>
                        </tr>
                      </thead>
                      <tbody class="divide-y divide-outline">
                        <tr v-for="row in groupPerformanceRows" :key="row.id">
                          <td class="max-w-72 px-3 py-3 font-medium text-foreground"><span class="block truncate" :title="row.name">{{ row.name }}</span></td>
                          <td class="px-3 py-3 text-right font-mono tabular-nums text-foreground">{{ row.averageLatency }}</td>
                          <td class="px-3 py-3 text-right font-mono font-semibold tabular-nums" :class="row.successRateClass">{{ row.successRate }}</td>
                          <td class="px-3 py-3 text-right font-mono tabular-nums text-foreground-muted">{{ row.samples }}</td>
                        </tr>
                      </tbody>
                    </table>
                  </div>
                </section>

                <ModelMarketplacePerformanceCharts :points="entry.monitorStatus?.timeline" />
              </div>

              <div v-else class="mt-6 space-y-8">
                <section>
                  <div class="mb-3 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                    <h3 class="flex items-center gap-2 text-sm font-semibold text-foreground">
                      <Icon name="terminal" size="sm" class="text-foreground-subtle" />
                      {{ t('modelMarketplace.details.callExample') }}
                    </h3>
                    <div class="flex min-w-0 gap-1 overflow-x-auto rounded-panel bg-surface-subtle p-0.5">
                      <button
                        v-for="language in codeLanguages"
                        :key="language.value"
                        type="button"
                        class="h-8 shrink-0 rounded-control px-3 text-xs text-foreground-muted transition-[background-color,color,box-shadow]"
                        :class="activeLanguage === language.value ? 'bg-surface-raised text-foreground shadow-sm' : 'hover:text-foreground'"
                        @click="activeLanguage = language.value"
                      >{{ language.label }}</button>
                    </div>
                  </div>

                  <div class="mb-3 flex gap-1 overflow-x-auto">
                    <button
                      v-for="protocol in availableProtocols"
                      :key="protocol"
                      type="button"
                      class="h-8 shrink-0 rounded-control px-3 text-xs transition-colors"
                      :class="activeProtocol === protocol ? 'bg-foreground text-surface' : 'text-foreground-muted hover:bg-surface-subtle hover:text-foreground'"
                      @click="activeProtocol = protocol"
                    >{{ protocol }}</button>
                  </div>

                  <div class="relative overflow-hidden rounded-panel border border-outline bg-surface-subtle">
                    <button
                      type="button"
                      class="absolute right-3 top-3 z-10 inline-flex h-8 w-8 items-center justify-center rounded-control bg-surface-raised text-foreground-muted shadow-sm transition-colors hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus"
                      :title="t('modelMarketplace.details.copyExample')"
                      :aria-label="t('modelMarketplace.details.copyExample')"
                      @click="copyApiExample"
                    >
                      <Icon name="copy" size="sm" />
                    </button>
                    <pre class="max-h-[440px] overflow-auto p-4 pr-14 text-xs leading-6 text-foreground sm:p-5 sm:pr-16"><code>{{ apiExample }}</code></pre>
                  </div>
                  <p class="mt-2 text-xs text-foreground-subtle">{{ t('modelMarketplace.details.apiKeyHint') }}</p>
                </section>

                <section>
                  <h3 class="mb-3 flex items-center gap-2 text-sm font-semibold text-foreground">
                    <Icon name="key" size="sm" class="text-foreground-subtle" />
                    {{ t('modelMarketplace.details.authentication') }}
                  </h3>
                  <div class="rounded-panel border border-outline px-4 py-3 text-xs leading-6 text-foreground-muted">
                    <p>{{ authenticationDescription }}</p>
                    <p class="mt-1 text-foreground-subtle">{{ t('modelMarketplace.details.authenticationHint') }}</p>
                  </div>
                </section>

                <section>
                  <h3 class="mb-3 flex items-center gap-2 text-sm font-semibold text-foreground">
                    <Icon name="calculator" size="sm" class="text-foreground-subtle" />
                    {{ t('modelMarketplace.details.supportedParameters') }}
                  </h3>
                  <dl class="divide-y divide-outline rounded-panel border border-outline sm:hidden">
                    <div v-for="parameter in apiParameters" :key="`mobile-parameter-${parameter.name}`" class="p-3">
                      <div class="flex items-start justify-between gap-3">
                        <dt class="font-mono text-xs font-semibold text-foreground">{{ parameter.name }}</dt>
                        <span class="shrink-0 rounded-control bg-surface-subtle px-2 py-0.5 font-mono text-[10px] text-foreground-muted">{{ parameter.type }}</span>
                      </div>
                      <dd class="mt-1.5 text-xs leading-5 text-foreground-muted">{{ parameter.description }}</dd>
                      <dd class="mt-1 font-mono text-[10px] text-foreground-subtle">{{ parameter.range }}</dd>
                    </div>
                  </dl>
                  <div class="hidden overflow-x-auto rounded-panel border border-outline sm:block">
                    <table class="w-full min-w-[720px] text-left text-xs">
                      <thead class="border-b border-outline bg-surface-subtle text-foreground-muted">
                        <tr>
                          <th class="px-3 py-2.5 font-medium">{{ t('modelMarketplace.details.parameter') }}</th>
                          <th class="px-3 py-2.5 font-medium">{{ t('modelMarketplace.details.type') }}</th>
                          <th class="px-3 py-2.5 font-medium">{{ t('modelMarketplace.details.defaultRange') }}</th>
                          <th class="px-3 py-2.5 font-medium">{{ t('modelMarketplace.details.description') }}</th>
                        </tr>
                      </thead>
                      <tbody class="divide-y divide-outline">
                        <tr v-for="parameter in apiParameters" :key="parameter.name">
                          <td class="px-3 py-2.5 font-mono text-foreground">{{ parameter.name }}</td>
                          <td class="px-3 py-2.5"><span class="rounded-control bg-surface-subtle px-2 py-1 font-mono text-[10px] text-foreground-muted">{{ parameter.type }}</span></td>
                          <td class="px-3 py-2.5 font-mono text-foreground-muted">{{ parameter.range }}</td>
                          <td class="px-3 py-2.5 text-foreground-muted">{{ parameter.description }}</td>
                        </tr>
                      </tbody>
                    </table>
                  </div>
                </section>

                <section v-if="groups.length">
                  <h3 class="mb-3 flex items-center gap-2 text-sm font-semibold text-foreground">
                    <Icon name="clock" size="sm" class="text-foreground-subtle" />
                    {{ t('modelMarketplace.details.rateLimits') }}
                  </h3>
                  <dl class="divide-y divide-outline rounded-panel border border-outline sm:hidden">
                    <div v-for="group in groups" :key="`mobile-limit-${group.id}`" class="flex items-center justify-between gap-3 px-3 py-2.5">
                      <dt class="min-w-0 truncate text-xs text-foreground">{{ group.name }}</dt>
                      <dd class="shrink-0 font-mono text-xs tabular-nums text-foreground">
                        {{ group.rpm_limit ? `${group.rpm_limit.toLocaleString()} RPM` : t('modelMarketplace.details.unlimited') }}
                      </dd>
                    </div>
                  </dl>
                  <div class="hidden overflow-x-auto rounded-panel border border-outline sm:block">
                    <table class="w-full min-w-[420px] text-left text-xs">
                      <thead class="border-b border-outline bg-surface-subtle text-foreground-muted">
                        <tr><th class="px-3 py-2.5 font-medium">{{ t('modelMarketplace.details.group') }}</th><th class="px-3 py-2.5 text-right font-medium">RPM</th></tr>
                      </thead>
                      <tbody class="divide-y divide-outline">
                        <tr v-for="group in groups" :key="group.id">
                          <td class="px-3 py-2.5 text-foreground">{{ group.name }}</td>
                          <td class="px-3 py-2.5 text-right font-mono tabular-nums text-foreground">{{ group.rpm_limit ? group.rpm_limit.toLocaleString() : t('modelMarketplace.details.unlimited') }}</td>
                        </tr>
                      </tbody>
                    </table>
                  </div>
                  <p class="mt-2 text-[10px] text-foreground-subtle">{{ t('modelMarketplace.details.rateLimitHint') }}</p>
                </section>
              </div>
            </main>
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
import ModelMarketplacePerformanceCharts from '@/components/user/ModelMarketplacePerformanceCharts.vue'
import ModelMonitorTimeline from '@/components/user/ModelMonitorTimeline.vue'
import { platformIconClass } from '@/utils/platformColors'
import { useAppStore } from '@/stores/app'
import { useClipboard } from '@/composables/useClipboard'
import { billingCategory, imagePriceRows, scaledPrice, type MarketplaceGroupOption, type MarketplaceModelEntry } from '@/views/user/modelMarketplace'

type DetailTab = 'overview' | 'performance' | 'api'
type ApiProtocol = 'anthropic' | 'openai' | 'gemini'
type CodeLanguage = 'curl' | 'python' | 'typescript' | 'javascript'

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
const appStore = useAppStore()
const { copyToClipboard } = useClipboard()
const closeButton = ref<HTMLButtonElement | null>(null)
const drawerPanel = ref<HTMLElement | null>(null)
const activeTab = ref<DetailTab>('overview')
const activeProtocol = ref<ApiProtocol>('openai')
const activeLanguage = ref<CodeLanguage>('curl')
const titleId = `model-marketplace-detail-${Math.random().toString(36).slice(2)}`
let previousFocus: HTMLElement | null = null
let previousBodyOverflow = ''

const detailTabs = computed(() => [
  { value: 'overview' as const, label: t('modelMarketplace.details.overview'), icon: 'infoCircle' as const },
  ...(props.monitorEnabled
    ? [{ value: 'performance' as const, label: t('modelMarketplace.details.performance'), icon: 'badge' as const }]
    : []),
  { value: 'api' as const, label: 'API', icon: 'terminal' as const },
])
const codeLanguages = [
  { value: 'curl' as const, label: 'cURL' },
  { value: 'python' as const, label: 'Python' },
  { value: 'typescript' as const, label: 'TypeScript' },
  { value: 'javascript' as const, label: 'JavaScript' },
]

const providerLabel = computed(() => props.entry ? labelOrFallback(`modelMarketplace.providers.${props.entry.platform}`, props.entry.platform.toUpperCase()) : '')
const billingLabel = computed(() => props.entry ? t(`modelMarketplace.billing.${billingCategory(props.entry.pricing)}`) : '')
const monitorPoints = computed(() => props.entry?.monitorStatus?.timeline ?? [])
const latencyValues = computed(() => monitorPoints.value.flatMap(point => point.latency_ms == null ? [] : [point.latency_ms]))
const averageLatency = computed(() => latencyValues.value.length ? formatLatency(latencyValues.value.reduce((sum, value) => sum + value, 0) / latencyValues.value.length) : '—')
const latestLatency = computed(() => props.entry?.monitorStatus?.latency_ms == null ? '—' : formatLatency(props.entry.monitorStatus.latency_ms))
const lastCheckedAt = computed(() => {
  const value = props.entry?.monitorStatus?.last_checked_at
  if (!value) return t('modelMarketplace.monitor.noData')
  return new Intl.DateTimeFormat(locale.value, {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  }).format(new Date(value))
})
const monitorAvailability = computed(() => {
  const value = props.entry?.monitorStatus?.availability_7d
  return value == null ? '—' : `${value.toFixed(2)}%`
})
const availabilityTextClass = computed(() => {
  const value = props.entry?.monitorStatus?.availability_7d
  if (value == null) return 'text-foreground'
  if (value >= 99) return 'text-success-foreground'
  if (value >= 95) return 'text-warning-foreground'
  return 'text-danger-foreground'
})

const pricingRows = computed(() => {
  const pricing = props.entry?.pricing
  if (!pricing) return []
  if (pricing.billing_mode === 'image') {
    return imagePriceRows(pricing, props.activeGroup).map(row => ({
      key: row.key,
      label: row.tier,
      rawValue: row.rawValue,
      scale: 1,
    }))
  }
  const primaryRows = pricing.billing_mode === 'token'
    ? [
        ['input', t('modelMarketplace.price.input'), pricing.input_price, 1_000_000],
        ['output', t('modelMarketplace.price.output'), pricing.output_price, 1_000_000],
        ['cache-read', t('modelMarketplace.price.cacheRead'), pricing.cache_read_price, 1_000_000],
        ['cache-write', t('modelMarketplace.price.cacheWrite'), pricing.cache_write_price, 1_000_000],
      ] as const
    : [['request', t('modelMarketplace.price.request'), pricing.per_request_price ?? pricing.image_output_price, 1]] as const
  const rows = [
    ...primaryRows,
    ['image-input', t('modelMarketplace.price.imageInput'), pricing.image_input_price, 1_000_000] as const,
  ]
  return rows.filter(([, , value]) => value != null).map(([key, label, value, scale]) => ({ key, label, rawValue: value, scale }))
})
const basePricingRows = computed(() => pricingRows.value.map(row => ({ ...row, value: scaledPrice(row.rawValue, row.scale, 1) })))
const primaryBasePricingRows = computed(() => basePricingRows.value.filter(row => row.key === 'input' || row.key === 'output' || row.key === 'request' || row.key.startsWith('image-')))
const secondaryBasePricingRows = computed(() => basePricingRows.value.filter(row => row.key === 'cache-read' || row.key === 'cache-write' || row.key === 'image-input'))
const groupPricingRows = computed(() => props.groups.map(group => ({
  ...group,
  prices: props.entry?.pricing?.billing_mode === 'image'
    ? imagePriceRows(props.entry.pricing, group).map(row => ({ key: row.key, label: row.tier, value: scaledPrice(row.effectiveValue, 1, 1) }))
    : pricingRows.value.map(row => ({ key: row.key, label: row.label, value: scaledPrice(row.rawValue, row.scale, group.effectiveRate) })),
})))

const groupPerformanceRows = computed(() => props.groups.map(group => {
  const points = monitorPoints.value.filter(point => point.group_id === group.id || (!point.group_id && point.group_name === group.name))
  const latencies = points.flatMap(point => point.latency_ms == null ? [] : [point.latency_ms])
  const successCount = points.filter(point => point.status === 'operational' || point.status === 'degraded').length
  const successValue = points.length ? (successCount / points.length) * 100 : null
  return {
    id: group.id,
    name: group.name,
    averageLatency: latencies.length ? formatLatency(latencies.reduce((sum, value) => sum + value, 0) / latencies.length) : '—',
    successRate: successValue == null ? '—' : `${successValue.toFixed(2)}%`,
    successRateClass: successValue == null ? 'text-foreground-muted' : successValue >= 99 ? 'text-success-foreground' : successValue >= 95 ? 'text-warning-foreground' : 'text-danger-foreground',
    samples: points.length || '—',
  }
}))

const availableProtocols = computed<ApiProtocol[]>(() => {
  if (props.entry?.platform === 'anthropic') return ['anthropic', 'openai']
  if (props.entry?.platform === 'gemini') return ['gemini', 'openai']
  return ['openai']
})
const apiBaseUrl = computed(() => (appStore.apiBaseUrl || (typeof window !== 'undefined' ? window.location.origin : '')).replace(/\/+$/, ''))
const requestDefinition = computed<{ url: string; headers: Record<string, string>; body: Record<string, unknown> }>(() => {
  const model = props.entry?.name ?? 'model-name'
  if (activeProtocol.value === 'anthropic') {
    const headers: Record<string, string> = { 'x-api-key': 'YOUR_API_KEY', 'anthropic-version': '2023-06-01', 'Content-Type': 'application/json' }
    return {
      url: `${apiBaseUrl.value}/v1/messages`,
      headers,
      body: { model, max_tokens: 1024, messages: [{ role: 'user', content: 'Explain quantum entanglement in one paragraph.' }] },
    }
  }
  if (activeProtocol.value === 'gemini') {
    const headers: Record<string, string> = { 'x-goog-api-key': 'YOUR_API_KEY', 'Content-Type': 'application/json' }
    return {
      url: `${apiBaseUrl.value}/v1beta/models/${model}:generateContent`,
      headers,
      body: { contents: [{ parts: [{ text: 'Explain quantum entanglement in one paragraph.' }] }] },
    }
  }
  const headers: Record<string, string> = { Authorization: 'Bearer YOUR_API_KEY', 'Content-Type': 'application/json' }
  return {
    url: `${apiBaseUrl.value}/v1/chat/completions`,
    headers,
    body: { model, messages: [{ role: 'user', content: 'Explain quantum entanglement in one paragraph.' }], stream: false },
  }
})

const apiExample = computed(() => {
  const request = requestDefinition.value
  const body = JSON.stringify(request.body, null, 2)
  if (activeLanguage.value === 'curl') {
    const headers = Object.entries(request.headers).map(([key, value]) => `  -H '${key}: ${value}' \\`).join('\n')
    return `curl '${request.url}' \\\n${headers}\n  -d '${body}'`
  }
  if (activeLanguage.value === 'python') {
    const pythonBody = body.replace(/\bfalse\b/g, 'False').replace(/\btrue\b/g, 'True').replace(/\bnull\b/g, 'None')
    return `import requests\n\nresponse = requests.post(\n    '${request.url}',\n    headers=${formatObject(request.headers, 4)},\n    json=${pythonBody.split('\n').join('\n    ')}\n)\n\nprint(response.json())`
  }
  const typePrefix = activeLanguage.value === 'typescript' ? `type ApiResponse = Record<string, unknown>\n\n` : ''
  const responseLine = activeLanguage.value === 'typescript' ? 'const data = await response.json() as ApiResponse' : 'const data = await response.json()'
  return `${typePrefix}const response = await fetch('${request.url}', {\n  method: 'POST',\n  headers: ${formatObject(request.headers, 2)},\n  body: JSON.stringify(${body.split('\n').join('\n  ')})\n})\n\n${responseLine}\nconsole.log(data)`
})

const authenticationDescription = computed(() => activeProtocol.value === 'anthropic'
  ? t('modelMarketplace.details.authAnthropic')
  : activeProtocol.value === 'gemini'
    ? t('modelMarketplace.details.authGemini')
    : t('modelMarketplace.details.authOpenAI'))

const apiParameters = computed(() => {
  if (activeProtocol.value === 'anthropic') {
    return [
      parameter('max_tokens', 'integer', '≥ 1', 'maxTokens'),
      parameter('temperature', 'number', '0 – 1', 'temperature'),
      parameter('top_p', 'number', '0 – 1', 'topP'),
      parameter('top_k', 'integer', '≥ 0', 'topK'),
      parameter('stop_sequences', 'array', '—', 'stop'),
      parameter('stream', 'boolean', 'false', 'stream'),
      parameter('system', 'string | array', '—', 'system'),
      parameter('tools', 'array', '—', 'tools'),
      parameter('tool_choice', 'object', 'auto', 'toolChoice'),
    ]
  }
  if (activeProtocol.value === 'gemini') {
    return [
      parameter('temperature', 'number', '0 – 2', 'temperature'),
      parameter('topP', 'number', '0 – 1', 'topP'),
      parameter('topK', 'integer', '≥ 0', 'topK'),
      parameter('maxOutputTokens', 'integer', '≥ 1', 'maxTokens'),
      parameter('stopSequences', 'array', '—', 'stop'),
      parameter('responseMimeType', 'string', 'text/plain', 'responseFormat'),
      parameter('tools', 'array', '—', 'tools'),
    ]
  }
  return [
    parameter('temperature', 'number', '0 – 2', 'temperature'),
    parameter('top_p', 'number', '0 – 1', 'topP'),
    parameter('max_tokens', 'integer', '≥ 1', 'maxTokens'),
    parameter('frequency_penalty', 'number', '-2 – 2', 'frequencyPenalty'),
    parameter('presence_penalty', 'number', '-2 – 2', 'presencePenalty'),
    parameter('stop', 'array', '—', 'stop'),
    parameter('seed', 'integer', '—', 'seed'),
    parameter('stream', 'boolean', 'false', 'stream'),
    parameter('response_format', 'object', '—', 'responseFormat'),
    parameter('tools', 'array', '—', 'tools'),
    parameter('tool_choice', 'string | object', 'auto', 'toolChoice'),
  ]
})

function parameter(name: string, type: string, range: string, descriptionKey: string) {
  return { name, type, range, description: t(`modelMarketplace.details.parameterDescriptions.${descriptionKey}`) }
}
function labelOrFallback(key: string, fallback: string) { const value = t(key); return value === key ? fallback : value }
function formatRate(value: number) { return Number(value.toFixed(4)).toString() }
function formatLatency(value: number) { return value >= 1000 ? `${(value / 1000).toFixed(2)} s` : `${Math.round(value)} ms` }
function intervalLabel(min: number, max: number | null) { return max == null ? `${min.toLocaleString()}+` : `${min.toLocaleString()}–${max.toLocaleString()}` }
function intervalPrice(value: number | null) { return scaledPrice(value, 1_000_000, props.showEffectivePrices ? props.activeGroup?.effectiveRate ?? 1 : 1) }
function formatObject(value: Record<string, string>, indent: number) {
  const padding = ' '.repeat(indent)
  return JSON.stringify(value, null, 2).split('\n').map((line, index) => index === 0 ? line : `${padding}${line}`).join('\n')
}
async function copyModelName() {
  if (props.entry) await copyToClipboard(props.entry.name, t('modelMarketplace.copySuccess', { name: props.entry.name }))
}
async function copyApiExample() { await copyToClipboard(apiExample.value, t('modelMarketplace.details.exampleCopied')) }
function handleKeydown(event: KeyboardEvent) {
  if (!props.entry) return
  if (event.key === 'Escape') {
    event.preventDefault()
    emit('close')
    return
  }
  if (event.key !== 'Tab' || !drawerPanel.value) return
  const focusable = [...drawerPanel.value.querySelectorAll<HTMLElement>('button:not([disabled]),a[href],select,[tabindex]:not([tabindex="-1"])')]
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
    activeTab.value = 'overview'
    activeLanguage.value = 'curl'
    activeProtocol.value = entry.platform === 'anthropic' ? 'anthropic' : entry.platform === 'gemini' ? 'gemini' : 'openai'
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
watch(availableProtocols, (protocols) => {
  if (!protocols.includes(activeProtocol.value)) activeProtocol.value = protocols[0] ?? 'openai'
})
onBeforeUnmount(() => {
  document.removeEventListener('keydown', handleKeydown)
  document.body.style.overflow = previousBodyOverflow
})
</script>

<style scoped>
.marketplace-detail-enter-active,
.marketplace-detail-leave-active { transition: opacity 160ms ease; }
.marketplace-detail-enter-active aside,
.marketplace-detail-leave-active aside { transition: transform 180ms ease; }
.marketplace-detail-enter-from,
.marketplace-detail-leave-to { opacity: 0; }
.marketplace-detail-enter-from aside,
.marketplace-detail-leave-to aside { transform: translateX(100%); }
</style>
