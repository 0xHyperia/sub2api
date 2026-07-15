import type {
  UserAvailableGroup,
  UserMarketplacePlatform,
  UserSupportedModelPricing,
} from '@/api/channels'

export interface MarketplaceModelEntry {
  key: string
  name: string
  platform: string
  groups: UserAvailableGroup[]
  pricing: UserSupportedModelPricing | null
}

export interface MarketplaceGroupOption extends UserAvailableGroup {
  effectiveRate: number
}

export function buildMarketplaceEntries(platforms: UserMarketplacePlatform[]): MarketplaceModelEntry[] {
  return platforms.flatMap((section) =>
    section.supported_models.map((model) => ({
      key: `${section.platform}::${model.name}`,
      name: model.name,
      platform: section.platform || model.platform,
      groups: model.groups,
      pricing: model.pricing,
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
  return [...groups.values()].sort((a, b) => a.name.localeCompare(b.name))
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
