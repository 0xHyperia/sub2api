import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import type { PaymentOrder } from '@/types/payment'
import AdminOrdersView from '../AdminOrdersView.vue'

const {
  getOrders,
  getOrder,
  cancelOrder,
  retryRecharge,
  refundOrder,
  queryRefund,
  showError,
  showSuccess,
} = vi.hoisted(() => ({
  getOrders: vi.fn(),
  getOrder: vi.fn(),
  cancelOrder: vi.fn(),
  retryRecharge: vi.fn(),
  refundOrder: vi.fn(),
  queryRefund: vi.fn(),
  showError: vi.fn(),
  showSuccess: vi.fn(),
}))

vi.mock('@/api/admin/payment', () => {
  const paymentAPI = {
    getOrders,
    getOrder,
    cancelOrder,
    retryRecharge,
    refundOrder,
    queryRefund,
  }
  return { adminPaymentAPI: paymentAPI, default: paymentAPI }
})

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ showError, showSuccess }),
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key }),
  }
})

function order(id: number, outTradeNo: string): PaymentOrder {
  return {
    id,
    user_id: 7,
    amount: 10,
    pay_amount: 10,
    fee_rate: 0,
    payment_type: 'alipay',
    out_trade_no: outTradeNo,
    status: 'PENDING',
    order_type: 'balance',
    created_at: '2026-07-12T00:00:00Z',
    expires_at: '2026-07-12T01:00:00Z',
    refund_amount: 0,
  }
}

function ordersResponse(...items: PaymentOrder[]) {
  return { data: { items, total: items.length } }
}

function deferred<T>() {
  let resolve!: (value: T) => void
  let reject!: (reason?: unknown) => void
  const promise = new Promise<T>((resolvePromise, rejectPromise) => {
    resolve = resolvePromise
    reject = rejectPromise
  })
  return { promise, resolve, reject }
}

function mountView() {
  return mount(AdminOrdersView, {
    global: {
      stubs: {
        AppLayout: { template: '<main><slot /></main>' },
        TablePageLayout: {
          template: '<div><slot name="filters" /><slot name="table" /><slot name="pagination" /></div>',
        },
        OrderTable: {
          props: ['orders', 'loading'],
          template: `
            <section data-testid="orders-table" :data-loading="String(loading)">
              <article v-for="row in orders" :key="row.id">
                <span>{{ row.out_trade_no }}</span>
                <slot name="actions" :row="row" />
              </article>
            </section>
          `,
        },
        Select: true,
        Pagination: true,
        BaseDialog: true,
        ConfirmDialog: true,
        AdminRefundDialog: true,
        OrderStatusBadge: true,
        Icon: true,
      },
    },
  })
}

function activeConfirmDialog(wrapper: ReturnType<typeof mountView>) {
  const dialog = wrapper.findComponent(ConfirmDialog)
  if (!dialog.exists() || !dialog.props('show')) {
    throw new Error('active confirmation dialog not found')
  }
  return dialog
}

