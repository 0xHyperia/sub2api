import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const source = readFileSync(
  resolve(dirname(fileURLToPath(import.meta.url)), '../BatchImageGuideView.vue'),
  'utf8',
)

describe('batch image mobile layout', () => {
  it('uses a task summary card with primary and overflow actions', () => {
    expect(source).toContain('<template #mobile-card="{ row }">')
    expect(source).toContain('statusBadgeClass(displayJob(row))')
    expect(source).toContain('displayJob(row).success_count')
    expect(source).toContain('displayJob(row).fail_count')
    expect(source).toContain('costLabel(displayJob(row))')
    expect(source).toContain('@click.stop="selectJob(row.id)"')
    expect(source).toContain('@click.stop="downloadJob(row)"')
    expect(source).toContain('@click.stop="toggleMoreMenu(row, $event)"')
  })

  it('keeps fixed task cell widths desktop-only', () => {
    expect(source).toContain('min-w-0 items-start gap-1 lg:w-[220px]')
    expect(source).toContain('min-w-0 text-center lg:max-w-[180px]')
    expect(source).not.toContain('class="flex w-[220px]')
  })

  it('renders result cards on phones and preserves the wide table for desktop', () => {
    expect(source).toContain('class="grid gap-3 lg:hidden"')
    expect(source).toContain('grid-cols-[84px_minmax(0,1fr)]')
    expect(source).toContain('itemDisplayStatusBadgeClass(item)')
    expect(source).toContain('itemResultClass(item)')
    expect(source).toContain('class="hidden overflow-x-auto rounded-panel border border-outline bg-surface lg:block"')
    expect(source).toContain('class="w-full min-w-[860px] table-fixed')
  })
})
