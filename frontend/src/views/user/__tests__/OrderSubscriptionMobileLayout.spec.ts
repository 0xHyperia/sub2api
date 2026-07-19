import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const testDir = dirname(fileURLToPath(import.meta.url))
const ordersSource = readFileSync(resolve(testDir, '../UserOrdersView.vue'), 'utf8')
const orderTableSource = readFileSync(resolve(testDir, '../../../components/payment/OrderTable.vue'), 'utf8')
const subscriptionsSource = readFileSync(resolve(testDir, '../SubscriptionsView.vue'), 'utf8')

describe('orders and subscriptions mobile layouts', () => {
  it('opts the user order page into a business summary without changing admin defaults', () => {
    expect(ordersSource).toContain('<OrderTable :orders="orders" :loading="loading" user-mobile-card>')
    expect(orderTableSource).toContain('<template v-if="userMobileCard" #mobile-card="{ row }">')
    expect(orderTableSource).toContain('userMobileCard?: boolean')
    expect(orderTableSource).toContain('orderTypeLabel(row)')
    expect(orderTableSource).toContain('paymentMethodLabel(row.payment_type)')
    expect(orderTableSource).toContain('row.pay_amount.toFixed(2)')
    expect(orderTableSource).toContain('formatDate(row.created_at)')
  })

  it('provides an inline order detail action and preserves transactional actions', () => {
    expect(orderTableSource).toContain('toggleMobileDetails(row)')
    expect(orderTableSource).toContain('isMobileExpanded(row)')
    expect(orderTableSource).toContain('<slot name="actions" :row="row" />')
    expect(ordersSource).toContain("row.status === 'PENDING' || canRequestRefund(row)")
    expect(ordersSource).toContain('handleCancel(row.id)')
    expect(ordersSource).toContain('openRefundDialog(row)')
  })

  it('keeps the existing subscription quota card and places purchase beside the phone title', () => {
    expect(subscriptionsSource).toContain('subscription.group?.name')
    expect(subscriptionsSource).toContain('formatExpirationDate(subscription.expires_at)')
    expect(subscriptionsSource).toContain('getProgressWidth(subscription.daily_usage_usd')
    expect(subscriptionsSource).toContain('getProgressWidth(subscription.weekly_usage_usd')
    expect(subscriptionsSource).toContain('getProgressWidth(subscription.monthly_usage_usd')
    expect(subscriptionsSource).toContain('page-header mb-0 flex items-start justify-between gap-3')
    expect(subscriptionsSource).toContain('btn btn-primary btn-icon shrink-0')
    expect(subscriptionsSource).toContain('btn btn-secondary btn-sm w-full self-start sm:w-auto')
  })
})
