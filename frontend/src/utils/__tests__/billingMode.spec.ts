import { describe, expect, it } from 'vitest'

import {
  BILLING_MODE_IMAGE,
  BILLING_MODE_PER_REQUEST,
  BILLING_MODE_TOKEN,
  BILLING_MODE_VIDEO,
  getBillingModeBadgeClass,
  getDisplayBillingMode,
  isImageUsage
} from '../billingMode'

describe('getBillingModeBadgeClass', () => {
  it.each([
    [BILLING_MODE_PER_REQUEST, 'bg-brand-subtle'],
    [BILLING_MODE_IMAGE, 'bg-info-subtle'],
    [BILLING_MODE_VIDEO, 'bg-warning-subtle'],
    ['token', 'bg-info-subtle']
  ])('uses a quiet semantic surface for %s', (mode, expectedClass) => {
    const classes = getBillingModeBadgeClass(mode)

    expect(classes).toContain(expectedClass)
    expect(classes).toContain('border-')
    expect(classes).not.toMatch(/(?:purple|pink|blue|amber)-\d+/)
  })
})

describe('billingMode helpers', () => {
  it('prefers explicit video mode over image_count', () => {
    expect(
      getDisplayBillingMode({ image_count: 1, billing_mode: BILLING_MODE_VIDEO })
    ).toBe(BILLING_MODE_VIDEO)
    expect(isImageUsage({ image_count: 1, billing_mode: BILLING_MODE_VIDEO })).toBe(false)
  })

  it('infers image when image_count set and mode missing', () => {
    expect(getDisplayBillingMode({ image_count: 2, billing_mode: null })).toBe(BILLING_MODE_IMAGE)
  })

  it('keeps token mode even with image_count', () => {
    expect(
      getDisplayBillingMode({ image_count: 1, billing_mode: BILLING_MODE_TOKEN })
    ).toBe(BILLING_MODE_TOKEN)
  })
})
