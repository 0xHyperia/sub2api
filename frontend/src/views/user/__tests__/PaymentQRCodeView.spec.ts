import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

const routeState = vi.hoisted(() => ({
  query: {} as Record<string, unknown>,
}))
const routerPush = vi.hoisted(() => vi.fn())
const pollOrderStatus = vi.hoisted(() => vi.fn())
const cancelOrder = vi.hoisted(() => vi.fn())
const toCanvas = vi.hoisted(() => vi.fn())

vi.mock('vue-router', () => ({
  useRoute: () => routeState,
  useRouter: () => ({ push: routerPush }),
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  }
})

vi.mock('@/stores/payment', () => ({
  usePaymentStore: () => ({
    pollOrderStatus,
  }),
}))

vi.mock('@/stores', () => ({
  useAppStore: () => ({
    showError: vi.fn(),
  }),
}))

vi.mock('@/api/payment', () => ({
  paymentAPI: {
    cancelOrder,
  },
}))

vi.mock('qrcode', () => ({
  default: {
    toCanvas,
  },
}))

import PaymentQRCodeView from '../PaymentQRCodeView.vue'

function mountView() {
  return mount(PaymentQRCodeView, {
    global: {
      stubs: {
        AppLayout: {
          template: '<main><slot /></main>',
        },
        Icon: true,
      },
    },
  })
}

describe('PaymentQRCodeView', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    routeState.query = {}
    routerPush.mockReset()
    pollOrderStatus.mockReset()
    cancelOrder.mockReset()
    toCanvas.mockReset()
    toCanvas.mockResolvedValue(undefined)
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('shows a missing-parameter error without starting a payment wait loop', async () => {
    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.text()).toContain('payment.qr.missingParams')
    expect(wrapper.find('canvas').exists()).toBe(false)

    await vi.advanceTimersByTimeAsync(60_000)
    expect(pollOrderStatus).not.toHaveBeenCalled()

    await wrapper.get('button').trigger('click')
    expect(routerPush).toHaveBeenCalledWith('/purchase')
  })

  it('keeps the QR canvas responsive inside the narrow payment surface', async () => {
    routeState.query = {
      order_id: '42',
      qr: 'https://pay.example.com/42',
      payment_type: 'custom-pay',
      expires_at: '2099-01-01T00:00:00.000Z',
    }

    const wrapper = mountView()
    await flushPromises()

    const canvas = wrapper.get('canvas')
    expect(canvas.classes()).toContain('w-full')
    expect(canvas.classes()).toContain('max-w-64')
    expect(canvas.attributes('role')).toBe('img')
    expect(toCanvas).toHaveBeenCalled()

    wrapper.unmount()
  })

  it('surfaces repeated status failures and lets the user retry', async () => {
    routeState.query = {
      order_id: '42',
      pay_url: 'https://pay.example.com/42',
      payment_type: 'custom-pay',
      expires_at: '2099-01-01T00:00:00.000Z',
    }
    pollOrderStatus.mockRejectedValue(new Error('temporarily unavailable'))

    const wrapper = mountView()
    await flushPromises()
    await vi.advanceTimersByTimeAsync(9000)
    await flushPromises()

    expect(pollOrderStatus).toHaveBeenCalledTimes(3)
    expect(wrapper.text()).toContain('payment.qr.statusUnavailable')

    pollOrderStatus.mockResolvedValue({ status: 'PENDING' })
    const retryButton = wrapper.findAll('button').find(button => button.text().includes('common.retry'))
    expect(retryButton).toBeDefined()
    await retryButton!.trigger('click')
    await flushPromises()

    expect(pollOrderStatus).toHaveBeenCalledTimes(4)
    expect(wrapper.text()).not.toContain('payment.qr.statusUnavailable')
    wrapper.unmount()
  })
})
