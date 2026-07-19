import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'

import type { PaymentOrder } from '@/types/payment'
import AdminOrderTable from '../AdminOrderTable.vue'

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key })
  }
})

const DataTableStub = {
  props: ['data'],
  template: `
    <div>
      <div v-for="row in data" :key="row.id">
        <slot name="mobile-card" :row="row" :index="0" :selected="false" :expanded="false" />
        <div :data-test="'desktop-order-status-' + row.id"><slot name="cell-status" :value="row.status" :row="row" /></div>
      </div>
    </div>
  `
}

function order(overrides: Partial<PaymentOrder>): PaymentOrder {
  return {
    id: 1,
    user_id: 12,
    amount: 100,
    pay_amount: 108,
    currency: 'CNY',
    fee_rate: 8,
    payment_type: 'stripe',
    out_trade_no: 'order-1',
    status: 'COMPLETED',
    order_type: 'subscription',
    created_at: '2026-07-19T02:00:00Z',
    expires_at: '2026-07-19T02:30:00Z',
    refund_amount: 0,
    ...overrides
  }
}

function mountTable() {
  return mount(AdminOrderTable, {
    props: {
      orders: [
        order({ id: 1, status: 'COMPLETED' }),
        order({ id: 2, status: 'PENDING', order_type: 'balance', payment_type: 'alipay' }),
        order({ id: 3, status: 'FAILED', payment_type: 'wxpay' })
      ],
      loading: false,
      page: 1,
      pageSize: 20,
      total: 3
    },
    global: {
      stubs: {
        DataTable: DataTableStub,
        Pagination: true,
        Select: true,
        Icon: true
      }
    }
  })
}

describe('AdminOrderTable mobile cards', () => {
  it('prioritizes order identity, user, amount, payment method, status and time', () => {
    const wrapper = mountTable()
    const card = wrapper.get('[data-test="admin-order-mobile-card-1"]')

    expect(card.text()).toContain('payment.admin.subscriptionOrder')
    expect(card.text()).toContain('payment.status.completed')
    expect(card.text()).toContain('#1')
    expect(card.text()).toContain('#12')
    expect(card.text()).toContain('¥108.00')
    expect(card.text()).toContain('$100.00')
    expect(card.text()).toContain('payment.methods.stripe')
    expect(card.text()).toContain('2026')
    expect(wrapper.get('[data-test="desktop-order-status-1"]').text()).toContain('payment.status.completed')
  })

  it('preserves view and status-specific cancel, retry and refund actions', async () => {
    const wrapper = mountTable()

    await wrapper.get('[data-test="admin-order-mobile-detail-1"]').trigger('click')
    await wrapper.get('[data-test="admin-order-mobile-refund-1"]').trigger('click')
    await wrapper.get('[data-test="admin-order-mobile-cancel-2"]').trigger('click')
    await wrapper.get('[data-test="admin-order-mobile-retry-3"]').trigger('click')

    expect(wrapper.emitted('detail')?.[0]?.[0]).toMatchObject({ id: 1 })
    expect(wrapper.emitted('refund')?.[0]?.[0]).toMatchObject({ id: 1 })
    expect(wrapper.emitted('cancel')?.[0]?.[0]).toMatchObject({ id: 2 })
    expect(wrapper.emitted('retry')?.[0]?.[0]).toMatchObject({ id: 3 })
  })
})
