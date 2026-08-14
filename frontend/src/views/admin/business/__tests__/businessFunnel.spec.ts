import { describe, expect, it } from 'vitest'
import { businessFunnelRate } from '../businessFunnel'

const steps = [
  { key: 'registered', count: 35 },
  { key: 'activated', count: 7 },
  { key: 'first_paid', count: 8 },
  { key: 'repurchased', count: 3 },
]

describe('businessFunnelRate', () => {
  it('uses registrations for first-use and first-payment conversions', () => {
    expect(businessFunnelRate(steps, 'activated', 7)).toEqual({ base: 'registered', rate: 20 })
    expect(businessFunnelRate(steps, 'first_paid', 8)).toEqual({
      base: 'registered',
      rate: 8 * 100 / 35,
    })
  })

  it('uses first-paid users for repurchase conversion', () => {
    expect(businessFunnelRate(steps, 'repurchased', 3)).toEqual({
      base: 'first_paid',
      rate: 37.5,
    })
  })
})
