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
    expect(source).toContain('hidden min-w-0 gap-3 lg:flex')
    expect(source).toContain('w-44 flex-none')
    expect(source).not.toContain('h-10 min-w-40 text-sm')
  })

  it('replaces the mobile filter sidebar with a compact toolbar and filter sheet', () => {
    expect(source).toContain('class="hidden h-fit lg:sticky')
    expect(source).toContain('class="space-y-2.5 lg:hidden"')
    expect(source).toContain('<ModelMarketplaceFilterDrawer')
    expect(source).toContain(':result-count="mobileFilterResultCount"')
    expect(source).toContain('activeFilterChips')
  })

  it('uses incremental loading instead of page navigation', () => {
    expect(source).toContain('IntersectionObserver')
    expect(source).toContain('visibleEntries')
    expect(source).toContain('loadMoreSentinel')
    expect(source).not.toContain('currentPage')
    expect(source).not.toContain('totalPages')
  })

  it('uses compact responsive cards and a dedicated detail drawer', () => {
    expect(source).toContain('xl:grid-cols-3 2xl:grid-cols-4')
    expect(source).toContain('min-h-[216px]')
    expect(source).toContain('sm:min-h-[230px]')
    expect(source).toContain('<ModelMarketplaceDetailDrawer')
    expect(source).toContain("cardBillingCategory(entry) === 'usage'")
    expect(source).not.toContain('<GroupBadge')
  })
})
