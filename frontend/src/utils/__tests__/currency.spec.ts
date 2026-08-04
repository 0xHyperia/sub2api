import { describe, expect, it } from 'vitest'
import { formatExchangeRate, roundHalfUp } from '../currency'

describe('currency formatting', () => {
  it('uses ROUND_HALF_UP at two decimal places', () => {
    expect(roundHalfUp(6.7549)).toBe(6.75)
    expect(roundHalfUp(6.755)).toBe(6.76)
    expect(roundHalfUp(1.005)).toBe(1.01)
  })

  it('always displays a valid exchange rate with two decimals', () => {
    expect(formatExchangeRate(6.7526)).toBe('6.75')
    expect(formatExchangeRate(6.755)).toBe('6.76')
    expect(formatExchangeRate(0)).toBe('—')
  })
})
