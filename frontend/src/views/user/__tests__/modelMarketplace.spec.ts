import { describe, expect, it } from 'vitest'
import type { UserMarketplacePlatform } from '@/api/channels'
import {
  buildMarketplaceEntries,
  buildMarketplaceGroups,
  effectiveRateForEntry,
  scaledPrice,
} from '../modelMarketplace'

const platforms: UserMarketplacePlatform[] = [{
    platform: 'openai',
    groups: [
      {
        id: 1,
        name: 'Default',
        platform: 'openai',
        subscription_type: 'standard',
        rate_multiplier: 1.2,
        peak_rate_enabled: false,
        peak_start: '',
        peak_end: '',
        peak_rate_multiplier: 1,
        is_exclusive: false,
      },
      {
        id: 2,
        name: 'VIP',
        platform: 'openai',
        subscription_type: 'standard',
        rate_multiplier: 0.9,
        peak_rate_enabled: false,
        peak_start: '',
        peak_end: '',
        peak_rate_multiplier: 1,
        is_exclusive: true,
      },
    ],
    supported_models: [{
      name: 'gpt-test',
      platform: 'openai',
      groups: [],
      pricing: {
        billing_mode: 'token',
        input_price: 0.000002,
        output_price: 0.000008,
        cache_write_price: null,
        cache_read_price: null,
        image_output_price: null,
        per_request_price: null,
        intervals: [],
      },
    }],
}]

describe('model marketplace data', () => {
  it('flattens platform sections into model cards', () => {
    const entries = buildMarketplaceEntries(platforms)
    expect(entries).toHaveLength(1)
    expect(entries[0]).toMatchObject({ name: 'gpt-test', platform: 'openai' })
  })

  it('uses user-specific rates and the lowest accessible rate by default', () => {
    const entries = buildMarketplaceEntries([{ ...platforms[0], supported_models: [{ ...platforms[0].supported_models[0], groups: platforms[0].groups }] }])
    const rates = { 1: 1.1 }
    expect(buildMarketplaceGroups(entries, rates).find((group) => group.id === 1)?.effectiveRate).toBe(1.1)
    expect(effectiveRateForEntry(entries[0], null, rates)).toBe(0.9)
    expect(effectiveRateForEntry(entries[0], 1, rates)).toBe(1.1)
  })

  it('formats prices after applying the selected group rate', () => {
    expect(scaledPrice(0.000002, 1_000_000, 1.1)).toBe('$2.2')
    expect(scaledPrice(null, 1_000_000, 1)).toBe('-')
  })
})
