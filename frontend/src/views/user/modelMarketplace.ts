import type {
  UserAvailableGroup,
  UserMarketplacePlatform,
  UserSupportedModelPricing,
  UserModelMonitorSummary,
} from '@/api/channels'

export interface MarketplaceModelEntry {
  key: string
  name: string
  platform: string
  groups: UserAvailableGroup[]
  pricing: UserSupportedModelPricing | null
  monitorStatus: UserModelMonitorSummary | null
  displayOrder: number
  label: string
}

export interface MarketplaceGroupOption extends UserAvailableGroup {
  effectiveRate: number
}

export type MarketplaceBillingCategory = 'usage' | 'request' | 'unpriced'

export function billingCategory(pricing: UserSupportedModelPricing | null): MarketplaceBillingCategory {
  if (!pricing) return 'unpriced'
  return pricing.billing_mode === 'token' ? 'usage' : 'request'
}

export function buildMarketplaceEntries(platforms: UserMarketplacePlatform[]): MarketplaceModelEntry[] {
  return platforms.flatMap((section) =>
    section.supported_models.map((model) => ({
      key: `${section.platform}::${model.name}`,
      name: model.name,
      platform: section.platform || model.platform,
      groups: model.groups,
      pricing: model.pricing,
      monitorStatus: model.monitor_status ?? null,
      displayOrder: model.monitor_status?.display_order ?? 0,
      label: model.monitor_status?.label ?? '',
    })),
  )
}

export function buildMarketplaceGroups(
  entries: MarketplaceModelEntry[],
  userGroupRates: Record<number, number>,
): MarketplaceGroupOption[] {
  const groups = new Map<number, MarketplaceGroupOption>()
  for (const entry of entries) {
    for (const group of entry.groups) {
      if (groups.has(group.id)) continue
      groups.set(group.id, {
        ...group,
        effectiveRate: userGroupRates[group.id] ?? group.rate_multiplier,
      })
    }
  }
  return [...groups.values()].sort(compareMarketplaceGroups)
}

export function compareMarketplaceGroups(a: MarketplaceGroupOption, b: MarketplaceGroupOption): number {
  if (a.effectiveRate !== b.effectiveRate) return a.effectiveRate - b.effectiveRate
  const byName = a.name.localeCompare(b.name)
  return byName !== 0 ? byName : a.id - b.id
}

export function compareMarketplaceDisplayOrder(a: MarketplaceModelEntry, b: MarketplaceModelEntry): number {
  return b.displayOrder - a.displayOrder
}

export function sortedEntryGroups(entry: MarketplaceModelEntry, userGroupRates: Record<number, number>): MarketplaceGroupOption[] {
  return entry.groups.map(group => ({ ...group, effectiveRate: userGroupRates[group.id] ?? group.rate_multiplier })).sort(compareMarketplaceGroups)
}

export function effectiveRateForEntry(
  entry: MarketplaceModelEntry,
  selectedGroupId: number | null,
  userGroupRates: Record<number, number>,
): number {
  const groups = selectedGroupId == null
    ? entry.groups
    : entry.groups.filter((group) => group.id === selectedGroupId)
  if (groups.length === 0) return 1
  return Math.min(...groups.map((group) => userGroupRates[group.id] ?? group.rate_multiplier))
}

export function primaryPrice(pricing: UserSupportedModelPricing | null): number | null {
  if (!pricing) return null
  if (pricing.billing_mode === 'token') return pricing.input_price
  if (pricing.billing_mode === 'image') return pricing.per_request_price ?? pricing.image_output_price
  return pricing.per_request_price
}

export function scaledPrice(value: number | null, scale: number, rate: number): string {
  if (value == null) return '-'
  const amount = value * scale * rate
  return `$${amount.toPrecision(10).replace(/\.?0+$/, '')}`
}
