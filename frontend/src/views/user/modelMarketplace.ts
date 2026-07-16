import type {
  UserAvailableGroup,
  UserMarketplacePlatform,
  UserModelMonitorTimelinePoint,
  UserSupportedModelPricing,
  UserModelMonitorSummary,
} from '@/api/channels'

export interface MarketplaceModelEntry {
  key: string
  name: string
  platform: string
  groups: UserAvailableGroup[]
  pricing: UserSupportedModelPricing | null
  capabilities: string[]
  monitorStatus: UserModelMonitorSummary | null
  displayOrder: number
  label: string
}

export interface MarketplaceGroupOption extends UserAvailableGroup {
  effectiveRate: number
}

export type MarketplaceBillingCategory = 'usage' | 'request' | 'unpriced'
export const MARKETPLACE_MODEL_CAPABILITIES = [
  'vision',
  'image_input',
  'audio_input',
  'video_input',
  'function_calling',
  'reasoning',
  'prompt_caching',
  'web_search',
  'pdf_input',
  'computer_use',
  'image_generation',
  'audio_output',
  'parallel_tools',
  'tool_choice',
  'structured_output',
  'assistant_prefill',
  'streaming',
  'system_messages',
  'url_context',
  'image_embedding',
  'service_tier',
] as const

export type MarketplaceModelCapability = typeof MARKETPLACE_MODEL_CAPABILITIES[number]

export const MARKETPLACE_CARD_CAPABILITY_ORDER: MarketplaceModelCapability[] = [
  'vision',
  'image_input',
  'audio_input',
  'video_input',
  'function_calling',
  'parallel_tools',
  'tool_choice',
  'structured_output',
  'assistant_prefill',
  'reasoning',
  'prompt_caching',
  'web_search',
  'pdf_input',
  'computer_use',
  'image_generation',
  'audio_output',
  'streaming',
  'system_messages',
  'url_context',
  'image_embedding',
  'service_tier',
]
export type MarketplaceMonitorSignalStatus = UserModelMonitorTimelinePoint['status'] | ''

export const DEFAULT_USD_TO_CNY_RATE = 7.2
const MARKETPLACE_PROVIDER_PRIORITY = ['openai', 'anthropic', 'gemini'] as const

export function billingCategory(pricing: UserSupportedModelPricing | null): MarketplaceBillingCategory {
  if (!pricing) return 'unpriced'
  return pricing.billing_mode === 'token' ? 'usage' : 'request'
}

export function inferMarketplaceModelCapabilities(
  entry: Pick<MarketplaceModelEntry, 'name' | 'platform' | 'pricing'> & { capabilities?: string[] },
): MarketplaceModelCapability[] {
  const declared = new Set(entry.capabilities ?? [])
  const catalogCapabilities = MARKETPLACE_MODEL_CAPABILITIES.filter(capability => declared.has(capability))
  if (catalogCapabilities.length > 0) return catalogCapabilities

  const name = entry.name.toLowerCase()
  const platform = entry.platform.toLowerCase()
  const imageModel = entry.pricing?.billing_mode === 'image' || /(image|imagine|dall-e|flux|midjourney|video)/.test(name)
  const capabilities: MarketplaceModelCapability[] = []

  if (/(vision|multimodal|omni|gpt-4o|gpt-5|gemini|qwen.*vl|claude-(3|4|sonnet|opus))/.test(name)) {
    capabilities.push('vision')
  }
  if (!imageModel && /^(openai|anthropic|gemini|vertex|bedrock|azure|deepseek|dashscope)$/.test(platform)) {
    capabilities.push('function_calling')
  }
  if (/(reason|thinking|deepseek-r1|(^|[-_.])o[134]([-_.]|$)|opus|pro|max)/.test(name)) {
    capabilities.push('reasoning')
  }
  if (entry.pricing?.cache_read_price != null || entry.pricing?.cache_write_price != null) {
    capabilities.push('prompt_caching')
  }
  if (imageModel) capabilities.push('image_generation')

  return [...new Set(capabilities)]
}

