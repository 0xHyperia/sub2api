import { describe, expect, it } from 'vitest'

import {
  BILLING_MODE_IMAGE,
  BILLING_MODE_PER_REQUEST,
  BILLING_MODE_VIDEO,
  getBillingModeBadgeClass
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
