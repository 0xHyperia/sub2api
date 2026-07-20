import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const testDir = dirname(fileURLToPath(import.meta.url))
const redeemSource = readFileSync(resolve(testDir, '../RedeemView.vue'), 'utf8')
const redeemFormSource = readFileSync(resolve(testDir, '../../../components/payment/RedeemCodeForm.vue'), 'utf8')
const affiliateSource = readFileSync(resolve(testDir, '../AffiliateView.vue'), 'utf8')
const supportSource = readFileSync(resolve(testDir, '../SupportTicketsView.vue'), 'utf8')
const conversationSource = readFileSync(resolve(testDir, '../../../components/tickets/TicketConversation.vue'), 'utf8')

describe('redeem, affiliate, and support mobile experience', () => {
  it('keeps redemption as the primary phone action with semantic result surfaces', () => {
    expect(redeemSource).toContain('<RedeemCodeForm :show-description="true"')
    expect(redeemFormSource).toContain('class="flex min-w-0 flex-col gap-2 sm:flex-row"')
    expect(redeemFormSource).toContain('class="btn btn-primary shrink-0 sm:min-w-28"')
    expect(redeemFormSource).toContain('bg-success-subtle')
    expect(redeemFormSource).toContain('bg-danger-subtle')
  })

  it('prioritizes transferable affiliate quota and summarizes invitees on phones', () => {
    expect(affiliateSource).toContain('order-first col-span-2')
    expect(affiliateSource).toContain('xl:order-none xl:col-span-1')
    expect(affiliateSource).toContain('<template #mobile-card="{ row }">')
    expect(affiliateSource).toContain('formatCurrency(row.total_rebate)')
    expect(affiliateSource).toContain('formatDateTime(row.created_at)')
    expect(affiliateSource).toContain('btn btn-primary w-full shrink-0 sm:w-auto')
  })

  it('presents tickets as a conversation inbox with compact phone controls', () => {
    expect(supportSource).toContain('grid-cols-[minmax(0,1fr)_auto]')
    expect(supportSource).toContain('class="btn btn-primary shrink-0"')
    expect(supportSource).toContain('grid-cols-[minmax(0,1fr)_104px_40px]')
    expect(supportSource).toContain('ticket.user_unread_count')
    expect(supportSource).toContain('name="chevronRight"')
    expect(supportSource).toContain('router.push(`/support/${ticket.number}`)')
  })

  it('keeps the conversation scrollable while the reply composer remains visible', () => {
    expect(conversationSource).toContain('data-test="ticket-message-viewport"')
    expect(conversationSource).toContain('data-test="ticket-composer"')
    expect(conversationSource).toContain('min-h-0 flex-1')
    expect(conversationSource).toContain('setInterval(() =>')
    expect(conversationSource).toContain('}, 5000)')
    expect(conversationSource).toContain("localText('客服', 'Support')")
    expect(conversationSource).toContain("localText('用户', 'Customer')")
  })
})
