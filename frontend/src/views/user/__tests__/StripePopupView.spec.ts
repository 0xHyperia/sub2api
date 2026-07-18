import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

const routeState = vi.hoisted(() => ({
  query: {
    order_id: '42',
    method: 'alipay',
    amount: '88.00',
  } as Record<string, unknown>,
}))
const loadStripe = vi.hoisted(() => vi.fn())
const confirmAlipayPayment = vi.hoisted(() => vi.fn())

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRoute: () => routeState,
  }
})

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  }
})

vi.mock('@stripe/stripe-js/pure', () => ({
  loadStripe,
}))

import StripePopupView from '../StripePopupView.vue'

describe('StripePopupView', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    confirmAlipayPayment.mockReset().mockResolvedValue({})
    loadStripe.mockReset().mockResolvedValue({ confirmAlipayPayment })
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('clears the handshake timeout as soon as valid initialization arrives', async () => {
    const wrapper = mount(StripePopupView, {
      global: {
        stubs: {
          Icon: true,
        },
      },
    })

    window.dispatchEvent(new MessageEvent('message', {
      origin: window.location.origin,
      data: {
        type: 'STRIPE_POPUP_INIT',
        clientSecret: 'pi_secret',
        publishableKey: 'pk_test_123',
      },
    }))
    await flushPromises()

    expect(loadStripe).toHaveBeenCalledWith('pk_test_123')
    expect(confirmAlipayPayment).toHaveBeenCalled()
    expect(vi.getTimerCount()).toBe(0)

    await vi.advanceTimersByTimeAsync(15_000)
    expect(wrapper.text()).not.toContain('payment.stripePopup.timeout')
    wrapper.unmount()
  })
})
