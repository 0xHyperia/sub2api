import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const testDir = dirname(fileURLToPath(import.meta.url))
const tableSource = readFileSync(resolve(testDir, '../affiliates/AdminAffiliateRecordsTable.vue'), 'utf8')

describe('admin affiliate record mobile cards', () => {
  it('summarizes invite relationships around both people and invitation time', () => {
    expect(tableSource).toContain('<template #mobile-card="{ row }">')
    expect(tableSource).toContain("props.type === 'invites'")
    expect(tableSource).toContain(':id="row.inviter_id"')
    expect(tableSource).toContain(':id="row.invitee_id"')
    expect(tableSource).toContain("t('admin.affiliates.records.invitedAt')")
  })

  it('summarizes rebates around recipient, amount, order, and time', () => {
    expect(tableSource).toContain("props.type === 'rebates'")
    expect(tableSource).toContain('formatAmount(row.rebate_amount)')
    expect(tableSource).toContain('#{{ row.order_id }}')
    expect(tableSource).toContain('<OrderStatusBadge :status="row.order_status" />')
    expect(tableSource).toContain("t('admin.affiliates.records.rebatedAt')")
  })

  it('shows atomic transfers as completed with initiation and completion timestamps', () => {
    expect(tableSource).toContain('formatAmount(row.amount)')
    expect(tableSource).toContain("t('admin.affiliates.records.transferCompleted')")
    expect(tableSource).toContain("t('admin.affiliates.records.initiatedAt')")
    expect(tableSource).toContain("t('admin.affiliates.records.completedAt')")
    expect(tableSource.match(/formatDateTime\(row\.created_at\)/g)?.length).toBeGreaterThanOrEqual(4)
  })

  it('preserves server sorting, filters, and pagination contracts', () => {
    expect(tableSource).toContain(':server-side-sort="true"')
    expect(tableSource).toContain('@sort="handleSort"')
    expect(tableSource).toContain('@update:page="handlePageChange"')
    expect(tableSource).toContain('@update:pageSize="handlePageSizeChange"')
    expect(tableSource).toContain('reloadFromFirstPage')
  })
})
