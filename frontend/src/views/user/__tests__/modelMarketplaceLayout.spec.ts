import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const source = readFileSync(
  resolve(dirname(fileURLToPath(import.meta.url)), '../ModelMarketplaceView.vue'),
  'utf8'
)

describe('model marketplace toolbar layout', () => {
  it('keeps desktop controls on one compact row without a full-width sort field', () => {
    expect(source).toContain('data-testid="marketplace-toolbar"')
    expect(source).toContain('compareMarketplaceProviders(a.value, b.value)')
    expect(source).toContain('compareMarketplaceProviders(a.platform, b.platform)')
    expect(source).toContain('hidden min-w-0 gap-3 lg:flex')
    expect(source).toContain('w-44 flex-none')
    expect(source).not.toContain('h-10 min-w-40 text-sm')
  })

  it('replaces the mobile filter sidebar with a compact toolbar and filter sheet', () => {
    expect(source).toContain('class="hidden h-fit rounded-panel border')
    expect(source).toContain('class="space-y-2.5 lg:hidden"')
    expect(source).toContain('<ModelMarketplaceFilterDrawer')
    expect(source).toContain(':result-count="mobileFilterResultCount"')
    expect(source).toContain(':capability-options="capabilityOptions"')
    expect(source).toContain(':selected-capability="selectedCapability"')
    expect(source).toContain('activeFilterChips')
  })

  it('provides a capability filter in the desktop sidebar', () => {
    expect(source).toContain('id="marketplace-capability-filter"')
    expect(source).toContain('v-for="capability in capabilityOptions"')
    expect(source).toContain('selectedCapability === capability.value')
    expect(source).toContain('inferMarketplaceModelCapabilities(entry).includes')
  })

  it('uses incremental loading instead of page navigation', () => {
    expect(source).toContain('IntersectionObserver')
    expect(source).toContain('visibleEntries')
    expect(source).toContain('loadMoreSentinel')
    expect(source).not.toContain('currentPage')
    expect(source).not.toContain('totalPages')
  })

  it('uses compact responsive cards and a dedicated detail drawer', () => {
    const capabilityTriggerClass = source.match(/class="group\/capability[^"]*"/)?.[0] ?? ''
    const cardActionClasses = [...source.matchAll(/class="(inline-flex h-6 w-6[^"]*)"[\s\S]*?@click\.stop="(?:copyModel\(entry\.name\)|openDetails\(entry\))"/g)]
      .map(match => match[1])

    expect(source).toContain('xl:grid-cols-3 2xl:grid-cols-4')
    expect(source).toContain('data-testid="marketplace-model-card"')
    expect(source).toContain('min-h-[152px]')
    expect(source).toContain('cardCapabilityBadges(entry)')
    expect(source).toContain("<Icon :name=\"capability.icon\"")
    expect(source).toContain('group/capability')
    expect(source).toContain('marketplace-capability-tooltip')
    expect(capabilityTriggerClass).toContain('hover:text-black')
    expect(capabilityTriggerClass).not.toContain('hover:bg-')
    expect(capabilityTriggerClass).not.toContain('hover:border-')
    expect(source).toContain(':aria-describedby="`capability-${entry.key}-${capability.key}`"')
    expect(source).toContain('@click.stop="openDetails(entry)"')
    expect(source).toContain('<Icon name="eye" size="xs"')
    expect(cardActionClasses).toHaveLength(2)
    expect(cardActionClasses.every(classes => classes.includes('hover:text-black'))).toBe(true)
    expect(cardActionClasses.every(classes => !classes.includes('hover:bg-') && !classes.includes('hover:border-'))).toBe(true)
    expect(source).toContain('monitorSignalPoints(entry)')
    expect(source).toContain('marketplace-status-bar')
    expect(source).toContain('monitorLatency(entry)')
    expect(source).not.toContain('<ModelMonitorTimeline compact')
    expect(source).toContain('grid-cols-[24px_minmax(0,1fr)_52px]')
    expect(source).toContain(':class="platformIconClass(entry.platform)"')
    expect(source).toContain('text-orange-600 dark:text-orange-400')
    expect(source).not.toContain('platformBadgeClass')
    expect(source).not.toContain('rounded-panel border"\n                  :class="platformBadgeClass')
    expect(source).not.toContain('hasMonitorTimeline(entry)')
    expect(source).toContain('divide-x divide-outline')
    expect(source).toContain('cardGroups(entry)')
    expect(source).toContain('marketplace-card-groups')
    expect(source).toContain("'border-black bg-black text-white'")
    expect(source).not.toContain('[active, ...sorted.filter')
    expect(source).toContain('overflow-x-auto')
    expect(source).toContain('max-w-40 truncate')
    expect(source).toContain('class="card h-[152px] animate-pulse')
    expect(source).not.toContain('.slice(0, 2)')
    expect(source).not.toContain("t('modelMarketplace.details.moreGroups'")
    expect(source).toContain('marketplace-rate-tooltip')
    expect(source).toContain('group/rate')
    expect(source).toContain('<ModelMarketplaceDetailDrawer')
    expect(source).toContain("cardBillingCategory(entry) === 'usage'")
    expect(source).toContain("t('modelMarketplace.realtimeRate')")
    expect(source).toContain('cardRealtimeRate(entry)')
    expect(source).not.toContain('<GroupBadge')
    expect(source).not.toContain('min-h-20')
  })
})
