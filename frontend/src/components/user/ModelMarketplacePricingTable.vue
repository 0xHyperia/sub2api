<template>
  <div v-if="variant === 'base' && columns.length" class="space-y-3">
    <dl class="grid grid-cols-2 gap-3">
      <div v-for="column in columns.filter(item => item.primary)" :key="column.key" class="min-w-0 rounded-panel border border-outline px-3 py-3">
        <dt class="text-xs text-foreground-muted">{{ column.label }}</dt>
        <dd class="mt-2 flex flex-wrap items-baseline gap-x-1.5 gap-y-1">
          <span class="break-all font-mono text-sm font-semibold tabular-nums text-foreground">{{ formatPrice(rows[0]?.values[column.key] ?? null, column.scale) }}</span>
          <span v-if="column.scale === 1000000" class="text-[10px] text-foreground-subtle">/ 1M</span>
        </dd>
      </div>
    </dl>
    <dl v-if="columns.some(item => !item.primary)" class="space-y-2 rounded-panel border border-outline px-3 py-3">
      <div v-for="column in columns.filter(item => !item.primary)" :key="column.key" class="flex items-baseline justify-between gap-3">
        <dt class="text-xs text-foreground-muted">{{ column.label }}</dt>
        <dd class="flex flex-wrap justify-end gap-x-1.5 font-mono text-xs tabular-nums text-foreground">
          <span>{{ formatPrice(rows[0]?.values[column.key] ?? null, column.scale) }}</span>
          <span class="text-[10px] text-foreground-subtle">/ 1M</span>
        </dd>
      </div>
    </dl>
  </div>
  <div v-else-if="columns.length || variant === 'groups'" class="overflow-x-auto" tabindex="0" role="region" :aria-label="t('modelMarketplace.details.pricing')">
    <table class="w-full text-left text-xs" :class="variant === 'groups' ? 'min-w-[520px]' : columns.length >= 4 ? 'min-w-[440px]' : ''">
      <thead class="border-b border-outline bg-surface-subtle text-foreground-muted">
        <tr>
          <th scope="col" class="px-3 py-2 font-medium">{{ t(variant === 'groups' ? 'modelMarketplace.details.group' : 'modelMarketplace.details.tier') }}</th>
          <th v-if="variant === 'groups'" scope="col" class="px-2 py-2 text-right font-medium">{{ t('modelMarketplace.details.multiplier') }}</th>
          <th v-for="column in columns" :key="column.key" scope="col" class="whitespace-nowrap px-2 py-2 text-right font-medium">{{ column.label }}</th>
        </tr>
      </thead>
      <tbody class="divide-y divide-outline">
        <tr v-for="(row, index) in rows" :key="index" :data-testid="row.group ? `group-pricing-${row.group.id}` : undefined" class="transition-colors duration-150 hover:bg-surface-subtle motion-reduce:transition-none">
          <th scope="row" class="min-w-20 max-w-40 break-words px-3 py-2.5 font-medium text-foreground-muted">
            {{ row.label }}
            <p v-if="row.group && !row.group.pricing" class="mt-1 text-[10px] font-normal text-foreground-subtle">{{ t('modelMarketplace.noPricing') }}</p>
            <div v-if="row.group?.pricing?.time_pricing?.periods.length" class="mt-2 text-[10px] font-normal text-foreground-muted">
              <p>{{ row.group.pricing.time_pricing.timezone }} · {{ t(row.group.pricing.time_pricing.weekdays_only ? 'modelMarketplace.details.weekdaysOnly' : 'modelMarketplace.details.everyDay') }}</p>
              <p v-for="(period, periodIndex) in row.group.pricing.time_pricing.periods" :key="periodIndex" class="mt-1">{{ period.start_time }} ~ {{ period.end_time }} · {{ period.multiplier }}×</p>
            </div>
          </th>
          <td v-if="variant === 'groups'" class="whitespace-nowrap px-2 py-2.5 text-right font-mono text-foreground-muted">{{ Number(row.group?.effectiveRate.toFixed(4)) }}×</td>
          <td v-for="column in columns" :key="column.key" class="whitespace-nowrap px-2 py-2.5 text-right font-mono tabular-nums text-foreground">
            {{ formatPrice(row.values[column.key] ?? null, column.scale, row.group ?? group) }}
          </td>
        </tr>
      </tbody>
    </table>
  </div>
  <p v-else class="px-3 py-3 text-xs text-foreground-subtle">{{ t('modelMarketplace.noPricing') }}</p>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { UserSupportedModelPricing } from '@/api/channels'
