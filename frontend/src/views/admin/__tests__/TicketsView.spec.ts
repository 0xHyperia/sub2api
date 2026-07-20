import { flushPromises, mount } from '@vue/test-utils'
import { ref } from 'vue'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import type { Ticket, TicketCategory } from '@/types'
import TicketsView from '../TicketsView.vue'

const { listTickets, listCategories, routerPush } = vi.hoisted(() => ({
  listTickets: vi.fn(),
  listCategories: vi.fn(),
  routerPush: vi.fn()
}))

vi.mock('@/api/admin/tickets', () => ({
  adminTicketsAPI: {
    list: listTickets,
    categories: listCategories,
    updateCategory: vi.fn(),
    createCategory: vi.fn(),
    reorderCategories: vi.fn()
  }
}))

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: routerPush })
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({ locale: ref('zh-CN') })
  }
})

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ showError: vi.fn(), showSuccess: vi.fn() })
}))

const category: TicketCategory = {
  id: 3,
  code: 'billing',
  name_zh: '账单问题',
  name_en: 'Billing',
  active: true,
  sort_order: 10,
  created_at: '2026-07-19T00:00:00Z',
  updated_at: '2026-07-19T00:00:00Z'
}

const ticket: Ticket = {
  id: 7,
  number: 'T-20260719-7',
  user_id: 42,
  user_email: 'customer@example.com',
  user_name: 'Customer',
  category_id: category.id,
  category,
  subject: '充值到账后余额没有更新',
  status: 'open',
  user_unread_count: 0,
  admin_unread_count: 3,
  last_actor_type: 'user',
  last_message_at: '2026-07-19T03:30:00Z',
  created_at: '2026-07-19T03:00:00Z',
  updated_at: '2026-07-19T03:30:00Z'
}

const mountView = () => mount(TicketsView, {
  global: {
    stubs: {
      AppLayout: { template: '<div><slot /></div>' },
      BaseDialog: { props: ['show'], template: '<div v-if="show"><slot /></div>' },
      Icon: true
    }
  }
})

describe('admin TicketsView responsive inbox', () => {
  beforeEach(() => {
    listTickets.mockReset()
    listCategories.mockReset()
    routerPush.mockReset()
    listCategories.mockResolvedValue([category])
    listTickets.mockResolvedValue({
      items: [ticket],
      total: 1,
      page: 1,
      page_size: 20,
      pages: 2
    })
  })

  it('renders a mobile inbox item and preserves the desktop table', async () => {
    const wrapper = mountView()
    await flushPromises()

    const mobileList = wrapper.get('[data-test="mobile-ticket-list"]')
    const mobileTicket = wrapper.get('[data-test="mobile-ticket-T-20260719-7"]')
    expect(mobileList.classes()).toContain('md:hidden')
    expect(mobileTicket.text()).toContain('T-20260719-7')
    expect(mobileTicket.text()).toContain('充值到账后余额没有更新')
    expect(mobileTicket.text()).toContain('3')
    expect(mobileTicket.text()).toContain('待处理')
    expect(mobileTicket.text()).toContain('Customer')
    expect(mobileTicket.text()).toContain('customer@example.com')
    expect(mobileTicket.text()).toContain('账单问题')

    await mobileTicket.trigger('click')
    expect(routerPush).toHaveBeenCalledWith('/admin/tickets/T-20260719-7')

    const desktopTable = wrapper.get('[data-test="desktop-ticket-table"]')
    expect(desktopTable.classes()).toEqual(expect.arrayContaining(['hidden', 'md:block']))
    expect(desktopTable.get('table').classes()).toContain('min-w-[800px]')
  })

  it('keeps search and pagination wired to the existing list request', async () => {
    const wrapper = mountView()
    await flushPromises()

    const searchInput = wrapper.get('input[placeholder="搜索工单号、标题或用户"]')
    await searchInput.setValue('T-20260719')
    await searchInput.trigger('keyup', { key: 'Enter' })
    await flushPromises()

    expect(listTickets).toHaveBeenLastCalledWith(expect.objectContaining({
      page: 1,
      search: 'T-20260719'
    }))

    const paginationButtons = wrapper.get('nav[aria-label="分页"]').findAll('button')
    await paginationButtons[1].trigger('click')
    await flushPromises()
    expect(listTickets).toHaveBeenLastCalledWith(expect.objectContaining({ page: 2 }))
  })

  it('uses compact mobile controls and collapsed category editing', async () => {
    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.get('.sm\\:hidden').classes()).toContain('grid-cols-[minmax(0,1fr)_40px]')
    expect(wrapper.text()).not.toContain('新增分类')

    const manageButton = wrapper.get('button[title="管理分类"]')
    await manageButton.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('新增分类')
    expect(wrapper.find('[data-test="category-editor-3"]').exists()).toBe(false)
    await wrapper.get('button[title="编辑"]').trigger('click')
    const editor = wrapper.get('[data-test="category-editor-3"]')
    expect((editor.findAll('input')[0].element as HTMLInputElement).value).toBe('账单问题')
  })
})
