import type { BusinessFunnelStep } from '@/api/businessAnalytics'

export type BusinessFunnelRateBase = 'registered' | 'first_paid'

export function businessFunnelRate(
  steps: BusinessFunnelStep[],
  key: string,
  value: number,
): { base: BusinessFunnelRateBase; rate: number | null } {
  const base: BusinessFunnelRateBase = key === 'repurchased' ? 'first_paid' : 'registered'
  const denominator = steps.find(step => step.key === base)?.count ?? 0
  return { base, rate: denominator > 0 ? (value * 100) / denominator : null }
}