import { imagePriceRows, scaledCurrencyPrice, scaledPrice, scaledRechargePrice, type MarketplaceGroupOption } from '@/views/user/modelMarketplace'

const props = withDefaults(defineProps<{
  pricing: UserSupportedModelPricing | null
  group?: MarketplaceGroupOption
  groups?: MarketplaceGroupOption[]
  variant?: 'base' | 'tiers' | 'groups'
  showRechargePrices: boolean
  balanceRechargeMultiplier: number
  officialUsdToCnyRate: number
}>(), { variant: 'tiers', groups: () => [] })
const { t } = useI18n()

type PriceRow = { label: string; values: Record<string, number | null | undefined>; group?: MarketplaceGroupOption }
function priceValues(pricing: UserSupportedModelPricing | null | undefined): PriceRow['values'] {
  if (!pricing) return {}
  return pricing.billing_mode === 'token'
    ? { input: pricing.input_price, output: pricing.output_price, 'cache-read': pricing.cache_read_price, 'cache-write': pricing.cache_write_price, 'image-input': pricing.image_input_price }
    : { request: pricing.per_request_price ?? pricing.image_output_price }
}
const rows = computed<PriceRow[]>(() => {
  if (props.variant === 'groups') {
    return props.groups.map(group => ({ label: group.name, values: priceValues(group.pricing), group }))
  }
  const pricing = props.pricing
  if (!pricing) return []
  if (props.variant === 'base') return [{ label: '', values: priceValues(pricing) }]
  if (pricing.billing_mode === 'image' && props.group) {
    return imagePriceRows(pricing, props.group).map(row => ({ label: row.tier, values: { request: row.rawValue } }))
  }
  const tiers = pricing.intervals.length ? pricing.intervals : [null]
  return tiers.map(tier => ({
    label: tier
      ? tier.tier_label || (tier.max_tokens == null ? `${tier.min_tokens.toLocaleString()}+` : `${tier.min_tokens.toLocaleString()}–${tier.max_tokens.toLocaleString()}`)
      : t('modelMarketplace.details.baseTier'),
    // Token intervals already contain the billing resolver's absolute prices.
    values: pricing.billing_mode === 'token'
      ? {
          input: (tier ?? pricing).input_price,
          output: (tier ?? pricing).output_price,
          'cache-read': (tier ?? pricing).cache_read_price,
          'cache-write': (tier ?? pricing).cache_write_price,
          'image-input': pricing.image_input_price,
        }
      : { request: tier?.per_request_price ?? pricing.per_request_price ?? pricing.image_output_price },
  }))
})

const columns = computed(() => {
  const pricing = props.pricing ?? props.groups.find(group => group.pricing)?.pricing
  const candidates = pricing?.billing_mode === 'token'
    ? [
        { key: 'input', label: t('modelMarketplace.price.input'), scale: 1_000_000, primary: true },
        { key: 'output', label: t('modelMarketplace.price.output'), scale: 1_000_000, primary: true },
        { key: 'cache-read', label: t('modelMarketplace.price.cacheRead'), scale: 1_000_000 },
        { key: 'cache-write', label: t('modelMarketplace.price.cacheWrite'), scale: 1_000_000 },
        { key: 'image-input', label: t('modelMarketplace.price.imageInput'), scale: 1_000_000 },
      ]
    : [{ key: 'request', label: t(props.pricing?.billing_mode === 'image' ? 'modelMarketplace.price.perImage' : 'modelMarketplace.price.perRequest'), scale: 1, primary: true }]
  return candidates.filter(column => rows.value.some(row => {
    const value = row.values[column.key]
    return value != null && (column.primary || value > 0)
  }))
})

function formatPrice(value: number | null, scale: number, group?: MarketplaceGroupOption) {
  if (group) {
    return props.showRechargePrices
      ? scaledRechargePrice(value, scale, group.effectiveRate, props.balanceRechargeMultiplier)
      : scaledPrice(value, scale, group.effectiveRate)
  }
  return props.showRechargePrices
    ? scaledCurrencyPrice(value, scale, props.officialUsdToCnyRate)
    : scaledPrice(value, scale, 1)
}
</script>
