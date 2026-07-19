import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const testDir = dirname(fileURLToPath(import.meta.url))
const usageViewSource = readFileSync(resolve(testDir, '../UsageView.vue'), 'utf8')
const usageTableSource = readFileSync(resolve(testDir, '../../../components/admin/usage/UsageTable.vue'), 'utf8')
const errorTableSource = readFileSync(resolve(testDir, '../../../components/user/UserErrorRequestsTable.vue'), 'utf8')
const drawerSource = readFileSync(resolve(testDir, '../../../components/user/UsageFilterDrawer.vue'), 'utf8')

describe('user usage mobile experience', () => {
  it('keeps the mobile toolbar focused and moves advanced filters into a sheet', () => {
    expect(usageViewSource).toContain('usage-mobile-records-toolbar sm:hidden')
    expect(usageViewSource).toContain(':class="errorViewEnabled ? \'grid-cols-2\' : \'grid-cols-1\'"')
    expect(usageViewSource).toContain('usage-filters hidden sm:block')
    expect(usageViewSource).toContain('<UsageFilterDrawer')
    expect(usageViewSource).toContain('activeFilterChips')
    expect(usageViewSource).toContain('@click="removeActiveFilter(chip.key)"')
  })

  it('uses a closeable, draft-based bottom sheet for advanced filters', () => {
    expect(drawerSource).toContain('max-h-[88dvh]')
    expect(drawerSource).toContain("event.key === 'Escape'")
    expect(drawerSource).toContain("document.body.style.overflow = 'hidden'")
    expect(drawerSource).toContain("emit('applyUsage', { ...usageDraft })")
    expect(drawerSource).toContain("emit('applyErrors', { ...errorDraft })")
  })

  it('replaces generic mobile rows with business summaries and detail actions', () => {
    expect(usageTableSource).toContain('<template v-if="userMobileCard" #mobile-card="{ row }">')
    expect(usageViewSource).toContain('user-mobile-card')
    expect(usageTableSource).toContain('row.api_key?.name')
    expect(usageTableSource).toContain('mobileTokenLabel(row)')
    expect(usageTableSource).toContain('row.actual_cost?.toFixed(6)')
    expect(usageTableSource).toContain('toggleMobileDetails(row)')
    expect(errorTableSource).toContain('<template #mobile-card="{ row }">')
    expect(errorTableSource).toContain("openDetail(row.id)")
  })
})
