import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const testDir = dirname(fileURLToPath(import.meta.url))
const statsSource = readFileSync(resolve(testDir, '../UserDashboardStats.vue'), 'utf8')
const chartsSource = readFileSync(resolve(testDir, '../UserDashboardCharts.vue'), 'utf8')

describe('user dashboard responsive presentation', () => {
  it('gives balance and daily cost full-width priority on narrow phones', () => {
    expect(statsSource).toContain('dashboard-metric dashboard-metric-balance')
    expect(statsSource).toContain('dashboard-metric dashboard-metric-api-keys')
    expect(statsSource).toContain('dashboard-metric dashboard-metric-requests')
    expect(statsSource).toContain('dashboard-metric dashboard-metric-cost')
    expect(statsSource).toContain('@media (max-width: 479px)')
    expect(statsSource).toMatch(/\.dashboard-metric-balance,\s*\.dashboard-metric-cost\s*{\s*grid-column: 1 \/ -1;/)
    expect(statsSource).toMatch(/\.dashboard-metric-value\s*{[\s\S]*white-space: nowrap;/)
  })

  it('uses a ranked business summary on phones and retains the desktop table', () => {
    expect(chartsSource).toContain('v-for="(model, index) in mobileModels"')
    expect(chartsSource).toContain('class="divide-y divide-outline sm:hidden"')
    expect(chartsSource).toContain('class="hidden max-h-48 overflow-auto sm:block"')
    expect(chartsSource).toContain('.sort((a, b) => (b.actual_cost ?? 0) - (a.actual_cost ?? 0))')
    expect(chartsSource).toContain('.slice(0, 5)')
  })
})
