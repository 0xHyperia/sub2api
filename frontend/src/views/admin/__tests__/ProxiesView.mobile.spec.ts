import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { parse } from 'vue/compiler-sfc'

const filename = resolve(process.cwd(), 'src/views/admin/ProxiesView.vue')
const source = readFileSync(filename, 'utf8')
const { descriptor, errors } = parse(source, { filename })
const template = descriptor.template?.content || ''

describe('ProxiesView mobile proxy cards', () => {
  it('keeps a business-focused mobile collection and the desktop DataTable', () => {
    expect(errors).toEqual([])
    expect(template).toMatch(/data-mobile-layout="proxy-cards"[^>]*md:hidden/)
    expect(template).toMatch(/data-desktop-layout="proxies-table"[^>]*hidden[^>]*md:block/)
    expect(template).toContain('<DataTable')
    expect(template).toContain('#cell-select')
    expect(template).toContain('@change="toggleSelectRow(row.id, $event)"')
  })

  it('masks the mobile URL without changing the full URL copy implementation', () => {
    expect(source).toContain('maskedProxyUrl(row)')
    expect(source).toMatch(/const maskedProxyUrl = \(row: Proxy\): string =>\s*`\$\{row\.protocol\}:\/\/\$\{maskProxyHost\(row\.host\)\}:\$\{row\.port\}`/)
    expect(source).toMatch(/function buildProxyUrl\(row: any\)[\s\S]*buildAuthPart\(row\)/)
  })

  it('keeps test and edit visible while placing low-frequency actions in More', () => {
    expect(template).toContain('data-mobile-action="test"')
    expect(template).toContain('@click="handleTestConnection(row)"')
    expect(template).toContain('data-mobile-action="edit"')
    expect(template).toContain('@click="handleEdit(row)"')

    const more = template.match(/<details class="group relative shrink-0">([\s\S]*?)<\/details>/)?.[1] || ''
    expect(more).toContain('@click="handleQualityCheck(row)"')
    expect(more).toContain('@click="copyProxyUrl(row)"')
    expect(more).toContain('@click="handleDelete(row)"')
  })

  it('surfaces health, scheduling, account association, and recent failures', () => {
    expect(template).toContain('row.latency_status')
    expect(template).toContain('qualityOverallLabel(row.quality_status)')
    expect(template).toContain('fallbackModeLabel(row.fallback_mode)')
    expect(template).toContain('@click="openAccountsModal(row)"')
    expect(template).toContain('proxyIssueSummary(row)')
  })
})
