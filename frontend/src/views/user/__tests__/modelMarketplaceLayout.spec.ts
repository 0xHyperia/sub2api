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
    expect(source).toContain('xl:flex-row xl:items-center xl:justify-between')
    expect(source).toContain('sm:flex sm:flex-nowrap')
    expect(source).toContain('sm:w-44 sm:flex-none')
    expect(source).not.toContain('h-10 min-w-40 text-sm')
  })

  it('uses a deliberate two-row action layout on narrow screens', () => {
    expect(source).toContain('grid-cols-[minmax(0,1fr)_40px]')
    expect(source).toContain('col-span-2 flex min-h-10')
  })
})
