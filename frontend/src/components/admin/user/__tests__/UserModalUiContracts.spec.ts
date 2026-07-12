import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const testDir = dirname(fileURLToPath(import.meta.url))
const componentPaths = [
  '../UserBalanceModal.vue',
  '../UserApiKeysModal.vue',
  '../UserBalanceHistoryModal.vue',
  '../UserAllowedGroupsModal.vue'
]

describe.each(componentPaths)('%s UI contract', (relativePath) => {
  const source = readFileSync(resolve(testDir, relativePath), 'utf8')

  it('uses the semantic surface palette without legacy color scales', () => {
    expect(source).not.toMatch(/(?:bg|border|text)-(?:blue|gray|green|emerald|red|purple|dark|primary)-/)
  })

  it('keeps panel radii at eight pixels or below', () => {
    expect(source).not.toMatch(/rounded-(?:xl|2xl|3xl|full)/)
    expect(source).not.toContain('bg-gradient')
  })
})
