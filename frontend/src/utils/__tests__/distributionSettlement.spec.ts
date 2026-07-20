import { describe, expect, it } from 'vitest'
import { calculatePlatformCredit, calculateWithdrawalPreview } from '../distributionSettlement'

describe('distribution settlement calculations', () => {
  it('previews percentage and fixed withdrawal fees', () => {
    expect(calculateWithdrawalPreview(100, 100, 2)).toEqual({ fee: 3, payout: 97 })
  })

  it('never reports a negative payout', () => {
    expect(calculateWithdrawalPreview(1, 0, 2)).toEqual({ fee: 2, payout: 0 })
  })

  it('converts commission using the platform CNY rate', () => {
    expect(calculatePlatformCredit(100, 2)).toBe(50)
    expect(calculatePlatformCredit(100, 0)).toBe(100)
  })
})
