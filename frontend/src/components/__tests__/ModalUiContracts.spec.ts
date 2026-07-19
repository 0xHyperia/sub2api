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

describe.each([
  {
    relativePath: '../admin/user/UserPlatformQuotaModal.vue',
    mobileLayout: 'platform-cards',
    desktopLayout: 'quota-table',
    mobileActions: ['@click="onReset(row.platform, quotaWindow)"']
  },
  {
    relativePath: '../admin/ErrorPassthroughRulesModal.vue',
    mobileLayout: 'rule-cards',
    desktopLayout: 'rules-table',
    mobileActions: ['@click="toggleEnabled(rule)"', '@click="handleEdit(rule)"', '@click="handleDelete(rule)"']
  },
  {
    relativePath: '../admin/TLSFingerprintProfilesModal.vue',
    mobileLayout: 'profile-cards',
    desktopLayout: 'profiles-table',
    mobileActions: ['@click="handleEdit(profile)"', '@click="handleDelete(profile)"']
  }
])('$relativePath responsive collection contract', ({ relativePath, mobileLayout, desktopLayout, mobileActions }) => {
  const source = readFileSync(resolve(testDir, relativePath), 'utf8')

  it('uses a mobile card collection while retaining the desktop table', () => {
    expect(source).toContain(`data-mobile-layout="${mobileLayout}"`)
    expect(source).toContain(`data-desktop-layout="${desktopLayout}"`)
    expect(source).toMatch(new RegExp(`data-mobile-layout="${mobileLayout}"[^>]*sm:hidden`))
    expect(source).toMatch(new RegExp(`data-desktop-layout="${desktopLayout}"[^>]*hidden[^>]*sm:block`))
  })

  it('keeps editing actions available in the mobile collection', () => {
    for (const action of mobileActions) expect(source).toContain(action)
  })
})
