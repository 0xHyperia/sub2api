import { describe, expect, it } from 'vitest'
import type { UserMarketplacePlatform } from '@/api/channels'
import {
  buildMarketplaceEntries,
  buildMarketplaceGroups,
  billingCategory,
  compareMarketplaceDisplayOrder,
  compareMarketplaceModelRecency,
  compareMarketplaceProviders,
  effectiveRateForEntry,
  imagePriceRows,
  inferMarketplaceModelCapabilities,
  modelAvailabilityBarClass,
  officialDiscount,
  recentModelSuccessRates,
  realtimeRate,
  scaledCurrencyPrice,
  scaledRechargePrice,
  scaledPrice,
  sortedEntryGroups,
  visibleMarketplaceGroupCount,
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
        image_rate_independent: false,
        image_rate_multiplier: 1,
        image_price_1k: null,
        image_price_2k: null,
        image_price_4k: null,
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
        image_rate_independent: false,
        image_rate_multiplier: 1,
        image_price_1k: null,
        image_price_2k: null,
        image_price_4k: null,
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
        image_input_price: null,
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

  it('maps presentation metadata and sorts higher priorities first', () => {
    const featured = buildMarketplaceEntries([{
      ...platforms[0],
      supported_models: [{
        ...platforms[0].supported_models[0],
        monitor_status: {
          status: '',
          latency_ms: null,
          availability_7d: null,
          last_checked_at: null,
          timeline: [],
          display_order: 100,
          label: 'New',
        },
      }],
    }])[0]
    const regular = { ...featured, key: 'openai::regular', name: 'regular', displayOrder: 0, label: '' }
    expect(featured).toMatchObject({ displayOrder: 100, label: 'New' })
    expect([regular, featured].sort(compareMarketplaceDisplayOrder).map(entry => entry.name)).toEqual(['gpt-test', 'regular'])
  })

  it('prioritizes OpenAI, Anthropic, and Gemini providers in that order', () => {
    expect(['vertex', 'gemini', 'anthropic', 'openai', 'azure'].sort(compareMarketplaceProviders)).toEqual([
      'openai',
      'anthropic',
      'gemini',
      'azure',
      'vertex',
    ])
    expect(['gemini', 'claude', 'openai'].sort(compareMarketplaceProviders)).toEqual([
      'openai',
      'claude',
      'gemini',
    ])
  })

  it('sorts each model family from newest generation to oldest', () => {
    expect(['gpt-5.2', 'gpt-5.6-sol', 'gpt-5.4-mini', 'gpt-5.5'].sort(compareMarketplaceModelRecency)).toEqual([
      'gpt-5.6-sol',
      'gpt-5.5',
      'gpt-5.4-mini',
      'gpt-5.2',
    ])
    expect(['claude-haiku-4-5-20251001', 'claude-sonnet-5', 'claude-opus-4-8'].sort(compareMarketplaceModelRecency)).toEqual([
      'claude-sonnet-5',
      'claude-opus-4-8',
      'claude-haiku-4-5-20251001',
    ])
    expect(['gemini-2.5-pro', 'gemini-3.5-flash', 'gemini-3.1-pro-preview'].sort(compareMarketplaceModelRecency)).toEqual([
      'gemini-3.5-flash',
      'gemini-3.1-pro-preview',
      'gemini-2.5-pro',
    ])
  })

  it('uses user-specific rates and the lowest accessible rate by default', () => {
    const entries = buildMarketplaceEntries([{ ...platforms[0], supported_models: [{ ...platforms[0].supported_models[0], groups: platforms[0].groups }] }])
    const rates = { 1: 1.1 }
    expect(buildMarketplaceGroups(entries, rates).find((group) => group.id === 1)?.effectiveRate).toBe(1.1)
    expect(effectiveRateForEntry(entries[0], null, rates)).toBe(0.9)
    expect(effectiveRateForEntry(entries[0], 1, rates)).toBe(1.1)
    expect(buildMarketplaceGroups(entries, rates).map(group => group.id)).toEqual([2, 1])
    expect(sortedEntryGroups(entries[0], rates).map(group => group.id)).toEqual([2, 1])
  })

  it('formats prices after applying the selected group rate', () => {
    expect(scaledPrice(0.000002, 1_000_000, 1.1)).toBe('$2.20')
    expect(scaledPrice(0.00000002, 1_000_000, 0.1)).toBe('$0.002')
    expect(scaledPrice(0, 1_000_000, 0.1)).toBe('$0.00')
    expect(scaledRechargePrice(0.00000002, 1_000_000, 0.1, 1)).toBe('¥0.002')
    expect(scaledCurrencyPrice(0.00000002, 1_000_000, 7.2)).toBe('¥0.144')
    expect(scaledPrice(null, 1_000_000, 1)).toBe('-')
  })

  it('formats top-up prices using the configured balance multiplier', () => {
    expect(scaledRechargePrice(0.0000072, 1_000_000, 1, 1)).toBe('¥7.20')
    expect(scaledRechargePrice(1, 1, 1, 0.2)).toBe('¥5.00')
    expect(scaledRechargePrice(1, 1, 1, 0)).toBe('¥1.00')
    expect(scaledRechargePrice(null, 1, 1, 0.2)).toBe('-')
  })

  it('formats official prices using the currency exchange rate', () => {
    expect(scaledCurrencyPrice(0.000001, 1_000_000, 7.2)).toBe('¥7.20')
    expect(scaledCurrencyPrice(1, 1, 0)).toBe('¥7.20')
    expect(scaledCurrencyPrice(null, 1, 7.2)).toBe('-')
  })

  it('uses configured image tier prices with the user-specific group rate', () => {
    const pricing = {
      ...platforms[0].supported_models[0].pricing!,
      billing_mode: 'image' as const,
      per_request_price: 0.03,
    }
    const entry = buildMarketplaceEntries([{
      ...platforms[0],
      supported_models: [{
        ...platforms[0].supported_models[0],
        pricing,
        groups: [{
          ...platforms[0].groups[0],
          image_price_1k: 0.04,
          image_price_2k: 0.08,
          image_price_4k: 0.12,
        }],
      }],
    }])[0]
    const group = sortedEntryGroups(entry, { 1: 0.75 })[0]

    expect(group.effectiveRate).toBe(0.75)
    expect(imagePriceRows(pricing, group)).toEqual([
      { key: 'image-1k', tier: '1K', rawValue: 0.04, effectiveValue: 0.03, usesGroupPrice: true },
      { key: 'image-2k', tier: '2K', rawValue: 0.08, effectiveValue: 0.06, usesGroupPrice: true },
      { key: 'image-4k', tier: '4K', rawValue: 0.12, effectiveValue: 0.09, usesGroupPrice: true },
    ])
  })

  it('uses the independent image multiplier instead of user and group rates', () => {
    const pricing = {
      ...platforms[0].supported_models[0].pricing!,
      billing_mode: 'image' as const,
      per_request_price: 0.05,
    }
    const entry = buildMarketplaceEntries([{
      ...platforms[0],
      supported_models: [{
        ...platforms[0].supported_models[0],
        pricing,
        groups: [{
          ...platforms[0].groups[0],
          image_rate_independent: true,
          image_rate_multiplier: 0.5,
        }],
      }],
    }])[0]
    const group = sortedEntryGroups(entry, { 1: 0.25 })[0]

    expect(group.effectiveRate).toBe(0.5)
    expect(effectiveRateForEntry(entry, 1, { 1: 0.25 })).toBe(0.5)
    expect(imagePriceRows(pricing, group).map(row => row.effectiveValue)).toEqual([0.025, 0.025, 0.025])
  })

  it('converts the effective group rate using the current balance recharge rate', () => {
    expect(realtimeRate(0.1, 1, 7.2)).toBeCloseTo(0.0138889)
    expect(realtimeRate(0.15, 1.5, 7.2)).toBeCloseTo(0.0138889)
    expect(realtimeRate(0.1, 0, 0)).toBeCloseTo(0.0138889)
  })

  it('converts an official-price multiplier to the Chinese discount scale', () => {
    expect(officialDiscount(0.1483)).toBeCloseTo(1.483)
  })

  it('reserves room for the overflow control before deciding which group chips fit', () => {
    expect(visibleMarketplaceGroupCount([50, 50, 50], 161, 30)).toBe(2)
    expect(visibleMarketplaceGroupCount([50, 50, 50], 130, 30)).toBe(1)
    expect(visibleMarketplaceGroupCount([50, 50, 50], 80, 30)).toBe(0)
    expect(visibleMarketplaceGroupCount([50, 50], 106, 30)).toBe(2)
  })

  it('maps technical billing modes to the two marketplace categories', () => {
    const tokenPricing = platforms[0].supported_models[0].pricing
    expect(billingCategory(tokenPricing)).toBe('usage')
    expect(billingCategory(tokenPricing ? { ...tokenPricing, billing_mode: 'image' } : null)).toBe('request')
    expect(billingCategory(tokenPricing ? { ...tokenPricing, billing_mode: 'per_request' } : null)).toBe('request')
    expect(billingCategory(null)).toBe('unpriced')
  })

  it('prefers declared catalog capabilities in the stable display order', () => {
    const pricing = platforms[0].supported_models[0].pricing
    expect(inferMarketplaceModelCapabilities({
      name: 'catalog-model',
      platform: 'openai',
      pricing,
      capabilities: ['service_tier', 'vision', 'function_calling', 'prompt_caching'],
    })).toEqual(['vision', 'function_calling', 'prompt_caching', 'service_tier'])
  })

  it('derives conservative capability markers when catalog metadata is unavailable', () => {
    const pricing = platforms[0].supported_models[0].pricing
    expect(inferMarketplaceModelCapabilities({ name: 'gpt-5-codex-mini', platform: 'openai', pricing })).toEqual([
      'vision',
      'function_calling',
    ])
    expect(inferMarketplaceModelCapabilities({
      name: 'gpt-image-2',
      platform: 'openai',
      pricing: pricing ? { ...pricing, billing_mode: 'image' } : null,
    })).toEqual(['image_generation'])
    expect(inferMarketplaceModelCapabilities({ name: 'deepseek-r1', platform: 'deepseek', pricing })).toEqual([
      'function_calling',
      'reasoning',
    ])
  })

  it('builds a three-point compact monitor signal from the latest hourly buckets', () => {
    const buckets = [
      { started_at: '2026-07-16T01:00:00Z', success_rate: 99.9, ttft_ms: null },
      { started_at: '2026-07-16T02:00:00Z', success_rate: 90, ttft_ms: null },
      { started_at: '2026-07-16T03:00:00Z', success_rate: 69, ttft_ms: null },
    ]
    expect(recentModelSuccessRates(buckets)).toEqual([99.9, 90, 69])
    expect(recentModelSuccessRates(buckets.slice(0, 1))).toEqual([null, null, 99.9])
    expect(recentModelSuccessRates([])).toEqual([null, null, null])
    expect(recentModelSuccessRates(buckets, 0)).toEqual([])
  })

  it('uses the same lenient colors as model availability', () => {
    expect(modelAvailabilityBarClass(null)).toBe('bg-outline-strong')
    expect(modelAvailabilityBarClass(30)).toBe('bg-danger')
    expect(modelAvailabilityBarClass(30.1)).toBe('bg-warning')
    expect(modelAvailabilityBarClass(69.9)).toBe('bg-warning')
    expect(modelAvailabilityBarClass(70)).toBe('bg-success/70')
    expect(modelAvailabilityBarClass(90)).toBe('bg-success/70')
    expect(modelAvailabilityBarClass(99.9)).toBe('bg-success')
  })
})
