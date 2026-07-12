import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const testDir = dirname(fileURLToPath(import.meta.url))
const componentPaths = [
  '../account/CreateAccountModal.vue',
  '../account/EditAccountModal.vue',
  '../account/BulkEditAccountModal.vue',
  '../admin/user/UserPlatformQuotaModal.vue',
  '../admin/ErrorPassthroughRulesModal.vue',
  '../admin/TLSFingerprintProfilesModal.vue',
  '../admin/monitor/MonitorTemplateManagerDialog.vue',
  '../admin/monitor/MonitorAdvancedRequestConfig.vue'
]

describe.each(componentPaths)('%s UI contract', (relativePath) => {
  const source = readFileSync(resolve(testDir, relativePath), 'utf8')

  it('does not use a native confirmation prompt', () => {
    expect(source).not.toMatch(/(?:window\.)?confirm\s*\(/)
  })

  it('does not force a two- or three-column grid at the narrowest breakpoint', () => {
    expect(source).not.toMatch(/(?:^|[\s"'`])grid-cols-[23]\b/m)
  })
})

describe.each([
  '../account/CreateAccountModal.vue',
  '../account/EditAccountModal.vue',
  '../account/BulkEditAccountModal.vue',
  '../admin/user/UserPlatformQuotaModal.vue'
])('%s confirmation contract', (relativePath) => {
  it('uses the shared confirmation dialog', () => {
    const source = readFileSync(resolve(testDir, relativePath), 'utf8')
    expect(source).toContain('<ConfirmDialog')
  })
})
