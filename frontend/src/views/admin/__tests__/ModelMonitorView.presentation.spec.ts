import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const source = readFileSync(
  resolve(dirname(fileURLToPath(import.meta.url)), '../ModelMonitorView.vue'),
  'utf8',
)

describe('model monitor presentation controls', () => {
  it('supports label and display priority configuration', () => {
    expect(source).toContain('openPresentation(row)')
    expect(source).toContain('presentationLabel')
    expect(source).toContain('presentationOrder')
    expect(source).toContain('display_order:')
  })

  it('silently refreshes visible monitor data every 30 seconds', () => {
    expect(source).toContain('window.setInterval')
    expect(source).toContain('30_000')
    expect(source).toContain("document.visibilityState === 'visible'")
    expect(source).toContain('void load(true)')
    expect(source).toContain('window.clearInterval')
  })

  it('uses business-specific mobile monitor cards and a history timeline', () => {
    expect(source).toContain('<template #mobile-card="{ row }">')
    expect(source).toContain('activeFilterCount')
    expect(source).toContain('grid grid-cols-2 gap-px')
    expect(source).toContain('mobile-history-')
    expect(source).toContain('class="divide-y divide-outline sm:hidden"')
    expect(source).toContain('class="hidden overflow-x-auto sm:block"')
  })
})