describe('AdminOrdersView', () => {
  beforeEach(() => {
    for (const mock of [
      getOrders,
      getOrder,
      cancelOrder,
      retryRecharge,
      refundOrder,
      queryRefund,
      showError,
      showSuccess,
    ]) {
      mock.mockReset()
    }
    cancelOrder.mockResolvedValue(undefined)
  })

  afterEach(() => {
    vi.clearAllTimers()
    vi.useRealTimers()
  })

  it('cancels an order only after the destructive confirmation is accepted', async () => {
    getOrders.mockResolvedValue(ordersResponse(order(41, 'pending-order')))
    const wrapper = mountView()
    await flushPromises()

    await wrapper.get('button[aria-label="payment.orders.cancel"]').trigger('click')
    let dialog = activeConfirmDialog(wrapper)
    expect(dialog.props('danger')).toBe(true)
    expect(dialog.props('message')).toBe('payment.confirmCancel')
    expect(cancelOrder).not.toHaveBeenCalled()

    dialog.vm.$emit('cancel')
    await flushPromises()
    expect(cancelOrder).not.toHaveBeenCalled()

    await wrapper.get('button[aria-label="payment.orders.cancel"]').trigger('click')
    dialog = activeConfirmDialog(wrapper)
    dialog.vm.$emit('confirm')
    await flushPromises()

    expect(cancelOrder).toHaveBeenCalledOnce()
    expect(cancelOrder).toHaveBeenCalledWith(41)
    expect(showSuccess).toHaveBeenCalledWith('payment.admin.orderCancelled')
    wrapper.unmount()
  })

  it('does not let a late response overwrite the newest filtered result', async () => {
    vi.useFakeTimers()
    const oldRequest = deferred<ReturnType<typeof ordersResponse>>()
    const newRequest = deferred<ReturnType<typeof ordersResponse>>()
    getOrders
      .mockImplementationOnce(() => oldRequest.promise)
      .mockImplementationOnce(() => newRequest.promise)

    const wrapper = mountView()
    await wrapper.get('input[type="search"]').setValue('new-filter')
    vi.advanceTimersByTime(300)
    await flushPromises()

    expect(getOrders).toHaveBeenNthCalledWith(2, expect.objectContaining({ keyword: 'new-filter' }))
    newRequest.resolve(ordersResponse(order(2, 'new-result')))
    await flushPromises()
    expect(wrapper.text()).toContain('new-result')
    expect(wrapper.get('[data-testid="orders-table"]').attributes('data-loading')).toBe('false')

    oldRequest.resolve(ordersResponse(order(1, 'old-result')))
    await flushPromises()
    expect(wrapper.text()).toContain('new-result')
    expect(wrapper.text()).not.toContain('old-result')
    wrapper.unmount()
  })

  it('keeps loading while a response for an outdated filter is discarded', async () => {
    vi.useFakeTimers()
    const oldRequest = deferred<ReturnType<typeof ordersResponse>>()
    const newRequest = deferred<ReturnType<typeof ordersResponse>>()
    getOrders
      .mockImplementationOnce(() => oldRequest.promise)
      .mockImplementationOnce(() => newRequest.promise)

    const wrapper = mountView()
    await wrapper.get('input[type="search"]').setValue('current-filter')

    oldRequest.resolve(ordersResponse(order(1, 'outdated-result')))
    await flushPromises()
    expect(wrapper.get('[data-testid="orders-table"]').attributes('data-loading')).toBe('true')
    expect(wrapper.text()).not.toContain('outdated-result')

    vi.advanceTimersByTime(300)
    await flushPromises()
    newRequest.resolve(ordersResponse(order(2, 'current-result')))
    await flushPromises()

    expect(wrapper.get('[data-testid="orders-table"]').attributes('data-loading')).toBe('false')
    expect(wrapper.text()).toContain('current-result')
    wrapper.unmount()
  })

  it('shows a persistent first-load error and recovers after retry', async () => {
    getOrders
      .mockRejectedValueOnce(new Error('orders unavailable'))
      .mockResolvedValueOnce(ordersResponse(order(9, 'recovered-order')))

    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.get('[role="alert"]').text()).toContain('orders unavailable')
    expect(wrapper.find('[data-testid="orders-table"]').exists()).toBe(false)

    await wrapper.get('.orders-load-error button').trigger('click')
    await flushPromises()

    expect(getOrders).toHaveBeenCalledTimes(2)
    expect(wrapper.find('[role="alert"]').exists()).toBe(false)
    expect(wrapper.text()).toContain('recovered-order')
    wrapper.unmount()
  })

  it('keeps existing orders visible when a refresh fails', async () => {
    getOrders
      .mockResolvedValueOnce(ordersResponse(order(3, 'cached-order')))
      .mockRejectedValueOnce(new Error('refresh unavailable'))

    const wrapper = mountView()
    await flushPromises()
    await wrapper.get('button[aria-label="common.refresh"]').trigger('click')
    await flushPromises()

    expect(wrapper.get('[role="alert"]').text()).toContain('refresh unavailable')
    expect(wrapper.text()).toContain('cached-order')
    expect(wrapper.get('[data-testid="orders-table"]').attributes('data-loading')).toBe('false')
    wrapper.unmount()
  })
})
