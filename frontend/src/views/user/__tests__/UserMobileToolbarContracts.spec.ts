import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it } from 'vitest'

const readSource = (path: string) => readFileSync(resolve(process.cwd(), path), 'utf8')

describe('user mobile toolbar contracts', () => {
  it('keeps page titles and primary actions in one compact heading row', () => {
    const dashboard = readSource('src/views/user/DashboardView.vue')
    const orders = readSource('src/views/user/UserOrdersView.vue')
    const subscriptions = readSource('src/views/user/SubscriptionsView.vue')
    const support = readSource('src/views/user/SupportTicketsView.vue')

    expect(dashboard).toContain('flex items-start justify-between gap-3 border-b')
    expect(orders).toContain('page-header flex items-start justify-between gap-3')
    expect(subscriptions).toContain('page-header mb-0 flex items-start justify-between gap-3')
    expect(subscriptions).toContain('btn btn-primary btn-icon shrink-0')
    expect(support).toContain('grid-cols-[minmax(0,1fr)_auto]')
  })

  it('pairs search, filter, and refresh controls instead of stacking each control', () => {
    const keys = readSource('src/views/user/KeysView.vue')
    const channels = readSource('src/views/user/AvailableChannelsView.vue')
    const support = readSource('src/views/user/SupportTicketsView.vue')
    const orders = readSource('src/views/user/UserOrdersView.vue')

    expect(keys).toContain('class="flex min-w-0 items-center gap-2 sm:hidden"')
    expect(keys).toContain('v-if="mobileFiltersOpen" class="grid grid-cols-2 gap-2 sm:hidden"')
    expect(channels).toContain('class="flex items-center gap-2 sm:justify-between"')
    expect(support).toContain('grid-cols-[minmax(0,1fr)_104px_40px]')
    expect(orders).toContain('class="flex items-center gap-2 rounded-panel')
  })

  it('uses a mobile filter sheet and compact action row for batch image jobs', () => {
    const source = readSource('src/views/user/BatchImageGuideView.vue')

    expect(source).toContain('grid-cols-[40px_40px_minmax(0,1fr)]')
    expect(source).toContain(':show="showMobileFilterModal"')
    expect(source).toContain('const mobileFilterDraft = reactive')
    expect(source).toContain('function applyMobileFilterDraft()')
  })

  it('prevents known mobile min-content and measurement overflows', () => {
    const affiliate = readSource('src/views/user/AffiliateView.vue')
    const marketplaceGroups = readSource('src/components/user/ModelMarketplaceCardGroups.vue')

    expect(affiliate.match(/class="min-w-0 space-y-2"/g)).toHaveLength(2)
    expect(marketplaceGroups).toContain('flex max-w-full items-center gap-1.5 overflow-hidden')
  })

  it('keeps analytics time controls in compact mobile grids', () => {
    const dashboardCharts = readSource('src/components/user/dashboard/UserDashboardCharts.vue')
    const usage = readSource('src/views/user/UsageView.vue')

    expect(dashboardCharts).toContain('grid-cols-[minmax(0,1fr)_112px_40px]')
    expect(usage).toContain('grid-template-columns: minmax(0, 1fr) 112px;')
  })
})
