import { describe, expect, it } from 'vitest'
import type { UserMarketplacePlatform } from '@/api/channels'
import {
  buildMarketplaceEntries,
  buildMarketplaceGroups,
  billingCategory,
  compareMarketplaceDisplayOrder,
  effectiveRateForEntry,
  inferMarketplaceModelCapabilities,
  recentMonitorStatuses,
  realtimeRate,
  scaledPrice,
  sortedEntryGroups,
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
    expect(scaledPrice(0.000002, 1_000_000, 1.1)).toBe('$2.2')
    expect(scaledPrice(null, 1_000_000, 1)).toBe('-')
  })

  it('converts the effective group rate using the current balance recharge rate', () => {
    expect(realtimeRate(0.1, 1, 7.2)).toBeCloseTo(0.0138889)
    expect(realtimeRate(0.15, 1.5, 7.2)).toBeCloseTo(0.0138889)
    expect(realtimeRate(0.1, 0, 0)).toBeCloseTo(0.0138889)
  })

  it('maps technical billing modes to the two marketplace categories', () => {
    const tokenPricing = platforms[0].supported_models[0].pricing
    expect(billingCategory(tokenPricing)).toBe('usage')
    expect(billingCategory(tokenPricing ? { ...tokenPricing, billing_mode: 'image' } : null)).toBe('request')
    expect(billingCategory(tokenPricing ? { ...tokenPricing, billing_mode: 'per_request' } : null)).toBe('request')
    expect(billingCategory(null)).toBe('unpriced')
  })

  it('derives conservative capability markers from model metadata', () => {
    const pricing = platforms[0].supported_models[0].pricing
    expect(inferMarketplaceModelCapabilities({ name: 'gpt-5-codex-mini', platform: 'openai', pricing })).toEqual([
      'chat',
      'tools',
      'vision',
      'coding',
      'fast',
    ])
    expect(inferMarketplaceModelCapabilities({
      name: 'gpt-image-2',
      platform: 'openai',
      pricing: pricing ? { ...pricing, billing_mode: 'image' } : null,
    })).toEqual(['image'])
    expect(inferMarketplaceModelCapabilities({ name: 'deepseek-r1', platform: 'deepseek', pricing })).toEqual([
      'chat',
      'tools',
      'reasoning',
    ])
  })

  it('builds a three-point compact monitor signal from the latest checks', () => {
    const points = [
      { status: 'failed' as const, latency_ms: 300, checked_at: '2026-07-16T03:00:00Z' },
      { status: 'degraded' as const, latency_ms: 200, checked_at: '2026-07-16T02:00:00Z' },
      { status: 'operational' as const, latency_ms: 100, checked_at: '2026-07-16T01:00:00Z' },
      { status: 'operational' as const, latency_ms: 90, checked_at: '2026-07-16T00:00:00Z' },
    ]
    expect(recentMonitorStatuses(points)).toEqual(['operational', 'degraded', 'failed'])
    expect(recentMonitorStatuses(points.slice(0, 1))).toEqual(['', '', 'failed'])
    expect(recentMonitorStatuses([], 3, 'operational')).toEqual(['', '', 'operational'])
  })
})
