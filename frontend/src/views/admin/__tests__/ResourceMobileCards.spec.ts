import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const viewsDirectory = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const source = (name: string) => readFileSync(resolve(viewsDirectory, name), 'utf8')

describe('admin resource mobile cards', () => {
  it('organizes group operations around capacity, accounts, usage, and subscription limits', () => {
    const groups = source('GroupsView.vue')

    expect(groups).toContain('<template #mobile-card="{ row }">')
    expect(groups).toContain("t('admin.groups.accountsAvailable')")
    expect(groups).toContain("t('admin.groups.usageToday')")
    expect(groups).toContain('<GroupCapacityBadge')
    expect(groups).toContain("t('admin.groups.subscription.title')")
  })

  it('organizes channels around platform, group, pricing, status, and primary actions', () => {
    const channels = source('ChannelsView.vue')

    expect(channels).toContain('<template #mobile-card="{ row }">')
    expect(channels).toContain('channelPlatforms(row)')
    expect(channels).toContain("t('admin.channels.columns.groups', 'Groups')")
    expect(channels).toContain("t('admin.channels.columns.pricing', 'Pricing')")
    expect(channels).toContain('toggleChannelStatus(row)')
  })

  it('keeps subscription identity, quota progress, expiry, and lifecycle actions together', () => {
    const subscriptions = source('SubscriptionsView.vue')

    expect(subscriptions).toContain('<template #mobile-card="{ row }">')
    expect(subscriptions).toContain('row.daily_usage_usd')
    expect(subscriptions).toContain('row.weekly_usage_usd')
    expect(subscriptions).toContain('row.monthly_usage_usd')
    expect(subscriptions).toContain('handleExtend(row)')
    expect(subscriptions).toContain('handleRevoke(row)')
  })

  it('treats announcements as a publishing queue with audience, schedule, and read status', () => {
    const announcements = source('AnnouncementsView.vue')

    expect(announcements).toContain('<template #mobile-card="{ row }">')
    expect(announcements).toContain('targetingSummary(row.targeting)')
    expect(announcements).toContain('row.starts_at')
    expect(announcements).toContain('openReadStatus(row)')
  })

  it('uses compact audit search, advanced filters, and event cards on mobile', () => {
    const audit = source('AuditLogView.vue')

    expect(audit).toContain('activeFilterCount')
    expect(audit).toContain('<template #mobile-card="{ row }">')
    expect(audit).toContain('authMethodLabel(row.auth_method)')
    expect(audit).toContain('openDetail(row.id)')
    expect(audit).not.toContain('  /60')
  })

  it('organizes channel monitors around current health and immediate operations', () => {
    const monitors = source('ChannelMonitorView.vue')

    expect(monitors).toContain('<template #mobile-card="{ row }">')
    expect(monitors).toContain('row.primary_status')
    expect(monitors).toContain('formatAvailability(row)')
    expect(monitors).toContain('handleRunNow(row)')
    expect(monitors).toContain('handleDuplicate(row)')
  })
})
