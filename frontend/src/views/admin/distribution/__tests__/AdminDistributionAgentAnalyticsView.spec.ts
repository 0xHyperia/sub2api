import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

const listAgents = vi.hoisted(() => vi.fn())
const showError = vi.hoisted(() => vi.fn())
vi.mock('@/api/admin/distribution', () => ({ listAgents }))
vi.mock('@/stores/app', () => ({ useAppStore: () => ({ showError }) }))
vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return { ...actual, useI18n: () => ({ t: (key: string) => key.split('.').at(-1) || key }) }
})

import AdminDistributionAgentAnalyticsView from '../AdminDistributionAgentAnalyticsView.vue'

const agent = (id: number) => ({
  id, user_id: id, level_id: 1, depth: 1, promotion_code: `CODE${id}`, effective_rate_bps: 100,
  max_child_rate_bps: 50, can_recruit_subagents: true, can_view_promotion_stats: true, status: 'active',
  available_cny: '0', frozen_cny: '0', reserved_cny: '0', debt_cny: '0', total_earned_cny: '0', total_withdrawn_cny: '0',
  email: `agent${id}@example.com`, username: `agent${id}`, customer_count: 1, team_count: 0,
  paying_customer_count: 0, customer_paid_cny: '0', this_month_commission_cny: '0', period_customer_count: 1,
  period_paying_customers: 0, period_customer_paid_cny: '0', period_commission_cny: '0', team_customer_count: 0,
  team_paying_customers: 0, team_customer_paid_cny: '0', team_commission_cny: '0', period_activated_customers: 0,
  period_cohort_paid_customers: 0, period_all_paying_customers: 0, period_repurchase_customers: 0,
  period_active_customers: 0, total_team_customers: 0, period_reward_cny: '0', total_converted_cny: '0',
})

function mountView() {
  return mount(AdminDistributionAgentAnalyticsView, {
    global: { stubs: {
      AppLayout: { template: '<div><slot /></div>' }, Icon: true, AdminDistributionNav: true,
      DistributionAnalyticsRange: true, RouterLink: { template: '<a><slot /></a>' },
    } },
  })
}

describe('AdminDistributionAgentAnalyticsView request state', () => {
  beforeEach(() => { vi.clearAllMocks(); listAgents.mockResolvedValue({ items: [agent(1)], total: 1 }) })

  it('removes old rows and shows an inline retry when a new filter request fails', async () => {
    const wrapper = mountView()
    await flushPromises()
    expect(wrapper.text()).toContain('agent1@example.com')

    listAgents.mockRejectedValueOnce(new Error('unavailable'))
    await wrapper.findAll('.segments button')[1].trigger('click')
    await flushPromises()

    expect(wrapper.text()).not.toContain('agent1@example.com')
    expect(wrapper.find('[role="alert"]').text()).toContain('unavailable')
    expect(wrapper.find('[role="alert"]').text()).toContain('retry')
  })

  it('ignores a late response from an older filter request', async () => {
    const wrapper = mountView()
    await flushPromises()
    let resolveOld!: (value: unknown) => void
    const old = new Promise(resolve => { resolveOld = resolve })
    listAgents.mockReturnValueOnce(old).mockResolvedValueOnce({ items: [agent(3)], total: 1 })
    await wrapper.findAll('.segments button')[1].trigger('click')
    await wrapper.findAll('.segments button')[2].trigger('click')
    await flushPromises()
    resolveOld({ items: [agent(2)], total: 1 })
    await flushPromises()

    expect(wrapper.text()).toContain('agent3@example.com')
    expect(wrapper.text()).not.toContain('agent2@example.com')
  })
})
