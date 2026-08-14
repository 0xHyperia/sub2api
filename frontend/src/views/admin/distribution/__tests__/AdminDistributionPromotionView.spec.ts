import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import AdminDistributionPromotionView from '../AdminDistributionPromotionView.vue'

const api = vi.hoisted(() => ({
  getPromotionAnalytics: vi.fn(),
  listPromotionVisits: vi.fn(),
  lookupAgents: vi.fn(),
}))

vi.mock('@/api/admin/distribution', () => api)
vi.mock('@/stores/app', () => ({ useAppStore: () => ({ showError: vi.fn() }) }))

const analytics = {
  summary: { total_visits: 12, unique_visitors: 8, bot_visits: 2, converted_visitors: 3, tracked_registrations: 3, registrations: 4, direct_registrations: 3, persisted_registrations: 1, untracked_direct: 1, conversion_rate: 37.5 },
  daily: [{ date: '2026-07-22', visits: 12, visitors: 8, registrations: 3 }],
  sources: [{ source: '直接访问', visits: 12, conversions: 2, registrations: 99 }],
  meta: { cohort: 'visit', tracking_enabled: true, attribution_enabled: true, attribution_days: 30, attribution_model: 'first_touch', bot_filter_enabled: true, detail_retention_days: 180, raw_retention_days: 367, timezone: 'Asia/Hong_Kong' },
}

function mountView() {
  return mount(AdminDistributionPromotionView, {
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        BaseDialog: { template: '<div><slot /></div>' },
        Icon: true,
        LoadingSpinner: true,
        AdminDistributionNav: true,
        RemoteEntityCombobox: true,
        RouterLink: { template: '<a><slot /></a>' },
      },
    },
  })
}

