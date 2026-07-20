import { flushPromises, mount } from '@vue/test-utils'
import { ref } from 'vue'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import type { Ticket, TicketCategory, TicketMessage } from '@/types'
import TicketConversation from '../TicketConversation.vue'

const { getTicket, markRead, refreshNotifications } = vi.hoisted(() => ({
  getTicket: vi.fn(),
  markRead: vi.fn(),
  refreshNotifications: vi.fn()
}))

vi.mock('@/api/tickets', () => ({
  ticketsAPI: {
    get: getTicket,
    markRead,
    capabilities: vi.fn().mockResolvedValue({
      attachments_available: false,
      max_file_bytes: 10 * 1024 * 1024,
      max_files_per_message: 5,
      max_total_bytes: 25 * 1024 * 1024,
      allowed_extensions: []
    }),
    reply: vi.fn(),
    close: vi.fn(),
    reopen: vi.fn(),
    attachmentURL: vi.fn()
  }
}))

vi.mock('@/api/admin/tickets', () => ({
  adminTicketsAPI: {
    get: vi.fn(),
    markRead: vi.fn(),
    reply: vi.fn(),
    close: vi.fn(),
    reopen: vi.fn()
  }
}))

vi.mock('@/stores/ticketNotifications', () => ({
  useTicketNotificationStore: () => ({ refresh: refreshNotifications })
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ showError: vi.fn(), showSuccess: vi.fn() })
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return { ...actual, useI18n: () => ({ locale: ref('zh-CN') }) }
})

const category: TicketCategory = {
  id: 1,
  code: 'account',
  name_zh: '账户问题',
  name_en: 'Account',
  active: true,
  sort_order: 10,
  created_at: '2026-07-20T00:00:00Z',
  updated_at: '2026-07-20T00:00:00Z'
}

const message = (id: number, sender_type: TicketMessage['sender_type'], content: string): TicketMessage => ({
  id,
  ticket_id: 7,
  sender_type,
  event_type: '',
  content,
  attachments: [],
  created_at: `2026-07-20T00:0${id}:00Z`
})

const ticket = (messages: TicketMessage[]): Ticket => ({
  id: 7,
  number: '20260720-TEST',
  user_id: 42,
  user_email: 'customer@example.com',
  category_id: category.id,
  category,
  subject: '测试工单',
  status: 'open',
  user_unread_count: 0,
  admin_unread_count: 1,
  last_actor_type: 'user',
  last_message_at: '2026-07-20T00:02:00Z',
  messages,
  created_at: '2026-07-20T00:00:00Z',
  updated_at: '2026-07-20T00:02:00Z'
})

describe('TicketConversation', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    getTicket.mockReset()
    markRead.mockReset().mockResolvedValue(undefined)
    refreshNotifications.mockReset()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('uses role labels and reveals new messages after the five-second refresh', async () => {
    getTicket
      .mockResolvedValueOnce(ticket([message(1, 'user', '需要帮助'), message(2, 'admin', '正在处理')]))
      .mockResolvedValueOnce(ticket([message(1, 'user', '需要帮助'), message(2, 'admin', '正在处理'), message(3, 'admin', '请查看最新回复')]))

    const wrapper = mount(TicketConversation, {
      props: { number: '20260720-TEST' },
      global: { stubs: { Icon: true } }
    })
    await flushPromises()

    expect(wrapper.text()).toContain('用户')
    expect(wrapper.text()).toContain('客服')
    expect(wrapper.get('[data-test="ticket-composer"]').exists()).toBe(true)

    const viewport = wrapper.get('[data-test="ticket-message-viewport"]').element as HTMLElement
    Object.defineProperties(viewport, {
      scrollHeight: { configurable: true, value: 1000 },
      clientHeight: { configurable: true, value: 300 }
    })
    viewport.scrollTop = 100

    await vi.advanceTimersByTimeAsync(5000)
    await flushPromises()

    expect(getTicket).toHaveBeenCalledTimes(2)
    expect(wrapper.get('[data-test="new-message-divider"]').text()).toContain('以下为新消息')
    expect(wrapper.get('[data-test="jump-to-latest"]').text()).toContain('跳到最新消息（1）')
  })
})