export function recentMonitorStatuses(
  points: UserModelMonitorTimelinePoint[] | null | undefined,
  limit = 3,
  fallbackStatus: MarketplaceMonitorSignalStatus = '',
): MarketplaceMonitorSignalStatus[] {
  const safeLimit = Math.max(0, Math.floor(limit))
  const recent = [...(points ?? [])]
    .slice(0, safeLimit)
    .reverse()
    .map(point => point.status)
  if (recent.length === 0 && safeLimit > 0 && fallbackStatus) recent.push(fallbackStatus)
  const empty = Array<MarketplaceMonitorSignalStatus>(Math.max(0, safeLimit - recent.length)).fill('')
  return [...empty, ...recent]
}

export function buildMarketplaceEntries(platforms: UserMarketplacePlatform[]): MarketplaceModelEntry[] {
  return platforms.flatMap((section) =>
    section.supported_models.map((model) => ({
      key: `${section.platform}::${model.name}`,
      name: model.name,
      platform: section.platform || model.platform,
      groups: model.groups,
      pricing: model.pricing,
      capabilities: model.capabilities ?? [],
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

export function compareMarketplaceProviders(a: string, b: string): number {
  const aPriority = MARKETPLACE_PROVIDER_PRIORITY.indexOf(a.toLowerCase() as typeof MARKETPLACE_PROVIDER_PRIORITY[number])
  const bPriority = MARKETPLACE_PROVIDER_PRIORITY.indexOf(b.toLowerCase() as typeof MARKETPLACE_PROVIDER_PRIORITY[number])
  const normalizedAPriority = aPriority === -1 ? MARKETPLACE_PROVIDER_PRIORITY.length : aPriority
  const normalizedBPriority = bPriority === -1 ? MARKETPLACE_PROVIDER_PRIORITY.length : bPriority
  return normalizedAPriority - normalizedBPriority || a.localeCompare(b)
}

function marketplaceModelVersion(modelName: string): [number, number] | null {
  const normalized = modelName.trim().toLowerCase()
  const patterns = [
    /^gpt-(\d+)(?:[.-](\d+))?/,
    /^claude-[a-z]+-(\d+)(?:-(\d+))?/,
    /^gemini-(\d+)(?:[.-](\d+))?/,
    /^(?:grok|glm)-(\d+)(?:[.-](\d+))?/,
    /^deepseek-v(\d+)(?:[.-](\d+))?/,
    /^kimi-k(\d+)(?:[.-](\d+))?/,
    /^minimax-m(\d+)(?:[.-](\d+))?/,
    /^qwen(\d+)(?:[.-](\d+))?/,
  ]
  for (const pattern of patterns) {
    const match = normalized.match(pattern)
    if (match) return [Number(match[1]), Number(match[2] ?? 0)]
  }
  return null
}

export function compareMarketplaceModelRecency(a: string, b: string): number {
  const aVersion = marketplaceModelVersion(a)
  const bVersion = marketplaceModelVersion(b)
  if (!aVersion && !bVersion) return 0
  if (!aVersion) return 1
  if (!bVersion) return -1
  return bVersion[0] - aVersion[0] || bVersion[1] - aVersion[1]
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

export function realtimeRate(
  groupRate: number,
  balanceRechargeMultiplier: number,
  usdToCnyRate: number,
): number {
  const rechargeMultiplier = Number.isFinite(balanceRechargeMultiplier) && balanceRechargeMultiplier > 0
    ? balanceRechargeMultiplier
    : 1
  const exchangeRate = Number.isFinite(usdToCnyRate) && usdToCnyRate > 0
    ? usdToCnyRate
    : DEFAULT_USD_TO_CNY_RATE
  return groupRate / rechargeMultiplier / exchangeRate
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