describe('AdminDistributionPromotionView', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    api.getPromotionAnalytics.mockResolvedValue(analytics)
    api.listPromotionVisits.mockResolvedValue({ items: [], total: 0, page: 1, page_size: 20 })
    api.lookupAgents.mockResolvedValue([])
  })

  it('keeps analytics usable when visit detail fails independently', async () => {
    api.listPromotionVisits.mockRejectedValueOnce(new Error('detail failed'))
    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.text()).toContain('访问明细加载失败')
    expect(wrapper.find('[aria-label="推广核心指标"]').exists()).toBe(true)
    expect(wrapper.text()).not.toContain('推广概览加载失败')
  })

  it('keeps visit detail usable when analytics fails independently', async () => {
    api.getPromotionAnalytics.mockRejectedValueOnce(new Error('analytics failed'))
    api.listPromotionVisits.mockResolvedValueOnce({
      items: [{ id: 1, agent_id: 2, agent_email: 'agent@example.com', source: 'google', device_type: 'mobile', visited_at: '2026-07-22T10:00:00Z' }],
      total: 1,
      page: 1,
      page_size: 20,
    })
    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.text()).toContain('推广概览加载失败')
    expect(wrapper.text()).toContain('agent@example.com')
    expect(wrapper.text()).not.toContain('访问明细加载失败')
  })

  it('renders the effective tracking policy returned with analytics', async () => {
    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.text()).toContain('正在采集')
    expect(wrapper.text()).toContain('归因窗口 30 天')
    expect(wrapper.text()).toContain('首次访问归因')
    expect(wrapper.text()).toContain('汇总排除机器人')
    expect(wrapper.text()).toContain('明细保留 180 天')
    expect(wrapper.text()).toContain('原始记录 367 天后删除')
    expect(wrapper.text()).toContain('去标识化')
    expect(wrapper.text()).toContain('尚未注册')
    expect(wrapper.text()).toContain('访客转化率')
    expect(wrapper.text()).toContain('已转化访客 ÷ 独立访客')
    expect(wrapper.text()).toContain('注册结果')
    expect(wrapper.text()).toContain('携推广码注册但未匹配到有效访问')
    expect(wrapper.text()).toContain('访问来源排行')
    expect(wrapper.text()).toContain('访问转化率')
    expect(wrapper.text()).toContain('12 / 2')
    expect(wrapper.text()).not.toContain('12 / 99')
    expect(wrapper.find('[aria-label="推广核心指标"]').exists()).toBe(true)
  })

  it('applies a source directly from the source ranking', async () => {
    const wrapper = mountView()
    await flushPromises()

    await wrapper.get('[aria-label="筛选来源 直接访问"]').trigger('click')
    await flushPromises()

    expect(api.getPromotionAnalytics).toHaveBeenLastCalledWith(expect.objectContaining({ source: '直接访问' }))
  })

  it('renders an inline date error and focuses the invalid field', async () => {
    const wrapper = mountView()
    await flushPromises()
    const dates = wrapper.findAll('input[type="date"]')
    const focus = vi.spyOn(dates[1].element as HTMLInputElement, 'focus')
    await dates[0].setValue('2026-07-23')
    await dates[1].setValue('2026-07-22')
    await wrapper.findAll('button').find((button) => button.text() === '应用筛选')!.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('结束日期不能早于开始日期')
    expect(dates[1].attributes('aria-invalid')).toBe('true')
    expect(focus).toHaveBeenCalled()
  })

  it('groups sparse 90-day trend rows by week', async () => {
    const wrapper = mountView()
    await flushPromises()
    api.getPromotionAnalytics.mockResolvedValueOnce({
      ...analytics,
      daily: [
        { date: '2026-07-20', visits: 4, visitors: 3, conversions: 1, registrations: 1 },
        { date: '2026-07-21', visits: 5, visitors: 4, conversions: 1, registrations: 2 },
      ],
    })

    await wrapper.findAll('button').find((button) => button.text() === '90天')!.trigger('click')
    await flushPromises()

    expect(wrapper.get('[aria-label="访问与注册趋势"]').findAll('[role="listitem"]')).toHaveLength(1)
    expect(wrapper.text()).toContain('按周聚合')
  })

  it('lists daily trend dates from newest to oldest', async () => {
    api.getPromotionAnalytics.mockResolvedValueOnce({
      ...analytics,
      daily: [
        { date: '2026-07-20', visits: 4, visitors: 3, conversions: 1, registrations: 1 },
        { date: '2026-07-22', visits: 5, visitors: 4, conversions: 1, registrations: 2 },
        { date: '2026-07-21', visits: 3, visitors: 2, conversions: 0, registrations: 0 },
      ],
    })
    const wrapper = mountView()
    await flushPromises()

    const rows = wrapper.get('[aria-label="访问与注册趋势"]').findAll('[role="listitem"]')
    expect(rows.map((row) => row.attributes('aria-label').split('，')[0])).toEqual([
      '07/22',
      '07/21',
      '07/20',
    ])
  })

  it('refreshes the current page and clamps it when the result set shrinks', async () => {
    api.listPromotionVisits.mockResolvedValueOnce({ items: [], total: 41, page: 1, page_size: 20 })
    const wrapper = mountView()
    await flushPromises()

    api.listPromotionVisits.mockResolvedValueOnce({ items: [], total: 41, page: 2, page_size: 20 })
    await wrapper.findAll('button').find((button) => button.text() === '下一页')!.trigger('click')
    await flushPromises()

    api.listPromotionVisits
      .mockResolvedValueOnce({ items: [], total: 20, page: 2, page_size: 20 })
      .mockResolvedValueOnce({ items: [], total: 20, page: 1, page_size: 20 })
    await wrapper.get('[aria-label="刷新当前页数据"]').trigger('click')
    await flushPromises()

    const recentCalls = api.listPromotionVisits.mock.calls.slice(-2)
    expect(recentCalls[0][0].page).toBe(2)
    expect(recentCalls[1][0].page).toBe(1)
  })
})
